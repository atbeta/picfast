package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
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
	AllowOauthRegistration   *bool            `json:"allow_oauth_registration"`
	AllowUserImageProcessing *bool            `json:"allow_user_image_processing"`
	RequireEmailVerification *bool            `json:"require_email_verification"`
	UserInitialCapacity      *int64           `json:"user_initial_capacity"`
	DefaultImageTTL          *string          `json:"default_image_ttl"`
	GuestImageTTL            *string          `json:"guest_image_ttl"`
	ModerationMode           *string          `json:"moderation_mode"`
	FooterText1              *string          `json:"footer_text_1"`
	FooterLink1              *string          `json:"footer_link_1"`
	FooterText2              *string          `json:"footer_text_2"`
	FooterLink2              *string          `json:"footer_link_2"`
	AnalyticsProvider        *string          `json:"analytics_provider"`
	AnalyticsConfig          *json.RawMessage `json:"analytics_config"`
	ThemeConfig              *json.RawMessage `json:"theme_config"`
	DefaultCopyFormat        *string          `json:"default_copy_format"`
	CopyTemplate             *string          `json:"copy_template"`
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
	if req.AllowOauthRegistration != nil {
		h.setter.SetAllowOauthRegistration(*req.AllowOauthRegistration)
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
	if req.GuestImageTTL != nil {
		if *req.GuestImageTTL == "0" || *req.GuestImageTTL == "" {
			h.setter.SetGuestImageTTL(0)
		} else {
			d, err := time.ParseDuration(*req.GuestImageTTL)
			if err != nil {
				Fail(w, http.StatusBadRequest, "invalid guest_image_ttl format")
				return
			}
			h.setter.SetGuestImageTTL(d)
		}
	}
	if req.ModerationMode != nil {
		oldMode := h.config.AppSnapshot().ModerationMode
		newMode := strings.TrimSpace(*req.ModerationMode)
		if isModerationActive(oldMode) && !isModerationActive(newMode) {
			if err := h.queries.ApproveAllPendingImages(r.Context()); err != nil {
				Fail(w, http.StatusInternalServerError, "failed to auto-approve pending images")
				return
			}
		}
		h.setter.SetModerationMode(newMode)
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
	if req.ThemeConfig != nil {
		themeConfig := cloneRawMessage(*req.ThemeConfig)
		if len(themeConfig) == 0 {
			themeConfig = json.RawMessage(`{}`)
		}
		if err := validateThemeConfig(themeConfig); err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}
		h.setter.SetThemeConfig(themeConfig)
	}
	if req.DefaultCopyFormat != nil || req.CopyTemplate != nil {
		format := h.config.AppSnapshot().DefaultCopyFormat
		template := h.config.AppSnapshot().CopyTemplate
		if req.DefaultCopyFormat != nil {
			format = strings.TrimSpace(*req.DefaultCopyFormat)
		}
		if req.CopyTemplate != nil {
			template = *req.CopyTemplate
		}
		if err := validateCopySettings(format, template); err != nil {
			Fail(w, http.StatusBadRequest, err.Error())
			return
		}
		h.setter.SetDefaultCopyFormat(format)
		h.setter.SetCopyTemplate(template)
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
		AllowOauthRegistration:   app.AllowOauthRegistration,
		AllowUserImageProcessing: app.AllowUserImageProcessing,
		RequireEmailVerification: app.RequireEmailVerification,
		UserInitialCapacity:      app.UserInitialCapacity,
		DefaultImageTtl:          app.DefaultImageTTL.String(),
		GuestImageTtl:            app.GuestImageTTL.String(),
		ModerationMode:           app.ModerationMode,
		SiteDescription:          app.SiteDescription,
		FaviconUrl:               app.FaviconURL,
		FooterText1:              app.FooterText1,
		FooterLink1:              app.FooterLink1,
		FooterText2:              app.FooterText2,
		FooterLink2:              app.FooterLink2,
		AnalyticsProvider:        app.AnalyticsProvider,
		AnalyticsConfig:          normalizedRawMessage(app.AnalyticsConfig),
		ThemeConfig:              normalizedRawMessage(app.ThemeConfig),
		DefaultCopyFormat:        app.DefaultCopyFormat,
		CopyTemplate:             app.CopyTemplate,
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
		"allow_oauth_registration":    app.AllowOauthRegistration,
		"allow_user_image_processing": app.AllowUserImageProcessing,
		"require_email_verification":  app.RequireEmailVerification,
		"user_initial_capacity":       app.UserInitialCapacity,
		"default_image_ttl":           app.DefaultImageTTL.String(),
		"guest_image_ttl":             app.GuestImageTTL.String(),
		"moderation_mode":             app.ModerationMode,
		"footer_text_1":               app.FooterText1,
		"footer_link_1":               app.FooterLink1,
		"footer_text_2":               app.FooterText2,
		"footer_link_2":               app.FooterLink2,
		"analytics_provider":          app.AnalyticsProvider,
		"analytics_config":            normalizedRawMessage(app.AnalyticsConfig),
		"theme_config":                normalizedRawMessage(app.ThemeConfig),
		"default_copy_format":           app.DefaultCopyFormat,
		"copy_template":                 app.CopyTemplate,
	}
	if includeMailReady {
		resp["require_email_verification"] = app.RequireEmailVerification && h.mailReady
		resp["email_verification_ready"] = h.mailReady
	}
	return resp
}

