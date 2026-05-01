package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/atbeta/picfast/internal/domain"
)

type OSSStorage struct {
	bucket     *oss.Bucket
	bucketName string
	endpoint   string
	url        string
}

func init() {
	Register(string(domain.StrategyTypeOSS), func(cfg json.RawMessage) (Storage, error) {
		return NewOSSStorage(cfg)
	})
	RegisterValidator(string(domain.StrategyTypeOSS), func(cfg json.RawMessage) error {
		var c domain.OSSStrategyConfig
		if err := json.Unmarshal(cfg, &c); err != nil {
			return err
		}
		if c.Endpoint == "" || c.Bucket == "" || c.AccessKey == "" || c.SecretKey == "" {
			return fmt.Errorf("endpoint, bucket, access_key, and secret_key are required for OSS storage")
		}
		return nil
	})
}

func NewOSSStorage(cfg json.RawMessage) (*OSSStorage, error) {
	var c domain.OSSStrategyConfig
	if err := json.Unmarshal(cfg, &c); err != nil {
		return nil, err
	}
	if err := ValidateConfig(string(domain.StrategyTypeOSS), cfg); err != nil {
		return nil, err
	}
	client, err := oss.New(c.Endpoint, c.AccessKey, c.SecretKey)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(c.Bucket)
	if err != nil {
		return nil, err
	}
	return &OSSStorage{
		bucket:     bucket,
		bucketName: c.Bucket,
		endpoint:   c.Endpoint,
		url:        c.URL,
	}, nil
}

func (s *OSSStorage) Write(ctx context.Context, path string, data []byte, contentType string) error {
	return s.bucket.PutObject(path, bytes.NewReader(data), oss.ContentType(contentType))
}

func (s *OSSStorage) Read(ctx context.Context, path string) ([]byte, error) {
	body, err := s.bucket.GetObject(path)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return io.ReadAll(body)
}

func (s *OSSStorage) Delete(ctx context.Context, path string) error {
	return s.bucket.DeleteObject(path)
}

func (s *OSSStorage) URL(pathname string) string {
	if s.url != "" {
		return joinPublicURL(s.url, pathname)
	}
	u, err := url.Parse(s.endpoint)
	if err != nil || u.Host == "" {
		return joinPublicURL(s.endpoint, pathname)
	}
	return joinPublicURL(u.Scheme+"://"+s.bucketName+"."+u.Host, strings.TrimLeft(pathname, "/"))
}

func (s *OSSStorage) HealthCheck(ctx context.Context) HealthResult {
	_, err := s.bucket.ListObjects(oss.MaxKeys(1))
	if err != nil {
		return HealthResult{Healthy: false, Error: "bucket unreachable: " + err.Error()}
	}
	return HealthResult{Healthy: true}
}

func (s *OSSStorage) Close() error { return nil }
