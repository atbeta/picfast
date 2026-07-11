package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/atbeta/picfast/internal/config"
	picservice "github.com/atbeta/picfast/internal/service"
	"github.com/atbeta/picfast/internal/service/maintenance"
	"github.com/atbeta/picfast/internal/sqlc"
	"github.com/atbeta/picfast/internal/version"
	"github.com/davidbyttow/govips/v2/vips"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func runMaintenanceCommand(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printMaintenanceUsage(stderr)
		return 2
	}

	switch args[0] {
	case "doctor":
		return runMaintenanceDoctor(ctx, args[1:], stdout, stderr)
	case "backup":
		return runMaintenanceBackup(ctx, args[1:], stdout, stderr)
	case "inspect":
		return runMaintenanceInspect(args[1:], stdout, stderr)
	case "restore":
		return runMaintenanceRestore(ctx, args[1:], stdout, stderr)
	case "repair-thumbnails":
		return runMaintenanceRepairThumbnails(ctx, args[1:], stdout, stderr)
	case "recalc-phash":
		return runRecalcPHash(ctx, args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printMaintenanceUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown maintenance command: %s\n\n", args[0])
		printMaintenanceUsage(stderr)
		return 2
	}
}

func printMaintenanceUsage(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  picfast maintenance doctor [flags]
  picfast maintenance backup [flags]
  picfast maintenance inspect <backup.tar|backup.tar.gz> [flags]
  picfast maintenance restore <backup.tar|backup.tar.gz> [flags]
  picfast maintenance repair-thumbnails [flags]
  picfast maintenance recalc-phash [flags]

Commands:
  doctor              Verify image objects and thumbnails without changing data
  backup              Create a versioned backup archive
  inspect             Validate backup manifest and checksums
  restore             Restore a backup archive after guarded preflight checks
  repair-thumbnails   Rebuild missing thumbnails from readable source objects
  recalc-phash        Recompute perceptual hashes for images missing one`)
}

func runMaintenanceDoctor(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("picfast maintenance doctor", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "write the report as JSON")
	limit := fs.Int("limit", 500, "maximum images to scan in this run")
	offset := fs.Int("offset", 0, "image offset for this run")
	all := fs.Bool("all", false, "scan all images in batches")
	batchSize := fs.Int("batch-size", 500, "batch size when --all is used")
	skipObjects := fs.Bool("skip-objects", false, "skip source object reads")
	skipThumbnails := fs.Bool("skip-thumbnails", false, "skip thumbnail checks")
	objectTimeout := fs.Duration("object-timeout", 15*time.Second, "timeout for each source object read")
	if err := parseInterspersedFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected doctor argument: %s\n", fs.Arg(0))
		return 2
	}
	if *limit <= 0 || *offset < 0 || *batchSize <= 0 {
		fmt.Fprintln(stderr, "limit and batch-size must be positive, and offset must be >= 0")
		return 2
	}
	if *objectTimeout <= 0 {
		fmt.Fprintln(stderr, "object-timeout must be positive")
		return 2
	}
	progress := !*jsonOutput

	if progress {
		fmt.Fprintln(stderr, "Loading PicFast config...")
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}
	if progress {
		fmt.Fprintln(stderr, "Connecting to database...")
	}
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(stderr, "failed to connect to database: %v\n", err)
		return 1
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(stderr, "failed to ping database: %v\n", err)
		return 1
	}

	source := maintenance.NewPGInventorySource(pool)
	verifier := maintenance.Verifier{ThumbnailDir: cfg.Storage.ThumbnailDir, ObjectTimeout: *objectTimeout}
	report := maintenance.NewReport()
	scanLimit := int32(*limit)
	scanOffset := int32(*offset)
	if *all {
		scanLimit = int32(*batchSize)
	}

	if progress {
		fmt.Fprintf(stderr, "Scanning images offset=%d limit=%d all=%t skip_objects=%t skip_thumbnails=%t...\n", scanOffset, scanLimit, *all, *skipObjects, *skipThumbnails)
	}
	for {
		items, err := source.ListImages(ctx, maintenance.InventoryOptions{Limit: scanLimit, Offset: scanOffset})
		if err != nil {
			fmt.Fprintf(stderr, "failed to load image inventory: %v\n", err)
			return 1
		}
		if len(items) == 0 {
			break
		}
		if progress {
			fmt.Fprintf(stderr, "Checking batch offset=%d count=%d...\n", scanOffset, len(items))
		}
		for _, item := range items {
			if !*skipObjects {
				report.AddObjectCheck(verifier.VerifyObject(ctx, item))
			}
			if !*skipThumbnails {
				report.AddThumbnailCheck(verifier.VerifyThumbnail(item))
			}
		}
		if !*all || len(items) < int(scanLimit) {
			break
		}
		scanOffset += scanLimit
	}

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(stderr, "failed to encode report: %v\n", err)
			return 1
		}
		return maintenanceDoctorExitCode(report)
	}

	printDoctorReport(stdout, report)
	return maintenanceDoctorExitCode(report)
}

func printDoctorReport(w io.Writer, report maintenance.Report) {
	fmt.Fprintf(w, "PicFast maintenance doctor\n")
	fmt.Fprintf(w, "Generated at: %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	if report.Objects.Total > 0 {
		fmt.Fprintf(w, "Objects:    total=%d ok=%d failed=%d\n", report.Objects.Total, report.Objects.OK, report.Objects.Failed)
	} else {
		fmt.Fprintln(w, "Objects:    skipped")
	}
	if report.Thumbnails.Total > 0 {
		fmt.Fprintf(w, "Thumbnails: total=%d ok=%d skipped=%d failed=%d\n", report.Thumbnails.Total, report.Thumbnails.OK, report.Thumbnails.Skipped, report.Thumbnails.Failed)
	} else {
		fmt.Fprintln(w, "Thumbnails: skipped")
	}
	if len(report.Findings) == 0 {
		fmt.Fprintln(w, "\nNo findings.")
		return
	}

	fmt.Fprintln(w, "\nFindings:")
	for _, finding := range report.Findings {
		parts := []string{
			fmt.Sprintf("image=%d", finding.ImageID),
			"key=" + finding.Key,
			"kind=" + finding.Kind,
			"status=" + finding.Status,
		}
		if finding.Path != "" {
			parts = append(parts, "path="+finding.Path)
		}
		if finding.Error != "" {
			parts = append(parts, "error="+finding.Error)
		}
		fmt.Fprintln(w, "  - "+strings.Join(parts, " "))
	}
}

func maintenanceDoctorExitCode(report maintenance.Report) int {
	if report.Objects.Failed > 0 || report.Thumbnails.Failed > 0 {
		return 1
	}
	return 0
}

func runMaintenanceBackup(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("picfast maintenance backup", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "write the result as JSON")
	output := fs.String("output", "picfast-backup.tar.gz", "backup archive path (.tar or .tar.gz)")
	databaseOnly := fs.Bool("database-only", false, "include only the database dump and metadata")
	allowMissingObjects := fs.Bool("allow-missing-objects", false, "write object inventory warnings instead of failing on unreadable objects")
	limit := fs.Int("limit", 500, "maximum images to scan when --all=false")
	offset := fs.Int("offset", 0, "image offset when --all=false")
	all := fs.Bool("all", true, "scan all images in batches")
	batchSize := fs.Int("batch-size", 500, "batch size when --all is used")
	objectTimeout := fs.Duration("object-timeout", 15*time.Second, "timeout for each source object read")
	pgDump := fs.String("pg-dump", "pg_dump", "pg_dump executable path")
	pgDumpContainer := fs.String("pg-dump-container", "", "Docker container that provides pg_dump, such as picfast-db")
	if err := parseInterspersedFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected backup argument: %s\n", fs.Arg(0))
		return 2
	}
	if *limit <= 0 || *offset < 0 || *batchSize <= 0 {
		fmt.Fprintln(stderr, "limit and batch-size must be positive, and offset must be >= 0")
		return 2
	}
	if *objectTimeout <= 0 {
		fmt.Fprintln(stderr, "object-timeout must be positive")
		return 2
	}

	progress := !*jsonOutput
	if progress {
		fmt.Fprintln(stderr, "Loading PicFast config...")
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}
	if progress {
		fmt.Fprintln(stderr, "Connecting to database...")
	}
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(stderr, "failed to connect to database: %v\n", err)
		return 1
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(stderr, "failed to ping database: %v\n", err)
		return 1
	}
	migrationVersion, err := currentMigrationVersion(ctx, pool)
	if err != nil {
		fmt.Fprintf(stderr, "failed to read migration version: %v\n", err)
		return 1
	}

	tmpDir, err := os.MkdirTemp("", "picfast-pgdump-*")
	if err != nil {
		fmt.Fprintf(stderr, "failed to create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)
	dumpPath := filepath.Join(tmpDir, "database.dump")
	if progress {
		fmt.Fprintln(stderr, "Running pg_dump...")
	}
	if err := runPGDump(ctx, *pgDump, *pgDumpContainer, cfg.Database.URL, dumpPath); err != nil {
		fmt.Fprintf(stderr, "failed to run pg_dump: %v\n", err)
		return 1
	}

	var items []maintenance.InventoryItem
	if !*databaseOnly {
		if progress {
			fmt.Fprintln(stderr, "Loading object inventory...")
		}
		items, err = loadMaintenanceInventory(ctx, maintenance.NewPGInventorySource(pool), *all, int32(*limit), int32(*offset), int32(*batchSize), progress, stderr)
		if err != nil {
			fmt.Fprintf(stderr, "failed to load image inventory: %v\n", err)
			return 1
		}
	}

	writer := maintenance.BackupWriter{}
	if progress {
		fmt.Fprintf(stderr, "Writing backup archive %s...\n", *output)
	}
	result, err := writer.Write(ctx, items, maintenance.BackupOptions{
		OutputPath:          *output,
		DatabaseDumpPath:    dumpPath,
		AppVersion:          version.Version,
		MigrationVersion:    migrationVersion,
		Features:            currentBackupFeatures(),
		IncludeObjects:      !*databaseOnly,
		AllowMissingObjects: *allowMissingObjects,
		ObjectTimeout:       *objectTimeout,
	})
	if err != nil {
		fmt.Fprintf(stderr, "failed to write backup: %v\n", err)
		return 1
	}

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(stderr, "failed to encode result: %v\n", err)
			return 1
		}
		return 0
	}
	printBackupResult(stdout, result)
	return 0
}

func runMaintenanceInspect(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("picfast maintenance inspect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "write the result as JSON")
	if err := parseInterspersedFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "inspect requires exactly one backup archive path")
		return 2
	}
	result, err := maintenance.InspectArchive(fs.Arg(0))
	if err != nil {
		fmt.Fprintf(stderr, "failed to inspect backup: %v\n", err)
		return 1
	}
	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(stderr, "failed to encode result: %v\n", err)
			return 1
		}
		if result.OK() {
			return 0
		}
		return 1
	}
	printInspectResult(stdout, result)
	if result.OK() {
		return 0
	}
	return 1
}

type restoreCLIResult struct {
	GeneratedAt      time.Time                         `json:"generated_at"`
	ArchivePath      string                            `json:"archive_path"`
	Apply            bool                              `json:"apply"`
	Force            bool                              `json:"force"`
	SkipObjects      bool                              `json:"skip_objects"`
	Manifest         maintenance.Manifest              `json:"manifest"`
	Preflight        []string                          `json:"preflight,omitempty"`
	Warnings         []string                          `json:"warnings,omitempty"`
	DatabaseRestored bool                              `json:"database_restored"`
	Objects          *maintenance.RestoreObjectsResult `json:"objects,omitempty"`
}

func runMaintenanceRestore(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("picfast maintenance restore", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "write the result as JSON")
	apply := fs.Bool("apply", false, "restore database and objects; without this flag the command is a dry run")
	force := fs.Bool("force", false, "allow restore into a non-empty target database")
	skipObjects := fs.Bool("skip-objects", false, "restore only the database")
	objectTimeout := fs.Duration("object-timeout", 15*time.Second, "timeout for each restored object write")
	pgRestore := fs.String("pg-restore", "pg_restore", "pg_restore executable path")
	pgRestoreContainer := fs.String("pg-restore-container", "", "Docker container that provides pg_restore, such as picfast-db")
	if err := parseInterspersedFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "restore requires exactly one backup archive path")
		return 2
	}
	if *objectTimeout <= 0 {
		fmt.Fprintln(stderr, "object-timeout must be positive")
		return 2
	}

	archivePath := fs.Arg(0)
	progress := !*jsonOutput
	result := restoreCLIResult{
		GeneratedAt: time.Now().UTC(),
		ArchivePath: archivePath,
		Apply:       *apply,
		Force:       *force,
		SkipObjects: *skipObjects,
	}

	if progress {
		fmt.Fprintln(stderr, "Inspecting backup archive...")
	}
	inspect, err := maintenance.InspectArchive(archivePath)
	if err != nil {
		fmt.Fprintf(stderr, "failed to inspect backup: %v\n", err)
		return 1
	}
	result.Manifest = inspect.Manifest
	if !inspect.OK() {
		result.Preflight = append(result.Preflight, "backup archive checksum validation failed")
	}
	if !*skipObjects && result.Manifest.Objects.Mode == "database_only" {
		result.Warnings = append(result.Warnings, "backup is database_only; object restore will be skipped")
	}

	if progress {
		fmt.Fprintln(stderr, "Loading PicFast config...")
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}
	if progress {
		fmt.Fprintln(stderr, "Connecting to database...")
	}
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(stderr, "failed to connect to database: %v\n", err)
		return 1
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(stderr, "failed to ping database: %v\n", err)
		return 1
	}

	if err := checkPGRestore(ctx, *pgRestore, *pgRestoreContainer); err != nil {
		result.Preflight = append(result.Preflight, "pg_restore is not available: "+err.Error())
	}
	nonEmpty, err := databaseHasData(ctx, pool)
	if err != nil {
		result.Preflight = append(result.Preflight, "failed to inspect target database: "+err.Error())
	} else if nonEmpty && !*force {
		result.Preflight = append(result.Preflight, "target database is not empty; pass --force to allow destructive restore")
	}
	targetMigration, err := databaseMigrationVersionIfPresent(ctx, pool)
	if err != nil {
		result.Preflight = append(result.Preflight, "failed to inspect target migration version: "+err.Error())
	} else if targetMigration != nil && *targetMigration != result.Manifest.MigrationVersion {
		result.Preflight = append(result.Preflight, fmt.Sprintf("target migration version %d does not match backup migration version %d", *targetMigration, result.Manifest.MigrationVersion))
	}
	if result.Manifest.FormatVersion != maintenance.BackupFormatVersion {
		result.Preflight = append(result.Preflight, "backup format version is not supported")
	}
	if unknown := unknownBackupFeatures(result.Manifest.Features); len(unknown) > 0 {
		result.Preflight = append(result.Preflight, "backup contains unsupported features: "+strings.Join(unknown, ", "))
	}

	if len(result.Preflight) > 0 {
		if *jsonOutput {
			writeJSONResult(stdout, stderr, result)
		} else {
			printRestoreResult(stdout, result)
		}
		return 1
	}
	if !*apply {
		result.Warnings = append(result.Warnings, "dry-run only; pass --apply to restore")
		if *jsonOutput {
			writeJSONResult(stdout, stderr, result)
		} else {
			printRestoreResult(stdout, result)
		}
		return 0
	}

	tmpDir, err := os.MkdirTemp("", "picfast-restore-*")
	if err != nil {
		fmt.Fprintf(stderr, "failed to create temp dir: %v\n", err)
		return 1
	}
	defer os.RemoveAll(tmpDir)
	if progress {
		fmt.Fprintln(stderr, "Extracting backup archive...")
	}
	if err := maintenance.ExtractArchive(archivePath, tmpDir); err != nil {
		fmt.Fprintf(stderr, "failed to extract backup: %v\n", err)
		return 1
	}
	dumpPath, err := extractedArchivePath(tmpDir, result.Manifest.Database.Path)
	if err != nil {
		fmt.Fprintf(stderr, "invalid database dump path in manifest: %v\n", err)
		return 1
	}
	if progress {
		fmt.Fprintln(stderr, "Running pg_restore...")
	}
	if err := runPGRestore(ctx, *pgRestore, *pgRestoreContainer, cfg.Database.URL, dumpPath); err != nil {
		fmt.Fprintf(stderr, "failed to run pg_restore: %v\n", err)
		return 1
	}
	result.DatabaseRestored = true

	if !*skipObjects && result.Manifest.Objects.Mode == "included" && result.Manifest.Objects.ManifestPath != "" {
		objectManifestPath, err := extractedArchivePath(tmpDir, result.Manifest.Objects.ManifestPath)
		if err != nil {
			fmt.Fprintf(stderr, "invalid object manifest path in manifest: %v\n", err)
			return 1
		}
		objects, err := maintenance.ReadObjectManifest(objectManifestPath)
		if err != nil {
			fmt.Fprintf(stderr, "failed to read object manifest: %v\n", err)
			return 1
		}
		if progress {
			fmt.Fprintf(stderr, "Restoring %d object entries...\n", len(objects))
		}
		objectResult, err := maintenance.RestoreObjects(ctx, pool, tmpDir, objects, maintenance.RestoreObjectsOptions{ObjectTimeout: *objectTimeout})
		if err != nil {
			fmt.Fprintf(stderr, "failed to restore objects: %v\n", err)
			return 1
		}
		result.Objects = &objectResult
	}

	if *jsonOutput {
		writeJSONResult(stdout, stderr, result)
	} else {
		printRestoreResult(stdout, result)
	}
	if result.Objects != nil && result.Objects.Failed > 0 {
		return 1
	}
	return 0
}

func runMaintenanceRepairThumbnails(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("picfast maintenance repair-thumbnails", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOutput := fs.Bool("json", false, "write the report as JSON")
	apply := fs.Bool("apply", false, "write missing thumbnails; without this flag the command is a dry run")
	limit := fs.Int("limit", 500, "maximum images to scan in this run")
	offset := fs.Int("offset", 0, "image offset for this run")
	all := fs.Bool("all", false, "scan all images in batches")
	batchSize := fs.Int("batch-size", 500, "batch size when --all is used")
	objectTimeout := fs.Duration("object-timeout", 15*time.Second, "timeout for each source object read")
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if fs.NArg() > 0 {
		fmt.Fprintf(stderr, "unexpected repair-thumbnails argument: %s\n", fs.Arg(0))
		return 2
	}
	if *limit <= 0 || *offset < 0 || *batchSize <= 0 {
		fmt.Fprintln(stderr, "limit and batch-size must be positive, and offset must be >= 0")
		return 2
	}
	if *objectTimeout <= 0 {
		fmt.Fprintln(stderr, "object-timeout must be positive")
		return 2
	}
	if *apply {
		vips.Startup(&vips.Config{
			ConcurrencyLevel: 2,
			MaxCacheFiles:    0,
			MaxCacheMem:      100 * 1024 * 1024,
			MaxCacheSize:     500,
		})
		defer vips.Shutdown()
	}

	progress := !*jsonOutput
	if progress {
		fmt.Fprintln(stderr, "Loading PicFast config...")
	}
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "failed to load config: %v\n", err)
		return 1
	}
	if progress {
		fmt.Fprintln(stderr, "Connecting to database...")
	}
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(stderr, "failed to connect to database: %v\n", err)
		return 1
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		fmt.Fprintf(stderr, "failed to ping database: %v\n", err)
		return 1
	}

	source := maintenance.NewPGInventorySource(pool)
	repairer := maintenance.ThumbnailRepairer{
		Verifier: maintenance.Verifier{
			ThumbnailDir:  cfg.Storage.ThumbnailDir,
			ObjectTimeout: *objectTimeout,
		},
		Generator: picservice.GenerateThumbnail,
	}
	result := maintenance.RepairResult{GeneratedAt: time.Now().UTC(), Apply: *apply}
	scanLimit := int32(*limit)
	scanOffset := int32(*offset)
	if *all {
		scanLimit = int32(*batchSize)
	}

	if progress {
		mode := "dry-run"
		if *apply {
			mode = "apply"
		}
		fmt.Fprintf(stderr, "Scanning images offset=%d limit=%d all=%t mode=%s...\n", scanOffset, scanLimit, *all, mode)
	}
	for {
		items, err := source.ListImages(ctx, maintenance.InventoryOptions{Limit: scanLimit, Offset: scanOffset})
		if err != nil {
			fmt.Fprintf(stderr, "failed to load image inventory: %v\n", err)
			return 1
		}
		if len(items) == 0 {
			break
		}
		if progress {
			fmt.Fprintf(stderr, "Repairing batch offset=%d count=%d...\n", scanOffset, len(items))
		}
		batch := repairer.Repair(ctx, items, maintenance.RepairOptions{Apply: *apply, ObjectTimeout: *objectTimeout})
		mergeRepairResult(&result, batch)
		if !*all || len(items) < int(scanLimit) {
			break
		}
		scanOffset += scanLimit
	}

	if *jsonOutput {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(stderr, "failed to encode report: %v\n", err)
			return 1
		}
		return maintenanceRepairExitCode(result)
	}

	printRepairReport(stdout, result)
	return maintenanceRepairExitCode(result)
}

func mergeRepairResult(dst *maintenance.RepairResult, src maintenance.RepairResult) {
	if dst.GeneratedAt.IsZero() {
		dst.GeneratedAt = src.GeneratedAt
	}
	dst.Apply = src.Apply
	dst.Total += src.Total
	dst.Repaired += src.Repaired
	dst.WouldRepair += src.WouldRepair
	dst.Skipped += src.Skipped
	dst.Failed += src.Failed
	dst.Items = append(dst.Items, src.Items...)
}

func printRepairReport(w io.Writer, result maintenance.RepairResult) {
	mode := "dry-run"
	if result.Apply {
		mode = "apply"
	}
	fmt.Fprintln(w, "PicFast maintenance repair-thumbnails")
	fmt.Fprintf(w, "Generated at: %s\n", result.GeneratedAt.Format("2006-01-02 15:04:05 MST"))
	fmt.Fprintf(w, "Mode: %s\n\n", mode)
	fmt.Fprintf(w, "Thumbnails: total=%d repaired=%d would_repair=%d skipped=%d failed=%d\n", result.Total, result.Repaired, result.WouldRepair, result.Skipped, result.Failed)
	if len(result.Items) == 0 {
		return
	}

	fmt.Fprintln(w, "\nItems:")
	for _, item := range result.Items {
		parts := []string{
			fmt.Sprintf("image=%d", item.ImageID),
			"key=" + item.Key,
			"status=" + string(item.Status),
		}
		if item.MD5 != "" {
			parts = append(parts, "md5="+item.MD5)
		}
		if item.Path != "" {
			parts = append(parts, "path="+item.Path)
		}
		if item.Error != "" {
			parts = append(parts, "error="+item.Error)
		}
		fmt.Fprintln(w, "  - "+strings.Join(parts, " "))
	}
}

func maintenanceRepairExitCode(result maintenance.RepairResult) int {
	if result.Failed > 0 {
		return 1
	}
	return 0
}

func loadMaintenanceInventory(ctx context.Context, source maintenance.InventorySource, all bool, limit, offset, batchSize int32, progress bool, stderr io.Writer) ([]maintenance.InventoryItem, error) {
	if all {
		limit = batchSize
		offset = 0
	}
	var allItems []maintenance.InventoryItem
	for {
		if progress {
			fmt.Fprintf(stderr, "Loading inventory batch offset=%d limit=%d...\n", offset, limit)
		}
		items, err := source.ListImages(ctx, maintenance.InventoryOptions{Limit: limit, Offset: offset})
		if err != nil {
			return nil, err
		}
		allItems = append(allItems, items...)
		if !all || len(items) < int(limit) {
			return allItems, nil
		}
		offset += limit
	}
}

func currentMigrationVersion(ctx context.Context, pool *pgxpool.Pool) (int, error) {
	var version int
	var dirty bool
	if err := pool.QueryRow(ctx, `SELECT version, dirty FROM schema_migrations LIMIT 1`).Scan(&version, &dirty); err != nil {
		return 0, err
	}
	if dirty {
		return 0, fmt.Errorf("database migration is dirty")
	}
	return version, nil
}

func runPGDump(ctx context.Context, pgDump, pgDumpContainer, databaseURL, outputPath string) error {
	if pgDumpContainer != "" {
		return runDockerPGDump(ctx, pgDumpContainer, databaseURL, outputPath)
	}
	cmd := exec.CommandContext(ctx, pgDump, "--format=custom", "--file", outputPath, databaseURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runDockerPGDump(ctx context.Context, container, databaseURL, outputPath string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}
	user := u.User.Username()
	if user == "" {
		user = "postgres"
	}
	password, _ := u.User.Password()
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		dbName = "postgres"
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	args := []string{"exec", "-i"}
	if password != "" {
		args = append(args, "-e", "PGPASSWORD="+password)
	}
	args = append(args, container, "pg_dump", "--format=custom", "-U", user, "-d", dbName)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdout = outFile
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func runPGRestore(ctx context.Context, pgRestore, pgRestoreContainer, databaseURL, dumpPath string) error {
	if pgRestoreContainer != "" {
		return runDockerPGRestore(ctx, pgRestoreContainer, databaseURL, dumpPath)
	}
	cmd := exec.CommandContext(ctx, pgRestore, "--clean", "--if-exists", "--no-owner", "--exit-on-error", "--dbname", databaseURL, dumpPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runDockerPGRestore(ctx context.Context, container, databaseURL, dumpPath string) error {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}
	user := u.User.Username()
	if user == "" {
		user = "postgres"
	}
	password, _ := u.User.Password()
	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		dbName = "postgres"
	}

	dumpFile, err := os.Open(dumpPath)
	if err != nil {
		return err
	}
	defer dumpFile.Close()

	args := []string{"exec", "-i"}
	if password != "" {
		args = append(args, "-e", "PGPASSWORD="+password)
	}
	args = append(args, container, "pg_restore", "--clean", "--if-exists", "--no-owner", "--exit-on-error", "-U", user, "-d", dbName)
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stdin = dumpFile
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return nil
}

func checkPGRestore(ctx context.Context, pgRestore, pgRestoreContainer string) error {
	var cmd *exec.Cmd
	if pgRestoreContainer != "" {
		cmd = exec.CommandContext(ctx, "docker", "exec", pgRestoreContainer, "pg_restore", "--version")
	} else {
		cmd = exec.CommandContext(ctx, pgRestore, "--version")
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func currentBackupFeatures() []string {
	return []string{
		"site_settings",
		"theme_config",
		"moderation",
		"email_verification",
		"password_reset",
		"analytics",
	}
}

func unknownBackupFeatures(features []string) []string {
	known := make(map[string]bool)
	for _, feature := range currentBackupFeatures() {
		known[feature] = true
	}
	var unknown []string
	for _, feature := range features {
		if !known[feature] {
			unknown = append(unknown, feature)
		}
	}
	return unknown
}

func printBackupResult(w io.Writer, result maintenance.BackupResult) {
	fmt.Fprintln(w, "PicFast maintenance backup")
	fmt.Fprintf(w, "Output: %s\n", result.OutputPath)
	fmt.Fprintf(w, "Format: %s v%d\n", result.Manifest.Format, result.Manifest.FormatVersion)
	fmt.Fprintf(w, "App version: %s\n", result.Manifest.AppVersion)
	fmt.Fprintf(w, "Migration version: %d\n", result.Manifest.MigrationVersion)
	fmt.Fprintf(w, "Database: %s\n", result.Manifest.Database.Path)
	fmt.Fprintf(w, "Objects: mode=%s count=%d bytes=%d\n", result.Manifest.Objects.Mode, result.Manifest.Objects.Count, result.Manifest.Objects.Bytes)
	if len(result.Warnings) > 0 {
		fmt.Fprintln(w, "\nWarnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintln(w, "  - "+warning)
		}
	}
}

func printInspectResult(w io.Writer, result maintenance.InspectResult) {
	status := "ok"
	if !result.OK() {
		status = "failed"
	}
	fmt.Fprintln(w, "PicFast maintenance inspect")
	fmt.Fprintf(w, "Archive: %s\n", result.ArchivePath)
	fmt.Fprintf(w, "Status: %s\n", status)
	fmt.Fprintf(w, "Format: %s v%d\n", result.Manifest.Format, result.Manifest.FormatVersion)
	fmt.Fprintf(w, "App version: %s\n", result.Manifest.AppVersion)
	fmt.Fprintf(w, "Migration version: %d\n", result.Manifest.MigrationVersion)
	fmt.Fprintf(w, "Checksums: verified=%d total=%d\n", result.VerifiedCount, result.ChecksumCount)
	if len(result.MissingChecksums) > 0 {
		fmt.Fprintln(w, "\nMissing checksum payloads:")
		for _, item := range result.MissingChecksums {
			fmt.Fprintln(w, "  - "+item)
		}
	}
	if len(result.ChecksumFailures) > 0 {
		fmt.Fprintln(w, "\nChecksum failures:")
		for _, item := range result.ChecksumFailures {
			fmt.Fprintln(w, "  - "+item)
		}
	}
	if len(result.Warnings) > 0 {
		fmt.Fprintln(w, "\nWarnings:")
		for _, warning := range result.Warnings {
			fmt.Fprintln(w, "  - "+warning)
		}
	}
}

func printRestoreResult(w io.Writer, result restoreCLIResult) {
	mode := "dry-run"
	if result.Apply {
		mode = "apply"
	}
	fmt.Fprintln(w, "PicFast maintenance restore")
	fmt.Fprintf(w, "Archive: %s\n", result.ArchivePath)
	fmt.Fprintf(w, "Mode: %s\n", mode)
	fmt.Fprintf(w, "Force: %t\n", result.Force)
	fmt.Fprintf(w, "Format: %s v%d\n", result.Manifest.Format, result.Manifest.FormatVersion)
	fmt.Fprintf(w, "App version: %s\n", result.Manifest.AppVersion)
	fmt.Fprintf(w, "Migration version: %d\n", result.Manifest.MigrationVersion)
	fmt.Fprintf(w, "Database restored: %t\n", result.DatabaseRestored)
	if result.Objects != nil {
		fmt.Fprintf(w, "Objects: total=%d restored=%d skipped=%d failed=%d\n", result.Objects.Total, result.Objects.Restored, result.Objects.Skipped, result.Objects.Failed)
	}
	if len(result.Preflight) > 0 {
		fmt.Fprintln(w, "\nPreflight blockers:")
		for _, item := range result.Preflight {
			fmt.Fprintln(w, "  - "+item)
		}
	}
	if len(result.Warnings) > 0 {
		fmt.Fprintln(w, "\nWarnings:")
		for _, item := range result.Warnings {
			fmt.Fprintln(w, "  - "+item)
		}
	}
	if result.Objects != nil && len(result.Objects.Items) > 0 {
		fmt.Fprintln(w, "\nObject results:")
		for _, item := range result.Objects.Items {
			parts := []string{
				fmt.Sprintf("image=%d", item.ImageID),
				"key=" + item.Key,
				"status=" + string(item.Status),
			}
			if item.Path != "" {
				parts = append(parts, "path="+item.Path)
			}
			if item.Error != "" {
				parts = append(parts, "error="+item.Error)
			}
			fmt.Fprintln(w, "  - "+strings.Join(parts, " "))
		}
	}
}

func writeJSONResult(stdout, stderr io.Writer, value any) bool {
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		fmt.Fprintf(stderr, "failed to encode result: %v\n", err)
		return false
	}
	return true
}

func databaseHasData(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	tables := []string{
		"users",
		"groups",
		"strategies",
		"group_strategies",
		"albums",
		"images",
		"api_tokens",
		"refresh_tokens",
		"image_moderations",
		"site_settings",
		"audit_logs",
	}
	for _, table := range tables {
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT to_regclass('public.`+table+`') IS NOT NULL`).Scan(&exists); err != nil {
			return false, err
		}
		if !exists {
			continue
		}
		var count int64
		if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM `+table).Scan(&count); err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}

func databaseMigrationVersionIfPresent(ctx context.Context, pool *pgxpool.Pool) (*int, error) {
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.schema_migrations') IS NOT NULL`).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	version, err := currentMigrationVersion(ctx, pool)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func extractedArchivePath(root, name string) (string, error) {
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

func parseInterspersedFlags(fs *flag.FlagSet, args []string) error {
	flagArgs := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}

		flagArgs = append(flagArgs, arg)
		name := strings.TrimLeft(arg, "-")
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		if name == "" {
			continue
		}
		if f := fs.Lookup(name); f != nil {
			if boolFlag, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && boolFlag.IsBoolFlag() {
				continue
			}
			if strings.Contains(arg, "=") {
				continue
			}
			if i+1 < len(args) {
				i++
				flagArgs = append(flagArgs, args[i])
			}
		}
	}
	return fs.Parse(append(flagArgs, positionals...))
}

