package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

// TestUploadEndpointsRequireWriteScope pins that API tokens missing the "write"
// scope cannot upload through any of the three upload entrypoints, while JWT
// sessions and guest uploads keep working (RequireScope passes when the request
// carries no scopes at all).
func TestUploadEndpointsRequireWriteScope(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	_, readToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "read only",
		"expires_in": "never",
		"scopes":     []string{"read"},
	})
	_, writeToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "read write",
		"expires_in": "never",
		"scopes":     []string{"read", "write"},
	})

	paths := []string{"/api/v1/upload", "/api/v1/sharex/upload", "/api/v1/flat/upload"}
	for _, path := range paths {
		t.Run("read token rejected "+path, func(t *testing.T) {
			rec := doReq(env.Router, uploadReq(t, path, "read.png", pngBytes(), readToken))
			if rec.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want 403; body: %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), "insufficient scope") {
				t.Fatalf("body = %s, want insufficient scope message", rec.Body.String())
			}
		})

		t.Run("write token accepted "+path, func(t *testing.T) {
			rec := doReq(env.Router, uploadReq(t, path, "write.png", pngBytes(), writeToken))
			if rec.Code >= 300 {
				t.Fatalf("status = %d, want 2xx; body: %s", rec.Code, rec.Body.String())
			}
		})
	}

	t.Run("jwt session unaffected", func(t *testing.T) {
		token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
		rec := doReq(env.Router, uploadReq(t, "/api/v1/upload", "jwt.png", pngBytes(), token))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("guest upload unaffected", func(t *testing.T) {
		rec := doReq(env.Router, uploadReq(t, "/api/v1/upload", "guest.png", pngBytes(), ""))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
	})
}
