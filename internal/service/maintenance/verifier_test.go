package maintenance

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/service/storage"
)

type fakeStorageFactory struct {
	store storage.Storage
	err   error
}

func (f fakeStorageFactory) New(string, []byte) (storage.Storage, error) {
	return f.store, f.err
}

type fakeStorage struct {
	data []byte
	err  error
}

func (s fakeStorage) Write(context.Context, string, []byte, string) error { return nil }
func (s fakeStorage) Read(context.Context, string) ([]byte, error)        { return s.data, s.err }
func (s fakeStorage) Delete(context.Context, string) error                { return nil }
func (s fakeStorage) URL(string) string                                   { return "" }
func (s fakeStorage) HealthCheck(context.Context) storage.HealthResult {
	return storage.HealthResult{Healthy: true}
}
func (s fakeStorage) Close() error { return nil }

func TestVerifyObjectOK(t *testing.T) {
	data := []byte("image-data")
	md5Sum := md5.Sum(data)
	sha1Sum := sha1.Sum(data)
	id := int64(1)
	item := InventoryItem{
		ImageID:        10,
		Key:            "abc",
		StrategyID:     &id,
		StrategyType:   "local",
		StrategyConfig: json.RawMessage(`{}`),
		ObjectPath:     "2026/05/cat.png",
		SizeBytes:      int64(len(data)),
		MD5:            hex.EncodeToString(md5Sum[:]),
		SHA1:           hex.EncodeToString(sha1Sum[:]),
	}
	verifier := Verifier{StorageFactory: fakeStorageFactory{store: fakeStorage{data: data}}}

	check := verifier.VerifyObject(context.Background(), item)
	if check.Status != ObjectStatusOK {
		t.Fatalf("Status = %q, want %q", check.Status, ObjectStatusOK)
	}
}

func TestVerifyObjectDetectsSizeMismatch(t *testing.T) {
	data := []byte("image-data")
	md5Sum := md5.Sum(data)
	sha1Sum := sha1.Sum(data)
	id := int64(1)
	item := InventoryItem{
		ImageID:        10,
		Key:            "abc",
		StrategyID:     &id,
		StrategyType:   "local",
		StrategyConfig: json.RawMessage(`{}`),
		ObjectPath:     "2026/05/cat.png",
		SizeBytes:      int64(len(data) + 1),
		MD5:            hex.EncodeToString(md5Sum[:]),
		SHA1:           hex.EncodeToString(sha1Sum[:]),
	}
	verifier := Verifier{StorageFactory: fakeStorageFactory{store: fakeStorage{data: data}}}

	check := verifier.VerifyObject(context.Background(), item)
	if check.Status != ObjectStatusSizeMismatch {
		t.Fatalf("Status = %q, want %q", check.Status, ObjectStatusSizeMismatch)
	}
}

// fakeStoreFS is an in-memory storage backend for verifier/repair tests.
type fakeStoreFS struct {
	files   map[string][]byte
	written map[string][]byte
}

func newFakeStoreFS() *fakeStoreFS {
	return &fakeStoreFS{files: map[string][]byte{}, written: map[string][]byte{}}
}

func (s *fakeStoreFS) Write(_ context.Context, path string, data []byte, _ string) error {
	s.written[path] = data
	return nil
}

func (s *fakeStoreFS) Read(_ context.Context, path string) ([]byte, error) {
	if b, ok := s.files[path]; ok {
		return b, nil
	}
	return nil, errors.New("object not found: " + path)
}

func (s *fakeStoreFS) Delete(context.Context, string) error { return nil }
func (s *fakeStoreFS) URL(string) string                    { return "" }
func (s *fakeStoreFS) HealthCheck(context.Context) storage.HealthResult {
	return storage.HealthResult{Healthy: true}
}
func (s *fakeStoreFS) Close() error { return nil }

func thumbnailItem(md5, ext, objectPath string) InventoryItem {
	id := int64(1)
	return InventoryItem{
		ImageID:        1,
		Key:            "abc",
		Extension:      ext,
		MD5:            md5,
		StrategyID:     &id,
		StrategyType:   "local",
		StrategyConfig: json.RawMessage(`{}`),
		ObjectPath:     objectPath,
	}
}

