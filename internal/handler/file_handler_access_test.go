package handler_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestPrivateThumbnailAllowsAdmin(t *testing.T) {
	env := newTestEnv(t)
	group, strategy, owner := env.seedSetup(t)
	admin := testutil.SeedUser(t, env.DB, group.ID, "admin-thumb@example.com", "password123", string(domain.RoleUser))
	makeAdmin(t, env, admin.ID)

	md5 := "11111111111111111111111111111111"
	if err := os.MkdirAll(env.Config.Storage.ThumbnailDir, 0o755); err != nil {
		t.Fatalf("mkdir thumbnail dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(env.Config.Storage.ThumbnailDir, md5+".png"), pngBytes(), 0o644); err != nil {
		t.Fatalf("write thumbnail file: %v", err)
	}

	_, err := env.DB.CreateImage(t.Context(), sqlc.CreateImageParams{
		UserID:     domain.PgInt8(owner.ID),
		AlbumID:    pgtype.Int8{},
		GroupID:    domain.PgInt8(group.ID),
		StrategyID: domain.PgInt8(strategy.ID),
		Key:        "admintestkey",
		Path:       ".",
		Name:       "admintest.png",
		OriginName: "admintest.png",
		SizeBytes:  int64(len(pngBytes())),
		Mimetype:   "image/png",
		Extension:  "png",
		Md5:        md5,
		Sha1:       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Width:      10,
		Height:     10,
		Permission: int16(domain.PermissionPrivate),
		UploadedIp: "127.0.0.1",
		ExpiresAt:  pgtype.Timestamptz{},
	})
	if err != nil {
		t.Fatalf("create private image: %v", err)
	}

	req := env.authReq(t, http.MethodGet, "/t/"+md5+".png", nil, admin.ID, domain.RoleAdmin, group.ID)
	rec := doReq(env.Router, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", rec.Code, rec.Body.String())
	}
}
