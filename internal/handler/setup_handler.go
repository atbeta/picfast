package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

const setupAdvisoryLockKey int64 = 0x7069636661737401

type SetupHandler struct {
	db     *sqlc.Queries
	pool   *pgxpool.Pool
	jwt    *JWTService
	config *config.Config
}

func NewSetupHandler(db *sqlc.Queries, pool *pgxpool.Pool, jwtSvc *JWTService, cfg *config.Config) *SetupHandler {
	return &SetupHandler{db: db, pool: pool, jwt: jwtSvc, config: cfg}
}

type SetupStatusResponse struct {
	Required bool `json:"required"`
}

type SetupAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	required, err := h.setupRequired(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to read setup status")
		return
	}
	Success(w, SetupStatusResponse{Required: required})
}

func (h *SetupHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
	var req SetupAdminRequest
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

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to start setup")
		return
	}
	defer tx.Rollback(r.Context())

	if _, err := tx.Exec(r.Context(), "SELECT pg_advisory_xact_lock($1)", setupAdvisoryLockKey); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to lock setup")
		return
	}

	qtx := sqlc.New(tx)
	userCount, err := qtx.CountUsers(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to read setup status")
		return
	}
	if userCount > 0 {
		Fail(w, http.StatusConflict, "setup has already been completed")
		return
	}

	group, err := qtx.GetDefaultGroup(r.Context())
	if err != nil {
		Fail(w, http.StatusInternalServerError, "no default group found")
		return
	}

	settings, _ := json.Marshal(domain.UserSettings{})
	user, err := qtx.CreateAdminUser(r.Context(), sqlc.CreateAdminUserParams{
		GroupID:       domain.PgInt8(group.ID),
		Email:         req.Email,
		Password:      string(hash),
		Name:          req.Name,
		Role:          string(domain.RoleAdmin),
		CapacityBytes: h.config.AppSnapshot().UserInitialCapacity,
		Settings:      settings,
		Status:        int16(domain.UserStatusActive),
		EmailVerified: true,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create admin user")
		return
	}

	tokens, err := h.generateTokens(r.Context(), qtx, user.ID, group.ID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate tokens")
		return
	}

	if err := tx.Commit(r.Context()); err != nil {
		if errors.Is(err, pgx.ErrTxClosed) {
			Fail(w, http.StatusConflict, "setup has already been completed")
			return
		}
		Fail(w, http.StatusInternalServerError, "failed to complete setup")
		return
	}

	Created(w, tokens)
}

func (h *SetupHandler) setupRequired(ctx context.Context) (bool, error) {
	count, err := h.db.CountUsers(ctx)
	if err != nil {
		return false, err
	}
	return count == 0, nil
}

func (h *SetupHandler) generateTokens(ctx context.Context, qtx *sqlc.Queries, userID int64, groupID int64) (*domain.AuthTokens, error) {
	accessToken, expiresIn, err := h.jwt.GenerateAccessToken(userID, domain.RoleAdmin, groupID)
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
