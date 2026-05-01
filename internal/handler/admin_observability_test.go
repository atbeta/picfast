package handler_test

import (
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
)

func TestAdminObservabilitySummary(t *testing.T) {
	env := newTestEnv(t)
	group, strategy, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO images (
			user_id, group_id, strategy_id, key, path, name, origin_name,
			size_bytes, mimetype, extension, md5, sha1, width, height,
			permission, uploaded_ip, moderation_status
		)
		VALUES ($1, $2, $3, 'obs-key', '2026/05/01', 'obs.png', 'obs.png',
			1234, 'image/png', 'png', '0123456789abcdef0123456789abcdef',
			'0123456789abcdef0123456789abcdef01234567', 100, 80,
			0, '127.0.0.1', 'pending')
	`, admin.ID, group.ID, strategy.ID)
	if err != nil {
		t.Fatalf("seed image: %v", err)
	}

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/observability/summary", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	data := respDataMap(t, parseResp(t, rec))
	health := nestedMap(t, data["health"])
	if got := nestedMap(t, health["database"])["healthy"]; got != true {
		t.Fatalf("database healthy = %v, want true", got)
	}
	if got := nestedMap(t, health["mail"])["healthy"]; got != true {
		t.Fatalf("mail healthy = %v, want true when email verification is disabled", got)
	}

	usage := nestedMap(t, data["usage"])
	if got := int64(usage["storage_bytes"].(float64)); got != 1234 {
		t.Fatalf("storage_bytes = %d, want 1234", got)
	}
	if got := int64(usage["uploads_24h"].(float64)); got != 1 {
		t.Fatalf("uploads_24h = %d, want 1", got)
	}
	if got := int64(usage["pending_moderation"].(float64)); got != 1 {
		t.Fatalf("pending_moderation = %d, want 1", got)
	}

	runtimeInfo := nestedMap(t, data["runtime"])
	if runtimeInfo["go_version"] == "" || runtimeInfo["goos"] == "" {
		t.Fatalf("missing runtime info: %#v", runtimeInfo)
	}

	config := nestedMap(t, data["config"])
	if config["metrics_enabled"] != true {
		t.Fatalf("metrics_enabled = %v, want true", config["metrics_enabled"])
	}

	strategies, ok := data["storage_strategies"].([]interface{})
	if !ok || len(strategies) != 1 {
		t.Fatalf("storage_strategies = %#v, want one item", data["storage_strategies"])
	}
}
