package service

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/atbeta/picfast/internal/config"
	"github.com/atbeta/picfast/internal/events"
	"github.com/atbeta/picfast/internal/service/storage"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/testutil"
)

// fakeStorage is an in-memory storage.Storage used to observe the overlap
// between the original write and the thumbnail write.
type fakeStorage struct {
	mu                    sync.Mutex
	written               map[string][]byte
	deleted               []string
	closed                bool
	thumbWriteDone        bool
	closedAfterThumbWrite bool

	// writeErrFor returns a write error for a path, or nil to succeed.
	writeErrFor func(path string) error
	// onWrite runs inside Write before data is recorded (barrier hook).
	onWrite func(path string)
}

func newFakeStorage() *fakeStorage {
	return &fakeStorage{written: map[string][]byte{}}
}

func (f *fakeStorage) Write(ctx context.Context, path string, data []byte, contentType string) error {
	if f.onWrite != nil {
		f.onWrite(path)
	}
	if f.writeErrFor != nil {
		if err := f.writeErrFor(path); err != nil {
			return err
		}
	}
	f.mu.Lock()
	f.written[path] = data
	if strings.HasPrefix(path, ThumbnailPrefix) {
		f.thumbWriteDone = true
	}
	f.mu.Unlock()
	return nil
}

func (f *fakeStorage) Read(ctx context.Context, path string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	data, ok := f.written[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (f *fakeStorage) Delete(ctx context.Context, path string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deleted = append(f.deleted, path)
	return nil
}

func (f *fakeStorage) URL(pathname string) string { return "" }

func (f *fakeStorage) HealthCheck(ctx context.Context) storage.HealthResult {
	return storage.HealthResult{}
}

func (f *fakeStorage) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
	// True only if the thumbnail goroutine finished writing before Close.
	f.closedAfterThumbWrite = f.thumbWriteDone
	return nil
}

func (f *fakeStorage) deletedPaths() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.deleted...)
}

func (f *fakeStorage) closedAfterThumbnail() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.closed && f.closedAfterThumbWrite
}

func newUploadServiceWithStorage(t *testing.T, store storage.Storage) (*UploadService, int64) {
	t.Helper()
	pool, q := testutil.SetupDB(t)
	group := testutil.SeedDefaultGroup(t, q)
	testutil.SeedStrategy(t, q, group.ID)
	user := testutil.SeedUser(t, q, group.ID, "uploader@example.com", "password123", "user")

	svc := NewUploadService(q, pool, &config.Config{}, events.NewCollectEmitter())
	svc.newStorage = func(sqlc.Strategy) (storage.Storage, error) {
		return store, nil
	}
	return svc, user.ID
}

// The original write and the thumbnail write must run concurrently: each
// blocks until the other has started, so a sequential implementation would
// hit the 5s timeout and fail the test.
func TestUploadOverlapsThumbnailWrite(t *testing.T) {
	origStarted := make(chan struct{})
	thumbStarted := make(chan struct{})
	var origOnce, thumbOnce sync.Once
	fake := newFakeStorage()
	fake.onWrite = func(path string) {
		if strings.HasPrefix(path, ThumbnailPrefix) {
			thumbOnce.Do(func() { close(thumbStarted) })
			select {
			case <-origStarted:
			case <-time.After(5 * time.Second):
				t.Error("thumbnail write timed out waiting for original write to start")
			}
		} else {
			origOnce.Do(func() { close(origStarted) })
			select {
			case <-thumbStarted:
			case <-time.After(5 * time.Second):
				t.Error("original write timed out waiting for thumbnail write to start")
			}
		}
	}

	svc, userID := newUploadServiceWithStorage(t, fake)

	fileData := createTestImage("png", 120, 60)
	res, err := svc.Store(context.Background(), UploadParams{
		FileData: fileData,
		FileName: "overlap.png",
		UserID:   &userID,
	})
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if res == nil {
		t.Fatal("expected result")
	}
	if !fake.closedAfterThumbnail() {
		t.Fatal("store closed before the thumbnail goroutine finished")
	}
	// Store re-encodes the upload, so the thumbnail md5 differs from the
	// input's; assert by the managed prefix instead.
	for path := range fake.written {
		if strings.HasPrefix(path, ThumbnailPrefix) {
			return
		}
	}
	t.Fatal("thumbnail not written")
}

// When the original write fails, the upload returns an error, the thumbnail
// object is deleted, and the thumbnail goroutine has finished before the
// store is closed.
func TestUploadOriginalWriteFailureCleansUpThumbnail(t *testing.T) {
	fake := newFakeStorage()
	fake.writeErrFor = func(path string) error {
		if strings.HasPrefix(path, ThumbnailPrefix) {
			return nil
		}
		return errors.New("injected original write failure")
	}

	svc, userID := newUploadServiceWithStorage(t, fake)

	_, err := svc.Store(context.Background(), UploadParams{
		FileData: createTestImage("png", 120, 60),
		FileName: "fail.png",
		UserID:   &userID,
	})
	if err == nil {
		t.Fatal("expected upload error")
	}

	deleted := false
	for _, p := range fake.deletedPaths() {
		if strings.HasPrefix(p, ThumbnailPrefix) {
			deleted = true
		}
	}
	if !deleted {
		t.Fatalf("thumbnail not deleted after original write failure; deleted = %v", fake.deletedPaths())
	}
	if !fake.closedAfterThumbnail() {
		t.Fatal("store closed before the thumbnail goroutine finished")
	}
}

// A panic inside thumbnail generation is converted into an error result:
// the upload still succeeds and the store is closed afterwards.
func TestUploadThumbnailPanicDoesNotCrash(t *testing.T) {
	fake := newFakeStorage()
	fake.onWrite = func(path string) {
		if strings.HasPrefix(path, ThumbnailPrefix) {
			panic("boom")
		}
	}

	svc, userID := newUploadServiceWithStorage(t, fake)

	res, err := svc.Store(context.Background(), UploadParams{
		FileData: createTestImage("png", 120, 60),
		FileName: "panic.png",
		UserID:   &userID,
	})
	if err != nil {
		t.Fatalf("upload should succeed despite thumbnail panic: %v", err)
	}
	if res == nil {
		t.Fatal("expected result")
	}
	if !fake.closed {
		t.Fatal("store was not closed")
	}
}
