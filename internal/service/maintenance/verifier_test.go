package maintenance

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"testing"

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

func TestVerifyThumbnail(t *testing.T) {
	dir := t.TempDir()
	item := InventoryItem{ImageID: 1, Key: "abc", Extension: "svg", MD5: "abc"}
	verifier := Verifier{ThumbnailDir: dir}

	check := verifier.VerifyThumbnail(item)
	if check.Status != ThumbnailStatusSkipped {
		t.Fatalf("Status = %q, want %q", check.Status, ThumbnailStatusSkipped)
	}

	item.Extension = "png"
	check = verifier.VerifyThumbnail(item)
	if check.Status != ThumbnailStatusMissing {
		t.Fatalf("Status = %q, want %q", check.Status, ThumbnailStatusMissing)
	}
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
