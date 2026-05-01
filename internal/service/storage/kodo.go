package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/domain"
	"github.com/qiniu/go-sdk/v7/auth"
	qiniustorage "github.com/qiniu/go-sdk/v7/storage"
)

type KodoStorage struct {
	mac      *auth.Credentials
	uploader *qiniustorage.FormUploader
	manager  *qiniustorage.BucketManager
	bucket   string
	domain   string
	private  bool
}

func init() {
	Register(string(domain.StrategyTypeKodo), func(cfg json.RawMessage) (Storage, error) {
		return NewKodoStorage(cfg)
	})
	RegisterValidator(string(domain.StrategyTypeKodo), func(cfg json.RawMessage) error {
		var c domain.KodoStrategyConfig
		if err := json.Unmarshal(cfg, &c); err != nil {
			return err
		}
		if c.AccessKey == "" || c.SecretKey == "" || c.Bucket == "" || c.Domain == "" {
			return fmt.Errorf("access_key, secret_key, bucket, and domain are required for Kodo storage")
		}
		if c.Zone != "" {
			if _, ok := qiniustorage.GetRegionByID(qiniustorage.RegionID(c.Zone)); !ok {
				return fmt.Errorf("unknown Kodo zone: %s", c.Zone)
			}
		}
		return nil
	})
}

func NewKodoStorage(cfg json.RawMessage) (*KodoStorage, error) {
	var c domain.KodoStrategyConfig
	if err := json.Unmarshal(cfg, &c); err != nil {
		return nil, err
	}
	if err := ValidateConfig(string(domain.StrategyTypeKodo), cfg); err != nil {
		return nil, err
	}
	mac := auth.New(c.AccessKey, c.SecretKey)
	qcfg := qiniustorage.Config{UseHTTPS: true, UseCdnDomains: false}
	if c.Zone != "" {
		region, _ := qiniustorage.GetRegionByID(qiniustorage.RegionID(c.Zone))
		qcfg.Zone = &region
	}
	return &KodoStorage{
		mac:      mac,
		uploader: qiniustorage.NewFormUploader(&qcfg),
		manager:  qiniustorage.NewBucketManager(mac, &qcfg),
		bucket:   c.Bucket,
		domain:   c.Domain,
		private:  c.Private,
	}, nil
}

func (s *KodoStorage) Write(ctx context.Context, path string, data []byte, contentType string) error {
	putPolicy := qiniustorage.PutPolicy{Scope: s.bucket + ":" + path}
	token := putPolicy.UploadToken(s.mac)
	ret := qiniustorage.PutRet{}
	return s.uploader.Put(ctx, &ret, token, path, bytes.NewReader(data), int64(len(data)), &qiniustorage.PutExtra{
		MimeType: contentType,
	})
}

func (s *KodoStorage) Read(ctx context.Context, path string) ([]byte, error) {
	resp, err := s.manager.Get(s.bucket, path, &qiniustorage.GetObjectInput{
		Context:         ctx,
		DownloadDomains: []string{s.domain},
		PresignUrl:      s.private,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	return io.ReadAll(resp)
}

func (s *KodoStorage) Delete(ctx context.Context, path string) error {
	return s.manager.Delete(s.bucket, path)
}

func (s *KodoStorage) URL(pathname string) string {
	pathname = strings.TrimLeft(pathname, "/")
	if s.private {
		return qiniustorage.MakePrivateURLv2(s.mac, s.domain, pathname, time.Now().Add(time.Hour).Unix())
	}
	return qiniustorage.MakePublicURLv2(s.domain, pathname)
}

func (s *KodoStorage) HealthCheck(ctx context.Context) HealthResult {
	_, _, err := s.manager.ListFilesWithContext(ctx, s.bucket, qiniustorage.ListInputOptionsLimit(1))
	if err != nil {
		return HealthResult{Healthy: false, Error: "bucket unreachable: " + err.Error()}
	}
	return HealthResult{Healthy: true}
}

func (s *KodoStorage) Close() error { return nil }
