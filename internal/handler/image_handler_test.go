package handler_test

import (
	"net/http"
	"testing"

	"github.com/pbeta/imgapi/internal/domain"
	"github.com/pbeta/imgapi/internal/testutil"
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
	req := uploadReq(t, "/api/v1/images", "test.png", pngBytes(), token)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("upload failed: %d; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	return data["key"].(string)
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
