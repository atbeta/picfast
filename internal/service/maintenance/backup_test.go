package maintenance

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupWriterAndInspectArchive(t *testing.T) {
	data := []byte("image-data")
	md5Sum := md5.Sum(data)
	sha1Sum := sha1.Sum(data)
	strategyID := int64(1)
	item := InventoryItem{
		ImageID:      10,
		Key:          "abc",
		StrategyID:   &strategyID,
		StrategyType: "local",
		ObjectPath:   "2026/05/cat.png",
		SizeBytes:    int64(len(data)),
		MimeType:     "image/png",
		Extension:    "png",
		MD5:          hex.EncodeToString(md5Sum[:]),
		SHA1:         hex.EncodeToString(sha1Sum[:]),
	}
	dir := t.TempDir()
	dumpPath := filepath.Join(dir, "database.dump")
	if err := os.WriteFile(dumpPath, []byte("dump"), 0644); err != nil {
		t.Fatalf("write dump: %v", err)
	}
	outputPath := filepath.Join(dir, "backup.tar.gz")
	writer := BackupWriter{StorageFactory: fakeStorageFactory{store: fakeStorage{data: data}}}

	result, err := writer.Write(context.Background(), []InventoryItem{item}, BackupOptions{
		OutputPath:       outputPath,
		DatabaseDumpPath: dumpPath,
		AppVersion:       "0.8.0",
		MigrationVersion: 24,
		IncludeObjects:   true,
	})
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if result.Manifest.Objects.Count != 1 {
		t.Fatalf("object count = %d, want 1", result.Manifest.Objects.Count)
	}

	inspect, err := InspectArchive(outputPath)
	if err != nil {
		t.Fatalf("InspectArchive() error = %v", err)
	}
	if !inspect.OK() {
		t.Fatalf("InspectArchive OK = false; missing=%v failures=%v", inspect.MissingChecksums, inspect.ChecksumFailures)
	}
	if inspect.VerifiedCount != 2 {
		t.Fatalf("VerifiedCount = %d, want 2", inspect.VerifiedCount)
	}
}
