# v0.0.8 - Phase 04: Virtualized data grid, staged edits, and import/export

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 12 open)
> **Estimated effort:** 3 days

---

## Overview

Phase 04 delivers the high-volume data grid, typed cell rendering, JSON/text cell inspector, safe staged transactional table editing (`UPDATE`/`INSERT`/`DELETE` preview + commit/discard), table filtering/sorting, and CSV/JSON/SQL import and export.

## Executive summary

- Virtualize both vertical rows and wide horizontal column windows using `ui/virtual` so 100,000+ rows scroll smoothly with constant DOM/display-list size.
- Stage all inline edits in a visible diff changeset before touching the database, with a SQL preview modal and transactional commit/rollback.
- Support everyday data workflows: quick `WHERE` filter bar, header click sorting, cell/row clipboard copy, and CSV/JSON/NDJSON/SQL export plus CSV import.

---

## Phase 4: Virtualized data grid, staged edits, and import/export

### 4.1 Virtualized result grid and typed cell rendering

- [ ] `QRY-37` Build the virtualized data grid in `examples/builder/quarry/ui/grid.go` using `ui/virtual` with sticky column headers (showing column name, data type subtitle, and `PK`/`FK` badge), sticky row-number gutter, vertical + horizontal wheel/trackpad scrolling, and drag-resizable column widths. Proof: `TestVirtualizedGridWindowing` loads 100,000 synthetic rows and verifies rendered row count stays bounded to viewport + overscan.
- [ ] `QRY-38` Implement typed cell rendering in `ui/grid.go`: styled `NULL` pill (`--qr-null-bg`/`--qr-null-fg`), right-aligned monospace numbers, boolean check/badge, formatted timestamps, compact JSON preview pill, and byte-length badge for BLOBs (`<BLOB 4.2 KB>`). Proof: `TestTypedCellRendering` asserts display output for every `CellValue` kind.
- [ ] `QRY-39` Add grid selection and keyboard navigation: single cell selection, row selection, arrow keys, `PageUp`/`PageDown`, `Home`/`End`, and copy (`Ctrl+C`) formatting single cells as plain text and multi-cell/row selections as TSV (`ownframe.UseMemory(true)` in tests). Proof: `TestGridNavigationAndCopy` passes.

### 4.2 Cell detail and JSON inspector drawer

- [ ] `QRY-40` Build the right-hand Cell Inspector drawer in `examples/builder/quarry/ui/inspector.go` that opens on double-click or toggle button, displaying full unwrapped text, pretty-printed/indented JSON with syntax colors (when cell content is valid JSON), byte size, and editable textarea when the table is writable. Proof: `TestJSONInspectorDrawer` pretty-prints a JSON payload from `audit_events` and applies an edit back to the staged changeset.

### 4.3 Table filter, sort, and pagination

- [ ] `QRY-41` Implement Table Data tab controls in `ui/grid.go`: quick `WHERE` filter bar (e.g. `status = 'paid' AND total > 100`), column header click sorting (cycles `ASC -> DESC -> none`), page-size selector (`100 / 500 / 1,000`), and `Prev / Next` pagination backed by safe parameterized/quoted queries. Proof: `TestTableFilterSortPagination` filters and sorts the seeded `orders` table and asserts row ordering and count.

### 4.4 Staged inline editing and transactional commit

- [ ] `QRY-42` Implement `examples/builder/quarry/sqlkit/changeset.go` tracking staged cell edits (`UpdateCell`), staged new rows (`InsertRow`), and staged deletions (`DeleteRow`), keyed by primary key (or SQLite `rowid` / Postgres `ctid` fallback). Generate dialect-correct SQL (`?` vs `$1` placeholders, proper identifier quoting) and refuse edits when `ConnectionConfig.ReadOnly` is true or no row identifier exists. Proof: `sqlkit/changeset_test.go` verifies SQL generation for both SQLite and PostgreSQL dialects.
- [ ] `QRY-43` Wire inline grid editing and staged-diff highlighting in `ui/grid.go`: double-clicking a writable cell opens an inline editor (with `Set NULL` button), modified cells tint with `--qr-diff-edit-bg`, inserted rows tint with `--qr-diff-insert-bg`, deleted rows tint with `--qr-diff-delete-bg`, and a floating Changeset Bar appears (`N pending changes`, `Preview SQL`, `Commit`, `Discard`). Proof: `TestStagedEditCommitAndDiscard` edits a row in the seeded SQLite DB, previews SQL, commits inside a transaction, and verifies the persisted value; also tests Discard restores original state.
- [ ] `QRY-44` Add transactional rollback handling on commit failure: if any statement in the staged changeset fails (e.g. constraint violation), roll back the entire transaction, keep the staged changes intact, and display the exact failing statement and DB error. Proof: `TestChangesetCommitRollbackOnError` triggers a `UNIQUE` / `CHECK` violation and verifies DB state is unchanged.

### 4.5 Import, export, and clipboard interop

- [ ] `QRY-45` Implement data export in `examples/builder/quarry/sqlkit/io.go` supporting CSV, JSON (pretty array), NDJSON, and SQL `INSERT INTO <table> ...` statements for either the current query result or a full table, with a clean Export Dialog (`ui/dialogs.go`) offering delimiter, header row, and NULL-representation options. Proof: `sqlkit/io_test.go` verifies round-trip CSV, valid JSON/NDJSON, and executable SQL `INSERT` output.
- [ ] `QRY-46` Implement CSV import in `sqlkit/io.go` + `ui/dialogs.go`: parse header row, infer SQLite/Postgres column types (`INTEGER`, `REAL`, `TEXT`, `BOOLEAN`), preview the first 10 rows, allow target table name (create new or append to existing), and batch-insert inside a single transaction. Proof: `TestCSVImportRoundTrip` imports a multi-type CSV into SQLite and verifies row count and types.
- [ ] `QRY-47` Add grid context menu (`Copy Value`, `Copy Row as JSON`, `Copy Row as INSERT`, `Copy Selection as Markdown Table`, `Set NULL`, `Duplicate Row`, `Delete Row`) using `ownframe.UseMemory(true)` in tests. Proof: `TestGridContextMenuCopyFormats` verifies JSON, `INSERT`, and Markdown output.
- [ ] `QRY-48` Run Phase 04 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/quarry/...`, `go test -v ./examples/builder/quarry/...`, and `make internal-import-check`. Proof: all commands exit 0 with results recorded beside this row.

---

## Dependencies

- `QRY-31` (Phase 03) must land before `QRY-37` and `QRY-41`.
- `QRY-42` must land before `QRY-40`, `QRY-43`, and `QRY-44`.
- All clipboard tests in `QRY-39` and `QRY-47` must call `ownframe.UseMemory(true)`.
