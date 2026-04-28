package handler

import (
	"encoding/json"
	"net/http"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	db *sqlc.Queries
}

func NewUserHandler(db *sqlc.Queries) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		Fail(w, http.StatusNotFound, "user not found")
		return
	}

	usedCapacity, _ := h.db.GetUserUsedCapacity(r.Context(), pgtype.Int8{Int64: userID, Valid: true})

	Success(w, map[string]interface{}{
		"id":             user.ID,
		"email":          user.Email,
		"name":           user.Name,
		"role":           user.Role,
		"status":         user.Status,
		"capacity_bytes": user.CapacityBytes,
		"used_bytes":     usedCapacity,
		"image_num":      user.ImageNum,
		"album_num":      user.AlbumNum,
		"settings":       json.RawMessage(user.Settings),
		"email_verified": user.EmailVerified,
		"created_at":     user.CreatedAt,
	})
}

type UpdateProfileRequest struct {
	Name     *string          `json:"name"`
	Password *string          `json:"password"`
	Settings *json.RawMessage `json:"settings"`
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKeyUserID).(int64)
	if !ok {
		Fail(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.db.GetUserByID(r.Context(), userID)
	if err != nil {
		Fail(w, http.StatusNotFound, "user not found")
		return
	}

	name := user.Name
	password := user.Password
	var settings json.RawMessage = user.Settings

	if req.Name != nil {
		name = *req.Name
	}
	if req.Password != nil {
		if len(*req.Password) < 8 {
			Fail(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		password = string(hash)
	}
	if req.Settings != nil {
		settings = *req.Settings
	}

	updated, err := h.db.UpdateUser(r.Context(), sqlc.UpdateUserParams{
		ID:            userID,
		Name:          name,
		Password:      password,
		GroupID:       user.GroupID,
		CapacityBytes: user.CapacityBytes,
		Status:        user.Status,
		Settings:      settings,
	})
	if err != nil {
		Fail(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	Success(w, map[string]interface{}{
		"id":         updated.ID,
		"email":      updated.Email,
		"name":       updated.Name,
		"settings":   json.RawMessage(updated.Settings),
		"updated_at": updated.UpdatedAt,
	})
}
