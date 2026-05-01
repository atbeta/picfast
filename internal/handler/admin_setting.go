package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/sqlc"
)

type AdminSettingHandler struct {
	config    *config.Config
	setter    *config.Setter
	queries   *sqlc.Queries
	mailReady bool
}

func NewAdminSettingHandler(cfg *config.Config, setter *config.Setter, queries *sqlc.Queries, mailReady bool) *AdminSettingHandler {
	return &AdminSettingHandler{config: cfg, setter: setter, queries: queries, mailReady: mailReady}
}

func (h *AdminSettingHandler) emailVerificationEnabled() bool {
	return h.config.App.RequireEmailVerification && h.mailReady
}

func (h *AdminSettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]interface{}{
		"app_name":                   h.config.App.Name,
		"app_url":                    h.config.Server.BaseURL,
		"allow_guest_upload":         h.config.App.AllowGuestUpload,
		"allow_registration":         h.config.App.AllowRegistration,
		"require_email_verification": h.emailVerificationEnabled(),
		"email_verification_ready":   h.mailReady,
		"user_initial_capacity":      h.config.App.UserInitialCapacity,
		"default_image_ttl":          h.config.App.DefaultImageTTL.String(),
		"moderation_mode":            h.config.App.ModerationMode,
	})
}

type updateSettingsRequest struct {
	AppName                  *string `json:"app_name"`
	AppURL                   *string `json:"app_url"`
	AllowGuestUpload         *bool   `json:"allow_guest_upload"`
	AllowRegistration        *bool   `json:"allow_registration"`
	RequireEmailVerification *bool   `json:"require_email_verification"`
	UserInitialCapacity      *int64  `json:"user_initial_capacity"`
	DefaultImageTTL          *string `json:"default_image_ttl"`
	ModerationMode           *string `json:"moderation_mode"`
}

func (h *AdminSettingHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req updateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AppName != nil {
		h.setter.SetAppName(*req.AppName)
	}
	if req.AppURL != nil {
		h.setter.SetBaseURL(*req.AppURL)
	}
	if req.AllowGuestUpload != nil {
		h.setter.SetAllowGuestUpload(*req.AllowGuestUpload)
	}
	if req.AllowRegistration != nil {
		h.setter.SetAllowRegistration(*req.AllowRegistration)
	}
	if req.RequireEmailVerification != nil {
		if *req.RequireEmailVerification && !h.mailReady {
			Fail(w, http.StatusBadRequest, "email verification requires a configured SMTP sender")
			return
		}
		h.setter.SetRequireEmailVerification(*req.RequireEmailVerification)
	}
	if req.UserInitialCapacity != nil {
		h.setter.SetUserInitialCapacity(*req.UserInitialCapacity)
	}
	if req.DefaultImageTTL != nil {
		if *req.DefaultImageTTL == "0" || *req.DefaultImageTTL == "" {
			h.setter.SetDefaultImageTTL(0)
		} else {
			d, err := time.ParseDuration(*req.DefaultImageTTL)
			if err != nil {
				Fail(w, http.StatusBadRequest, "invalid default_image_ttl format")
				return
			}
			h.setter.SetDefaultImageTTL(d)
		}
	}
	if req.ModerationMode != nil {
		h.setter.SetModerationMode(*req.ModerationMode)
	}
	if err := h.persist(r.Context()); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to persist settings")
		return
	}

	Success(w, map[string]interface{}{
		"app_name":                   h.config.App.Name,
		"app_url":                    h.config.Server.BaseURL,
		"allow_guest_upload":         h.config.App.AllowGuestUpload,
		"allow_registration":         h.config.App.AllowRegistration,
		"require_email_verification": h.emailVerificationEnabled(),
		"email_verification_ready":   h.mailReady,
		"user_initial_capacity":      h.config.App.UserInitialCapacity,
		"default_image_ttl":          h.config.App.DefaultImageTTL.String(),
		"moderation_mode":            h.config.App.ModerationMode,
	})
}

func (h *AdminSettingHandler) persist(ctx context.Context) error {
	_, err := h.queries.UpsertSiteSettings(ctx, sqlc.UpsertSiteSettingsParams{
		AppName:                  h.config.App.Name,
		AppUrl:                   h.config.Server.BaseURL,
		AllowGuestUpload:         h.config.App.AllowGuestUpload,
		AllowRegistration:        h.config.App.AllowRegistration,
		RequireEmailVerification: h.config.App.RequireEmailVerification,
		UserInitialCapacity:      h.config.App.UserInitialCapacity,
		DefaultImageTtl:          h.config.App.DefaultImageTTL.String(),
		ModerationMode:           h.config.App.ModerationMode,
	})
	return err
}
