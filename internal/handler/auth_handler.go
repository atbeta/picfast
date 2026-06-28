package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/clientip"
	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/events"
	mailservice "github.com/atbeta/picfast/internal/service/mail"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const (
	emailVerificationTokenTTL = 24 * time.Hour
	passwordResetTokenTTL     = time.Hour
	mailActionEmailCooldown   = 2 * time.Minute
	bcryptCost                = bcrypt.DefaultCost
)

type AuthHandler struct {
	db      *sqlc.Queries
	pool    *pgxpool.Pool
	jwt     *JWTService
	config  *config.Config
	mail    mailservice.Sender
	emitter events.Emitter
}

func NewAuthHandler(db *sqlc.Queries, pool *pgxpool.Pool, jwtSvc *JWTService, cfg *config.Config, sender mailservice.Sender, emitter events.Emitter) *AuthHandler {
	if sender == nil {
		sender = mailservice.NewNoopSender(false)
	}
	return &AuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg, mail: sender, emitter: emitter}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Language string `json:"language,omitempty"`
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
	Email    string `json:"email"`
	Language string `json:"language,omitempty"`
}

type ForgotPasswordRequest struct {
	Email    string `json:"email"`
	Language string `json:"language,omitempty"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

type RegisterResponse struct {
	RequiresEmailVerification bool               `json:"requires_email_verification"`
	VerificationEmailSent     bool               `json:"verification_email_sent"`
	Tokens                    *domain.AuthTokens `json:"tokens,omitempty"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	userCount, err := h.db.CountUsers(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to read setup status")
		return
	}
	if userCount == 0 {
		Fail(w, http.StatusConflict, "setup required")
		return
	}

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
	req.Language = strings.TrimSpace(req.Language)
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcryptCost)
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
			Password:      pgtype.Text{String: string(hash), Valid: true},
			Name:          req.Name,
			Role:          string(domain.RoleUser),
			CapacityBytes: resolveUserCapacity(group, app),
			Settings:      settings,
			Status:        int16(domain.UserStatusActive),
			EmailVerified: false,
			RegisteredIp:  clientip.FromRequest(r),
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

	ev := events.BuildUserRegistered(user.ID, user.Email, user.Name, domain.PgInt8Val(user.GroupID), requiresVerification)
	ev.Actor = events.UserActor(user.Email, user.Name, user.ID)
	events.EmitAsync(h.emitter, ev)

	if requiresVerification {
		resp.VerificationEmailSent = h.sendVerificationEmail(
			r.Context(),
			user.Email,
			user.Name,
			verificationToken,
			resolveEmailLanguage(req.Language, r.Header.Get("Accept-Language")),
		) == nil
		if !resp.VerificationEmailSent {
			slog.Warn("failed to send verification email after registration", "email", user.Email)
		}
		Created(w, resp)
		return
	}

	resp.Tokens = tokens
	setAccessTokenCookie(w, r, tokens.AccessToken, tokens.ExpiresIn)
	Created(w, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		Fail(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if user.Status != int16(domain.UserStatusActive) {
		h.auditAdminLogin(r, user, false, "frozen_user")
		Fail(w, http.StatusForbidden, "account is frozen")
		return
	}
	if h.emailVerificationEnabled() && !user.EmailVerified {
		h.auditAdminLogin(r, user, false, "email_unverified")
		Fail(w, http.StatusForbidden, "email verification required")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password.String), []byte(req.Password)); err != nil {
		h.auditAdminLogin(r, user, false, "invalid_credentials")
		Fail(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	groupID := domain.PgInt8Val(user.GroupID)

	tokens, err := h.generateTokens(r.Context(), h.db, user.ID, domain.UserRole(user.Role), groupID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	h.auditAdminLogin(r, user, true, "")
	writeAuthTokens(w, r, tokens)
}

func (h *AuthHandler) auditAdminLogin(r *http.Request, user sqlc.User, success bool, reason string) {
	if domain.UserRole(user.Role) != domain.RoleAdmin {
		return
	}
	action := "admin.auth.login.success"
	details := map[string]any{
		"email": user.Email,
	}
	if !success {
		action = "admin.auth.login.failed"
		details["reason"] = reason
	}
	ctx := context.WithValue(r.Context(), domain.ContextKeyUserID, user.ID)
	writeAuditLog(h.db, r.WithContext(ctx), action, "auth", strconv.FormatInt(user.ID, 10), user.Email, details)
}

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	const verifySuccessMessage = "email verified"
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
		SuccessMessage(w, verifySuccessMessage)
		return
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = h.db.DeleteEmailVerificationTokenByHash(r.Context(), tokenHash)
		Fail(w, http.StatusBadRequest, "verification token has expired")
		return
	}

	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		if _, err := qtx.MarkEmailVerificationTokenUsed(r.Context(), stored.ID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
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

	SuccessMessage(w, verifySuccessMessage)
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	const genericMessage = "if the account exists, a verification email has been sent"
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
	req.Language = strings.TrimSpace(req.Language)

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil || user.EmailVerified {
		SuccessMessage(w, genericMessage)
		return
	}

	if latest, qErr := h.db.GetLatestUnusedEmailVerificationTokenByUser(r.Context(), user.ID); qErr == nil {
		if cooldown := remainingCooldown(latest.CreatedAt, mailActionEmailCooldown); cooldown > 0 {
			setMailActionCooldownHeader(w, cooldown)
			SuccessMessage(w, genericMessage)
			return
		}
	} else if !errors.Is(qErr, pgx.ErrNoRows) {
		slog.Warn("failed to read latest email verification token", "email", user.Email, "error", qErr)
		SuccessMessage(w, genericMessage)
		return
	}

	plain, err := h.issueVerificationToken(r.Context(), h.db, user.ID)
	if err != nil {
		slog.Warn("failed to create verification token", "email", user.Email, "error", err)
		SuccessMessage(w, genericMessage)
		return
	}
	if err := h.sendVerificationEmail(
		r.Context(),
		user.Email,
		user.Name,
		plain,
		resolveEmailLanguage(req.Language, r.Header.Get("Accept-Language")),
	); err != nil {
		slog.Warn("failed to resend verification email", "email", user.Email, "error", err)
	}

	setMailActionCooldownHeader(w, mailActionEmailCooldown)
	SuccessMessage(w, genericMessage)
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	const genericMessage = "if the account exists, a password reset email has been sent"
	if h.mail == nil || !h.mail.Ready() {
		Fail(w, http.StatusServiceUnavailable, "password reset is not available")
		return
	}

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Language = strings.TrimSpace(req.Language)
	if req.Email == "" {
		Fail(w, http.StatusBadRequest, "email is required")
		return
	}
	if !isValidEmail(req.Email) {
		Fail(w, http.StatusBadRequest, "invalid email format")
		return
	}

	user, err := h.db.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		SuccessMessage(w, genericMessage)
		return
	}

	if latest, qErr := h.db.GetLatestUnusedPasswordResetTokenByUser(r.Context(), user.ID); qErr == nil {
		if cooldown := remainingCooldown(latest.CreatedAt, mailActionEmailCooldown); cooldown > 0 {
			setMailActionCooldownHeader(w, cooldown)
			SuccessMessage(w, genericMessage)
			return
		}
	} else if !errors.Is(qErr, pgx.ErrNoRows) {
		slog.Warn("failed to read latest password reset token", "email", user.Email, "error", qErr)
		SuccessMessage(w, genericMessage)
		return
	}

	plain, err := h.issuePasswordResetToken(r.Context(), h.db, user.ID)
	if err != nil {
		slog.Warn("failed to create password reset token", "email", user.Email, "error", err)
		SuccessMessage(w, genericMessage)
		return
	}
	if err := h.sendPasswordResetEmail(
		r.Context(),
		user.Email,
		user.Name,
		plain,
		resolveEmailLanguage(req.Language, r.Header.Get("Accept-Language")),
	); err != nil {
		slog.Warn("failed to send password reset email", "email", user.Email, "error", err)
	}

	setMailActionCooldownHeader(w, mailActionEmailCooldown)
	SuccessMessage(w, genericMessage)
}

func remainingCooldown(createdAt time.Time, cooldown time.Duration) time.Duration {
	remaining := cooldown - time.Since(createdAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func setMailActionCooldownHeader(w http.ResponseWriter, cooldown time.Duration) {
	seconds := int(math.Ceil(cooldown.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	w.Header().Set("X-Cooldown-Seconds", strconv.Itoa(seconds))
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		Fail(w, http.StatusBadRequest, "token is required")
		return
	}
	if len(req.NewPassword) < 8 || len(req.NewPassword) > 72 {
		Fail(w, http.StatusBadRequest, "password must be between 8 and 72 bytes")
		return
	}

	tokenHash := HashRefreshToken(req.Token)
	stored, err := h.db.GetPasswordResetTokenByHash(r.Context(), tokenHash)
	if err != nil {
		Fail(w, http.StatusBadRequest, "invalid reset token")
		return
	}
	if stored.UsedAt.Valid {
		Fail(w, http.StatusBadRequest, "reset token has already been used")
		return
	}
	if time.Now().After(stored.ExpiresAt) {
		_ = h.db.DeletePasswordResetTokenByHash(r.Context(), tokenHash)
		Fail(w, http.StatusBadRequest, "reset token has expired")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcryptCost)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		if _, err := qtx.MarkPasswordResetTokenUsed(r.Context(), stored.ID); err != nil {
			return err
		}
		if err := qtx.UpdateUserPasswordByID(r.Context(), sqlc.UpdateUserPasswordByIDParams{
			ID:       stored.UserID,
			Password: pgtype.Text{String: string(hash), Valid: true},
		}); err != nil {
			return err
		}
		if err := qtx.DeleteUnusedPasswordResetTokensByUser(r.Context(), stored.UserID); err != nil {
			return err
		}
		return qtx.DeleteAllUserRefreshTokens(r.Context(), stored.UserID)
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to reset password")
		return
	}

	SuccessMessage(w, "password reset successfully")
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			Fail(w, http.StatusBadRequest, "invalid request body")
			return
		}
	}

	if req.RefreshToken == "" {
		if c, err := r.Cookie(oauthRefreshCookie); err == nil {
			req.RefreshToken = c.Value
		}
	}
	if req.RefreshToken == "" {
		Fail(w, http.StatusUnauthorized, "refresh token required")
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

	writeAuthTokens(w, r, tokens)
	setRefreshTokenCookie(w, r, tokens.RefreshToken, h.config.JWT.RefreshTTL)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	h.db.DeleteAllUserRefreshTokens(r.Context(), userID)
	clearAccessTokenCookie(w, r)
	clearRefreshTokenCookie(w, r)
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

func (h *AuthHandler) issuePasswordResetToken(ctx context.Context, qtx *sqlc.Queries, userID int64) (string, error) {
	if err := qtx.DeleteUnusedPasswordResetTokensByUser(ctx, userID); err != nil {
		return "", err
	}
	plain, tokenHash, err := GenerateRefreshToken()
	if err != nil {
		return "", err
	}
	_, err = qtx.CreatePasswordResetToken(ctx, sqlc.CreatePasswordResetTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(passwordResetTokenTTL),
	})
	if err != nil {
		return "", err
	}
	return plain, nil
}

func (h *AuthHandler) sendVerificationEmail(ctx context.Context, toEmail, toName, token, language string) error {
	baseURL, app := h.resolveAppWebBaseURL()
	verifyURL := strings.TrimRight(baseURL, "/") + "/verify-email?token=" + token
	subject, body := buildVerificationEmail(app.Name, toName, verifyURL, language)

	return h.mail.Send(ctx, mailservice.Message{
		ToEmail: toEmail,
		ToName:  toName,
		Subject: subject,
		Text:    body,
	})
}

func (h *AuthHandler) sendPasswordResetEmail(ctx context.Context, toEmail, toName, token, language string) error {
	baseURL, app := h.resolveAppWebBaseURL()
	resetURL := strings.TrimRight(baseURL, "/") + "/reset-password?token=" + token
	subject, body := buildPasswordResetEmail(app.Name, toName, resetURL, language)

	return h.mail.Send(ctx, mailservice.Message{
		ToEmail: toEmail,
		ToName:  toName,
		Subject: subject,
		Text:    body,
	})
}

func resolveEmailLanguage(explicit, acceptLanguage string) string {
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(explicit)), "zh") {
		return "zh-CN"
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(explicit)), "en") {
		return "en-US"
	}
	if strings.Contains(strings.ToLower(acceptLanguage), "zh") {
		return "zh-CN"
	}
	return "en-US"
}

