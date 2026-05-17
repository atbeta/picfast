package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
)

func seedPendingModerationImage(t *testing.T, env *testEnv, ownerID, groupID, strategyID int64, key string) sqlc.Image {
	t.Helper()
	img := createTestImageWithModeration(t, env, ownerID, groupID, strategyID, key, "png", "pending", int16(domain.PermissionPublic))
	_, err := env.DB.CreateImageModeration(t.Context(), sqlc.CreateImageModerationParams{
		ImageID:  img.ID,
		Status:   "pending",
		Provider: "manual",
	})
	if err != nil {
		t.Fatalf("CreateImageModeration: %v", err)
	}
	return img
}

func TestAdminModeration(t *testing.T) {
	env, group, strategy, adminID := setupAdminEnv(t)
	owner := testutil.SeedUser(t, env.DB, group.ID, "owner@example.com", "password", string(domain.RoleUser))
	role := domain.RoleAdmin

	t.Run("non_admin forbidden", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/moderation/pending", nil, owner.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
	})

	pending := seedPendingModerationImage(t, env, owner.ID, group.ID, strategy.ID, "mod-pending")

	t.Run("list pending", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/admin/moderation/pending", nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		data := respDataMap(t, parseResp(t, rec))
		items, ok := data["items"].([]interface{})
		if !ok || len(items) == 0 {
			t.Fatal("expected pending items")
		}
		found := false
		for _, item := range items {
			m := nestedMap(t, item)
			if m["key"] == pending.Key {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("pending list missing key %q", pending.Key)
		}
	})

	t.Run("approve", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, fmt.Sprintf("/api/v1/admin/moderation/%d/approve", pending.ID), nil, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		updated, err := env.DB.GetImageByID(t.Context(), pending.ID)
		if err != nil {
			t.Fatalf("GetImageByID: %v", err)
		}
		if updated.ModerationStatus != "approved" {
			t.Errorf("moderation_status = %q, want approved", updated.ModerationStatus)
		}
	})

	toReject := seedPendingModerationImage(t, env, owner.ID, group.ID, strategy.ID, "mod-reject")

	t.Run("reject", func(t *testing.T) {
		body := map[string]interface{}{"reason": "policy violation"}
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, fmt.Sprintf("/api/v1/admin/moderation/%d/reject", toReject.ID), body, adminID, role, 0))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		updated, err := env.DB.GetImageByID(t.Context(), toReject.ID)
		if err != nil {
			t.Fatalf("GetImageByID: %v", err)
		}
		if updated.ModerationStatus != "rejected" {
			t.Errorf("moderation_status = %q, want rejected", updated.ModerationStatus)
		}
	})

	t.Run("approve invalid id", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/admin/moderation/not-a-id/approve", nil, adminID, role, 0))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})
}
