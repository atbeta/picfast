package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
)

type APITokenHandler struct {
	db *sqlc.Queries
}

func NewAPITokenHandler(db *sqlc.Queries) *APITokenHandler {
	return &APITokenHandler{db: db}
}

type createAPITokenRequest struct {
	Name      string   `json:"name"`
	ExpiresIn string   `json:"expires_in"` // e.g. "30d", "90d", "1y", "never"
	Scopes    []string `json:"scopes"`
}

type apiTokenResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"` // only shown on creation
	Scopes    []string  `json:"scopes"`
	LastUsed  time.Time `json:"last_used_at,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time `json:"created_at"`
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
	expiresAt := parseTokenExpiry(req.ExpiresIn)

	scopes, _ := json.Marshal(req.Scopes)
	if len(req.Scopes) == 0 {
		scopes = []byte(`["read","write"]`)
	}

	var expiresAtTz pgtype.Timestamptz
	if t, ok := expiresAt.(time.Time); ok {
		expiresAtTz = pgtype.Timestamptz{Time: t, Valid: true}
	}

	token, err := h.db.CreateAPIToken(r.Context(), sqlc.CreateAPITokenParams{
		UserID:    userID,
		Name:      req.Name,
		TokenHash: tokenHash,
		Scopes:    scopes,
		ExpiresAt: expiresAtTz,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to create token")
		return
	}

	var scopeList []string
	_ = json.Unmarshal(token.Scopes, &scopeList)

	resp := apiTokenResponse{
		ID:        token.ID,
		Name:      token.Name,
		Token:     plainToken, // ONLY shown once on creation
		Scopes:    scopeList,
		CreatedAt: token.CreatedAt,
	}
	if token.ExpiresAt.Valid {
		resp.ExpiresAt = token.ExpiresAt.Time
	}

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
		_ = json.Unmarshal(t.Scopes, &scopeList)
		items[i] = apiTokenResponse{
			ID:        t.ID,
			Name:      t.Name,
			Scopes:    scopeList,
			CreatedAt: t.CreatedAt,
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

	SuccessMessage(w, "token deleted")
}

// parseTokenExpiry converts an expiry string to an optional time.
// Supports Go duration strings ("24h", "72h"), compact forms ("30d", "90d", "1y"),
// and "never" or empty string for no expiry.
func parseTokenExpiry(s string) interface{} {
	if s == "" || s == "never" {
		return nil
	}
	// Try standard Go duration parsing first
	if d, err := time.ParseDuration(s); err == nil {
		return time.Now().Add(d)
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
		return nil
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n <= 0 {
		return nil
	}
	return time.Now().Add(time.Duration(n) * unit)
}