func (h *AuthHandler) resolveAppWebBaseURL() (string, config.AppConfig) {
	server, app := h.config.RuntimeSnapshot()
	if strings.TrimSpace(app.WebBaseURL) != "" {
		return strings.TrimSpace(app.WebBaseURL), app
	}
	return server.BaseURL, app
}

func buildVerificationEmail(appName, toName, verifyURL, language string) (subject, body string) {
	if language == "zh-CN" {
		subject = appName + " 邮箱验证"
		body = strings.TrimSpace("你好 " + toName + "，\n\n" +
			"欢迎使用 " + appName + "。请点击下方链接完成邮箱验证：\n\n" +
			verifyURL + "\n\n" +
			"该链接 24 小时内有效。\n")
		return subject, body
	}
	subject = "Verify your email for " + appName
	body = strings.TrimSpace("Hi " + toName + ",\n\n" +
		"Welcome to " + appName + ". Please verify your email by opening the link below:\n\n" +
		verifyURL + "\n\n" +
		"This link expires in 24 hours.\n")
	return subject, body
}

func buildPasswordResetEmail(appName, toName, resetURL, language string) (subject, body string) {
	if language == "zh-CN" {
		subject = appName + " 重置密码"
		body = strings.TrimSpace("你好 " + toName + "，\n\n" +
			"我们收到了重置 " + appName + " 账号密码的请求。请点击下方链接设置新密码：\n\n" +
			resetURL + "\n\n" +
			"该链接 1 小时内有效，且仅可使用一次。\n" +
			"如果这不是你的操作，请忽略此邮件。\n")
		return subject, body
	}
	subject = "Reset your password for " + appName
	body = strings.TrimSpace("Hi " + toName + ",\n\n" +
		"We received a request to reset your password for " + appName + ". Open the link below to set a new password:\n\n" +
		resetURL + "\n\n" +
		"This link expires in 1 hour and can only be used once.\n" +
		"If you did not request this, you can safely ignore this email.\n")
	return subject, body
}
