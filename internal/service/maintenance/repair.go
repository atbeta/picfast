package maintenance

import (
	"context"
	"time"
)

type ThumbnailGenerator func(data []byte, extension, thumbnailDir, md5Hash string) error

type RepairStatus string

const (
	RepairStatusSkippedExisting RepairStatus = "skipped_existing"
	RepairStatusSkippedType     RepairStatus = "skipped_type"
	RepairStatusDuplicate       RepairStatus = "duplicate"
	RepairStatusWouldRepair     RepairStatus = "would_repair"
	RepairStatusRepaired        RepairStatus = "repaired"
	RepairStatusMissingStrategy RepairStatus = "missing_strategy"
	RepairStatusInitFailed      RepairStatus = "storage_init_failed"
	RepairStatusReadFailed      RepairStatus = "read_failed"
	RepairStatusGenerateFailed  RepairStatus = "generate_failed"
)

type RepairOptions struct {
	Apply         bool
	ObjectTimeout time.Duration
}

type RepairResult struct {
	GeneratedAt time.Time          `json:"generated_at"`
	Apply       bool               `json:"apply"`
	Total       int64              `json:"total"`
	Repaired    int64              `json:"repaired"`
	WouldRepair int64              `json:"would_repair"`
	Skipped     int64              `json:"skipped"`
	Failed      int64              `json:"failed"`
	Items       []RepairItemResult `json:"items,omitempty"`
}

type RepairItemResult struct {
	ImageID int64        `json:"image_id"`
	Key     string       `json:"key"`
	MD5     string       `json:"md5,omitempty"`
	Path    string       `json:"path,omitempty"`
	Status  RepairStatus `json:"status"`
	Error   string       `json:"error,omitempty"`
}

type ThumbnailRepairer struct {
	Verifier  Verifier
	Generator ThumbnailGenerator
}

func (r ThumbnailRepairer) Repair(ctx context.Context, items []InventoryItem, opts RepairOptions) RepairResult {
	result := RepairResult{GeneratedAt: time.Now().UTC(), Apply: opts.Apply}
	handledMD5 := make(map[string]bool)

	for _, item := range items {
		result.Total++
		itemResult := RepairItemResult{
			ImageID: item.ImageID,
			Key:     item.Key,
			MD5:     item.MD5,
		}

		thumb := r.Verifier.VerifyThumbnail(item)
		itemResult.Path = thumb.Path
		if thumb.Status == ThumbnailStatusOK {
			itemResult.Status = RepairStatusSkippedExisting
			result.Skipped++
			result.Items = append(result.Items, itemResult)
			continue
		}
		if thumb.Status == ThumbnailStatusSkipped {
			itemResult.Status = RepairStatusSkippedType
			result.Skipped++
			result.Items = append(result.Items, itemResult)
			continue
		}
		if item.MD5 != "" && handledMD5[item.MD5] {
			itemResult.Status = RepairStatusDuplicate
			result.Skipped++
			result.Items = append(result.Items, itemResult)
			continue
		}
		if !opts.Apply {
			itemResult.Status = RepairStatusWouldRepair
			result.WouldRepair++
			if item.MD5 != "" {
				handledMD5[item.MD5] = true
			}
			result.Items = append(result.Items, itemResult)
			continue
		}

		data, status, errText := r.readObject(ctx, item, opts.ObjectTimeout)
		if status != "" {
			itemResult.Status = status
			itemResult.Error = errText
			result.Failed++
			result.Items = append(result.Items, itemResult)
			continue
		}
		if err := r.Generator(data, item.Extension, r.Verifier.ThumbnailDir, item.MD5); err != nil {
			itemResult.Status = RepairStatusGenerateFailed
			itemResult.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, itemResult)
			continue
		}
		itemResult.Status = RepairStatusRepaired
		result.Repaired++
		if item.MD5 != "" {
			handledMD5[item.MD5] = true
		}
		result.Items = append(result.Items, itemResult)
	}

	return result
}

func (r ThumbnailRepairer) readObject(ctx context.Context, item InventoryItem, timeout time.Duration) ([]byte, RepairStatus, string) {
	if item.StrategyID == nil || item.StrategyType == "" {
		return nil, RepairStatusMissingStrategy, ""
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	factory := r.Verifier.StorageFactory
	if factory == nil {
		factory = DefaultStorageFactory{}
	}
	store, err := factory.New(item.StrategyType, item.StrategyConfig)
	if err != nil {
		return nil, RepairStatusInitFailed, err.Error()
	}
	defer store.Close()

	data, err := store.Read(ctx, item.ObjectPath)
	if err != nil {
		return nil, RepairStatusReadFailed, err.Error()
	}
	return data, "", ""
}
