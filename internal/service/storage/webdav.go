package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/domain"
)

type WebDAVStorage struct {
	endpoint string
	username string
	password string
	url      string
	client   *http.Client
}

func init() {
	Register(string(domain.StrategyTypeWebDAV), func(cfg json.RawMessage) (Storage, error) {
		return NewWebDAVStorage(cfg)
	})
	RegisterValidator(string(domain.StrategyTypeWebDAV), func(cfg json.RawMessage) error {
		var c domain.WebDAVStrategyConfig
		if err := json.Unmarshal(cfg, &c); err != nil {
			return err
		}
		if c.Endpoint == "" || c.Username == "" || c.Password == "" {
			return fmt.Errorf("endpoint, username, and password are required for WebDAV storage")
		}
		return nil
	})
}

func NewWebDAVStorage(cfg json.RawMessage) (*WebDAVStorage, error) {
	var c domain.WebDAVStrategyConfig
	if err := json.Unmarshal(cfg, &c); err != nil {
		return nil, err
	}
	if err := ValidateConfig(string(domain.StrategyTypeWebDAV), cfg); err != nil {
		return nil, err
	}
	return &WebDAVStorage{
		endpoint: strings.TrimRight(c.Endpoint, "/"),
		username: c.Username,
		password: c.Password,
		url:      c.URL,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (s *WebDAVStorage) Write(ctx context.Context, path string, data []byte, contentType string) error {
	if err := s.ensureDirs(ctx, path); err != nil {
		return err
	}
	req, err := s.newRequest(ctx, http.MethodPut, s.objectURL(path), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webdav put failed: %s", resp.Status)
	}
	return nil
}

func (s *WebDAVStorage) Read(ctx context.Context, path string) ([]byte, error) {
	req, err := s.newRequest(ctx, http.MethodGet, s.objectURL(path), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("webdav get failed: %s", resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func (s *WebDAVStorage) Delete(ctx context.Context, path string) error {
	req, err := s.newRequest(ctx, http.MethodDelete, s.objectURL(path), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webdav delete failed: %s", resp.Status)
	}
	return nil
}

func (s *WebDAVStorage) URL(pathname string) string {
	if s.url != "" {
		return joinPublicURL(s.url, pathname)
	}
	return joinPublicURL(s.endpoint, pathname)
}

func (s *WebDAVStorage) HealthCheck(ctx context.Context) HealthResult {
	req, err := s.newRequest(ctx, "PROPFIND", s.endpoint, nil)
	if err != nil {
		return HealthResult{Healthy: false, Error: err.Error()}
	}
	req.Header.Set("Depth", "0")
	resp, err := s.client.Do(req)
	if err != nil {
		return HealthResult{Healthy: false, Error: err.Error()}
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return HealthResult{Healthy: false, Error: "endpoint unreachable: " + resp.Status}
	}
	return HealthResult{Healthy: true}
}

func (s *WebDAVStorage) Close() error { return nil }

func (s *WebDAVStorage) ensureDirs(ctx context.Context, objectPath string) error {
	parts := strings.Split(strings.Trim(objectPath, "/"), "/")
	if len(parts) <= 1 {
		return nil
	}
	current := ""
	for _, part := range parts[:len(parts)-1] {
		if part == "" {
			continue
		}
		current = strings.Trim(current+"/"+part, "/")
		req, err := s.newRequest(ctx, "MKCOL", s.objectURL(current), nil)
		if err != nil {
			return err
		}
		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}
		resp.Body.Close()
		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusMethodNotAllowed {
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("webdav mkcol failed: %s", resp.Status)
		}
	}
	return nil
}

func (s *WebDAVStorage) objectURL(pathname string) string {
	return joinPublicURL(s.endpoint, pathname)
}

func (s *WebDAVStorage) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.username, s.password)
	return req, nil
}
