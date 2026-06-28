package handler_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/testutil"
)

func pngBytes() []byte {
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
		0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
		0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
		0x00, 0x00, 0x02, 0x00, 0x01, 0xE2, 0x21, 0xBC,
		0x33, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
		0x44, 0xAE, 0x42, 0x60, 0x82,
	}
}

func uploadAndGetKey(t *testing.T, env *testEnv, token string) string {
	t.Helper()
	key, _ := uploadAndGetImage(t, env, token)
	return key
}

func uploadAndGetImage(t *testing.T, env *testEnv, token string) (key string, id int64) {
	t.Helper()
	req := uploadReq(t, "/api/v1/images", "test.png", pngBytes(), token)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload failed: %d; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	return data["key"].(string), int64(data["id"].(float64))
}

func TestImageUpload(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	t.Run("authenticated user", func(t *testing.T) {
		token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
		req := uploadReq(t, "/api/v1/images", "test.png", pngBytes(), token)
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
		img := respDataMap(t, parseResp(t, rec))
		if img["key"] == nil {
			t.Fatal("missing key")
		}

		u, _ := env.DB.GetUserByID(t.Context(), user.ID)
		if u.ImageNum != 1 {
			t.Fatalf("image_num = %d, want 1", u.ImageNum)
		}
	})

	t.Run("guest upload", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/upload", "guest.png", pngBytes(), "")
		rec := doReq(env.Router, req)

		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestGuestUploadCapacityLimit(t *testing.T) {
	env := newTestEnv(t)
	env.seedSetup(t)
	env.Config.App.GuestCapacityBytes = int64(len(pngBytes()) - 1)

	req := uploadReq(t, "/api/v1/upload", "guest-limit.png", pngBytes(), "")
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "guest storage capacity exceeded") {
		t.Fatalf("body = %s, want guest storage capacity exceeded", rec.Body.String())
	}
}

func TestImageList(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	for i := 0; i < 2; i++ {
		req := uploadReq(t, "/api/v1/images", "list.png", pngBytes(), token)
		if rec := doReq(env.Router, req); rec.Code != http.StatusCreated {
			t.Fatalf("upload %d failed: %d", i, rec.Code)
		}
	}

	req := env.authReq(t, http.MethodGet, "/api/v1/images", nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	items := data["items"].([]interface{})
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
}

func TestImageGet(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key := uploadAndGetKey(t, env, token)

	req := env.authReq(t, http.MethodGet, "/api/v1/images/"+key, nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	img := respDataMap(t, parseResp(t, rec))
	if img["mimetype"] != "image/png" {
		t.Fatalf("mimetype = %v, want image/png", img["mimetype"])
	}
}

func TestImageDelete(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key := uploadAndGetKey(t, env, token)

	req := env.authReq(t, http.MethodDelete, "/api/v1/images/"+key, nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}

	u, _ := env.DB.GetUserByID(t.Context(), user.ID)
	if u.ImageNum != 0 {
		t.Fatalf("image_num = %d, want 0", u.ImageNum)
	}
}

func TestImageDeleteOtherUser(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)
	other := testutil.SeedUser(t, env.DB, group.ID, "other2@example.com", "password123", string(domain.RoleUser))

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key := uploadAndGetKey(t, env, token)

	req := env.authReq(t, http.MethodDelete, "/api/v1/images/"+key, nil, other.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestImageBatchDelete(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)
	other := testutil.SeedUser(t, env.DB, group.ID, "batch-other@example.com", "password", string(domain.RoleUser))

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key1, _ := uploadAndGetImage(t, env, token)
	key2, _ := uploadAndGetImage(t, env, token)
	otherKey := uploadAndGetKey(t, env, env.generateToken(t, other.ID, domain.RoleUser, group.ID))

	t.Run("deletes own images", func(t *testing.T) {
		body := map[string]interface{}{"keys": []string{key1, key2}}
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/images/batch-delete", body, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		data := respDataMap(t, parseResp(t, rec))
		if int(data["deleted"].(float64)) != 2 {
			t.Errorf("deleted = %v, want 2", data["deleted"])
		}
		if int(data["failed"].(float64)) != 0 {
			t.Errorf("failed = %v, want 0", data["failed"])
		}
		u, _ := env.DB.GetUserByID(t.Context(), user.ID)
		if u.ImageNum != 0 {
			t.Errorf("image_num = %d, want 0", u.ImageNum)
		}
	})

	t.Run("counts other users images as failed", func(t *testing.T) {
		ownKey := uploadAndGetKey(t, env, token)
		body := map[string]interface{}{"keys": []string{ownKey, otherKey}}
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/images/batch-delete", body, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		data := respDataMap(t, parseResp(t, rec))
		if int(data["deleted"].(float64)) != 1 {
			t.Errorf("deleted = %v, want 1", data["deleted"])
		}
		if int(data["failed"].(float64)) != 1 {
			t.Errorf("failed = %v, want 1", data["failed"])
		}
	})

	t.Run("empty keys bad request", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/images/batch-delete", map[string]interface{}{"keys": []string{}}, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("read only api token forbidden", func(t *testing.T) {
		_, readToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
			"name": "read batch", "expires_in": "never", "scopes": []string{"read"},
		})
		body := map[string]interface{}{"keys": []string{uploadAndGetKey(t, env, token)}}
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodPost, "/api/v1/images/batch-delete", body, readToken))
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestImageUpdatePermission(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key := uploadAndGetKey(t, env, token)

	body := map[string]int{"permission": 0}
	req := env.authReq(t, http.MethodPatch, "/api/v1/images/"+key, body, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	img := respDataMap(t, parseResp(t, rec))
	if img["permission"] != nil && int(img["permission"].(float64)) != 0 {
		t.Fatalf("permission = %v, want 0", img["permission"])
	}
}

func TestImagePipeline(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key := uploadAndGetKey(t, env, token)

	req := env.authReq(t, http.MethodGet, "/api/v1/images/"+key+"/pipeline", nil, user.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	pipeline := respDataMap(t, parseResp(t, rec))
	if pipeline["upload"] != "completed" {
		t.Fatalf("upload = %v, want completed", pipeline["upload"])
	}
	if pipeline["processing"] == "" {
		t.Fatalf("processing is empty")
	}
	if pipeline["thumbnail"] == "" {
		t.Fatalf("thumbnail is empty")
	}
	if pipeline["moderation"] == "" {
		t.Fatalf("moderation is empty")
	}
}

func TestImagePipelinePrivateBlocked(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)
	other := testutil.SeedUser(t, env.DB, group.ID, "pipeline-blocked@example.com", "password123", string(domain.RoleUser))

	token := env.generateToken(t, user.ID, domain.RoleUser, group.ID)
	key := uploadAndGetKey(t, env, token)

	// Set image to private
	patchBody := map[string]int{"permission": 0}
	patchReq := env.authReq(t, http.MethodPatch, "/api/v1/images/"+key, patchBody, user.ID, domain.RoleUser, group.ID)
	patchRec := doReq(env.Router, patchReq)
	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch status = %d, want 200; body: %s", patchRec.Code, patchRec.Body.String())
	}

	// Other user should get 404
	req := env.authReq(t, http.MethodGet, "/api/v1/images/"+key+"/pipeline", nil, other.ID, domain.RoleUser, group.ID)
	rec := doReq(env.Router, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}
