package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atbeta/picfast/internal/domain"
)

type LocalStorage struct {
	root string
	url  string
}

func init() {
	Register(string(domain.StrategyTypeLocal), func(cfg json.RawMessage) (Storage, error) {
		return NewLocalStorage(cfg)
	})
	RegisterValidator(string(domain.StrategyTypeLocal), func(cfg json.RawMessage) error {
		var c domain.LocalStrategyConfig
		if err := json.Unmarshal(cfg, &c); err != nil {
			return err
		}
		if c.Root == "" || c.URL == "" {
			return fmt.Errorf("root and url are required for local storage")
		}
		return nil
	})
}

func NewLocalStorage(cfg json.RawMessage) (*LocalStorage, error) {
	var c domain.LocalStrategyConfig
	if err := json.Unmarshal(cfg, &c); err != nil {
		return nil, err
	}
	return &LocalStorage{root: c.Root, url: c.URL}, nil
}

func (s *LocalStorage) safePath(path string) (string, error) {
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("invalid path: contains '..'")
	}
	fullPath := filepath.Join(s.root, path)
	// Ensure the resolved path stays within root
	absRoot, err := filepath.Abs(s.root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if !strings.HasPrefix(absPath, absRoot+string(filepath.Separator)) && absPath != absRoot {
		return "", fmt.Errorf("invalid path: escapes root directory")
	}
	return fullPath, nil
}

func (s *LocalStorage) Write(ctx context.Context, path string, data []byte) error {
	fullPath, err := s.safePath(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}
	return os.WriteFile(fullPath, data, 0644)
}

func (s *LocalStorage) Read(ctx context.Context, path string) ([]byte, error) {
	fullPath, err := s.safePath(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(fullPath)
}

func (s *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath, err := s.safePath(path)
	if err != nil {
		return err
	}
	return os.Remove(fullPath)
}

func (s *LocalStorage) URL(pathname string) string {
	return strings.TrimRight(s.url, "/") + "/" + strings.TrimLeft(pathname, "/")
}

func (s *LocalStorage) Close() error { return nil }

func (s *LocalStorage) HealthCheck(ctx context.Context) HealthResult {
	// Try to create and remove a temporary file
	testPath := filepath.Join(s.root, ".healthcheck")
	if err := os.WriteFile(testPath, []byte("ok"), 0644); err != nil {
		return HealthResult{Healthy: false, Error: "not writable: " + err.Error()}
	}
	os.Remove(testPath)
	return HealthResult{Healthy: true}
}
