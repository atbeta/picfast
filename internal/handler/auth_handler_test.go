package handler_test

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/testutil"
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
		result := respDataMap(t, resp)
		tokens := nestedMap(t, result["tokens"])
		if tokens["access_token"] == nil || tokens["refresh_token"] == nil {
			t.Fatal("missing tokens in response")
		}
	})

	t.Run("requires verification when enabled", func(t *testing.T) {
		env.Config.App.RequireEmailVerification = true
		env.MailSender.ready = true
		env.MailSender.messages = nil
		env.rebuildRouter()
		defer func() {
			env.Config.App.RequireEmailVerification = false
			env.MailSender.ready = false
			env.MailSender.messages = nil
			env.rebuildRouter()
		}()

		body := map[string]string{
			"email":    "verify@example.com",
			"password": "password123",
			"name":     "Verify User",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}

		data := respDataMap(t, parseResp(t, rec))
		if data["requires_email_verification"] != true {
			t.Fatalf("requires_email_verification = %v, want true", data["requires_email_verification"])
		}
		if data["verification_email_sent"] != true {
			t.Fatalf("verification_email_sent = %v, want true", data["verification_email_sent"])
		}
		if len(env.MailSender.messages) != 1 {
			t.Fatalf("sent messages = %d, want 1", len(env.MailSender.messages))
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

	t.Run("invalid email", func(t *testing.T) {
		body := map[string]string{
			"email":    "not-an-email",
			"password": "password123",
			"name":     "Invalid Email",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("too long password", func(t *testing.T) {
		body := map[string]string{
			"email":    "long-password@example.com",
			"password": strings.Repeat("a", 73),
			"name":     "Long Password",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("registration disabled", func(t *testing.T) {
		env.Config.App.AllowRegistration = false
		env.rebuildRouter()
		defer func() {
			env.Config.App.AllowRegistration = true
			env.rebuildRouter()
		}()

		body := map[string]string{
			"email":    "nope@example.com",
			"password": "password123",
			"name":     "Nope",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", body)
		rec := doReq(env.Router, req)

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

	t.Run("unverified user blocked when verification is required", func(t *testing.T) {
		env.Config.App.RequireEmailVerification = true
		env.MailSender.ready = true
		env.rebuildRouter()
		defer func() {
			env.Config.App.RequireEmailVerification = false
			env.MailSender.ready = false
			env.rebuildRouter()
		}()

		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body: %s", rec.Code, rec.Body.String())
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

func TestAdminLoginAudit(t *testing.T) {
	env := newTestEnv(t)
	_, _, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	t.Run("records admin login success", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "password123",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		req.Header.Set("CF-Connecting-IP", "203.0.113.10")
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}

		var count int
		if err := env.Pool.QueryRow(t.Context(), `
			SELECT COUNT(*)
			FROM audit_logs
			WHERE action = 'admin.auth.login.success'
			  AND resource_type = 'auth'
			  AND resource_id = $1
			  AND actor_user_id = $2
			  AND ip = '203.0.113.10'
		`, strconv.FormatInt(admin.ID, 10), admin.ID).Scan(&count); err != nil {
			t.Fatalf("count login success audit logs: %v", err)
		}
		if count != 1 {
			t.Fatalf("success audit log count = %d, want 1", count)
		}
	})

	t.Run("records admin login failure", func(t *testing.T) {
		body := map[string]string{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}

		var count int
		if err := env.Pool.QueryRow(t.Context(), `
			SELECT COUNT(*)
			FROM audit_logs
			WHERE action = 'admin.auth.login.failed'
			  AND resource_type = 'auth'
			  AND resource_id = $1
			  AND actor_user_id = $2
			  AND details->>'reason' = 'invalid_credentials'
		`, strconv.FormatInt(admin.ID, 10), admin.ID).Scan(&count); err != nil {
			t.Fatalf("count login failure audit logs: %v", err)
		}
		if count != 1 {
			t.Fatalf("failure audit log count = %d, want 1", count)
		}
	})

	t.Run("skips ordinary user login", func(t *testing.T) {
		group := testutil.SeedDefaultGroup(t, env.DB)
		user := testutil.SeedUser(t, env.DB, group.ID, "ordinary@example.com", "password123", string(domain.RoleUser))
		body := map[string]string{
			"email":    user.Email,
			"password": "password123",
		}
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}

		var count int
		if err := env.Pool.QueryRow(t.Context(), `
			SELECT COUNT(*)
			FROM audit_logs
			WHERE action LIKE 'admin.auth.login.%'
			  AND resource_id = $1
		`, strconv.FormatInt(user.ID, 10)).Scan(&count); err != nil {
			t.Fatalf("count ordinary login audit logs: %v", err)
		}
		if count != 0 {
			t.Fatalf("ordinary user audit log count = %d, want 0", count)
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

		result := respDataMap(t, parseResp(t, regRec))
		tokens := nestedMap(t, result["tokens"])
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

func TestVerifyEmailFlow(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)
	env.Config.App.RequireEmailVerification = true
	env.MailSender.ready = true
	env.rebuildRouter()

	regBody := map[string]string{
		"email":    "verify-flow@example.com",
		"password": "password123",
		"name":     "Verify Flow",
	}
	regReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/register", regBody)
	regRec := doReq(env.Router, regReq)
	if regRec.Code != http.StatusCreated {
		t.Fatalf("register status = %d, want 201; body: %s", regRec.Code, regRec.Body.String())
	}
	if len(env.MailSender.messages) != 1 {
		t.Fatalf("sent messages = %d, want 1", len(env.MailSender.messages))
	}

	re := regexp.MustCompile(`token=([a-f0-9]+)`)
	match := re.FindStringSubmatch(env.MailSender.messages[0].Text)
	if len(match) != 2 {
		t.Fatalf("verification token not found in mail body: %s", env.MailSender.messages[0].Text)
	}

	verifyReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/verify-email", map[string]string{"token": match[1]})
	verifyRec := doReq(env.Router, verifyReq)
	if verifyRec.Code != http.StatusOK {
		t.Fatalf("verify status = %d, want 200; body: %s", verifyRec.Code, verifyRec.Body.String())
	}

	user, err := env.DB.GetUserByEmail(t.Context(), "verify-flow@example.com")
	if err != nil {
		t.Fatalf("load verified user: %v", err)
	}
	if !user.EmailVerified {
		t.Fatal("expected email_verified to be true after verification")
	}
}
