package handler_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/testutil"
)

func TestSetupStatus(t *testing.T) {
	env := newTestEnv(t)

	req := newJSONReq(t, http.MethodGet, "/api/v1/setup/status", nil)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	if data["required"] != true {
		t.Fatalf("required = %v, want true", data["required"])
	}

	_, _, _ = env.seedSetup(t)
	req = newJSONReq(t, http.MethodGet, "/api/v1/setup/status", nil)
	rec = doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	data = respDataMap(t, parseResp(t, rec))
	if data["required"] != false {
		t.Fatalf("required = %v, want false", data["required"])
	}
}

func TestSetupCreateAdmin(t *testing.T) {
	env := newTestEnv(t)
	group := testutil.SeedDefaultGroup(t, env.DB)
	_ = testutil.SeedStrategy(t, env.DB, group.ID)

	body := map[string]string{
		"email":    "owner@example.com",
		"password": "password123",
		"name":     "Owner",
	}
	req := newJSONReq(t, http.MethodPost, "/api/v1/setup/admin", body)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	if data["access_token"] == nil || data["refresh_token"] == nil {
		t.Fatal("missing tokens in setup response")
	}

	user, err := env.DB.GetUserByEmail(t.Context(), "owner@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if domain.UserRole(user.Role) != domain.RoleAdmin {
		t.Fatalf("role = %q, want admin", user.Role)
	}
	if !user.EmailVerified {
		t.Fatal("setup admin should be email verified")
	}

	req = newJSONReq(t, http.MethodPost, "/api/v1/setup/admin", body)
	rec = doReq(env.Router, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", rec.Code, rec.Body.String())
	}
}

func TestSetupCreateAdminRejectsInvalidInput(t *testing.T) {
	env := newTestEnv(t)

	cases := []struct {
		name string
		body map[string]string
	}{
		{
			name: "invalid email",
			body: map[string]string{"email": "bad", "password": "password123", "name": "Owner"},
		},
		{
			name: "short password",
			body: map[string]string{"email": "owner@example.com", "password": "short", "name": "Owner"},
		},
		{
			name: "missing name",
			body: map[string]string{"email": "owner@example.com", "password": "password123"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := newJSONReq(t, http.MethodPost, "/api/v1/setup/admin", tc.body)
			rec := doReq(env.Router, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestRegisterRequiresSetupOnEmptyInstance(t *testing.T) {
	env := newTestEnv(t)
	env.Config.App.AllowRegistration = true
	env.rebuildRouter()

	body := map[string]string{
		"email":    "new@example.com",
		"password": "password123",
		"name":     "New User",
	}
	req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", rec.Code, rec.Body.String())
	}
}

func TestSetupIncompleteBlocksGuestUpload(t *testing.T) {
	env := newTestEnv(t)
	env.Config.App.AllowGuestUpload = true
	env.rebuildRouter()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "setup.png")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write(pngBytes()); err != nil {
		t.Fatalf("write png: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", bytes.NewReader(body.Bytes()))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", rec.Code, rec.Body.String())
	}
}
