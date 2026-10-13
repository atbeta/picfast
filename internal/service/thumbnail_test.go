package service

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/atbeta/picfast/internal/service/storage"
)

func TestThumbnailObjectPath(t *testing.T) {
	const md5 = "0123456789abcdef0123456789abcdef"
	if got := ThumbnailObjectPath(md5); got != "_picfast/thumbnails/"+md5+".png" {
		t.Fatalf("ThumbnailObjectPath = %q", got)
	}
}

func newTestLocalStorage(t *testing.T, root string) storage.Storage {
	t.Helper()
	cfg, err := json.Marshal(domain.LocalStrategyConfig{Root: root, URL: "/i"})
	if err != nil {
		t.Fatalf("marshal local config: %v", err)
	}
	store, err := storage.New("local", cfg)
	if err != nil {
		t.Fatalf("storage.New: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestStoreThumbnailWritesPNGIntoStorage(t *testing.T) {
	root := t.TempDir()
	store := newTestLocalStorage(t, root)

	const md5 = "0123456789abcdef0123456789abcdef"
	data := createTestImage("png", 200, 200)
	if err := StoreThumbnail(context.Background(), store, data, "png", md5); err != nil {
		t.Fatalf("StoreThumbnail: %v", err)
	}

	written, err := os.ReadFile(filepath.Join(root, "_picfast", "thumbnails", md5+".png"))
	if err != nil {
		t.Fatalf("read thumbnail object: %v", err)
	}
	if len(written) < 8 || string(written[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("thumbnail is not a PNG (first bytes: %x)", written[:min(8, len(written))])
	}
}

func TestStoreThumbnailSkipsSVG(t *testing.T) {
	root := t.TempDir()
	store := newTestLocalStorage(t, root)

	const md5 = "fedcba9876543210fedcba9876543210"
	if err := StoreThumbnail(context.Background(), store, []byte("<svg/>"), "svg", md5); err != nil {
		t.Fatalf("StoreThumbnail: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "_picfast")); !os.IsNotExist(err) {
		t.Fatalf("expected no thumbnail writes, stat err = %v", err)
	}
}

func TestStoreThumbnailWriteFailure(t *testing.T) {
	// root is a file, so LocalStorage.Write fails inside MkdirAll.
	root := filepath.Join(t.TempDir(), "blocker")
	if err := os.WriteFile(root, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	store := newTestLocalStorage(t, root)

	const md5 = "0123456789abcdef0123456789abcdef"
	data := createTestImage("png", 200, 200)
	if err := StoreThumbnail(context.Background(), store, data, "png", md5); err == nil {
		t.Fatal("expected write failure error, got nil")
	}
}

func TestStoreThumbnailDecodeFailureIsBestEffort(t *testing.T) {
	store := newTestLocalStorage(t, t.TempDir())
	const md5 = "0123456789abcdef0123456789abcdef"
	if err := StoreThumbnail(context.Background(), store, []byte("not an image"), "png", md5); err != nil {
		t.Fatalf("expected nil for undecodable data, got %v", err)
	}
}
