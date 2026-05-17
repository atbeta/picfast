package handler_test

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

func seedStrategyForEnv(t *testing.T, env *testEnv, groupID int64) sqlc.Strategy {
	t.Helper()
	cfg := domain.LocalStrategyConfig{Root: env.Config.Storage.LocalRoot, URL: "/i"}
	configs, _ := json.Marshal(cfg)
	strategy, err := env.DB.CreateStrategy(t.Context(), sqlc.CreateStrategyParams{
		Name:         "Test Local",
		StrategyType: "local",
		Configs:      configs,
	})
	if err != nil {
		t.Fatalf("seed strategy: %v", err)
	}
	if err := env.DB.AddGroupStrategy(t.Context(), sqlc.AddGroupStrategyParams{
		GroupID:    groupID,
		StrategyID: strategy.ID,
	}); err != nil {
		t.Fatalf("add group strategy: %v", err)
	}
	return strategy
}

type fileAccessActor struct {
	name   string
	userID int64
	role   domain.UserRole
	authed bool
}

func TestFileHandlerImageAccess(t *testing.T) {
	env := newTestEnv(t)
	group, _, owner := env.seedSetup(t)
	strategy := seedStrategyForEnv(t, env, group.ID)
	other := testutil.SeedUser(t, env.DB, group.ID, "other@example.com", "password", string(domain.RoleUser))
	admin := testutil.SeedUser(t, env.DB, group.ID, "admin-access@example.com", "password", string(domain.RoleUser))
	makeAdmin(t, env, admin.ID)

	actors := map[string]fileAccessActor{
		"anonymous": {name: "anonymous", authed: false},
		"owner":     {name: "owner", userID: owner.ID, role: domain.RoleUser, authed: true},
		"other":     {name: "other", userID: other.ID, role: domain.RoleUser, authed: true},
		"admin":     {name: "admin", userID: admin.ID, role: domain.RoleAdmin, authed: true},
	}

	assertImage := func(t *testing.T, img sqlc.Image, actor fileAccessActor, want int) {
		t.Helper()
		var req *http.Request
		if actor.authed {
			req = env.authReq(t, http.MethodGet, "/i/"+img.Key+".png", nil, actor.userID, actor.role, group.ID)
		} else {
			req = newJSONReq(t, http.MethodGet, "/i/"+img.Key+".png", nil)
		}
		rec := doReq(env.Router, req)
		if rec.Code != want {
			t.Fatalf("%s: status = %d, want %d; body: %s", actor.name, rec.Code, want, rec.Body.String())
		}
	}

	t.Run("private permission", func(t *testing.T) {
		img := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "private-perm", "png", int16(domain.PermissionPrivate))
		assertImage(t, img, actors["anonymous"], http.StatusNotFound)
		assertImage(t, img, actors["other"], http.StatusNotFound)
		assertImage(t, img, actors["owner"], http.StatusOK)
		assertImage(t, img, actors["admin"], http.StatusOK)
	})

	t.Run("public permission", func(t *testing.T) {
		img := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "public-perm", "png", int16(domain.PermissionPublic))
		assertImage(t, img, actors["anonymous"], http.StatusOK)
		assertImage(t, img, actors["other"], http.StatusOK)
	})

	t.Run("pending moderation", func(t *testing.T) {
		img := createTestImageWithModeration(t, env, owner.ID, group.ID, strategy.ID, "pending-mod", "png", "pending", int16(domain.PermissionPrivate))
		assertImage(t, img, actors["other"], http.StatusNotFound)
		assertImage(t, img, actors["owner"], http.StatusOK)
	})

	t.Run("rejected moderation", func(t *testing.T) {
		img := createTestImageWithModeration(t, env, owner.ID, group.ID, strategy.ID, "rejected-mod", "png", "rejected", int16(domain.PermissionPrivate))
		assertImage(t, img, actors["other"], http.StatusNotFound)
		assertImage(t, img, actors["owner"], http.StatusOK)
		assertImage(t, img, actors["admin"], http.StatusOK)
	})

	t.Run("private approved still denied to non owner", func(t *testing.T) {
		img := createTestImageWithModeration(t, env, owner.ID, group.ID, strategy.ID, "priv-approved", "png", "approved", int16(domain.PermissionPrivate))
		assertImage(t, img, actors["other"], http.StatusNotFound)
		assertImage(t, img, actors["owner"], http.StatusOK)
	})

	t.Run("public approved accessible to non owner", func(t *testing.T) {
		img := createTestImageWithModeration(t, env, owner.ID, group.ID, strategy.ID, "pub-approved", "png", "approved", int16(domain.PermissionPublic))
		assertImage(t, img, actors["other"], http.StatusOK)
	})
}