func TestVerifyThumbnail(t *testing.T) {
	dir := t.TempDir()
	const md5 = "0123456789abcdef0123456789abcdef"

	t.Run("skipped", func(t *testing.T) {
		item := InventoryItem{ImageID: 1, Key: "abc", Extension: "svg", MD5: "abc"}
		verifier := Verifier{ThumbnailDir: dir}
		check := verifier.VerifyThumbnail(context.Background(), item)
		if check.Status != ThumbnailStatusSkipped {
			t.Fatalf("Status = %q, want %q", check.Status, ThumbnailStatusSkipped)
		}
	})

	t.Run("ok in strategy storage", func(t *testing.T) {
		store := newFakeStoreFS()
		objPath := service.ThumbnailObjectPath(md5)
		store.files[objPath] = []byte("thumb")
		verifier := Verifier{
			ThumbnailDir:   dir,
			StorageFactory: fakeStorageFactory{store: store},
		}
		check := verifier.VerifyThumbnail(context.Background(), thumbnailItem(md5, "png", "cat.png"))
		if check.Status != ThumbnailStatusOK {
			t.Fatalf("Status = %q, want %q (err %q)", check.Status, ThumbnailStatusOK, check.Error)
		}
		if check.Path != objPath {
			t.Fatalf("Path = %q, want %q", check.Path, objPath)
		}
	})

	t.Run("legacy fallback", func(t *testing.T) {
		legacyPath := filepath.Join(dir, md5+".png")
		if err := os.WriteFile(legacyPath, []byte("old thumb"), 0o644); err != nil {
			t.Fatalf("write legacy thumbnail: %v", err)
		}
		verifier := Verifier{
			ThumbnailDir:   dir,
			StorageFactory: fakeStorageFactory{store: newFakeStoreFS()},
		}
		check := verifier.VerifyThumbnail(context.Background(), thumbnailItem(md5, "png", "cat.png"))
		if check.Status != ThumbnailStatusLegacy {
			t.Fatalf("Status = %q, want %q", check.Status, ThumbnailStatusLegacy)
		}
		if check.Path != legacyPath {
			t.Fatalf("Path = %q, want %q", check.Path, legacyPath)
		}
	})

	t.Run("missing", func(t *testing.T) {
		verifier := Verifier{
			ThumbnailDir:   t.TempDir(),
			StorageFactory: fakeStorageFactory{store: newFakeStoreFS()},
		}
		check := verifier.VerifyThumbnail(context.Background(), thumbnailItem(md5, "png", "cat.png"))
		if check.Status != ThumbnailStatusMissing {
			t.Fatalf("Status = %q, want %q", check.Status, ThumbnailStatusMissing)
		}
	})

	t.Run("store init failure is unreadable", func(t *testing.T) {
		verifier := Verifier{
			ThumbnailDir:   dir,
			StorageFactory: fakeStorageFactory{err: errors.New("init failed")},
		}
		check := verifier.VerifyThumbnail(context.Background(), thumbnailItem(md5, "png", "cat.png"))
		if check.Status != ThumbnailStatusUnreadable {
			t.Fatalf("Status = %q, want %q", check.Status, ThumbnailStatusUnreadable)
		}
		if check.Error == "" {
			t.Fatal("expected error detail")
		}
	})
}

func TestRepairThumbnailsDryRunDedupesByMD5(t *testing.T) {
	dir := t.TempDir()
	items := []InventoryItem{
		{ImageID: 1, Key: "a", Extension: "png", MD5: "same"},
		{ImageID: 2, Key: "b", Extension: "png", MD5: "same"},
	}
	repairer := ThumbnailRepairer{Verifier: Verifier{ThumbnailDir: dir}}

	result := repairer.Repair(context.Background(), items, RepairOptions{})
	if result.WouldRepair != 1 {
		t.Fatalf("WouldRepair = %d, want 1", result.WouldRepair)
	}
	if result.Skipped != 1 {
		t.Fatalf("Skipped = %d, want 1", result.Skipped)
	}
	if result.Items[1].Status != RepairStatusDuplicate {
		t.Fatalf("second status = %q, want %q", result.Items[1].Status, RepairStatusDuplicate)
	}
}

func TestRepairThumbnailsWritesIntoStorage(t *testing.T) {
	const md5 = "0123456789abcdef0123456789abcdef"

	runRepair := func(t *testing.T, setup func(t *testing.T, store *fakeStoreFS) Verifier) (RepairResult, *fakeStoreFS) {
		t.Helper()
		store := newFakeStoreFS()
		store.files["2026/05/cat.png"] = []byte("original image bytes")
		verifier := setup(t, store)
		var writtenVia []string
		repairer := ThumbnailRepairer{
			Verifier: verifier,
			Generator: func(ctx context.Context, st storage.Storage, data []byte, ext, m string) error {
				if st != store {
					t.Fatal("generator did not receive the strategy store")
				}
				writtenVia = append(writtenVia, m)
				return st.Write(ctx, service.ThumbnailObjectPath(m), []byte("png"), "image/png")
			},
		}
		item := thumbnailItem(md5, "png", "2026/05/cat.png")
		result := repairer.Repair(context.Background(), []InventoryItem{item}, RepairOptions{Apply: true})
		if len(writtenVia) != 1 {
			t.Fatalf("generator calls = %d, want 1", len(writtenVia))
		}
		return result, store
	}

	t.Run("missing", func(t *testing.T) {
		result, store := runRepair(t, func(t *testing.T, store *fakeStoreFS) Verifier {
			return Verifier{ThumbnailDir: t.TempDir(), StorageFactory: fakeStorageFactory{store: store}}
		})
		if result.Repaired != 1 || result.Items[0].Status != RepairStatusRepaired {
			t.Fatalf("result = %+v", result)
		}
		if _, ok := store.written[service.ThumbnailObjectPath(md5)]; !ok {
			t.Fatal("thumbnail object was not written into strategy storage")
		}
	})

	t.Run("legacy", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, md5+".png"), []byte("old"), 0o644); err != nil {
			t.Fatalf("write legacy thumbnail: %v", err)
		}
		result, store := runRepair(t, func(t *testing.T, store *fakeStoreFS) Verifier {
			return Verifier{ThumbnailDir: dir, StorageFactory: fakeStorageFactory{store: store}}
		})
		if result.Repaired != 1 || result.Items[0].Status != RepairStatusRepaired {
			t.Fatalf("result = %+v", result)
		}
		if _, ok := store.written[service.ThumbnailObjectPath(md5)]; !ok {
			t.Fatal("thumbnail object was not written into strategy storage")
		}
	})
}
