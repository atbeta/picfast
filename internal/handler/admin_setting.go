package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
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
	return h.config.AppSnapshot().RequireEmailVerification && h.mailReady
}

func (h *AdminSettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	Success(w, h.settingsResponse(true))
}

type updateSettingsRequest struct {
	AppName                  *string          `json:"app_name"`
	AppURL                   *string          `json:"app_url"`
	SiteDescription          *string          `json:"site_description"`
	FaviconURL               *string          `json:"favicon_url"`
	AllowGuestUpload         *bool            `json:"allow_guest_upload"`
	GuestCapacityBytes       *int64           `json:"guest_capacity_bytes"`
	AllowRegistration        *bool            `json:"allow_registration"`
	AllowUserImageProcessing *bool            `json:"allow_user_image_processing"`
	RequireEmailVerification *bool            `json:"require_email_verification"`
	UserInitialCapacity      *int64           `json:"user_initial_capacity"`
	DefaultImageTTL          *string          `json:"default_image_ttl"`
	ModerationMode           *string          `json:"moderation_mode"`
	FooterText1              *string          `json:"footer_text_1"`
	FooterLink1              *string          `json:"footer_link_1"`
	FooterText2              *string          `json:"footer_text_2"`
	FooterLink2              *string          `json:"footer_link_2"`
	AnalyticsProvider        *string          `json:"analytics_provider"`
	AnalyticsConfig          *json.RawMessage `json:"analytics_config"`
}

