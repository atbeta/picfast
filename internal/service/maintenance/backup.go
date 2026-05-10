package maintenance

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ObjectManifestEntry struct {
	ImageID      int64     `json:"image_id"`
	Key          string    `json:"key"`
	StrategyID   *int64    `json:"strategy_id,omitempty"`
	StrategyName string    `json:"strategy_name,omitempty"`
	StrategyType string    `json:"strategy_type,omitempty"`
	ObjectPath   string    `json:"object_path"`
	ArchivePath  string    `json:"archive_path,omitempty"`
	SizeBytes    int64     `json:"size_bytes"`
	MimeType     string    `json:"mimetype"`
	Extension    string    `json:"extension"`
	MD5          string    `json:"md5"`
	SHA1         string    `json:"sha1"`
	CreatedAt    time.Time `json:"created_at"`
}

type ChecksumEntry struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type BackupOptions struct {
	OutputPath          string
	DatabaseDumpPath    string
	AppVersion          string
	MigrationVersion    int
	Features            []string
	IncludeObjects      bool
	AllowMissingObjects bool
	ObjectTimeout       time.Duration
}

type BackupResult struct {
	OutputPath string   `json:"output_path"`
	Manifest   Manifest `json:"manifest"`
	Warnings   []string `json:"warnings,omitempty"`
}

type BackupWriter struct {
	StorageFactory StorageFactory
}

func (w BackupWriter) Write(ctx context.Context, items []InventoryItem, opts BackupOptions) (BackupResult, error) {
	if opts.OutputPath == "" {
		return BackupResult{}, fmt.Errorf("output path is required")
	}
	if opts.DatabaseDumpPath == "" {
		return BackupResult{}, fmt.Errorf("database dump path is required")
	}

	tmpDir, err := os.MkdirTemp("", "picfast-backup-*")
	if err != nil {
		return BackupResult{}, err
	}
	defer os.RemoveAll(tmpDir)

	manifest := NewManifest(opts.AppVersion, opts.MigrationVersion)
	manifest.Features = opts.Features
	manifest.Generator = "picfast maintenance backup"
	if !opts.IncludeObjects {
		manifest.Objects.Mode = "database_only"
		manifest.Objects.ManifestPath = ""
		manifest.Objects.ChecksumPath = "checksums.jsonl"
	}

	checksums := make([]ChecksumEntry, 0, len(items)+1)
	dbChecksum, err := copyFileWithChecksum(filepath.Join(tmpDir, manifest.Database.Path), opts.DatabaseDumpPath)
	if err != nil {
		return BackupResult{}, fmt.Errorf("copy database dump: %w", err)
	}
	checksums = append(checksums, dbChecksum)

	var warnings []string
	if opts.IncludeObjects {
		objectsDir := filepath.Join(tmpDir, "objects")
		if err := os.MkdirAll(objectsDir, 0755); err != nil {
			return BackupResult{}, err
		}
		objectManifest, objectChecksums, objectBytes, objectWarnings, err := w.writeObjects(ctx, objectsDir, items, opts)
		if err != nil {
			return BackupResult{}, err
		}
		warnings = append(warnings, objectWarnings...)
		manifest.Objects.Count = int64(len(objectChecksums))
		manifest.Objects.Bytes = objectBytes
		checksums = append(checksums, objectChecksums...)
		if err := writeJSONL(filepath.Join(tmpDir, manifest.Objects.ManifestPath), objectManifest); err != nil {
			return BackupResult{}, err
		}
	}

	if err := writeJSONL(filepath.Join(tmpDir, manifest.Objects.ChecksumPath), checksums); err != nil {
		return BackupResult{}, err
	}
	if err := manifest.Validate(); err != nil {
		return BackupResult{}, err
	}
	if err := writeJSON(filepath.Join(tmpDir, "manifest.json"), manifest); err != nil {
		return BackupResult{}, err
	}
	if err := writeTarArchive(opts.OutputPath, tmpDir); err != nil {
		return BackupResult{}, err
	}

	return BackupResult{OutputPath: opts.OutputPath, Manifest: manifest, Warnings: warnings}, nil
}

