package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	errUnknownSettingsKey = errors.New("unknown settings key")
	errInvalidSettingsType = errors.New("invalid settings value type")
)

var validThemeModes = map[string]bool{"light": true, "dark": true, "system": true}
var validDensities = map[string]bool{"compact": true, "comfortable": true, "spacious": true}
var validMotions = map[string]bool{"none": true, "subtle": true, "playful": true}
var validCopyFormats = map[string]bool{"url": true, "markdown": true, "html": true, "bbcode": true, "thumbnail": true, "custom": true}
var validImageFormats = map[string]bool{"": true, "origin": true, "jpeg": true, "jpg": true, "png": true, "webp": true}

var knownUserSettingsKeys = map[string]bool{
	"default_strategy":   true,
	"default_album":      true,
	"default_permission": true,
	"image_processing":   true,
	"default_copy_format": true,
	"copy_template":       true,
	"theme_override":      true,
	"language":            true,
}

func validateUserSettings(raw json.RawMessage) error {
	if len(raw) == 0 || string(raw) == "null" || string(raw) == "{}" {
		return nil
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return errInvalidSettingsType
	}

	for key := range m {
		if !knownUserSettingsKeys[key] {
			return fmt.Errorf("%w: %s", errUnknownSettingsKey, key)
		}
	}

	if rawTO, ok := m["theme_override"]; ok {
		var to domain.ThemeOverride
		if err := json.Unmarshal(rawTO, &to); err != nil {
			return fmt.Errorf("theme_override: %w", errInvalidSettingsType)
		}
		for k := range mustMap(rawTO) {
			switch k {
			case "preset", "mode", "density", "motion":
			default:
				return fmt.Errorf("%w: theme_override.%s", errUnknownSettingsKey, k)
			}
		}
		if to.Mode != "" && !validThemeModes[to.Mode] {
			return fmt.Errorf("theme_override.mode: must be light, dark, or system")
		}
		if to.Density != "" && !validDensities[to.Density] {
			return fmt.Errorf("theme_override.density: must be compact, comfortable, or spacious")
		}
		if to.Motion != "" && !validMotions[to.Motion] {
			return fmt.Errorf("theme_override.motion: must be none, subtle, or playful")
		}
	}

	if rawCF, ok := m["default_copy_format"]; ok {
		var cf string
		if err := json.Unmarshal(rawCF, &cf); err != nil {
			return fmt.Errorf("default_copy_format: %w", errInvalidSettingsType)
		}
		if cf != "" && !validCopyFormats[cf] {
			return fmt.Errorf("default_copy_format: must be one of url, markdown, html, bbcode, thumbnail, custom")
		}
	}

	if rawIP, ok := m["image_processing"]; ok {
		if err := validateImageProcessing(rawIP); err != nil {
			return err
		}
	}

	return nil
}

func validateImageProcessing(raw json.RawMessage) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(raw, &m); err != nil {
		return fmt.Errorf("image_processing: %w", errInvalidSettingsType)
	}
	for key := range m {
		switch key {
		case "image_save_quality", "image_save_format", "is_strip_exif", "is_enable_watermark", "watermark_configs":
		default:
			return fmt.Errorf("%w: image_processing.%s", errUnknownSettingsKey, key)
		}
	}
	if rawFmt, ok := m["image_save_format"]; ok {
		var fmtStr string
		if err := json.Unmarshal(rawFmt, &fmtStr); err != nil {
			return fmt.Errorf("image_processing.image_save_format: %w", errInvalidSettingsType)
		}
		if !validImageFormats[fmtStr] {
			return fmt.Errorf("image_processing.image_save_format: must be origin, jpeg, png, or webp")
		}
	}
	return nil
}

func mustMap(raw json.RawMessage) map[string]json.RawMessage {
	var m map[string]json.RawMessage
	json.Unmarshal(raw, &m)
	return m
}

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

	usedCapacity, err := h.db.GetUserUsedCapacity(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		slog.Warn("failed to load user used capacity", "error", err, "user_id", userID)
		usedCapacity = 0
	}

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
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcryptCost)
		if err != nil {
			Fail(w, http.StatusInternalServerError, "failed to hash password")
			return
		}
		password = pgtype.Text{String: string(hash), Valid: true}
	}
	if req.Settings != nil {
		if err := validateUserSettings(*req.Settings); err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}
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

	usedCapacity, err := h.db.GetUserUsedCapacity(r.Context(), pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		slog.Warn("failed to load user used capacity", "error", err, "user_id", userID)
		usedCapacity = 0
	}

	Success(w, map[string]interface{}{
		"id":             updated.ID,
		"email":          updated.Email,
		"name":           updated.Name,
		"role":           updated.Role,
		"status":         updated.Status,
		"capacity_bytes": updated.CapacityBytes,
		"used_bytes":     usedCapacity,
		"image_num":      updated.ImageNum,
		"album_num":      updated.AlbumNum,
		"settings":       json.RawMessage(updated.Settings),
		"updated_at":     updated.UpdatedAt,
	})
}
