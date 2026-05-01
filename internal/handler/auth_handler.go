package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	mailservice "github.com/atbeta/picfast/internal/service/mail"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const emailVerificationTokenTTL = 24 * time.Hour

type AuthHandler struct {
	db     *sqlc.Queries
	pool   *pgxpool.Pool
	jwt    *JWTService
	config *config.Config
	mail   mailservice.Sender
}

func NewAuthHandler(db *sqlc.Queries, pool *pgxpool.Pool, jwtSvc *JWTService, cfg *config.Config, sender mailservice.Sender) *AuthHandler {
	if sender == nil {
		sender = mailservice.NewNoopSender(false)
	}
	return &AuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg, mail: sender}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}

type ResendVerificationRequest struct {
	Email string `json:"email"`
}

type RegisterResponse struct {
	RequiresEmailVerification bool               `json:"requires_email_verification"`
	VerificationEmailSent     bool               `json:"verification_email_sent"`
	Tokens                    *domain.AuthTokens `json:"tokens,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	app := h.config.AppSnapshot()
	if !app.AllowRegistration {
		Fail(w, http.StatusForbidden, "registration is disabled")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" || req.Password == "" || req.Name == "" {
		Fail(w, http.StatusBadRequest, "email, password and name are required")
		return
	}
	if !isValidEmail(req.Email) {
		Fail(w, http.StatusBadRequest, "invalid email format")
		return
	}
	if len(req.Password) < 8 || len(req.Password) > 72 {
		Fail(w, http.StatusBadRequest, "password must be between 8 and 72 bytes")
		return
	}

	existing, _ := h.db.GetUserByEmail(r.Context(), req.Email)
	if existing.ID != 0 {
		Fail(w, http.StatusConflict, "email already registered")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	group, err := h.db.GetDefaultGroup(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "no default group found")
		return
	}

	settings, _ := json.Marshal(domain.UserSettings{})

	requiresVerification := h.emailVerificationEnabled()
	resp := RegisterResponse{
		RequiresEmailVerification: requiresVerification,
	}
	var (
		tokens            *domain.AuthTokens
		verificationToken string
		user              sqlc.User
	)
	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		user, err = qtx.CreateUser(r.Context(), sqlc.CreateUserParams{
			GroupID:       domain.PgInt8(group.ID),
			Email:         req.Email,
			Password:      string(hash),
			Name:          req.Name,
			Role:          string(domain.RoleUser),
			CapacityBytes: app.UserInitialCapacity,
			Settings:      settings,
			Status:        int16(domain.UserStatusActive),
			EmailVerified: false,
			RegisteredIp:  r.RemoteAddr,
		})
		if err != nil {
			return err
		}

		if requiresVerification {
			if err := qtx.DeleteUnusedEmailVerificationTokensByUser(r.Context(), user.ID); err != nil {
				return err
			}
			plain, tokenHash, err := GenerateRefreshToken()
			if err != nil {
				return err
			}
			verificationToken = plain
			_, err = qtx.CreateEmailVerificationToken(r.Context(), sqlc.CreateEmailVerificationTokenParams{
				UserID:    user.ID,
				TokenHash: tokenHash,
				ExpiresAt: time.Now().Add(emailVerificationTokenTTL),
			})
			return err
		}

		tokens, err = h.generateTokens(r.Context(), qtx, user.ID, domain.RoleUser, group.ID)
		return err
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	if requiresVerification {
		resp.VerificationEmailSent = h.sendVerificationEmail(r.Context(), user.Email, user.Name, verificationToken) == nil
		if !resp.VerificationEmailSent {
			slog.Warn("failed to send verification email after registration", "email", user.Email)
		}
		Created(w, resp)
		return
	}

	resp.Tokens = tokens
	Created(w, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		Fail(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if user.Status != int16(domain.UserStatusActive) {
		Fail(w, http.StatusForbidden, "account is frozen")
		return
	}
	if h.emailVerificationEnabled() && !user.EmailVerified {
		Fail(w, http.StatusForbidden, "email verification required")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		Fail(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	groupID := domain.PgInt8Val(user.GroupID)

	tokens, err := h.generateTokens(r.Context(), h.db, user.ID, domain.UserRole(user.Role), groupID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	Success(w, tokens)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req VerifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Token) == "" {
		Fail(w, http.StatusBadRequest, "token is required")
		return
	}

	tokenHash := HashRefreshToken(req.Token)
	stored, err := h.db.GetEmailVerificationTokenByHash(r.Context(), tokenHash)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid verification token")
		return
	}
	if stored.UsedAt.Valid {
		Fail(w, http.StatusBadRequest, "verification token has already been used")
		return
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = h.db.DeleteEmailVerificationTokenByHash(r.Context(), tokenHash)
		Fail(w, http.StatusBadRequest, "verification token has expired")
		return
	}

	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		if _, err := qtx.MarkEmailVerificationTokenUsed(r.Context(), stored.ID); err != nil {
			return err
		}
		if err := qtx.UpdateUserEmailVerified(r.Context(), stored.UserID); err != nil {
			return err
		}
		return qtx.DeleteUnusedEmailVerificationTokensByUser(r.Context(), stored.UserID)
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to verify email")
		return
	}

	SuccessMessage(w, "email verified")
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if !h.emailVerificationEnabled() {
		Fail(w, http.StatusServiceUnavailable, "email verification is not available")
		return
	}

	var req ResendVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(req.Email) == "" {
		Fail(w, http.StatusBadRequest, "email is required")
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user.EmailVerified {
		SuccessMessage(w, "if the account exists, a verification email has been sent")
		return
	}

	plain, err := h.issueVerificationToken(r.Context(), h.db, user.ID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create verification token")
		return
	}
	if err := h.sendVerificationEmail(r.Context(), user.Email, user.Name, plain); err != nil {
		slog.Warn("failed to resend verification email", "email", user.Email, "error", err)
		Fail(w, http.StatusInternalServerError, "failed to send verification email")
		return
	}

	SuccessMessage(w, "verification email sent")
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tokenHash := HashRefreshToken(req.RefreshToken)
	stored, err := h.db.GetRefreshTokenByHash(r.Context(), tokenHash)
	if err != nil {
		Fail(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	if time.Now().After(stored.ExpiresAt) {
		h.db.DeleteRefreshToken(r.Context(), tokenHash)
		Fail(w, http.StatusUnauthorized, "refresh token expired")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), stored.UserID)
	if err != nil {
		Fail(w, http.StatusUnauthorized, "user not found")
		return
	}
	if h.emailVerificationEnabled() && !user.EmailVerified {
		h.db.DeleteRefreshToken(r.Context(), tokenHash)
		Fail(w, http.StatusForbidden, "email verification required")
		return
	}

	groupID := domain.PgInt8Val(user.GroupID)

	var tokens *domain.AuthTokens
	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		if err := qtx.DeleteRefreshToken(r.Context(), tokenHash); err != nil {
			return err
		}
		tokens, err = h.generateTokens(r.Context(), qtx, user.ID, domain.UserRole(user.Role), groupID)
		return err
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	Success(w, tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.db.DeleteAllUserRefreshTokens(r.Context(), userID)
	SuccessMessage(w, "logged out")
}

func (h *AuthHandler) generateTokens(ctx context.Context, qtx *sqlc.Queries, userID int64, role domain.UserRole, groupID int64) (*domain.AuthTokens, error) {
	accessToken, expiresIn, err := h.jwt.GenerateAccessToken(userID, role, groupID)
	if err != nil {
		return nil, err
	}

	plain, hash, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	_, err = qtx.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(h.config.JWT.RefreshTTL),
	})
	if err != nil {
		return nil, errors.New("failed to store refresh token")
	}

	return &domain.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: plain,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

func (h *AuthHandler) emailVerificationEnabled() bool {
	return h.config.AppSnapshot().RequireEmailVerification && h.mail != nil && h.mail.Ready()
}

func isValidEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}

func (h *AuthHandler) issueVerificationToken(ctx context.Context, qtx *sqlc.Queries, userID int64) (string, error) {
	if err := qtx.DeleteUnusedEmailVerificationTokensByUser(ctx, userID); err != nil {
		return "", err
	}
	plain, tokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", err
	}
	_, err = qtx.CreateEmailVerificationToken(ctx, sqlc.CreateEmailVerificationTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(emailVerificationTokenTTL),
	})
	if err != nil {
		return "", err
	}
	return plain, nil
}

func (h *AuthHandler) sendVerificationEmail(ctx context.Context, toEmail, toName, token string) error {
	server, app := h.config.RuntimeSnapshot()
	verifyURL := strings.TrimRight(server.BaseURL, "/") + "/verify-email?token=" + token
	body := strings.TrimSpace("Hi " + toName + ",\n\n" +
		"Welcome to " + app.Name + ". Please verify your email by opening the link below:\n\n" +
		verifyURL + "\n\n" +
		"This link expires in 24 hours.\n")

	return h.mail.Send(ctx, mailservice.Message{
		ToEmail: toEmail,
		ToName:  toName,
		Subject: "Verify your email for " + app.Name,
		Text:    body,
	})
}
