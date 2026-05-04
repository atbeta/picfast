package handler_test

import (
	"net/http"
	"testing"
)

func TestEmailVerificationEndpoints(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)

	t.Run("verify email rejects invalid token", func(t *testing.T) {
		body := map[string]string{"token": "placeholder"}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/verify-email", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("resend verification unavailable when mail is not ready", func(t *testing.T) {
		body := map[string]string{"email": "test@example.com"}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/resend-verification", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("resend verification returns generic success when send fails", func(t *testing.T) {
		env.Config.App.RequireEmailVerification = true
		env.MailSender.ready = true
		env.MailSender.failSend = true
		env.MailSender.messages = nil
		env.rebuildRouter()
		t.Cleanup(func() { env.MailSender.failSend = false })

		body := map[string]string{"email": "test@example.com"}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/resend-verification", body)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		if len(env.MailSender.messages) != 0 {
			t.Fatalf("messages = %d, want 0 when sender fails", len(env.MailSender.messages))
		}
	})
}
