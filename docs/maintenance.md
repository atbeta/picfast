# PicFast Maintenance Design

PicFast maintenance covers two related needs for private deployments:

- Backup and migration: move or restore an instance with confidence.
- Storage doctor: verify that database records, source objects, thumbnails, and storage strategies are still consistent.

The first implementation should be small, but the format and service boundaries must be versioned so future features can evolve without replacing the foundation.

## Goals

- Support same-version disaster recovery first.
- Keep backup artifacts inspectable without starting PicFast.
- Share inventory, verification, reporting, and checksum logic between backup and storage doctor.
- Prefer CLI for destructive or long-running operations. The admin UI can surface status, docs, and reports later.
- Make future cross-version migration possible through explicit format versions and feature flags.

## Non-Goals For v1

- No arbitrary cross-major-version restore promise.
- No browser-triggered restore.
- No marketplace or remote backup provider integration.
- No direct manipulation of third-party buckets beyond normal PicFast storage APIs.
- No guarantee that a backup produced with one PostgreSQL toolchain can be restored by an older or incompatible `pg_restore`. v1 assumes operators have a compatible PostgreSQL client/server environment, and restore preflight should check it where possible.

## Backup Format v1

Recommended archive layout:

```text
picfast-backup/
  manifest.json
  database.dump
  objects/
  objects.jsonl
  checksums.jsonl
```

`manifest.json` is the compatibility contract:

```json
{
  "format": "picfast-backup",
  "format_version": 1,
  "app_version": "0.8.0",
  "migration_version": 24,
  "created_at": "2026-05-10T12:00:00Z",
  "features": ["site_settings", "theme_config", "moderation"],
  "database": {
    "mode": "pg_dump",
    "path": "database.dump"
  },
  "objects": {
    "mode": "included",
    "manifest_path": "objects.jsonl",
    "checksum_path": "checksums.jsonl",
    "count": 1234,
    "bytes": 987654321
  },
  "config": {
    "included": true,
    "redacted": false
  }
}
```

`objects.jsonl` and `checksums.jsonl` have distinct responsibilities:

- `objects.jsonl` is the object inventory. Each line describes one logical PicFast object from the database: image ID, key, strategy, object path, size, mimetype, extension, and backup-local path when object data is included.
- `checksums.jsonl` is the archive integrity ledger. Each line records the digest of a payload stored in the backup archive, such as `database.dump` or an object payload under `objects/`.
- For included objects, v1 should write one inventory line and one checksum line for every object payload. The inventory is the source of truth for object identity; the checksum ledger is the source of truth for archive integrity.
- `manifest.objects.count` and `manifest.objects.bytes` describe payloads actually included in the archive, not every logical image row. Missing or externally managed objects may still appear in `objects.jsonl` without `archive_path`.
- Restore must report an error when an inventory entry references an included object without a matching checksum entry, or when checksum entries refer to unknown object payloads.
- Future incremental or remote-object backups may make object checksums optional for objects not included in the archive, but v1 should keep the one-to-one rule for included payloads.

Compatibility rules:

- `format_version` is the backup reader contract. Increment it only when readers need new behavior.
- `app_version` and `migration_version` describe the source instance.
- v1 restore should support same-version restores first. Newer PicFast versions may restore older v1 backups after running normal migrations.
- Unknown `features` should not break inspection.
- Restore must treat required unknown features as incompatible. Optional unknown features may produce warnings when the manifest explicitly marks them optional. If the manifest does not distinguish feature criticality, unknown features should be treated as restore-blocking.
- Sensitive config export must be explicit. Operational backups may include secrets; shareable exports should be redacted.

Archive shape:

- `picfast-backup.tar.zst` is the recommended distribution artifact, but implementations should avoid requiring a full extracted copy plus a full compressed copy for large instances.
- v1 may start with a simple archive writer, but the design should leave room for streaming tar creation and streaming restore preflight.
- Restore docs should call out temporary disk requirements clearly, especially when object payloads are included.

## Shared Service Boundaries

Use one internal foundation for backup and doctor:

```text
internal/service/maintenance
  Manifest      // backup format and compatibility metadata
  Inventory     // database-backed list of images, objects, strategies, thumbnails
  Verifier      // object existence, size, hash, thumbnail checks
  Reporter      // machine-readable summary for CLI and future admin UI
  Backup        // archive writer built on inventory and verifier
  Restore       // preflight + restore executor, added after backup v1
  Repair        // thumbnail rebuild and future object repair tools
```

