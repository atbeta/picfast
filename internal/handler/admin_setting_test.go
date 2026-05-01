package handler_test

import (
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
