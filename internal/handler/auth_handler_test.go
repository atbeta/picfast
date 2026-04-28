package handler_test

import (
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/router"
)

func TestRegister(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)

	t.Run("success", func(t *testing.T) {
		body := map[string]string{
			"email":    "new@example.com",
			"password": "password123",
			"name":     "New User",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
		resp := parseResp(t, rec)
		if !resp.Status {
			t.Fatal("expected status true")
		}
		tokens := respDataMap(t, resp)
		if tokens["access_token"] == nil || tokens["refresh_token"] == nil {
			t.Fatal("missing tokens in response")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
			"name":     "Dup",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", rec.Code)
		}
	})

	t.Run("short password", func(t *testing.T) {
		body := map[string]string{
			"email":    "short@example.com",
			"password": "1234567",
			"name":     "Short",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("registration disabled", func(t *testing.T) {
		env2 := newTestEnv(t)
		env2.Config.App.AllowRegistration = false
		env2.Router = router.New(env2.DB, env2.Pool, env2.Config, env2.JWT, nil)

		body := map[string]string{
			"email":    "nope@example.com",
			"password": "password123",
			"name":     "Nope",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env2.Router, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})
}

func TestLogin(t *testing.T) {
	env := newTestEnv(t)
	_, _, user := env.seedSetup(t)

	t.Run("success", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		tokens := respDataMap(t, parseResp(t, rec))
		if tokens["access_token"] == nil {
			t.Fatal("missing access_token")
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})

	t.Run("frozen user", func(t *testing.T) {
		_, err := env.Pool.Exec(t.Context(), "UPDATE users SET status = 0 WHERE id = $1", user.ID)
		if err != nil {
			t.Fatalf("freeze user: %v", err)
		}

		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})
}

func TestRefresh(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)

	t.Run("success", func(t *testing.T) {
		regBody := map[string]string{
			"email":    "refresh@example.com",
			"password": "password123",
			"name":     "Refresh",
		}
		regReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", regBody)
		regRec := doReq(env.Router, regReq)

		tokens := respDataMap(t, parseResp(t, regRec))
		refreshToken := tokens["refresh_token"].(string)

		refreshBody := map[string]string{"refresh_token": refreshToken}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/refresh", refreshBody)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		body := map[string]string{"refresh_token": "invalid-token"}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/refresh", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestLogout(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)

	// Login first to get a valid token
	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	loginReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", loginBody)
	loginRec := doReq(env.Router, loginReq)
	tokens := respDataMap(t, parseResp(t, loginRec))
	accessToken := tokens["access_token"].(string)

	req := newJSONReq(t, http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
}