The first code milestone should implement only the safe pieces:

- Manifest model and validation.
- Inventory model and read-only database source.
- Verifier/report model for object and thumbnail checks.

## Inventory Model

Each inventory item should preserve enough data to locate and verify the real object without depending on UI state:

- image ID and public key
- `strategy_id`, strategy name, strategy type, strategy config
- object path derived from `path + name`
- expected size, md5, sha1, mimetype, extension
- thumbnail path derived from `md5 + ".png"` when thumbnails are expected

The inventory layer should not mutate data. It exists so backup, doctor, and future repair commands all scan the same source of truth.

## Verification Model

Object verification checks:

- missing strategy
- storage initialization failure
- object read failure
- size mismatch
- md5 mismatch
- sha1 mismatch

Thumbnail verification checks:

- skipped for `svg` and `ico`
- missing thumbnail
- thumbnail file exists but is not readable

Verification rules are part of the maintenance contract. If thumbnail policy changes later, such as generating thumbnails for `ico`, the rule should evolve through `format_version` or an explicit verifier rule version so old reports and backups remain understandable.

Reports should include counts and per-item findings. The CLI can later support `--json` output using the same structs.

Full doctor scans can read every database row and every object payload. Operators should prefer maintenance windows for large instances, and the implementation should support batching, progress output, and rate limits before promoting full scans as always-safe production checks.

## CLI Roadmap

Recommended commands:

```bash
picfast maintenance doctor
picfast maintenance doctor --json --limit 500
picfast maintenance doctor --all --batch-size 500
picfast maintenance backup --output picfast-backup.tar.zst
picfast maintenance inspect picfast-backup.tar.zst
picfast maintenance restore picfast-backup.tar.zst
picfast maintenance repair-thumbnails
```

`doctor` should start as a read-only command. By default it may scan a bounded batch rather than the full instance; operators can use `--all` explicitly for full scans. `--skip-objects` and `--skip-thumbnails` are useful when checking only one side of the consistency model.

Current first slice:

```bash
picfast maintenance doctor --limit 500
picfast maintenance doctor --json --limit 500
picfast maintenance doctor --all --batch-size 500
picfast maintenance doctor --object-timeout 15s
picfast maintenance backup --output picfast-backup.tar.gz
picfast maintenance backup --database-only --output picfast-db-only.tar.gz
picfast maintenance backup --database-only --pg-dump-container picfast-db
picfast maintenance inspect picfast-backup.tar.gz
picfast maintenance restore picfast-backup.tar.gz
picfast maintenance restore picfast-backup.tar.gz --apply --force
picfast maintenance restore picfast-backup.tar.gz --apply --force --pg-restore-container picfast-db
```

The command exits with status `1` when findings are present, so it can be used in scripts. Use `--json` for machine-readable reports.

`backup` v1 uses `pg_dump --format=custom` for the database section and therefore requires a compatible `pg_dump` executable in the runtime environment. In local Docker development, `--pg-dump-container picfast-db` can use the PostgreSQL container's bundled `pg_dump` instead of requiring PostgreSQL client tools on the host. Use `--database-only` when source objects are managed separately or when historical object records are already known to be incomplete. Use `inspect` after moving a backup archive to validate manifest and checksum integrity before restore work begins.

`restore` defaults to a dry run. It verifies the archive, checks the target database, reports blockers, and exits without changing data unless `--apply` is present. Restoring into a non-empty database is blocked unless `--force` is also present. In local Docker development, `--pg-restore-container picfast-db` uses the PostgreSQL container's bundled `pg_restore`, which avoids installing PostgreSQL client tools on the host. Object payload restore runs after the database restore and uses the restored strategy configuration, so local paths, bucket credentials, and mounted volumes must already be valid on the target server.

Implementation order:

1. `doctor`: read-only inventory and verification.
2. `backup inspect`: parse and validate manifest.
3. `backup`: same-version export with manifest.
4. `restore --dry-run`: preflight only.
5. `restore`: guarded destructive restore.
6. `repair-thumbnails`: rebuild missing thumbnails.

## Restore Safety

Restore should always run preflight checks before changing anything:

- target database is reachable
- compatible PostgreSQL dump/restore tooling is available
- target version and backup format are compatible
- target instance is empty or `--force` is explicitly passed
- required storage roots are writable
- archive checksums are valid

The first restore implementation should be same-version only. Cross-version behavior should be enabled deliberately after the v1 backup format has real usage.
