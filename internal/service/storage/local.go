package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	root string
	url  string
}

func NewLocalStorage(root, url string) *LocalStorage {
	return &LocalStorage{root: root, url: url}
}

func (s *LocalStorage) Write(ctx context.Context, path string, data []byte) error {
	fullPath := filepath.Join(s.root, path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *LocalStorage) Read(ctx context.Context, path string) ([]byte, error) {
	fullPath := filepath.Join(s.root, path)
	return os.ReadFile(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(s.root, path)
	return os.Remove(fullPath)
}

func (s *LocalStorage) URL(pathname string) string {
	return strings.TrimRight(s.url, "/") + "/" + strings.TrimLeft(pathname, "/")
}