func (h *AdminSettingHandler) Update(w http.ResponseWriter, r *http.Request) {
	before := h.settingsResponse(false)
	var req updateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		Fail(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.AppName != nil {
		h.setter.SetAppName(*req.AppName)
	}
	if req.AppURL != nil {
		if err := validateOptionalURL(*req.AppURL); err != nil {
			Fail(w, http.StatusBadRequest, "invalid app_url")
			return
		}
		h.setter.SetBaseURL(*req.AppURL)
	}
	if req.SiteDescription != nil {
		h.setter.SetSiteDescription(*req.SiteDescription)
	}
	if req.FaviconURL != nil {
		if err := validateOptionalURL(*req.FaviconURL); err != nil {
			Fail(w, http.StatusBadRequest, "invalid favicon_url")
			return
		}
		h.setter.SetFaviconURL(*req.FaviconURL)
	}
	if req.AllowGuestUpload != nil {
		h.setter.SetAllowGuestUpload(*req.AllowGuestUpload)
	}
	if req.GuestCapacityBytes != nil {
		if *req.GuestCapacityBytes < 0 {
			Fail(w, http.StatusBadRequest, "guest_capacity_bytes must be >= 0")
			return
		}
		h.setter.SetGuestCapacityBytes(*req.GuestCapacityBytes)
	}
	if req.AllowRegistration != nil {
		h.setter.SetAllowRegistration(*req.AllowRegistration)
	}
	if req.AllowUserImageProcessing != nil {
		h.setter.SetAllowUserImageProcessing(*req.AllowUserImageProcessing)
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
	if req.FooterLink1 != nil {
		if err := validateOptionalURL(*req.FooterLink1); err != nil {
			Fail(w, http.StatusBadRequest, "invalid footer_link_1")
			return
		}
	}
	if req.FooterLink2 != nil {
		if err := validateOptionalURL(*req.FooterLink2); err != nil {
			Fail(w, http.StatusBadRequest, "invalid footer_link_2")
			return
		}
	}
	if req.FooterText1 != nil || req.FooterLink1 != nil || req.FooterText2 != nil || req.FooterLink2 != nil {
		_, app := h.config.RuntimeSnapshot()
		text1 := app.FooterText1
		link1 := app.FooterLink1
		text2 := app.FooterText2
		link2 := app.FooterLink2
		if req.FooterText1 != nil {
			text1 = *req.FooterText1
		}
		if req.FooterLink1 != nil {
			link1 = *req.FooterLink1
		}
		if req.FooterText2 != nil {
			text2 = *req.FooterText2
		}
		if req.FooterLink2 != nil {
			link2 = *req.FooterLink2
		}
		h.setter.SetFooterItems(text1, link1, text2, link2)
	}
	if req.AnalyticsProvider != nil || req.AnalyticsConfig != nil {
		_, app := h.config.RuntimeSnapshot()
		provider := app.AnalyticsProvider
		analyticsConfig := cloneRawMessage(app.AnalyticsConfig)
		if req.AnalyticsProvider != nil {
			provider = strings.TrimSpace(*req.AnalyticsProvider)
		}
		if req.AnalyticsConfig != nil {
			analyticsConfig = cloneRawMessage(*req.AnalyticsConfig)
		}
		if len(analyticsConfig) == 0 {
			analyticsConfig = json.RawMessage(`{}`)
		}
		if err := validateAnalytics(provider, analyticsConfig); err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}
		h.setter.SetAnalytics(provider, analyticsConfig)
	}
	if err := h.persist(r.Context()); err != nil {
		Fail(w, http.StatusInternalServerError, "failed to persist settings")
		return
	}
	after := h.settingsResponse(false)
	_, app := h.config.RuntimeSnapshot()
	writeAuditLog(h.queries, r, "admin.settings.update", "site_settings", "1", app.Name, map[string]any{
		"before": before,
		"after":  after,
	})

	Success(w, h.settingsResponse(true))
}

func (h *AdminSettingHandler) persist(ctx context.Context) error {
	server, app := h.config.RuntimeSnapshot()
	_, err := h.queries.UpsertSiteSettings(ctx, sqlc.UpsertSiteSettingsParams{
		AppName:                  app.Name,
		AppUrl:                   server.BaseURL,
		AllowGuestUpload:         app.AllowGuestUpload,
		GuestCapacityBytes:       app.GuestCapacityBytes,
		AllowRegistration:        app.AllowRegistration,
		AllowUserImageProcessing: app.AllowUserImageProcessing,
		RequireEmailVerification: app.RequireEmailVerification,
		UserInitialCapacity:      app.UserInitialCapacity,
		DefaultImageTtl:          app.DefaultImageTTL.String(),
		ModerationMode:           app.ModerationMode,
		SiteDescription:          app.SiteDescription,
		FaviconUrl:               app.FaviconURL,
		FooterText1:              app.FooterText1,
		FooterLink1:              app.FooterLink1,
		FooterText2:              app.FooterText2,
		FooterLink2:              app.FooterLink2,
		AnalyticsProvider:        app.AnalyticsProvider,
		AnalyticsConfig:          normalizedRawMessage(app.AnalyticsConfig),
	})
	return err
}

func (h *AdminSettingHandler) settingsResponse(includeMailReady bool) map[string]interface{} {
	server, app := h.config.RuntimeSnapshot()
	resp := map[string]interface{}{
		"app_name":                    app.Name,
		"app_url":                     server.BaseURL,
		"site_description":            app.SiteDescription,
		"favicon_url":                 app.FaviconURL,
		"allow_guest_upload":          app.AllowGuestUpload,
		"guest_capacity_bytes":        app.GuestCapacityBytes,
		"allow_registration":          app.AllowRegistration,
		"allow_user_image_processing": app.AllowUserImageProcessing,
		"require_email_verification":  app.RequireEmailVerification,
		"user_initial_capacity":       app.UserInitialCapacity,
		"default_image_ttl":           app.DefaultImageTTL.String(),
		"moderation_mode":             app.ModerationMode,
		"footer_text_1":              app.FooterText1,
		"footer_link_1":              app.FooterLink1,
		"footer_text_2":              app.FooterText2,
		"footer_link_2":              app.FooterLink2,
		"analytics_provider":          app.AnalyticsProvider,
		"analytics_config":            normalizedRawMessage(app.AnalyticsConfig),
	}
	if includeMailReady {
		resp["require_email_verification"] = app.RequireEmailVerification && h.mailReady
		resp["email_verification_ready"] = h.mailReady
	}
	return resp
}

func validateOptionalURL(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	u, err := url.ParseRequestURI(value)
	if err != nil {
		return err
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("unsupported url scheme")
	}
	return nil
}

func validateAnalytics(provider string, raw json.RawMessage) error {
	switch provider {
	case "", "plausible", "umami", "ga4", "baidu", "custom":
	default:
		return &badRequestError{"invalid analytics_provider"}
	}
	if provider == "" {
		return nil
	}
	var cfg map[string]any
	if err := json.Unmarshal(normalizedRawMessage(raw), &cfg); err != nil {
		return &badRequestError{"invalid analytics_config"}
	}
	required := map[string][]string{
		"plausible": {"domain"},
		"umami":     {"script_url", "website_id"},
		"ga4":       {"measurement_id"},
		"baidu":     {"site_id"},
		"custom":    {"script"},
	}
	for _, key := range required[provider] {
		if strings.TrimSpace(stringValue(cfg[key])) == "" {
			return &badRequestError{"analytics_config missing " + key}
		}
	}
	if scriptURL := stringValue(cfg["script_url"]); scriptURL != "" {
		if err := validateOptionalURL(scriptURL); err != nil {
			return &badRequestError{"invalid analytics_config script_url"}
		}
	}
	return nil
}

func stringValue(v any) string {
	s, _ := v.(string)
	return s
}

func normalizedRawMessage(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" || string(raw) == "null" {
		return json.RawMessage(`{}`)
	}
	return raw
}

func cloneRawMessage(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	cp := make([]byte, len(raw))
	copy(cp, raw)
	return cp
}

type badRequestError struct {
	message string
}

func (e *badRequestError) Error() string {
	return e.message
}
