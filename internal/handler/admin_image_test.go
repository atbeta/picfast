package handler_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5"
)

func TestAdminImages(t *testing.T) {
	env, group, _, adminID := setupAdminEnv(t)
	owner := testutil.SeedUser(t, env.DB, group.ID, "img-owner@example.com", "password", string(domain.RoleUser))
	role := domain.RoleAdmin

	token := env.generateToken(t, owner.ID, domain.RoleUser, group.ID)
	_, imageID := uploadAndGetImage(t, env, token)
	uploadAndGetKey(t, env, token)

	t.Run("non_admin forbidden", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/images", nil, owner.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	t.Run("list", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/images", nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		data := respDataMap(t, parseResp(t, rec))
		items, ok := data["items"].([]interface{})
		if !ok || len(items) < 2 {
			t.Fatalf("items = %v, want at least 2", data["items"])
		}
		if total, _ := data["total"].(float64); total < 2 {
			t.Errorf("total = %v, want at least 2", data["total"])
		}
	})

	t.Run("delete", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/admin/images/%d", imageID), nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		_, err := env.DB.GetImageByID(t.Context(), imageID)
		if !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("GetImageByID after delete = %v, want pgx.ErrNoRows", err)
		}
	})

	t.Run("delete invalid id", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, "/api/v1/admin/images/not-a-id", nil, adminID, role, 0))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}
