package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db     *sqlc.Queries
	pool   *pgxpool.Pool
	jwt    *JWTService
	config *config.Config
}

func NewAuthHandler(db *sqlc.Queries, pool *pgxpool.Pool, jwtSvc *JWTService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}
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

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !h.config.App.AllowRegistration {
		Fail(w, http.StatusForbidden, "registration is disabled")
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		Fail(w, http.StatusBadRequest, "email, password and name are required")
		return
	}
	if len(req.Password) < 8 {
		Fail(w, http.StatusBadRequest, "password must be at least 8 characters")
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

	var tokens *domain.AuthTokens
	err = sqlc.RunInTx(r.Context(), h.pool, func(qtx *sqlc.Queries) error {
		user, err := qtx.CreateUser(r.Context(), sqlc.CreateUserParams{
			GroupID:       domain.PgInt8(group.ID),
			Email:         req.Email,
			Password:      string(hash),
			Name:          req.Name,
			Role:          string(domain.RoleUser),
			CapacityBytes: h.config.App.UserInitialCapacity,
			Settings:      settings,
			Status:        int16(domain.UserStatusActive),
			EmailVerified: false,
			RegisteredIp:  r.RemoteAddr,
		})
		if err != nil {
			return err
		}
		tokens, err = h.generateTokens(r.Context(), qtx, user.ID, domain.RoleUser, group.ID)
		return err
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	Created(w, tokens)
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
