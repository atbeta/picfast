package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
)

type strategyListPayload struct {
	Status bool `json:"status"`
	Data   []struct {
		ID           int64           `json:"id"`
		Name         string          `json:"name"`
		StrategyType string          `json:"strategy_type"`
		Configs      json.RawMessage `json:"configs"`
	} `json:"data"`
}

// TestUserStrategyListOmitsConfigs pins the security contract of the
// user-facing strategy endpoint: it must never return storage credentials and
// must only list the strategies bound to the caller's group.
func TestUserStrategyListOmitsConfigs(t *testing.T) {
	env := newTestEnv(t)
	group, ownStrategy, user := env.seedSetup(t)

	// A strategy from another group that carries credentials.
	cfg, err := json.Marshal(domain.S3StrategyConfig{
		Endpoint:        "https://s3.example.com",
		Bucket:          "foreign-bucket",
		AccessKeyID:     "AKIA_FOREIGN",
		SecretAccessKey: "foreign-secret",
	})
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	if _, err := env.DB.CreateStrategy(t.Context(), sqlc.CreateStrategyParams{
		Name:         "Foreign S3",
		StrategyType: string(domain.StrategyTypeS3),
		Configs:      cfg,
	}); err != nil {
		t.Fatalf("create foreign strategy: %v", err)
	}

	assertSlimList := func(t *testing.T, rec *httptest.ResponseRecorder) {
		t.Helper()
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		raw := rec.Body.String()
		for _, leaked := range []string{
			"configs", "secret_key", "access_key", "foreign-secret", "AKIA_FOREIGN", "Foreign S3",
		} {
			if strings.Contains(raw, leaked) {
				t.Fatalf("response leaks %q: %s", leaked, raw)
			}
		}

		var payload strategyListPayload
		if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if !payload.Status {
			t.Fatalf("status field = false; body: %s", raw)
		}
		if len(payload.Data) != 1 {
			t.Fatalf("strategies = %d, want 1 (only the caller's group); body: %s", len(payload.Data), raw)
		}
		if payload.Data[0].ID != ownStrategy.ID {
			t.Fatalf("strategy id = %d, want %d", payload.Data[0].ID, ownStrategy.ID)
		}
	}

	t.Run("jwt session", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/strategies", nil, user.ID, domain.RoleUser, group.ID))
		assertSlimList(t, rec)
	})

	t.Run("read-scoped api token", func(t *testing.T) {
		_, plainToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
			"name":       "strategy reader",
			"expires_in": "never",
			"scopes":     []string{"read"},
		})
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/strategies", nil, plainToken))
		assertSlimList(t, rec)
	})

	t.Run("admin still sees configs so the console can edit them", func(t *testing.T) {
		makeAdmin(t, env, user.ID)
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/strategies", nil, user.ID, domain.RoleAdmin, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "configs") {
			t.Fatalf("admin response should keep configs; body: %s", rec.Body.String())
		}
	})
}