func (w BackupWriter) writeObjects(ctx context.Context, objectsDir string, items []InventoryItem, opts BackupOptions) ([]ObjectManifestEntry, []ChecksumEntry, int64, []string, error) {
	factory := w.StorageFactory
	if factory == nil {
		factory = DefaultStorageFactory{}
	}
	manifest := make([]ObjectManifestEntry, 0, len(items))
	checksums := make([]ChecksumEntry, 0, len(items))
	var totalBytes int64
	var warnings []string
	writtenMD5 := make(map[string]string)

	for _, item := range items {
		entry := ObjectManifestEntry{
			ImageID:      item.ImageID,
			Key:          item.Key,
			StrategyID:   item.StrategyID,
			StrategyName: item.StrategyName,
			StrategyType: item.StrategyType,
			ObjectPath:   item.ObjectPath,
			SizeBytes:    item.SizeBytes,
			MimeType:     item.MimeType,
			Extension:    item.Extension,
			MD5:          item.MD5,
			SHA1:         item.SHA1,
			CreatedAt:    item.CreatedAt,
		}
		if item.MD5 != "" {
			if archivePath, ok := writtenMD5[item.MD5]; ok {
				entry.ArchivePath = archivePath
				manifest = append(manifest, entry)
				continue
			}
		}
		if item.StrategyID == nil || item.StrategyType == "" {
			warning := fmt.Sprintf("image %d has no storage strategy", item.ImageID)
			if !opts.AllowMissingObjects {
				return nil, nil, 0, warnings, fmt.Errorf("%s", warning)
			}
			warnings = append(warnings, warning)
			manifest = append(manifest, entry)
			continue
		}
		readCtx := ctx
		if opts.ObjectTimeout > 0 {
			var cancel context.CancelFunc
			readCtx, cancel = context.WithTimeout(ctx, opts.ObjectTimeout)
			defer cancel()
		}
		store, err := factory.New(item.StrategyType, item.StrategyConfig)
		if err != nil {
			warning := fmt.Sprintf("image %d storage init failed: %v", item.ImageID, err)
			if !opts.AllowMissingObjects {
				return nil, nil, 0, warnings, fmt.Errorf("%s", warning)
			}
			warnings = append(warnings, warning)
			manifest = append(manifest, entry)
			continue
		}
		data, err := store.Read(readCtx, item.ObjectPath)
		closeErr := store.Close()
		if err != nil {
			warning := fmt.Sprintf("image %d object read failed: %v", item.ImageID, err)
			if !opts.AllowMissingObjects {
				return nil, nil, 0, warnings, fmt.Errorf("%s", warning)
			}
			warnings = append(warnings, warning)
			manifest = append(manifest, entry)
			continue
		}
		if closeErr != nil {
			warnings = append(warnings, fmt.Sprintf("image %d storage close failed: %v", item.ImageID, closeErr))
		}
		archivePath := filepath.ToSlash(filepath.Join("objects", fmt.Sprintf("%d-%s-%s", item.ImageID, item.Key, filepath.Base(item.ObjectPath))))
		targetPath := filepath.Join(objectsDir, filepath.Base(archivePath))
		checksum, err := writeBytesWithChecksum(targetPath, archivePath, data)
		if err != nil {
			return nil, nil, 0, warnings, err
		}
		totalBytes += int64(len(data))
		entry.ArchivePath = archivePath
		if item.MD5 != "" {
			writtenMD5[item.MD5] = archivePath
		}
		manifest = append(manifest, entry)
		checksums = append(checksums, checksum)
	}
	return manifest, checksums, totalBytes, warnings, nil
}

func copyFileWithChecksum(dst, src string) (ChecksumEntry, error) {
	in, err := os.Open(src)
	if err != nil {
		return ChecksumEntry{}, err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return ChecksumEntry{}, err
	}
	defer out.Close()
	return copyWithChecksum(out, filepath.Base(dst), in)
}

func writeBytesWithChecksum(dst, archivePath string, data []byte) (ChecksumEntry, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return ChecksumEntry{}, err
	}
	out, err := os.Create(dst)
	if err != nil {
		return ChecksumEntry{}, err
	}
	defer out.Close()
	return copyWithChecksum(out, archivePath, bytes.NewReader(data))
}

func copyWithChecksum(dst io.Writer, archivePath string, src io.Reader) (ChecksumEntry, error) {
	hash := sha256.New()
	n, err := io.Copy(io.MultiWriter(dst, hash), src)
	if err != nil {
		return ChecksumEntry{}, err
	}
	return ChecksumEntry{Path: archivePath, SizeBytes: n, SHA256: hex.EncodeToString(hash.Sum(nil))}, nil
}

func writeJSON(path string, value any) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	return enc.Encode(value)
}

func writeJSONL[T any](path string, values []T) error {
	if path == "" {
		return nil
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	enc := json.NewEncoder(file)
	for _, value := range values {
		if err := enc.Encode(value); err != nil {
			return err
		}
	}
	return nil
}

func writeTarArchive(outputPath, root string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil && filepath.Dir(outputPath) != "." {
		return err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	var writer io.WriteCloser = file
	if strings.HasSuffix(outputPath, ".gz") || strings.HasSuffix(outputPath, ".tgz") {
		gz := gzip.NewWriter(file)
		defer gz.Close()
		writer = gz
	}
	tw := tar.NewWriter(writer)
	defer tw.Close()

	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tw, file)
		return err
	})
}
