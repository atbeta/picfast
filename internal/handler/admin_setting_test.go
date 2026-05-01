package handler_test

import (
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func TestAdminSettingsEmailVerificationState(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/settings", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	data := respDataMap(t, parseResp(t, rec))
	if data["require_email_verification"] != false {
		t.Fatalf("require_email_verification = %v, want false", data["require_email_verification"])
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

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if !env.Config.App.RequireEmailVerification {
		t.Fatal("expected config require_email_verification to be updated")
	}
}
