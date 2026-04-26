package handler_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/testutil"
)

func TestAlbumCreate(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	t.Run("success", func(t *testing.T) {
		body := map[string]string{"name": "My Album", "intro": "album intro"}
		req := env.authReq(t, http.MethodPost, "/api/v1/albums", body, user.ID, domain.RoleUser, group.ID)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
		album := respDataMap(t, parseResp(t, rec))
		if album["name"] != "My Album" {
			t.Fatalf("name = %v, want My Album", album["name"])
		}

		u, _ := env.DB.GetUserByID(t.Context(), user.ID)
		if u.AlbumNum != 1 {
			t.Fatalf("album_num = %d, want 1", u.AlbumNum)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		body := map[string]string{"intro": "no name"}
		req := env.authReq(t, http.MethodPost, "/api/v1/albums", body, user.ID, domain.RoleUser, group.ID)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		body := map[string]string{"name": "No Auth"}
		req := newJSONReq(t, http.MethodPost, "/api/v1/albums", body)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAlbumList(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	testutil.SeedAlbum(t, env.DB, user.ID, "Album A")
	testutil.SeedAlbum(t, env.DB, user.ID, "Album B")

	req := env.authReq(t, http.MethodGet, "/api/v1/albums", nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	data := respDataMap(t, parseResp(t, rec))
	items := data["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	if int(data["total"].(float64)) != 2 {
		t.Fatalf("total = %v, want 2", data["total"])
	}
}

func TestAlbumUpdate(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)
	album := testutil.SeedAlbum(t, env.DB, user.ID, "Original")

	body := map[string]string{"name": "Updated", "intro": "new intro"}
	req := env.authReq(t, http.MethodPut, fmt.Sprintf("/api/v1/albums/%d", album.ID), body, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	result := respDataMap(t, parseResp(t, rec))
	if result["name"] != "Updated" {
		t.Fatalf("name = %v, want Updated", result["name"])
	}
}

func TestAlbumDelete(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)
	album := testutil.SeedAlbum(t, env.DB, user.ID, "To Delete")
	env.DB.IncrementUserAlbumNum(t.Context(), user.ID)

	req := env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/albums/%d", album.ID), nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	u, _ := env.DB.GetUserByID(t.Context(), user.ID)
	if u.AlbumNum != 0 {
		t.Fatalf("album_num = %d, want 0", u.AlbumNum)
	}
}

func TestAlbumDeleteOtherUser(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	other := testutil.SeedUser(t, env.DB, group.ID, "other@example.com", "password123", string(domain.RoleUser))
	album := testutil.SeedAlbum(t, env.DB, other.ID, "Other's Album")

	req := env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/albums/%d", album.ID), nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
