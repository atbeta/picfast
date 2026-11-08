package handler

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type APITokenHandler struct {
	db *sqlc.Queries
}

func NewAPITokenHandler(db *sqlc.Queries) *APITokenHandler {
	return &APITokenHandler{db: db}
}

type createAPITokenRequest struct {
	Name       string   `json:"name"`
	ExpiresIn  string   `json:"expires_in"` // e.g. "30d", "90d", "1y", "never"
	Scopes     []string `json:"scopes"`
	StrategyID *int64   `json:"strategy_id"`
	AlbumID    *int64   `json:"album_id"`
}

type apiTokenResponse struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Token      string    `json:"token,omitempty"` // only shown on creation
	Scopes     []string  `json:"scopes"`
	StrategyID *int64    `json:"strategy_id,omitempty"`
	AlbumID    *int64    `json:"album_id,omitempty"`
	LastUsed   time.Time `json:"last_used_at,omitempty"`
	ExpiresAt  time.Time `json:"expires_at,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

func (h *APITokenHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createAPITokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		Fail(w, http.StatusBadRequest, "name is required")
		return
	}

	// Generate random token (32 bytes -> 64 hex chars)
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	plainToken := "img_" + hex.EncodeToString(b)
	hash := sha256.Sum256([]byte(plainToken))
	tokenHash := hex.EncodeToString(hash[:])

	// Parse expiration: supports Go duration strings plus "d" (days) and "y" (years).
	expiresAt, hasExpiry := parseTokenExpiry(req.ExpiresIn)

	if len(req.Scopes) == 0 {
		req.Scopes = []string{"read", "write"}
	}
	strategyID, albumID, err := h.resolveTokenUploadTarget(r.Context(), userID, req.StrategyID, req.AlbumID)
	if err != nil {
		Fail(w, http.StatusBadRequest, err.Error())
		return
	}
	scopes, err := json.Marshal(req.Scopes)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to encode token scopes")
		return
	}

	var expiresAtTz pgtype.Timestamptz
	if hasExpiry {
		expiresAtTz = pgtype.Timestamptz{Time: expiresAt, Valid: true}
	}

	token, err := h.db.CreateAPIToken(r.Context(), sqlc.CreateAPITokenParams{
		UserID:     userID,
		Name:       req.Name,
		TokenHash:  tokenHash,
		Scopes:     scopes,
		ExpiresAt:  expiresAtTz,
		StrategyID: strategyID,
		AlbumID:    albumID,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	var scopeList []string
	if err := json.Unmarshal(token.Scopes, &scopeList); err != nil || scopeList == nil {
		scopeList = []string{"read", "write"}
	}

	resp := apiTokenResponse{
		ID:         token.ID,
		Name:       token.Name,
		Token:      plainToken, // ONLY shown once on creation
		Scopes:     scopeList,
		StrategyID: domain.PgInt8PtrVal(token.StrategyID),
		AlbumID:    domain.PgInt8PtrVal(token.AlbumID),
		CreatedAt:  token.CreatedAt,
	}
	if token.ExpiresAt.Valid {
		resp.ExpiresAt = token.ExpiresAt.Time
	}
	writeAuditLog(h.db, r, "api_token.create", "api_token", fmt.Sprintf("%d", token.ID), token.Name, map[string]any{
		"scopes":      scopeList,
		"expires_at":  token.ExpiresAt.Time,
		"strategy_id": domain.PgInt8PtrVal(token.StrategyID),
		"album_id":    domain.PgInt8PtrVal(token.AlbumID),
	})

	Created(w, resp)
}

func (h *APITokenHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	tokens, err := h.db.ListAPITokensByUser(r.Context(), userID)
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to list tokens")
		return
	}

	items := make([]apiTokenResponse, len(tokens))
	for i, t := range tokens {
		var scopeList []string
		if err := json.Unmarshal(t.Scopes, &scopeList); err != nil || scopeList == nil {
			scopeList = []string{"read", "write"}
		}
		items[i] = apiTokenResponse{
			ID:         t.ID,
			Name:       t.Name,
			Scopes:     scopeList,
			StrategyID: domain.PgInt8PtrVal(t.StrategyID),
			AlbumID:    domain.PgInt8PtrVal(t.AlbumID),
			CreatedAt:  t.CreatedAt,
		}
		if t.LastUsedAt.Valid {
			items[i].LastUsed = t.LastUsedAt.Time
		}
		if t.ExpiresAt.Valid {
			items[i].ExpiresAt = t.ExpiresAt.Time
		}
	}

	Success(w, items)
}

func (h *APITokenHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	idStr := chi.URLParam(r, "id")
	var id int64
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		Fail(w, http.StatusBadRequest, "invalid token id")
		return
	}

	if err := h.db.DeleteAPIToken(r.Context(), sqlc.DeleteAPITokenParams{
		ID:     id,
		UserID: userID,
	}); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to delete token")
		return
	}
	writeAuditLog(h.db, r, "api_token.delete", "api_token", fmt.Sprintf("%d", id), "", nil)

	SuccessMessage(w, "token deleted")
}

func (h *APITokenHandler) resolveTokenUploadTarget(ctx context.Context, userID int64, strategyID, albumID *int64) (pgtype.Int8, pgtype.Int8, error) {
	var strategy, album pgtype.Int8
	if strategyID == nil && albumID == nil {
		return strategy, album, nil
	}

	user, err := h.db.GetUserByID(ctx, userID)
	if err != nil {
		return strategy, album, fmt.Errorf("user not found")
	}

	if strategyID != nil {
		if *strategyID <= 0 {
			return strategy, album, fmt.Errorf("invalid strategy_id")
		}
		if _, err := h.db.GetStrategyByID(ctx, *strategyID); err != nil {
			return strategy, album, fmt.Errorf("strategy not found")
		}
		if user.Role != string(domain.RoleAdmin) {
			if !user.GroupID.Valid {
				return strategy, album, fmt.Errorf("strategy not available for your group")
			}
			strategies, err := h.db.GetGroupStrategies(ctx, user.GroupID.Int64)
			if err != nil {
				return strategy, album, fmt.Errorf("strategy not available for your group")
			}
			found := false
			for _, item := range strategies {
				if item.ID == *strategyID {
					found = true
					break
				}
			}
			if !found {
				return strategy, album, fmt.Errorf("strategy not available for your group")
			}
		}
		strategy = domain.PgInt8(*strategyID)
	}

	if albumID != nil {
		if *albumID <= 0 {
			return strategy, album, fmt.Errorf("invalid album_id")
		}
		record, err := h.db.GetAlbumByID(ctx, *albumID)
		if err != nil || record.UserID != userID {
			return strategy, album, fmt.Errorf("album not found")
		}
		album = domain.PgInt8(*albumID)
	}

	return strategy, album, nil
}

// parseTokenExpiry converts an expiry string to an optional time.
// Supports Go duration strings ("24h", "72h"), compact forms ("30d", "90d", "1y"),
// and "never" or empty string for no expiry.
func parseTokenExpiry(s string) (time.Time, bool) {
	if s == "" || s == "never" {
		return time.Time{}, false
	}
	// Try standard Go duration parsing first
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(d), true
	}
	// Compact forms with day/year suffix
	var unit time.Duration
	switch {
	case len(s) > 1 && s[len(s)-1] == 'd':
		s = s[:len(s)-1]
		unit = 24 * time.Hour
	case len(s) > 1 && s[len(s)-1] == 'y':
		s = s[:len(s)-1]
		unit = 365 * 24 * time.Hour
	default:
		return time.Time{}, false
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n <= 0 {
		return time.Time{}, false
	}
	return time.Now().Add(time.Duration(n) * unit), true
}
