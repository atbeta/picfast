package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5"
)

func makeAdmin(t *testing.T, env *testEnv, userID int64) {
	t.Helper()
	_, err := env.Pool.Exec(t.Context(), "UPDATE users SET role = 'admin' WHERE id = $1", userID)
	if err != nil {
		t.Fatalf("make admin: %v", err)
	}
}

func setupAdminEnv(t *testing.T) (*testEnv, sqlc.Group, sqlc.Strategy, int64) {
	t.Helper()
	env := newTestEnv(t)
	group, strategy, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)
	return env, group, strategy, admin.ID
}

func TestAdminPprofDisabledByDefault(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/debug/pprof/", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestAdminPprofCanBeEnabled(t *testing.T) {
	env := newTestEnv(t)
	_, group, admin := env.seedSetup(t)
	makeAdmin(t, env, admin.ID)
	env.Config.Server.EnablePprof = true
	env.rebuildRouter()

	req := env.authReq(t, http.MethodGet, "/api/v1/admin/debug/pprof/", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminAuthorization(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	t.Run("non_admin cannot list users", func(t *testing.T) {
		req := env.authReq(t, http.MethodGet, "/api/v1/admin/users", nil, user.ID, domain.RoleUser, group.ID)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("non_admin cannot create group", func(t *testing.T) {
		body := map[string]interface{}{
			"name":       "Unauthorized Group",
			"is_default": false,
			"is_guest":   false,
		}
		req := env.authReq(t, http.MethodPost, "/api/v1/admin/groups", body, user.ID, domain.RoleUser, group.ID)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("non_admin cannot create strategy", func(t *testing.T) {
		body := map[string]interface{}{
			"name":          "Unauthorized",
			"strategy_type": "local",
			"configs":       map[string]interface{}{"root": "/tmp", "url": "/i"},
		}
		req := env.authReq(t, http.MethodPost, "/api/v1/admin/strategies", body, user.ID, domain.RoleUser, group.ID)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})
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

func TestAdminGroups(t *testing.T) {
	env, group, _, adminID := setupAdminEnv(t)
	role := domain.RoleAdmin

	t.Run("list", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/groups", nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		items, ok := parseResp(t, rec).Data.([]interface{})
		if !ok || len(items) < 1 {
			t.Fatalf("data = %T, want non-empty array", parseResp(t, rec).Data)
		}
	})

	t.Run("get", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, fmt.Sprintf("/api/v1/admin/groups/%d", group.ID), nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		data := respDataMap(t, parseResp(t, rec))
		if data["id"] != float64(group.ID) {
			t.Errorf("group id = %v, want %d", data["id"], group.ID)
		}
	})

	t.Run("update", func(t *testing.T) {
		createBody := map[string]interface{}{"name": "Updatable Group", "is_default": false, "is_guest": false}
		createRec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/groups", createBody, adminID, role, 0))
		updatableID := int64(respDataMap(t, parseResp(t, createRec))["id"].(float64))

		body := map[string]interface{}{
			"name":       "Updated Group Name",
			"is_default": false,
			"is_guest":   false,
			"configs":    map[string]interface{}{},
		}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/groups/%d", updatableID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		data := respDataMap(t, parseResp(t, rec))
		if data["name"] != "Updated Group Name" {
			t.Errorf("name = %v, want Updated Group Name", data["name"])
		}
	})

	t.Run("create requires name", func(t *testing.T) {
		body := map[string]interface{}{"name": "", "is_default": false, "is_guest": false}
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/groups", body, adminID, role, 0))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("delete succeeds", func(t *testing.T) {
		body := map[string]interface{}{"name": "Deletable Group", "is_default": false, "is_guest": false}
		createRec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/groups", body, adminID, role, 0))
		if createRec.Code != http.StatusCreated {
			t.Fatalf("create status = %d, want 201", createRec.Code)
		}
		newGroupID := int64(respDataMap(t, parseResp(t, createRec))["id"].(float64))

		deleteRec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/groups/%d", newGroupID), nil, adminID, role, 0))
		if deleteRec.Code != http.StatusOK {
			t.Fatalf("delete status = %d, want 200", deleteRec.Code)
		}
	})

	t.Run("cannot delete default group", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/groups/%d", group.ID), nil, adminID, role, 0))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "default or guest") {
			t.Fatalf("body = %q, want default or guest error", rec.Body.String())
		}
	})

	t.Run("cannot delete group with users", func(t *testing.T) {
		body := map[string]interface{}{"name": "Group With User", "is_default": false, "is_guest": false}
		createRec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/groups", body, adminID, role, 0))
		newGroupID := int64(respDataMap(t, parseResp(t, createRec))["id"].(float64))
		testutil.SeedUser(t, env.DB, newGroupID, "user@ingroup.com", "password123", string(domain.RoleUser))

		deleteRec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/groups/%d", newGroupID), nil, adminID, role, 0))
		if deleteRec.Code != http.StatusConflict {
			t.Fatalf("delete status = %d, want 409", deleteRec.Code)
		}
	})

	t.Run("set strategies", func(t *testing.T) {
		strategy := testutil.SeedStrategy(t, env.DB, group.ID)
		body := map[string]interface{}{"strategy_ids": []int64{strategy.ID}}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/groups/%d/strategies", group.ID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		strategies, _ := env.DB.GetGroupStrategies(t.Context(), group.ID)
		if len(strategies) != 1 || strategies[0].ID != strategy.ID {
			t.Fatalf("group strategies = %v, want [%d]", strategies, strategy.ID)
		}
	})
}

func TestAdminUsers(t *testing.T) {
	env, group, _, adminID := setupAdminEnv(t)
	role := domain.RoleAdmin

	t.Run("list", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/users", nil, adminID, role, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if respDataMap(t, parseResp(t, rec))["items"] == nil {
			t.Fatal("missing items")
		}
	})

	t.Run("get", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, fmt.Sprintf("/api/v1/admin/users/%d", adminID), nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if respDataMap(t, parseResp(t, rec))["email"] == nil {
			t.Error("email missing from response")
		}
	})

	t.Run("get not found", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/users/99999", nil, adminID, role, 0))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("update name", func(t *testing.T) {
		target := testutil.SeedUser(t, env.DB, group.ID, "target@example.com", "password", string(domain.RoleUser))
		newName := "Updated Name"
		body := map[string]interface{}{"name": newName}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if respDataMap(t, parseResp(t, rec))["name"] != newName {
			t.Errorf("name = %v, want %q", respDataMap(t, parseResp(t, rec))["name"], newName)
		}
	})

	t.Run("update password", func(t *testing.T) {
		target := testutil.SeedUser(t, env.DB, group.ID, "pw@example.com", "password", string(domain.RoleUser))
		body := map[string]interface{}{"password": "newpassword123"}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("update status", func(t *testing.T) {
		target := testutil.SeedUser(t, env.DB, group.ID, "status@example.com", "password", string(domain.RoleUser))
		body := map[string]interface{}{"status": 0}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if int(respDataMap(t, parseResp(t, rec))["status"].(float64)) != 0 {
			t.Errorf("status = %v, want 0", respDataMap(t, parseResp(t, rec))["status"])
		}
	})

	t.Run("delete user", func(t *testing.T) {
		target := testutil.SeedUser(t, env.DB, group.ID, "delete@user.com", "password", string(domain.RoleUser))
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", target.ID), nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		_, err := env.DB.GetUserByID(t.Context(), target.ID)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("GetUserByID after delete = %v, want pgx.ErrNoRows", err)
		}
	})

	t.Run("cannot delete admin user", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/users/%d", adminID), nil, adminID, role, 0))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})
}

func TestAdminStrategies(t *testing.T) {
	env, _, seeded, adminID := setupAdminEnv(t)
	role := domain.RoleAdmin

	t.Run("list and get", func(t *testing.T) {
		listRec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/strategies", nil, adminID, role, 0))
		if listRec.Code != http.StatusOK {
			t.Fatalf("list status = %d, want 200", listRec.Code)
		}
		list, ok := parseResp(t, listRec).Data.([]interface{})
		if !ok || len(list) < 1 {
			t.Fatalf("list data = %T, want non-empty array", parseResp(t, listRec).Data)
		}

		getRec := doReq(env.Router, env.authReq(t, http.MethodGet, fmt.Sprintf("/api/v1/admin/strategies/%d", seeded.ID), nil, adminID, role, 0))
		if getRec.Code != http.StatusOK {
			t.Fatalf("get status = %d, want 200", getRec.Code)
		}
		data := respDataMap(t, parseResp(t, getRec))
		if data["id"] != float64(seeded.ID) {
			t.Errorf("id = %v, want %d", data["id"], seeded.ID)
		}
	})

	t.Run("update", func(t *testing.T) {
		body := map[string]interface{}{
			"name":          "Updated Strategy",
			"strategy_type": "local",
			"configs":       map[string]interface{}{"root": "/new/path", "url": "/i"},
		}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/strategies/%d", seeded.ID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		updated, err := env.DB.GetStrategyByID(t.Context(), seeded.ID)
		if err != nil {
			t.Fatalf("GetStrategyByID: %v", err)
		}
		if updated.Name != "Updated Strategy" {
			t.Errorf("name = %q, want Updated Strategy", updated.Name)
		}
	})

	t.Run("update requires configs when type changes", func(t *testing.T) {
		body := map[string]interface{}{
			"name":          "Updated Strategy",
			"strategy_type": "webdav",
		}
		rec := doReq(env.Router, env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/admin/strategies/%d", seeded.ID), body, adminID, role, 0))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "configs is required when strategy_type changes") {
			t.Fatalf("body = %q, want configs required message", rec.Body.String())
		}
	})

	t.Run("create", func(t *testing.T) {
		body := map[string]interface{}{
			"name":          "New Local Strategy",
			"strategy_type": "local",
			"configs":       map[string]interface{}{"root": "/tmp/uploads", "url": "/i"},
		}
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/strategies", body, adminID, role, 0))
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201", rec.Code)
		}
	})

	t.Run("create validation", func(t *testing.T) {
		cases := []struct {
			name string
			body map[string]interface{}
		}{
			{
				name: "requires name",
				body: map[string]interface{}{
					"name": "", "strategy_type": "local",
					"configs": map[string]interface{}{"root": "/tmp", "url": "/i"},
				},
			},
			{
				name: "unknown type",
				body: map[string]interface{}{
					"name": "Bad Type", "strategy_type": "unknown_type", "configs": map[string]interface{}{},
				},
			},
			{
				name: "requires configs",
				body: map[string]interface{}{
					"name": "No Configs", "strategy_type": "local", "configs": nil,
				},
			},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/strategies", tc.body, adminID, role, 0))
				if rec.Code != http.StatusBadRequest {
					t.Fatalf("status = %d, want 400", rec.Code)
				}
			})
		}
	})

	t.Run("delete", func(t *testing.T) {
		body := map[string]interface{}{
			"name":          "To Delete",
			"strategy_type": "local",
			"configs":       map[string]interface{}{"root": "/tmp/del", "url": "/i"},
		}
		createRec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/strategies", body, adminID, role, 0))
		newID := int64(respDataMap(t, parseResp(t, createRec))["id"].(float64))

		deleteRec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/strategies/%d", newID), nil, adminID, role, 0))
		if deleteRec.Code != http.StatusOK {
			t.Fatalf("delete status = %d, want 200", deleteRec.Code)
		}
		_, err := env.DB.GetStrategyByID(t.Context(), newID)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("GetStrategyByID after delete = %v, want pgx.ErrNoRows", err)
		}
	})
}
