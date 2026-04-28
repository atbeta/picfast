package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/testutil"
)

func makeAdmin(t *testing.T, env *testEnv, userID int64) {
	t.Helper()
	_, err := env.Pool.Exec(t.Context(), "UPDATE users SET role = 'admin' WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("make admin: %v", err)
	}
}

func TestAdminListUsers(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/users", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	if data["items"] == nil {
		t.Fatal("missing items")
	}
}

func TestAdminCreateGroup(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"name":       "New Group",
		"is_default": true,
		"is_guest":   false,
		"configs":    map[string]interface{}{},
	}
	req := env.authReq(t, http.MethodPost, "/api/v1/admin/groups", body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}

	groups, _ := env.DB.ListGroups(t.Context())
	defaultCount := 0
	for _, g := range groups {
		if g.IsDefault {
			defaultCount++
		}
	}
	if defaultCount != 1 {
		t.Fatalf("default groups = %d, want 1", defaultCount)
	}
}

func TestAdminUpdateStrategy(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	strategy := testutil.SeedStrategy(t, env.DB, group.ID)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"name":          "Updated Strategy",
		"strategy_type": "local",
		"configs":       map[string]interface{}{"root": "/new/path", "url": "/i"},
	}
	req := env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/strategies/%d", strategy.ID), body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminNonAdminAccess(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/users", nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestAdminDeleteUser(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	target := testutil.SeedUser(t, env.DB, group.ID, "delete@me.com", "password123", string(domain.RoleUser))

	token := env.generateToken(t, admin.ID, domain.RoleAdmin, group.ID)
	req := newJSONReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminSetGroupStrategies(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	strategy := testutil.SeedStrategy(t, env.DB, group.ID)
	makeAdmin(t, env, admin.ID)

	body := map[string]interface{}{
		"strategy_ids": []int64{strategy.ID},
	}
	req := env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/groups/%d/strategies", group.ID), body, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	strategies, _ := env.DB.GetGroupStrategies(t.Context(), group.ID)
	if len(strategies) != 1 || strategies[0].ID != strategy.ID {
		t.Fatalf("group strategies = %v, want [%d]", strategies, strategy.ID)
	}
}
