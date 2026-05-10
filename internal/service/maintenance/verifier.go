package maintenance

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/atbeta/picfast/internal/service/storage"
)

type ObjectStatus string

const (
	ObjectStatusOK              ObjectStatus = "ok"
	ObjectStatusMissingStrategy ObjectStatus = "missing_strategy"
	ObjectStatusInitFailed      ObjectStatus = "storage_init_failed"
	ObjectStatusReadFailed      ObjectStatus = "read_failed"
	ObjectStatusSizeMismatch    ObjectStatus = "size_mismatch"
	ObjectStatusMD5Mismatch     ObjectStatus = "md5_mismatch"
	ObjectStatusSHA1Mismatch    ObjectStatus = "sha1_mismatch"
)

type ThumbnailStatus string

const (
	ThumbnailStatusOK         ThumbnailStatus = "ok"
	ThumbnailStatusSkipped    ThumbnailStatus = "skipped"
	ThumbnailStatusMissing    ThumbnailStatus = "missing"
	ThumbnailStatusUnreadable ThumbnailStatus = "unreadable"
)

type ObjectCheck struct {
	ImageID  int64        `json:"image_id"`
	Key      string       `json:"key"`
	Path     string       `json:"path"`
	Status   ObjectStatus `json:"status"`
	Expected int64        `json:"expected_size,omitempty"`
	Actual   int64        `json:"actual_size,omitempty"`
	Error    string       `json:"error,omitempty"`
}

type ThumbnailCheck struct {
	ImageID int64           `json:"image_id"`
	Key     string          `json:"key"`
	Path    string          `json:"path,omitempty"`
	Status  ThumbnailStatus `json:"status"`
	Error   string          `json:"error,omitempty"`
}

type StorageFactory interface {
	New(typ string, cfg []byte) (storage.Storage, error)
}

type DefaultStorageFactory struct{}

func (DefaultStorageFactory) New(typ string, cfg []byte) (storage.Storage, error) {
	return storage.New(typ, cfg)
}

type Verifier struct {
	StorageFactory StorageFactory
	ThumbnailDir   string
	ObjectTimeout  time.Duration
}

func (v Verifier) VerifyObject(ctx context.Context, item InventoryItem) ObjectCheck {
	check := ObjectCheck{
		ImageID:  item.ImageID,
		Key:      item.Key,
		Path:     item.ObjectPath,
		Expected: item.SizeBytes,
	}
	if item.StrategyID == nil || item.StrategyType == "" {
		check.Status = ObjectStatusMissingStrategy
		return check
	}
	if v.ObjectTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, v.ObjectTimeout)
		defer cancel()
	}

	factory := v.StorageFactory
	if factory == nil {
		factory = DefaultStorageFactory{}
	}
	store, err := factory.New(item.StrategyType, item.StrategyConfig)
	if err != nil {
		check.Status = ObjectStatusInitFailed
		check.Error = err.Error()
		return check
	}
	defer store.Close()

	data, err := store.Read(ctx, item.ObjectPath)
	if err != nil {
		check.Status = ObjectStatusReadFailed
		check.Error = err.Error()
		return check
	}
	check.Actual = int64(len(data))
	if check.Actual != item.SizeBytes {
		check.Status = ObjectStatusSizeMismatch
		return check
	}
	md5Sum := md5.Sum(data)
	if hex.EncodeToString(md5Sum[:]) != item.MD5 {
		check.Status = ObjectStatusMD5Mismatch
		return check
	}
	sha1Sum := sha1.Sum(data)
	if hex.EncodeToString(sha1Sum[:]) != item.SHA1 {
		check.Status = ObjectStatusSHA1Mismatch
		return check
	}
	check.Status = ObjectStatusOK
	return check
}

func (v Verifier) VerifyThumbnail(item InventoryItem) ThumbnailCheck {
	name := item.ThumbnailName()
	check := ThumbnailCheck{
		ImageID: item.ImageID,
		Key:     item.Key,
	}
	if name == "" {
		check.Status = ThumbnailStatusSkipped
		return check
	}
	check.Path = filepath.Join(v.ThumbnailDir, name)
	if _, err := os.Stat(check.Path); err != nil {
		if os.IsNotExist(err) {
			check.Status = ThumbnailStatusMissing
			return check
		}
		check.Status = ThumbnailStatusUnreadable
		check.Error = err.Error()
		return check
	}
	check.Status = ThumbnailStatusOK
	return check
}

type Report struct {
	GeneratedAt time.Time       `json:"generated_at"`
	Objects     ReportCounts    `json:"objects"`
	Thumbnails  ReportCounts    `json:"thumbnails"`
	Findings    []ReportFinding `json:"findings,omitempty"`
}

type ReportCounts struct {
	Total   int64 `json:"total"`
	OK      int64 `json:"ok"`
	Skipped int64 `json:"skipped,omitempty"`
	Failed  int64 `json:"failed"`
}

type ReportFinding struct {
	ImageID int64  `json:"image_id"`
	Key     string `json:"key"`
	Kind    string `json:"kind"`
	Status  string `json:"status"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewReport() Report {
	return Report{GeneratedAt: time.Now().UTC()}
}

func (r *Report) AddObjectCheck(check ObjectCheck) {
	r.Objects.Total++
	if check.Status == ObjectStatusOK {
		r.Objects.OK++
		return
	}
	r.Objects.Failed++
	r.Findings = append(r.Findings, ReportFinding{
		ImageID: check.ImageID,
		Key:     check.Key,
		Kind:    "object",
		Status:  string(check.Status),
		Path:    check.Path,
		Error:   check.Error,
	})
}

func (r *Report) AddThumbnailCheck(check ThumbnailCheck) {
	r.Thumbnails.Total++
	if check.Status == ThumbnailStatusOK {
		r.Thumbnails.OK++
		return
	}
	if check.Status == ThumbnailStatusSkipped {
		r.Thumbnails.Skipped++
		return
	}
	r.Thumbnails.Failed++
	r.Findings = append(r.Findings, ReportFinding{
		ImageID: check.ImageID,
		Key:     check.Key,
		Kind:    "thumbnail",
		Status:  string(check.Status),
		Path:    check.Path,
		Error:   check.Error,
	})
}

func VerifyBatch(ctx context.Context, verifier Verifier, items []InventoryItem) (Report, error) {
	report := NewReport()
	for _, item := range items {
		select {
		case <-ctx.Done():
			return report, fmt.Errorf("verify batch: %w", ctx.Err())
		default:
		}
		report.AddObjectCheck(verifier.VerifyObject(ctx, item))
		report.AddThumbnailCheck(verifier.VerifyThumbnail(item))
	}
	return report, nil
}
