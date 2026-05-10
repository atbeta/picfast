package maintenance

import (
	"errors"
	"time"
)

const (
	BackupFormat        = "picfast-backup"
	BackupFormatVersion = 1
)

type Manifest struct {
	Format            string          `json:"format"`
	FormatVersion     int             `json:"format_version"`
	AppVersion        string          `json:"app_version"`
	MigrationVersion  int             `json:"migration_version"`
	CreatedAt         time.Time       `json:"created_at"`
	Features          []string        `json:"features,omitempty"`
	Database          DatabaseSection `json:"database"`
	Objects           ObjectSection   `json:"objects"`
	Config            ConfigSection   `json:"config"`
	Generator         string          `json:"generator,omitempty"`
	MinimumAppVersion string          `json:"minimum_app_version,omitempty"`
}

type DatabaseSection struct {
	Mode string `json:"mode"`
	Path string `json:"path"`
}

type ObjectSection struct {
	Mode         string `json:"mode"`
	ManifestPath string `json:"manifest_path,omitempty"`
	ChecksumPath string `json:"checksum_path,omitempty"`
	Count        int64  `json:"count"`
	Bytes        int64  `json:"bytes"`
}

type ConfigSection struct {
	Included bool `json:"included"`
	Redacted bool `json:"redacted"`
}

func NewManifest(appVersion string, migrationVersion int) Manifest {
	return Manifest{
		Format:           BackupFormat,
		FormatVersion:    BackupFormatVersion,
		AppVersion:       appVersion,
		MigrationVersion: migrationVersion,
		CreatedAt:        time.Now().UTC(),
		Database: DatabaseSection{
			Mode: "pg_dump",
			Path: "database.dump",
		},
		Objects: ObjectSection{
			Mode:         "included",
			ManifestPath: "objects.jsonl",
			ChecksumPath: "checksums.jsonl",
		},
		Config: ConfigSection{
			Included: true,
		},
	}
}

func (m Manifest) Validate() error {
	if m.Format != BackupFormat {
		return errors.New("unsupported backup format")
	}
	if m.FormatVersion != BackupFormatVersion {
		return errors.New("unsupported backup format version")
	}
	if m.AppVersion == "" {
		return errors.New("app_version is required")
	}
	if m.MigrationVersion < 0 {
		return errors.New("migration_version must be >= 0")
	}
	if m.CreatedAt.IsZero() {
		return errors.New("created_at is required")
	}
	if m.Database.Mode == "" || m.Database.Path == "" {
		return errors.New("database mode and path are required")
	}
	if m.Objects.Mode == "" {
		return errors.New("objects mode is required")
	}
	if m.Objects.Count < 0 || m.Objects.Bytes < 0 {
		return errors.New("object count and bytes must be >= 0")
	}
	return nil
}
