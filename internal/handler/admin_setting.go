package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pbeta/imgapi/internal/config"
)

type AdminSettingHandler struct {
	config *config.Config
	setter *config.Setter
}

func NewAdminSettingHandler(cfg *config.Config, setter *config.Setter) *AdminSettingHandler {
	return &AdminSettingHandler{config: cfg, setter: setter}
}

func (h *AdminSettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	Success(w, map[string]interface{}{
		"app_name":                h.config.App.Name,
		"allow_guest_upload":      h.config.App.AllowGuestUpload,
		"allow_registration":      h.config.App.AllowRegistration,
		"user_initial_capacity":   h.config.App.UserInitialCapacity,
		"moderation_mode":         h.config.App.ModerationMode,
	})
}

type updateSettingsRequest struct {
	AppName              *string `json:"app_name"`
	AllowGuestUpload     *bool   `json:"allow_guest_upload"`
	AllowRegistration    *bool   `json:"allow_registration"`
	UserInitialCapacity  *int64  `json:"user_initial_capacity"`
	ModerationMode       *string `json:"moderation_mode"`
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
	if req.AllowGuestUpload != nil {
		h.setter.SetAllowGuestUpload(*req.AllowGuestUpload)
	}
	if req.AllowRegistration != nil {
		h.setter.SetAllowRegistration(*req.AllowRegistration)
	}
	if req.UserInitialCapacity != nil {
		h.setter.SetUserInitialCapacity(*req.UserInitialCapacity)
	}
	if req.ModerationMode != nil {
		h.setter.SetModerationMode(*req.ModerationMode)
	}

	Success(w, map[string]interface{}{
		"app_name":                h.config.App.Name,
		"allow_guest_upload":      h.config.App.AllowGuestUpload,
		"allow_registration":      h.config.App.AllowRegistration,
		"user_initial_capacity":   h.config.App.UserInitialCapacity,
		"moderation_mode":         h.config.App.ModerationMode,
	})
}