func runRecalcPHash(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("picfast maintenance recalc-phash", flag.ContinueOnError)
	batchSize := fs.Int("batch", 100, "Number of images to process per batch")
	dryRun := fs.Bool("dry-run", false, "Show what would be done without making changes")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(stderr, "cannot load config: %v\n", err)
		return 1
	}
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {
		fmt.Fprintf(stderr, "cannot connect: %v\n", err)
		return 1
	}
	defer pool.Close()

	db := sqlc.New(pool)

	maxID, err := db.GetMaxImageID(ctx)
	if err != nil {
		fmt.Fprintf(stderr, "failed to get max image id: %v\n", err)
		return 1
	}

	total := 0
	skipped := 0
	lastID := int64(0)
	for lastID < maxID {
		images, err := db.GetImagesForPHashRecalc(ctx, sqlc.GetImagesForPHashRecalcParams{
			AfterID:   lastID,
			BatchSize: int32(*batchSize),
		})
		if err != nil {
			fmt.Fprintf(stderr, "query error: %v\n", err)
			return 1
		}
		if len(images) == 0 {
			break
		}

		for _, img := range images {
			lastID = img.ID
			strategy, err := db.GetStrategyByID(ctx, img.StrategyID.Int64)
			if err != nil {
				fmt.Fprintf(stderr, "strategy lookup failed for image %d: %v\n", img.ID, err)
				skipped++
				continue
			}

			store, err := picservice.GetStorageForStrategy(strategy)
			if err != nil {
				fmt.Fprintf(stderr, "storage init failed for image %d: %v\n", img.ID, err)
				skipped++
				continue
			}

			pathname := img.Name
			if img.Path != "" && img.Path != "." {
				pathname = img.Path + "/" + img.Name
			}
			data, err := store.Read(ctx, pathname)
			_ = store.Close()
			if err != nil {
				fmt.Fprintf(stderr, "read failed for image %d: %v\n", img.ID, err)
				if !*dryRun {
					db.UpdateImagePHash(ctx, sqlc.UpdateImagePHashParams{
						ID:    img.ID,
						Phash: pgtype.Int8{Int64: -1, Valid: true},
					})
				}
				skipped++
				continue
			}

			phash, err := picservice.ComputePHash(data)
			if err != nil {
				fmt.Fprintf(stderr, "phash failed for image %d: %v\n", img.ID, err)
				if !*dryRun {
					db.UpdateImagePHash(ctx, sqlc.UpdateImagePHashParams{
						ID:    img.ID,
						Phash: pgtype.Int8{Int64: -1, Valid: true},
					})
				}
				skipped++
				continue
			}

			if !*dryRun {
				if err := db.UpdateImagePHash(ctx, sqlc.UpdateImagePHashParams{
					ID:    img.ID,
					Phash: pgtype.Int8{Int64: int64(phash), Valid: true},
				}); err != nil {
					fmt.Fprintf(stderr, "update phash failed for image %d: %v\n", img.ID, err)
					skipped++
					continue
				}
			}

			total++
			if total%10 == 0 {
				fmt.Fprintf(stdout, "processed %d images...\n", total)
			}
		}
	}

	fmt.Fprintf(stdout, "done: processed %d images, skipped %d\n", total, skipped)
	return 0
}
