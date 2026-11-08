package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
)

func createAPITokenViaAPI(t *testing.T, env *testEnv, userID, groupID int64, body map[string]interface{}) (tokenID int64, plainToken string) {
	t.Helper()
	rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/api-tokens", body, userID, domain.RoleUser, groupID))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create token status = %d, want 201; body: %s", rec.Code, rec.Body.String())
	}
	data := respDataMap(t, parseResp(t, rec))
	plainToken, _ = data["token"].(string)
	if plainToken == "" {
		t.Fatal("plain token missing from create response")
	}
	tokenID = int64(data["id"].(float64))
	return tokenID, plainToken
}

func TestAPITokenCRUD(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	tokenID, plainToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "integration token",
		"expires_in": "never",
		"scopes":     []string{"read", "write"},
	})

	t.Run("list", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/api-tokens", nil, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		resp := parseResp(t, rec)
		items, ok := resp.Data.([]interface{})
		if !ok || len(items) == 0 {
			t.Fatal("expected token list")
		}
	})

	t.Run("delete", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodDelete, fmt.Sprintf("/api/v1/api-tokens/%d", tokenID), nil, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("revoked token unauthorized", func(t *testing.T) {
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, plainToken))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAPITokenDualAuth(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	_, plainToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "dual-auth token",
		"expires_in": "never",
		"scopes":     []string{"read", "write"},
	})

	t.Run("x_api_token lists images", func(t *testing.T) {
		uploadReq := uploadReq(t, "/api/v1/images", "api-tok.png", pngBytes(), env.generateToken(t, user.ID, domain.RoleUser, group.ID))
		if rec := doReq(env.Router, uploadReq); rec.Code != http.StatusCreated {
			t.Fatalf("seed upload status = %d, want 201", rec.Code)
		}

		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, plainToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("bearer api token lists images", func(t *testing.T) {
		rec := doReq(env.Router, env.bearerAPITokenReq(t, http.MethodGet, "/api/v1/images", nil, plainToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("invalid token unauthorized", func(t *testing.T) {
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, "img_invalid"))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", rec.Code)
		}
	})
}

func TestAPITokenRequireScope(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	_, readOnlyToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":       "read only",
		"expires_in": "never",
		"scopes":     []string{"read"},
	})

	t.Run("read scope can list", func(t *testing.T) {
		rec := doReq(env.Router, env.apiTokenReq(t, http.MethodGet, "/api/v1/images", nil, readOnlyToken))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("read scope cannot upload", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/images", "scoped.png", pngBytes(), "")
		req.Header.Set("X-API-Token", readOnlyToken)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("jwt upload without scope restriction", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/images", "jwt.png", pngBytes(), env.generateToken(t, user.ID, domain.RoleUser, group.ID))
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
	})
}

func TestAPITokenUploadTarget(t *testing.T) {
	env := newTestEnv(t)
	group, strategyA, user := env.seedSetup(t)
	strategyB := seedStrategyForEnv(t, env, group.ID)

	var cfg domain.GroupConfig
	if err := json.Unmarshal(group.Configs, &cfg); err != nil {
		t.Fatalf("decode group config: %v", err)
	}
	cfg.DefaultStrategyID = strategyA.ID
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("encode group config: %v", err)
	}
	if _, err := env.DB.UpdateGroup(t.Context(), sqlc.UpdateGroupParams{
		ID:        group.ID,
		Name:      group.Name,
		IsDefault: group.IsDefault,
		IsGuest:   group.IsGuest,
		Configs:   raw,
	}); err != nil {
		t.Fatalf("set default strategy: %v", err)
	}

	album, err := env.DB.CreateAlbum(t.Context(), sqlc.CreateAlbumParams{UserID: user.ID, Name: "Notes", Intro: ""})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}
	otherAlbum, err := env.DB.CreateAlbum(t.Context(), sqlc.CreateAlbumParams{UserID: user.ID, Name: "Override", Intro: ""})
	if err != nil {
		t.Fatalf("create override album: %v", err)
	}
	foreignStrategy, err := env.DB.CreateStrategy(t.Context(), sqlc.CreateStrategyParams{
		Name:         "Foreign",
		StrategyType: "local",
		Configs:      strategyA.Configs,
	})
	if err != nil {
		t.Fatalf("create foreign strategy: %v", err)
	}
	other := testutil.SeedUser(t, env.DB, group.ID, "other-token@example.com", "password123", string(domain.RoleUser))
	foreignAlbum, err := env.DB.CreateAlbum(t.Context(), sqlc.CreateAlbumParams{UserID: other.ID, Name: "Theirs", Intro: ""})
	if err != nil {
		t.Fatalf("create foreign album: %v", err)
	}

	t.Run("rejects a strategy outside the group", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/api-tokens", map[string]interface{}{
			"name":        "foreign strategy",
			"strategy_id": foreignStrategy.ID,
		}, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("rejects another user's album", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/api-tokens", map[string]interface{}{
			"name":     "foreign album",
			"album_id": foreignAlbum.ID,
		}, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400; body: %s", rec.Code, rec.Body.String())
		}
	})

	_, plainToken := createAPITokenViaAPI(t, env, user.ID, group.ID, map[string]interface{}{
		"name":        "obsidian",
		"expires_in":  "never",
		"scopes":      []string{"read", "write"},
		"strategy_id": strategyB.ID,
		"album_id":    album.ID,
	})

	t.Run("list returns the binding", func(t *testing.T) {
		rec := doReq(env.Router, env.authReq(t, http.MethodGet, "/api/v1/api-tokens", nil, user.ID, domain.RoleUser, group.ID))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		body := rec.Body.String()
		if !bytes.Contains(rec.Body.Bytes(), []byte(fmt.Sprintf(`"strategy_id":%d`, strategyB.ID))) &&
			!bytes.Contains(rec.Body.Bytes(), []byte(fmt.Sprintf(`"strategy_id": %d`, strategyB.ID))) {
			t.Fatalf("list missing strategy %d: %s", strategyB.ID, body)
		}
		if !bytes.Contains(rec.Body.Bytes(), []byte(fmt.Sprintf(`"album_id":%d`, album.ID))) &&
			!bytes.Contains(rec.Body.Bytes(), []byte(fmt.Sprintf(`"album_id": %d`, album.ID))) {
			t.Fatalf("list missing album %d: %s", album.ID, body)
		}
	})

	assertStored := func(t *testing.T, key string, strategyID, albumID int64) {
		t.Helper()
		img, err := env.DB.GetImageByKey(t.Context(), key)
		if err != nil {
			t.Fatalf("load image: %v", err)
		}
		if !img.StrategyID.Valid || img.StrategyID.Int64 != strategyID {
			t.Fatalf("strategy = %v, want %d", img.StrategyID, strategyID)
		}
		if !img.AlbumID.Valid || img.AlbumID.Int64 != albumID {
			t.Fatalf("album = %v, want %d", img.AlbumID, albumID)
		}
	}

	t.Run("token binding fills a bare upload", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/images", "bound.png", pngBytes(), "")
		req.Header.Set("X-API-Token", plainToken)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
		data := respDataMap(t, parseResp(t, rec))
		key, _ := data["key"].(string)
		assertStored(t, key, strategyB.ID, album.ID)
	})

	t.Run("sharex upload uses the same binding", func(t *testing.T) {
		req := uploadReq(t, "/api/v1/sharex/upload", "sharex-bound.png", pngBytes(), "")
		req.Header.Set("X-API-Token", plainToken)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
		}
		img := latestImageByOrigin(t, env, "sharex-bound.png")
		assertStored(t, img.Key, strategyB.ID, album.ID)
	})

	t.Run("explicit request overrides the token", func(t *testing.T) {
		req := uploadReqWithFields(t, "/api/v1/images", "override.png", pngBytes(), map[string]string{
			"strategy_id": fmt.Sprint(strategyA.ID),
			"album_id":    fmt.Sprint(otherAlbum.ID),
		})
		req.Header.Set("X-API-Token", plainToken)
		rec := doReq(env.Router, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("status = %d, want 201; body: %s", rec.Code, rec.Body.String())
		}
		data := respDataMap(t, parseResp(t, rec))
		key, _ := data["key"].(string)
		assertStored(t, key, strategyA.ID, otherAlbum.ID)
	})
}

func latestImageByOrigin(t *testing.T, env *testEnv, origin string) sqlc.Image {
	t.Helper()
	rows, err := env.Pool.Query(t.Context(), `SELECT id, key, strategy_id, album_id FROM images WHERE origin_name = $1 ORDER BY id DESC LIMIT 1`, origin)
	if err != nil {
		t.Fatalf("query image: %v", err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatalf("image %q not found", origin)
	}
	var img sqlc.Image
	if err := rows.Scan(&img.ID, &img.Key, &img.StrategyID, &img.AlbumID); err != nil {
		t.Fatalf("scan image: %v", err)
	}
	return img
}

func uploadReqWithFields(t *testing.T, path, fileName string, fileData []byte, fields map[string]string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", fileName)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(fileData); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	for key, value := range fields {
		if err := w.WriteField(key, value); err != nil {
			t.Fatalf("write field %s: %v", key, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close form: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestAPITokenCreateValidation(t *testing.T) {
	env := newTestEnv(t)
	_, group, user := env.seedSetup(t)

	rec := doReq(env.Router, env.authReq(t, http.MethodPost, "/api/v1/api-tokens", map[string]interface{}{
		"name": "",
	}, user.ID, domain.RoleUser, group.ID))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}
