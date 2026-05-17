package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func createAPITokenViaAPI(t *testing.T, env *testEnv, userID, groupID int64, body map[string]interface{}) (tokenID int64, plainToken string) {
	t.Helper()
	rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/api-tokens", body, userID, domain.RoleUser, groupID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	plainToken, _ = data["token"].(string)
	if plainToken == "" {
		t.Fatal("plain token missing from create response")
	}
	tokenID = int64(data["id"].(float64))
	return tokenID, plainToken
}

func TestAPITokenCRUD(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	tokenID, plainToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "integration token",
		"expires_in": "never",
		"scopes":     []string{"read", "write"},
	})

	t.Run("list", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/api-tokens", nil, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		resp := parseResp(t, rec)
		items, ok := resp.Data.([]interface{})
		if !ok || len(items) == 0 {
			t.Fatal("expected token list")
		}
	})

	t.Run("delete", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/api-tokens/%d", tokenID), nil, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("revoked token unauthorized", func(t *testing.T) {
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, plainToken))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAPITokenDualAuth(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	_, plainToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "dual-auth token",
		"expires_in": "never",
		"scopes":     []string{"read", "write"},
	})

	t.Run("x_api_token lists images", func(t *testing.T) {
		uploadReq := uploadReq(t, "/api/v1/images", "api-tok.png", pngBytes(), env.generateToken(t, user.ID, domain.RoleUser, group.ID))
		if rec := doReq(env.Router, uploadReq); rec.Code != http.StatusCreated {
			t.Fatalf("seed upload status = %d, want 201", rec.Code)
		}

		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, plainToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("bearer api token lists images", func(t *testing.T) {
		rec := doReq(env.Router, env.bearerAPITokenReq(t, http.MethodGet, "/api/v1/images", nil, plainToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("invalid token unauthorized", func(t *testing.T) {
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, "img_invalid"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAPITokenRequireScope(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	_, readOnlyToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "read only",
		"expires_in": "never",
		"scopes":     []string{"read"},
	})

	t.Run("read scope can list", func(t *testing.T) {
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, readOnlyToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("read scope cannot upload", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/images", "scoped.png", pngBytes(), "")
		req.Header.Set("X-API-Token", readOnlyToken)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("jwt upload without scope restriction", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/images", "jwt.png", pngBytes(), env.generateToken(t, user.ID, domain.RoleUser, group.ID))
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestAPITokenCreateValidation(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/api-tokens", map[string]interface{}{
		"name": "",
	}, user.ID, domain.RoleUser, group.ID))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
