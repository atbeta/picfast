package storage

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/atbeta/picfast/internal/domain"
)

type S3Storage struct {
	client *s3.Client
	bucket string
	url    string
}

func NewS3Storage(cfg domain.S3StrategyConfig) (*S3Storage, error) {
	creds := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")

	region := cfg.Region
	if region == "" {
		region = "auto"
	}

	client := s3.New(s3.Options{
		Region:               region,
		Credentials:          creds,
		BaseEndpoint:         aws.String(cfg.Endpoint),
		UsePathStyle:         true,
		AuthSchemePreference: []string{"sigv4"},
	})

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
		url:    cfg.URL,
	}, nil
}

func (s *S3Storage) Write(ctx context.Context, path string, data []byte) error {
	slog.Info("s3 write starting", "bucket", s.bucket, "key", path, "size", len(data))
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(path),
		Body:          bytes.NewReader(data),
		ContentLength: aws.Int64(int64(len(data))),
		ContentType:   aws.String("application/octet-stream"),
	})
	if err != nil {
		slog.Error("s3 write failed", "bucket", s.bucket, "key", path, "error", err)
		return err
	}
	slog.Info("s3 write completed", "bucket", s.bucket, "key", path)
	return nil
}

func (s *S3Storage) Read(ctx context.Context, path string) ([]byte, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		slog.Error("s3 read failed", "bucket", s.bucket, "key", path, "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	n, readErr := buf.ReadFrom(resp.Body)
	if readErr != nil {
		slog.Error("s3 read body failed", "bucket", s.bucket, "key", path, "error", readErr)
		return nil, readErr
	}
	slog.Info("s3 read completed", "bucket", s.bucket, "key", path, "bytes", n)
	return buf.Bytes(), nil
}

func (s *S3Storage) Delete(ctx context.Context, path string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(path),
	})
	return err
}

func (s *S3Storage) URL(pathname string) string {
	if s.url != "" {
		return s.url + "/" + pathname
	}
	return fmt.Sprintf("https://%s.s3.%s/%s", s.bucket, "endpoint", pathname)
}

func (s *S3Storage) HealthCheck(ctx context.Context) HealthResult {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		return HealthResult{Healthy: false, Error: "bucket unreachable: " + err.Error()}
	}
	return HealthResult{Healthy: true}
}
