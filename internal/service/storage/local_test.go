package storage

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLocalStorageSafePath(t *testing.T) {
	root := t.TempDir()
	s := &LocalStorage{root: root, url: "http://localhost/"}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{name: "normal_relative_path", path: "images/photo.jpg", wantErr: false},
		{name: "single_parent_escape", path: "../etc/passwd", wantErr: true},
		{name: "double_parent_escape", path: "../../root/.ssh/id_rsa", wantErr: true},
		{name: "parent_in_middle", path: "images/../etc/passwd", wantErr: true},
		{name: "encoded_parent_segment", path: "images/%2e%2e/etc/passwd", wantErr: false}, // literal segment; no ".." substring
		{name: "absolute_segment_under_root", path: "/etc/passwd", wantErr: false},       // Join(root, "/etc/passwd") stays under root
		{name: "subdir_with_parent", path: "subdir/../other/file.jpg", wantErr: true},
		{name: "dotfile_allowed", path: ".well-known/test", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.safePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("safePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}

	t.Run("write rejects parent_escape", func(t *testing.T) {
		ctx := context.Background()
		if err := s.Write(ctx, "../outside.txt", []byte("x"), "text/plain"); err == nil {
			t.Error("Write with .. succeeded, want error")
		}
	})

	if runtime.GOOS != "windows" {
		t.Run("symlink_escape_not_blocked", func(t *testing.T) {
			outside := t.TempDir()
			target := filepath.Join(outside, "secret.txt")
			if err := os.WriteFile(target, []byte("secret"), 0o644); err != nil {
				t.Fatal(err)
			}
			linkPath := filepath.Join(root, "escape-link")
			if err := os.Symlink(target, linkPath); err != nil {
				t.Fatal(err)
			}
			got, err := s.Read(context.Background(), "escape-link")
			if err != nil {
				t.Fatalf("Read: %v", err)
			}
			if string(got) != "secret" {
				t.Errorf("Read = %q, want secret (documents symlink follow; harden safePath if undesired)", got)
			}
		})
	}
}

func TestLocalStorageCRUD(t *testing.T) {
	root := t.TempDir()
	s := &LocalStorage{root: root, url: "http://localhost/"}
	ctx := context.Background()

	t.Run("write read nested", func(t *testing.T) {
		data := []byte("nested content")
		if err := s.Write(ctx, "a/b/c/nested.txt", data, "text/plain"); err != nil {
			t.Fatalf("Write: %v", err)
		}
		got, err := s.Read(ctx, "a/b/c/nested.txt")
		if err != nil {
			t.Fatalf("Read: %v", err)
		}
		if string(got) != "nested content" {
			t.Errorf("Read = %q, want nested content", got)
		}
	})

	t.Run("write read delete", func(t *testing.T) {
		if err := s.Write(ctx, "to-delete.txt", []byte("delete me"), "text/plain"); err != nil {
			t.Fatalf("Write: %v", err)
		}
		if err := s.Delete(ctx, "to-delete.txt"); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := s.Read(ctx, "to-delete.txt"); err == nil {
			t.Error("Read after delete succeeded, want error")
		}
	})

	t.Run("read nonexistent", func(t *testing.T) {
		if _, err := s.Read(ctx, "nonexistent/file.txt"); err == nil {
			t.Error("Read nonexistent succeeded, want error")
		}
	})

	t.Run("delete nonexistent", func(t *testing.T) {
		if err := s.Delete(ctx, "nonexistent/file.txt"); err == nil {
			t.Error("Delete nonexistent succeeded, want error")
		}
	})
}

func TestLocalStorageURL(t *testing.T) {
	tests := []struct {
		root    string
		baseURL string
		path    string
		want    string
	}{
		{"/var/data", "http://cdn.example.com/", "images/photo.jpg", "http://cdn.example.com/images/photo.jpg"},
		{"/var/data", "http://cdn.example.com", "images/photo.jpg", "http://cdn.example.com/images/photo.jpg"},
		{"/var/data", "https://cdn.example.com/", "/images/photo.jpg", "https://cdn.example.com/images/photo.jpg"},
	}

	for _, tt := range tests {
		s := &LocalStorage{root: tt.root, url: tt.baseURL}
		if got := s.URL(tt.path); got != tt.want {
			t.Errorf("URL(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestLocalStorageHealthCheck(t *testing.T) {
	t.Run("writable root", func(t *testing.T) {
		root := t.TempDir()
		s := &LocalStorage{root: root, url: "http://localhost/"}
		result := s.HealthCheck(context.Background())
		if !result.Healthy {
			t.Errorf("healthy = false, want true: %s", result.Error)
		}
	})

	t.Run("readonly root", func(t *testing.T) {
		root := t.TempDir()
		os.Chmod(root, 0555)
		defer os.Chmod(root, 0755)

		s := &LocalStorage{root: root, url: "http://localhost/"}
		result := s.HealthCheck(context.Background())
		if result.Healthy {
			t.Error("healthy = true, want false for readonly root")
		}
		if result.Error == "" {
			t.Error("Error empty, want message")
		}
	})
}

func TestLocalStorageClose(t *testing.T) {
	s := &LocalStorage{root: t.TempDir(), url: "http://localhost/"}
	if err := s.Close(); err != nil {
		t.Errorf("Close error = %v, want nil", err)
	}
}
