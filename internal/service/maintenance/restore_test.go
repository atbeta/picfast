package maintenance

import (
	"archive/tar"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractArchiveRejectsUnsafePath(t *testing.T) {
	dir := t.TempDir()
	archivePath := filepath.Join(dir, "backup.tar")
	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	tw := tar.NewWriter(file)
	if err := tw.WriteHeader(&tar.Header{Name: "../escape.txt", Mode: 0644, Size: int64(len("bad"))}); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte("bad")); err != nil {
		t.Fatalf("write payload: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}

	if err := ExtractArchive(archivePath, filepath.Join(dir, "out")); err == nil {
		t.Fatal("ExtractArchive() expected error")
	}
	if _, err := os.Stat(filepath.Join(dir, "escape.txt")); !os.IsNotExist(err) {
		t.Fatalf("unsafe file was written: %v", err)
	}
}

func TestReadObjectManifest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "objects.jsonl")
	if err := os.WriteFile(path, []byte(`{"image_id":1,"key":"abc","object_path":"2026/05/a.png"}`+"\n"), 0644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	objects, err := ReadObjectManifest(path)
	if err != nil {
		t.Fatalf("ReadObjectManifest() error = %v", err)
	}
	if len(objects) != 1 || objects[0].ImageID != 1 || objects[0].Key != "abc" {
		t.Fatalf("objects = %#v", objects)
	}
}
