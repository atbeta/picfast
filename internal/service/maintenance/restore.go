package maintenance

import (
	"archive/tar"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RestoreObjectStatus string

const (
	RestoreObjectStatusRestored        RestoreObjectStatus = "restored"
	RestoreObjectStatusSkippedExternal RestoreObjectStatus = "skipped_external"
	RestoreObjectStatusDuplicate       RestoreObjectStatus = "duplicate"
	RestoreObjectStatusMissingStrategy RestoreObjectStatus = "missing_strategy"
	RestoreObjectStatusReadFailed      RestoreObjectStatus = "read_failed"
	RestoreObjectStatusInitFailed      RestoreObjectStatus = "storage_init_failed"
	RestoreObjectStatusWriteFailed     RestoreObjectStatus = "write_failed"
)

type RestoreObjectsOptions struct {
	ObjectTimeout  time.Duration
	StorageFactory StorageFactory
}

type RestoreObjectsResult struct {
	GeneratedAt time.Time                 `json:"generated_at"`
	Total       int64                     `json:"total"`
	Restored    int64                     `json:"restored"`
	Skipped     int64                     `json:"skipped"`
	Failed      int64                     `json:"failed"`
	Items       []RestoreObjectItemResult `json:"items,omitempty"`
}

type RestoreObjectItemResult struct {
	ImageID int64               `json:"image_id"`
	Key     string              `json:"key"`
	Path    string              `json:"path,omitempty"`
	Status  RestoreObjectStatus `json:"status"`
	Error   string              `json:"error,omitempty"`
}

func ExtractArchive(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader, closeFn, err := archiveReader(archivePath, file)
	if err != nil {
		return err
	}
	if closeFn != nil {
		defer closeFn()
	}

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		target, err := safeArchivePath(destDir, header.Name)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(out, tr)
		closeErr := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
	}
	return nil
}

func ReadObjectManifest(path string) ([]ObjectManifestEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return parseObjectManifestJSONL(file)
}

func RestoreObjects(ctx context.Context, pool *pgxpool.Pool, extractedRoot string, objects []ObjectManifestEntry, opts RestoreObjectsOptions) (RestoreObjectsResult, error) {
	result := RestoreObjectsResult{GeneratedAt: time.Now().UTC()}
	written := make(map[string]bool)
	factory := opts.StorageFactory
	if factory == nil {
		factory = DefaultStorageFactory{}
	}
	strategyCache := make(map[int64]struct {
		typ string
		cfg []byte
	})

	for _, object := range objects {
		result.Total++
		item := RestoreObjectItemResult{
			ImageID: object.ImageID,
			Key:     object.Key,
			Path:    object.ObjectPath,
		}
		if object.ArchivePath == "" {
			item.Status = RestoreObjectStatusSkippedExternal
			result.Skipped++
			result.Items = append(result.Items, item)
			continue
		}
		if written[object.ArchivePath] {
			item.Status = RestoreObjectStatusDuplicate
			result.Skipped++
			result.Items = append(result.Items, item)
			continue
		}
		if object.StrategyID == nil {
			item.Status = RestoreObjectStatusMissingStrategy
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}

		strategy, ok := strategyCache[*object.StrategyID]
		if !ok {
			var cfg json.RawMessage
			if err := pool.QueryRow(ctx, `SELECT strategy_type, configs FROM strategies WHERE id = $1`, *object.StrategyID).Scan(&strategy.typ, &cfg); err != nil {
				item.Status = RestoreObjectStatusMissingStrategy
				item.Error = err.Error()
				result.Failed++
				result.Items = append(result.Items, item)
				continue
			}
			strategy.cfg = cfg
			strategyCache[*object.StrategyID] = strategy
		}

		archivePath, err := safeArchivePath(extractedRoot, object.ArchivePath)
		if err != nil {
			item.Status = RestoreObjectStatusReadFailed
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		data, err := os.ReadFile(archivePath)
		if err != nil {
			item.Status = RestoreObjectStatusReadFailed
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		writeCtx := ctx
		var cancel context.CancelFunc
		if opts.ObjectTimeout > 0 {
			writeCtx, cancel = context.WithTimeout(ctx, opts.ObjectTimeout)
		}
		store, err := factory.New(strategy.typ, strategy.cfg)
		if err != nil {
			if cancel != nil {
				cancel()
			}
			item.Status = RestoreObjectStatusInitFailed
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		err = store.Write(writeCtx, object.ObjectPath, data, object.MimeType)
		if cancel != nil {
			cancel()
		}
		closeErr := store.Close()
		if err != nil {
			item.Status = RestoreObjectStatusWriteFailed
			item.Error = err.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		if closeErr != nil {
			item.Status = RestoreObjectStatusWriteFailed
			item.Error = closeErr.Error()
			result.Failed++
			result.Items = append(result.Items, item)
			continue
		}
		written[object.ArchivePath] = true
		item.Status = RestoreObjectStatusRestored
		result.Restored++
		result.Items = append(result.Items, item)
	}
	return result, nil
}

func safeArchivePath(root, name string) (string, error) {
	clean := filepath.Clean(name)
	if clean == "." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || filepath.IsAbs(clean) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	target := filepath.Join(root, clean)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("unsafe archive path: %s", name)
	}
	return target, nil
}