func TestFileHandlerThumbnailAccess(t *testing.T) {
	env := newTestEnv(t)
	group, _, owner := env.seedSetup(t)
	strategy := seedStrategyForEnv(t, env, group.ID)
	other := testutil.SeedUser(t, env.DB, group.ID, "thumb-other@example.com", "password", string(domain.RoleUser))
	admin := testutil.SeedUser(t, env.DB, group.ID, "admin-thumb@example.com", "password", string(domain.RoleUser))
	makeAdmin(t, env, admin.ID)

	actors := map[string]fileAccessActor{
		"anonymous": {name: "anonymous", authed: false},
		"owner":     {name: "owner", userID: owner.ID, role: domain.RoleUser, authed: true},
		"other":     {name: "other", userID: other.ID, role: domain.RoleUser, authed: true},
		"admin":     {name: "admin", userID: admin.ID, role: domain.RoleAdmin, authed: true},
	}

	assertThumb := func(t *testing.T, md5 string, actor fileAccessActor, want int) {
		t.Helper()
		var req *http.Request
		if actor.authed {
			req = env.authReq(t, http.MethodGet, "/t/"+md5+".png", nil, actor.userID, actor.role, group.ID)
		} else {
			req = newJSONReq(t, http.MethodGet, "/t/"+md5+".png", nil)
		}
		rec := doReq(env.Router, req)
		if rec.Code != want {
			t.Fatalf("%s: status = %d, want %d; body: %s", actor.name, rec.Code, want, rec.Body.String())
		}
	}

	seedPrivateThumb := func(t *testing.T, key, md5 string) {
		t.Helper()
		img := createTestImage(t, env, owner.ID, group.ID, strategy.ID, key, "png", int16(domain.PermissionPrivate))
		updateImageMD5(t, env, img.ID, md5)
		writeThumbnail(t, env, md5)
	}

	t.Run("private thumbnail", func(t *testing.T) {
		const md5 = "33333333333333333333333333333333"
		seedPrivateThumb(t, "thumb-private", md5)
		assertThumb(t, md5, actors["anonymous"], http.StatusNotFound)
		assertThumb(t, md5, actors["other"], http.StatusNotFound)
		assertThumb(t, md5, actors["owner"], http.StatusOK)
		assertThumb(t, md5, actors["admin"], http.StatusOK)
	})
}

func createTestImage(t *testing.T, env *testEnv, userID, groupID, strategyID int64, key, ext string, permission int16) sqlc.Image {
	t.Helper()
	img, err := env.DB.CreateImage(t.Context(), sqlc.CreateImageParams{
		UserID:     domain.PgInt8(userID),
		AlbumID:    pgtype.Int8{},
		GroupID:    domain.PgInt8(groupID),
		StrategyID: domain.PgInt8(strategyID),
		Key:        key,
		Path:       ".",
		Name:       key + "." + ext,
		OriginName: key + "." + ext,
		SizeBytes:  int64(len(pngBytes())),
		Mimetype:   "image/png",
		Extension:  ext,
		Md5:        "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Sha1:       "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Width:      10,
		Height:     10,
		Permission: permission,
		UploadedIp: "127.0.0.1",
		ExpiresAt:  pgtype.Timestamptz{},
	})
	if err != nil {
		t.Fatalf("create image: %v", err)
	}

	writeTestImageFile(t, env, img.Name)
	return img
}

func writeTestImageFile(t *testing.T, env *testEnv, name string) {
	t.Helper()
	if env.Config.Storage.LocalRoot == "" {
		return
	}
	fullPath := filepath.Join(env.Config.Storage.LocalRoot, name)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("mkdir storage dir: %v", err)
	}
	if err := os.WriteFile(fullPath, pngBytes(), 0644); err != nil {
		t.Fatalf("write storage file: %v", err)
	}
}

func createTestImageWithModeration(t *testing.T, env *testEnv, userID, groupID, strategyID int64, key, ext, modStatus string, permission int16) sqlc.Image {
	t.Helper()
	img := createTestImage(t, env, userID, groupID, strategyID, key, ext, permission)
	_, err := env.DB.UpdateImageModerationStatus(t.Context(), sqlc.UpdateImageModerationStatusParams{
		ID:               img.ID,
		ModerationStatus: modStatus,
	})
	if err != nil {
		t.Fatalf("update moderation status: %v", err)
	}
	img.ModerationStatus = modStatus
	return img
}

func updateImageMD5(t *testing.T, env *testEnv, imageID int64, md5 string) {
	t.Helper()
	_, err := env.Pool.Exec(t.Context(), "UPDATE images SET md5 = $1 WHERE id = $2", md5, imageID)
	if err != nil {
		t.Fatalf("update image md5: %v", err)
	}
}

func writeThumbnail(t *testing.T, env *testEnv, md5 string) {
	t.Helper()
	if err := os.MkdirAll(env.Config.Storage.ThumbnailDir, 0o755); err != nil {
		t.Fatalf("mkdir thumbnail dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(env.Config.Storage.ThumbnailDir, md5+".png"), pngBytes(), 0o644); err != nil {
		t.Fatalf("write thumbnail file: %v", err)
	}
}
