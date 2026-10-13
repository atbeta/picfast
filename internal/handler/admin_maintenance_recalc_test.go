package handler_test

import (
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/pglock"
)

// While another session holds the recalc advisory lock — another instance or
// an in-flight run — the endpoint must keep answering 409.
func TestRecalcPHashConflictWhenLockHeld(t *testing.T) {
	env := newTestEnv(t)
	group, _, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	conn, err := env.Pool.Acquire(t.Context())
	if err != nil {
		t.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()

	var locked bool
	if err := conn.QueryRow(t.Context(), "SELECT pg_try_advisory_lock($1)", pglock.KeyRecalcPHash).Scan(&locked); err != nil {
		t.Fatalf("pg_try_advisory_lock: %v", err)
	}
	if !locked {
		t.Fatal("could not acquire recalc advisory lock")
	}
	defer func() {
		if _, err := conn.Exec(t.Context(), "SELECT pg_advisory_unlock($1)", pglock.KeyRecalcPHash); err != nil {
			t.Errorf("unlock: %v", err)
		}
	}()

	req := env.authReq(t, http.MethodPost, "/api/v1/admin/maintenance/recalc-phash", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409; body: %s", rec.Code, rec.Body.String())
	}
}
