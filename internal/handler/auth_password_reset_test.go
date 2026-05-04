package handler_test

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestForgotAndResetPasswordFlow(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)
	env.Config.App.WebBaseURL = "http://localhost:5173"
	env.MailSender.ready = true
	env.rebuildRouter()

	forgotReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
		"email":    "test@example.com",
		"language": "zh-CN",
	})
	forgotRec := doReq(env.Router, forgotReq)
	if forgotRec.Code != http.StatusOK {
		t.Fatalf("forgot status = %d, want 200; body: %s", forgotRec.Code, forgotRec.Body.String())
	}
	if forgotRec.Header().Get("X-Cooldown-Seconds") == "" {
		t.Fatal("forgot response should include X-Cooldown-Seconds header")
	}
	if len(env.MailSender.messages) != 1 {
		t.Fatalf("sent messages = %d, want 1", len(env.MailSender.messages))
	}
	if !strings.Contains(env.MailSender.messages[0].Subject, "重置密码") {
		t.Fatalf("subject = %q, want zh reset password subject", env.MailSender.messages[0].Subject)
	}
	if !strings.Contains(env.MailSender.messages[0].Text, "http://localhost:5173/reset-password?token=") {
		t.Fatalf("reset mail link should use app.web_base_url, got: %s", env.MailSender.messages[0].Text)
	}

	re := regexp.MustCompile(`token=([a-f0-9]+)`)
	match := re.FindStringSubmatch(env.MailSender.messages[0].Text)
	if len(match) != 2 {
		t.Fatalf("reset token not found in mail body: %s", env.MailSender.messages[0].Text)
	}
	token := match[1]

	resetReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        token,
		"new_password": "newpassword123",
	})
	resetRec := doReq(env.Router, resetReq)
	if resetRec.Code != http.StatusOK {
		t.Fatalf("reset status = %d, want 200; body: %s", resetRec.Code, resetRec.Body.String())
	}

	oldLoginReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	})
	oldLoginRec := doReq(env.Router, oldLoginReq)
	if oldLoginRec.Code != http.StatusUnauthorized {
		t.Fatalf("old password login status = %d, want 401", oldLoginRec.Code)
	}

	newLoginReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    "test@example.com",
		"password": "newpassword123",
	})
	newLoginRec := doReq(env.Router, newLoginReq)
	if newLoginRec.Code != http.StatusOK {
		t.Fatalf("new password login status = %d, want 200; body: %s", newLoginRec.Code, newLoginRec.Body.String())
	}

	reuseReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/reset-password", map[string]string{
		"token":        token,
		"new_password": "anotherpassword123",
	})
	reuseRec := doReq(env.Router, reuseReq)
	if reuseRec.Code != http.StatusBadRequest {
		t.Fatalf("reuse token status = %d, want 400; body: %s", reuseRec.Code, reuseRec.Body.String())
	}
}

func TestForgotPasswordSafetyAndAvailability(t *testing.T) {
	env := newTestEnv(t)
	_, _, _ = env.seedSetup(t)

	t.Run("returns unavailable when mail sender is not ready", func(t *testing.T) {
		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
			"email": "test@example.com",
		})
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})

	t.Run("does not expose account existence", func(t *testing.T) {
		env.MailSender.ready = true
		env.MailSender.messages = nil
		env.rebuildRouter()

		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
			"email": "not-found@example.com",
		})
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		if len(env.MailSender.messages) != 0 {
			t.Fatalf("messages = %d, want 0 for non-existing account", len(env.MailSender.messages))
		}
	})

	t.Run("returns generic success when mail send fails", func(t *testing.T) {
		env.MailSender.ready = true
		env.MailSender.failSend = true
		env.MailSender.messages = nil
		env.rebuildRouter()
		t.Cleanup(func() { env.MailSender.failSend = false })

		req := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
			"email": "test@example.com",
		})
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		if len(env.MailSender.messages) != 0 {
			t.Fatalf("messages = %d, want 0 when sender fails", len(env.MailSender.messages))
		}
	})

	t.Run("applies email cooldown to avoid repeated sends", func(t *testing.T) {
		env.MailSender.ready = true
		env.MailSender.failSend = false
		env.MailSender.messages = nil
		env.rebuildRouter()
		user, err := env.DB.GetUserByEmail(t.Context(), "test@example.com")
		if err != nil {
			t.Fatalf("get user: %v", err)
		}
		if err := env.DB.DeleteUnusedPasswordResetTokensByUser(t.Context(), user.ID); err != nil {
			t.Fatalf("clear reset tokens: %v", err)
		}

		firstReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
			"email": "test@example.com",
		})
		firstRec := doReq(env.Router, firstReq)
		if firstRec.Code != http.StatusOK {
			t.Fatalf("first status = %d, want 200; body: %s", firstRec.Code, firstRec.Body.String())
		}

		secondReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
			"email": "test@example.com",
		})
		secondRec := doReq(env.Router, secondReq)
		if secondRec.Code != http.StatusOK {
			t.Fatalf("second status = %d, want 200; body: %s", secondRec.Code, secondRec.Body.String())
		}
		cooldown, err := strconv.Atoi(secondRec.Header().Get("X-Cooldown-Seconds"))
		if err != nil || cooldown < 1 || cooldown > 120 {
			t.Fatalf("cooldown header = %q, want 1..120", secondRec.Header().Get("X-Cooldown-Seconds"))
		}
		if len(env.MailSender.messages) != 1 {
			t.Fatalf("messages = %d, want 1 due to email cooldown", len(env.MailSender.messages))
		}
	})

	t.Run("is rate limited to protect mail service", func(t *testing.T) {
		env.MailSender.ready = true
		env.MailSender.messages = nil
		env.rebuildRouter()

		for i := 0; i < 3; i++ {
			req := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
				"email": "test@example.com",
			})
			req.RemoteAddr = "198.51.100.23:12345"
			rec := doReq(env.Router, req)
			if rec.Code != http.StatusOK {
				t.Fatalf("request %d status = %d, want 200; body: %s", i+1, rec.Code, rec.Body.String())
			}
		}

		blockedReq := newJSONReq(t, http.MethodPost, "/api/v1/auth/forgot-password", map[string]string{
			"email": "test@example.com",
		})
		blockedReq.RemoteAddr = "198.51.100.23:12345"
		blockedRec := doReq(env.Router, blockedReq)
		if blockedRec.Code != http.StatusTooManyRequests {
			t.Fatalf("blocked status = %d, want 429; body: %s", blockedRec.Code, blockedRec.Body.String())
		}
	})
}
