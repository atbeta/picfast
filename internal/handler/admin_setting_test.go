package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func TestAdminSettingsEmailVerificationState(t *testing.T) {
	env := newTestEnv(t)
	env.Config.App.RequireEmailVerification = true
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/settings", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	data := respDataMap(t, parseResp(t, rec))
	if data["require_email_verification"] != false {
		t.Fatalf("require_email_verification = %v, want false when mail is not ready", data["require_email_verification"])
	}
	if data["email_verification_ready"] != false {
		t.Fatalf("email_verification_ready = %v, want false", data["email_verification_ready"])
	}
}

func TestAdminSettingsUpdateEmailVerificationFlag(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]bool{
		"require_email_verification": true,
	}
	req := env.authReq(t, http.MethodPut, "/api/v1/admin/settings", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	if env.Config.App.RequireEmailVerification {
		t.Fatal("expected config require_email_verification to stay disabled without mail support")
	}
}

func TestAdminSettingsUpdateEmailVerificationFlagWhenMailReady(t *testing.T) {
	env := newTestEnv(t)
	env.MailSender.ready = true
	env.rebuildRouter()
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]bool{
		"require_email_verification": true,
	}
	req := env.authReq(t, http.MethodPut, "/api/v1/admin/settings", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if !env.Config.App.RequireEmailVerification {
		t.Fatal("expected config require_email_verification to be updated")
	}

	data := respDataMap(t, parseResp(t, rec))
	if data["require_email_verification"] != true {
		t.Fatalf("require_email_verification = %v, want true", data["require_email_verification"])
	}
}

func TestAdminSettingsPersistSiteSettings(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"allow_guest_upload":    true,
		"allow_registration":    true,
		"user_initial_capacity": int64(1024),
		"moderation_mode":       "manual",
	}
	req := env.authReq(t, http.MethodPut, "/api/v1/admin/settings", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	settings, err := env.DB.GetSiteSettings(req.Context())
	if err != nil {
		t.Fatalf("get persisted site settings: %v", err)
	}
	if !settings.AllowGuestUpload || !settings.AllowRegistration {
		t.Fatalf("persisted toggles = guest:%v registration:%v, want true/true", settings.AllowGuestUpload, settings.AllowRegistration)
	}
	if settings.UserInitialCapacity != 1024 {
		t.Fatalf("persisted user_initial_capacity = %d, want 1024", settings.UserInitialCapacity)
	}
	if settings.ModerationMode != "manual" {
		t.Fatalf("persisted moderation_mode = %q, want manual", settings.ModerationMode)
	}
}

func TestAdminSettingsPersistSiteMetadata(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"site_description":   "A private image hosting service for the team.",
		"favicon_url":        "https://img.example.com/favicon.ico",
		"icp_number":         "京ICP备12345678号-1",
		"icp_link":           "https://beian.miit.gov.cn/",
		"psb_number":         "京公网安备11000002000001号",
		"psb_link":           "https://www.beian.gov.cn/",
		"analytics_provider": "umami",
		"analytics_config": map[string]interface{}{
			"script_url": "https://analytics.example.com/script.js",
			"website_id": "site-123",
		},
	}
	req := env.authReq(t, http.MethodPut, "/api/v1/admin/settings", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	data := respDataMap(t, parseResp(t, rec))
	if data["site_description"] != "A private image hosting service for the team." {
		t.Fatalf("site_description = %v", data["site_description"])
	}
	if data["favicon_url"] != "https://img.example.com/favicon.ico" {
		t.Fatalf("favicon_url = %v", data["favicon_url"])
	}
	if data["icp_number"] != "京ICP备12345678号-1" {
		t.Fatalf("icp_number = %v", data["icp_number"])
	}
	if data["analytics_provider"] != "umami" {
		t.Fatalf("analytics_provider = %v", data["analytics_provider"])
	}

	settings, err := env.DB.GetSiteSettings(req.Context())
	if err != nil {
		t.Fatalf("get persisted site settings: %v", err)
	}
	if settings.SiteDescription != "A private image hosting service for the team." {
		t.Fatalf("persisted site_description = %q", settings.SiteDescription)
	}
	if settings.FaviconUrl != "https://img.example.com/favicon.ico" {
		t.Fatalf("persisted favicon_url = %q", settings.FaviconUrl)
	}
	var analytics map[string]string
	if err := json.Unmarshal(settings.AnalyticsConfig, &analytics); err != nil {
		t.Fatalf("unmarshal analytics config: %v", err)
	}
	if analytics["website_id"] != "site-123" {
		t.Fatalf("analytics website_id = %q", analytics["website_id"])
	}
}

func TestAdminSettingsRejectsInvalidAnalyticsProvider(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"analytics_provider": "unsupported",
	}
	req := env.authReq(t, http.MethodPut, "/api/v1/admin/settings", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
}

func TestPublicConfigIncludesSiteMetadata(t *testing.T) {
	env := newTestEnv(t)
	group, _, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"site_description":   "Public description",
		"favicon_url":        "https://img.example.com/site.ico",
		"icp_number":         "沪ICP备12345678号",
		"analytics_provider": "plausible",
		"analytics_config": map[string]interface{}{
			"domain":     "img.example.com",
			"script_url": "https://plausible.io/js/script.js",
		},
	}
	req := env.authReq(t, http.MethodPut, "/api/v1/admin/settings", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("settings status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	cfgReq := env.authReq(t, http.MethodGet, "/api/v1/config", nil, admin.ID, domain.RoleAdmin, group.ID)
	cfgRec := doReq(env.Router, cfgReq)
	if cfgRec.Code != http.StatusOK {
		t.Fatalf("config status = %d, want 200; body: %s", cfgRec.Code, cfgRec.Body.String())
	}
	data := respDataMap(t, parseResp(t, cfgRec))
	if data["site_description"] != "Public description" {
		t.Fatalf("site_description = %v", data["site_description"])
	}
	if data["favicon_url"] != "https://img.example.com/site.ico" {
		t.Fatalf("favicon_url = %v", data["favicon_url"])
	}
	if data["icp_number"] != "沪ICP备12345678号" {
		t.Fatalf("icp_number = %v", data["icp_number"])
	}
	if data["analytics_provider"] != "plausible" {
		t.Fatalf("analytics_provider = %v", data["analytics_provider"])
	}
	if data["github_url"] != "https://github.com/atbeta/picfast" {
		t.Fatalf("github_url = %v", data["github_url"])
	}
}
