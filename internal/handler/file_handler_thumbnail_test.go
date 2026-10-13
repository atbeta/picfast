package handler_test

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/service"
)

// strategyThumbPath returns the on-disk location a local strategy uses for the
// thumbnail object of md5.
func strategyThumbPath(env *testEnv, md5 string) string {
	return filepath.Join(env.Config.Storage.LocalRoot, "_picfast", "thumbnails", md5+".png")
}

func writeStrategyThumbnail(t *testing.T, env *testEnv, md5 string) {
	t.Helper()
	path := strategyThumbPath(env, md5)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir thumbnail object dir: %v", err)
	}
	if err := os.WriteFile(path, pngBytes(), 0o644); err != nil {
		t.Fatalf("write thumbnail object: %v", err)
	}
}

func TestFileHandlerThumbnailFromStrategyStore(t *testing.T) {
	env := newTestEnv(t)
	group, _, owner := env.seedSetup(t)
	strategy := seedStrategyForEnv(t, env, group.ID)
	const md5 = "44444444444444444444444444444444"
	img := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "thumb-store", "png", int16(domain.PermissionPublic))
	updateImageMD5(t, env, img.ID, md5)
	writeStrategyThumbnail(t, env, md5)

	req := newJSONReq(t, http.MethodGet, "/t/"+md5+".png", nil)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/png") {
		t.Fatalf("Content-Type = %q, want image/png", ct)
	}
	if etag := rec.Header().Get("ETag"); etag != `"t-`+md5+`"` {
		t.Fatalf("ETag = %q, want %q", etag, `"t-`+md5+`"`)
	}

	req = newJSONReq(t, http.MethodGet, "/t/"+md5+".png", nil)
	req.Header.Set("If-None-Match", `"t-`+md5+`"`)
	rec = doReq(env.Router, req)
	if rec.Code != http.StatusNotModified {
		t.Fatalf("If-None-Match status = %d, want 304", rec.Code)
	}
}

func TestFileHandlerThumbnailLazyRegen(t *testing.T) {
	env := newTestEnv(t)
	group, _, owner := env.seedSetup(t)
	strategy := seedStrategyForEnv(t, env, group.ID)
	const md5 = "55555555555555555555555555555555"
	img := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "thumb-regen", "png", int16(domain.PermissionPublic))
	updateImageMD5(t, env, img.ID, md5)

	req := newJSONReq(t, http.MethodGet, "/t/"+md5+".png", nil)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "image/png") {
		t.Fatalf("Content-Type = %q, want image/png", ct)
	}
	if _, err := os.Stat(strategyThumbPath(env, md5)); err != nil {
		t.Fatalf("expected regenerated thumbnail object to be written: %v", err)
	}
}

func TestDeleteRemovesStrategyAndLegacyThumbnail(t *testing.T) {
	env := newTestEnv(t)
	group, _, owner := env.seedSetup(t)
	strategy := seedStrategyForEnv(t, env, group.ID)
	const md5 = "66666666666666666666666666666666"
	img := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "thumb-del", "png", int16(domain.PermissionPublic))
	updateImageMD5(t, env, img.ID, md5)
	writeStrategyThumbnail(t, env, md5)
	writeThumbnail(t, env, md5) // legacy on-disk fallback

	deleter := service.NewDeleteService(env.DB, env.Pool, env.Config.Storage.ThumbnailDir, events.NewCollectEmitter())
	if err := deleter.DeleteImage(t.Context(), img.ID); err != nil {
		t.Fatalf("delete image: %v", err)
	}
	if _, err := os.Stat(strategyThumbPath(env, md5)); !os.IsNotExist(err) {
		t.Fatalf("strategy thumbnail should be removed, stat err = %v", err)
	}
	legacyPath := filepath.Join(env.Config.Storage.ThumbnailDir, md5+".png")
	if _, err := os.Stat(legacyPath); !os.IsNotExist(err) {
		t.Fatalf("legacy thumbnail should be removed, stat err = %v", err)
	}
}

func TestDeleteRetainsThumbnailForSameMD5Sibling(t *testing.T) {
	env := newTestEnv(t)
	group, _, owner := env.seedSetup(t)
	strategy := seedStrategyForEnv(t, env, group.ID)
	const md5 = "77777777777777777777777777777777"
	first := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "thumb-dup-a", "png", int16(domain.PermissionPublic))
	second := createTestImage(t, env, owner.ID, group.ID, strategy.ID, "thumb-dup-b", "png", int16(domain.PermissionPublic))
	// Identical content => same md5+sha1, so the object and thumbnail are shared.
	updateImageMD5(t, env, first.ID, md5)
	updateImageMD5(t, env, second.ID, md5)
	writeStrategyThumbnail(t, env, md5)
	writeThumbnail(t, env, md5)

	deleter := service.NewDeleteService(env.DB, env.Pool, env.Config.Storage.ThumbnailDir, events.NewCollectEmitter())
	if err := deleter.DeleteImage(t.Context(), first.ID); err != nil {
		t.Fatalf("delete image: %v", err)
	}
	if _, err := os.Stat(strategyThumbPath(env, md5)); err != nil {
		t.Fatalf("strategy thumbnail should be retained: %v", err)
	}
	legacyPath := filepath.Join(env.Config.Storage.ThumbnailDir, md5+".png")
	if _, err := os.Stat(legacyPath); err != nil {
		t.Fatalf("legacy thumbnail should be retained: %v", err)
	}
}
