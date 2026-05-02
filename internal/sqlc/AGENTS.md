# internal/sqlc rules

This directory contains sqlc inputs and generated Go code.

Allowed manual edits:

- `queries/*.sql`

Do not manually edit:

- `*.go`

The Go files in this directory are tracked generated files. They may appear in
commits only after running `make generate`; never edit them directly with an
editor, `apply_patch`, scripts, or formatting tools.

If a generated Go file needs to change, update the sqlc input instead:

- `internal/sqlc/queries/*.sql`
- `migrations/*.sql`
- `sqlc.yaml`

Then run `make generate` and review the generated diff.