func validateThemeConfig(raw json.RawMessage) error {
	var cfg map[string]any
	if err := json.Unmarshal(normalizedRawMessage(raw), &cfg); err != nil {
		return &badRequestError{"invalid theme_config"}
	}
	allowedKeys := map[string]bool{
		"preset": true, "mode": true, "tokens": true, "public": true, "custom_css": true,
	}
	for key := range cfg {
		if !allowedKeys[key] {
			return &badRequestError{"theme_config contains unknown field"}
		}
	}
	if css, ok := cfg["custom_css"].(string); ok && len(css) > 20000 {
		return &badRequestError{"theme_config custom_css is too large"}
	}
	if preset, ok := cfg["preset"].(string); ok && len(preset) > 80 {
		return &badRequestError{"theme_config preset is too long"}
	}
	if mode, ok := cfg["mode"].(string); ok && mode != "" {
		switch mode {
		case "light", "dark", "system":
		default:
			return &badRequestError{"theme_config mode is invalid"}
		}
	}
	if public, ok := cfg["public"].(map[string]any); ok {
		if backgroundImage := stringValue(public["background_image"]); backgroundImage != "" {
			if err := validateOptionalURL(backgroundImage); err != nil {
				return &badRequestError{"theme_config background_image is invalid"}
			}
		}
		if style := stringValue(public["background_style"]); style != "" {
			switch style {
			case "soft", "clean", "image":
			default:
				return &badRequestError{"theme_config background_style is invalid"}
			}
		}
		if logoShape := stringValue(public["logo_shape"]); logoShape != "" {
			switch logoShape {
			case "rounded", "circle", "square":
			default:
				return &badRequestError{"theme_config logo_shape is invalid"}
			}
		}
		for _, field := range []struct {
			key    string
			values []string
		}{
			{"upload_style", []string{"dashed", "solid", "glass"}},
			{"card_style", []string{"flat", "elevated", "glass"}},
			{"button_style", []string{"default", "pill", "sharp"}},
			{"density", []string{"compact", "comfortable", "spacious"}},
			{"motion", []string{"none", "subtle", "playful"}},
		} {
			raw := stringValue(public[field.key])
			if raw == "" {
				continue
			}
			ok := false
			for _, allowed := range field.values {
				if raw == allowed {
					ok = true
					break
				}
			}
			if !ok {
				return &badRequestError{"theme_config " + field.key + " is invalid"}
			}
		}
	}
	if tokens, ok := cfg["tokens"].(map[string]any); ok {
		for _, mode := range []string{"light", "dark"} {
			values, ok := tokens[mode].(map[string]any)
			if !ok {
				continue
			}
			for key, rawValue := range values {
				value, ok := rawValue.(string)
				if !ok || strings.TrimSpace(value) == "" {
					continue
				}
				if key == "radius" {
					if !isSafeCSSLength(value) {
						return &badRequestError{"theme_config radius is invalid"}
					}
					continue
				}
				if !isSafeCSSColor(value) {
					return &badRequestError{"theme_config color is invalid"}
				}
			}
		}
	}
	return nil
}

func validateCopySettings(format, template string) error {
	switch format {
	case "url", "markdown", "html", "bbcode", "thumbnail", "custom":
	default:
		return &badRequestError{"invalid default_copy_format"}
	}
	if len(template) > 4000 {
		return &badRequestError{"copy_template is too large"}
	}
	return nil
}

var (
	themeHexColorRe  = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	themeFuncColorRe = regexp.MustCompile(`^(?:rgb|rgba|hsl|hsla|oklch|oklab|lch|lab|color)\([0-9a-zA-Z\s.,%/+_-]+\)$`)
	themeLengthRe    = regexp.MustCompile(`^(?:0|[0-9]+(?:\.[0-9]+)?(?:px|rem|em|%|vh|vw|vmin|vmax|ch|ex))$`)
	themeNamedColors = map[string]bool{
		"transparent": true, "currentcolor": true, "black": true, "white": true,
		"red": true, "orange": true, "yellow": true, "green": true, "blue": true,
		"purple": true, "pink": true, "gray": true, "grey": true, "slate": true,
		"cyan": true, "teal": true, "lime": true, "indigo": true, "violet": true,
	}
)

func isSafeCSSColor(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, ";{}<>") || len(value) > 160 {
		return false
	}
	return themeHexColorRe.MatchString(value) || themeFuncColorRe.MatchString(value) || themeNamedColors[strings.ToLower(value)]
}

func isSafeCSSLength(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.ContainsAny(value, ";{}<>") || len(value) > 48 {
		return false
	}
	return themeLengthRe.MatchString(value)
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

func isModerationActive(mode string) bool {
	return mode != "" && mode != "disabled"
}
