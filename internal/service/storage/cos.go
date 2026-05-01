package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type COSStorage struct {
	client    *cos.Client
	bucketURL string
	url       string
}

func init() {
	Register(string(domain.StrategyTypeCOS), func(cfg json.RawMessage) (Storage, error) {
		return NewCOSStorage(cfg)
	})
	RegisterValidator(string(domain.StrategyTypeCOS), func(cfg json.RawMessage) error {
		var c domain.COSStrategyConfig
		if err := json.Unmarshal(cfg, &c); err != nil {
			return err
		}
		if c.BucketURL == "" || c.SecretID == "" || c.SecretKey == "" {
			return fmt.Errorf("bucket_url, secret_id, and secret_key are required for COS storage")
		}
		if _, err := url.ParseRequestURI(c.BucketURL); err != nil {
			return fmt.Errorf("bucket_url is invalid: %w", err)
		}
		return nil
	})
}

func NewCOSStorage(cfg json.RawMessage) (*COSStorage, error) {
	var c domain.COSStrategyConfig
	if err := json.Unmarshal(cfg, &c); err != nil {
		return nil, err
	}
	if err := ValidateConfig(string(domain.StrategyTypeCOS), cfg); err != nil {
		return nil, err
	}
	bucketURL, err := url.Parse(c.BucketURL)
	if err != nil {
		return nil, err
	}
	client := cos.NewClient(&cos.BaseURL{BucketURL: bucketURL}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  c.SecretID,
			SecretKey: c.SecretKey,
		},
	})
	return &COSStorage{client: client, bucketURL: c.BucketURL, url: c.URL}, nil
}

func (s *COSStorage) Write(ctx context.Context, path string, data []byte, contentType string) error {
	_, err := s.client.Object.Put(ctx, path, bytes.NewReader(data), &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{ContentType: contentType},
	})
	return err
}

func (s *COSStorage) Read(ctx context.Context, path string) ([]byte, error) {
	resp, err := s.client.Object.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (s *COSStorage) Delete(ctx context.Context, path string) error {
	_, err := s.client.Object.Delete(ctx, path)
	return err
}

func (s *COSStorage) URL(pathname string) string {
	if s.url != "" {
		return joinPublicURL(s.url, pathname)
	}
	return joinPublicURL(s.bucketURL, pathname)
}

func (s *COSStorage) HealthCheck(ctx context.Context) HealthResult {
	_, err := s.client.Bucket.Head(ctx)
	if err != nil {
		return HealthResult{Healthy: false, Error: "bucket unreachable: " + err.Error()}
	}
	return HealthResult{Healthy: true}
}

func (s *COSStorage) Close() error { return nil }
