# v0.0.8 - Phase 02: SQLite & PostgreSQL drivers and schema explorer

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 12 open)
> **Estimated effort:** 2-3 days

---

## Overview

Phase 02 implements the unified database `Driver` layer for both SQLite (`modernc.org/sqlite`) and PostgreSQL (`github.com/jackc/pgx/v5`), builds the collapsible schema explorer tree in the left sidebar, and adds the table Structure / DDL inspector tab.

## Executive summary

- Define a clean `Driver` interface with context cancellation, query row caps, read-only enforcement, and typed cell decoding (`NULL`, int, float, text, bool, time, JSON, bytes).
- Implement full SQLite and PostgreSQL introspection (schemas, tables, views, materialized views, indexes, triggers, foreign keys, and DDL).
- Provide a deterministic mock/stub PostgreSQL driver so all CI and headless tests exercise PostgreSQL UI paths without requiring a live Postgres server, alongside an opt-in `QUARRY_TEST_PG_DSN` integration suite.

---

## Phase 2: SQLite & PostgreSQL drivers and schema explorer

### 2.1 Unified driver interface and typed cell values

- [ ] `QRY-13` Implement `examples/builder/quarry/driver/driver.go` with the `Driver` interface, `CellValue` typed wrapper (`Null`, `Int`, `Float`, `Text`, `Bool`, `Time`, `JSON`, `Blob`), `SchemaNode`, `ColumnMeta`, `IndexMeta`, `ForeignKeyMeta`, and `QueryResult`. Enforce read-only mode by rejecting non-read statements (`INSERT`, `UPDATE`, `DELETE`, `DROP`, `ALTER`, `TRUNCATE`, `CREATE`) before execution when `ConnectionConfig.ReadOnly` is true. Proof: `driver/driver_test.go` verifies typed value formatting and read-only statement blocking.

### 2.2 SQLite driver implementation

- [ ] `QRY-14` Implement `examples/builder/quarry/driver/sqlite.go` using `modernc.org/sqlite` (`database/sql`) with connection pooling tuned for SQLite (WAL awareness, busy timeout, optional `?mode=ro`), `Ping`, `Query` with `maxRows` truncation detection, and `Exec`. Proof: `driver/sqlite_test.go` tests query execution, truncation flags, cancellation, and read-only mode against `driver/seed.go`.
- [ ] `QRY-15` Implement SQLite schema introspection in `driver/sqlite.go`: list tables, views, indexes, and triggers from `sqlite_master`; fetch columns via `PRAGMA table_info`, foreign keys via `PRAGMA foreign_key_list`, indexes via `PRAGMA index_list` / `index_info`, approximate row counts, and raw `CREATE TABLE` DDL. Proof: `driver/sqlite_test.go` verifies all metadata on the seeded demo database.
- [ ] `QRY-16` Implement SQLite `EXPLAIN QUERY PLAN` parsing in `driver/sqlite.go`, mapping SQLite's parent/id tree into the shared `ExplainPlan` structure with scan classification (`SCAN` vs `SEARCH USING INDEX`). Proof: `driver/sqlite_test.go` tests both full-table scan and indexed lookup plans.

### 2.3 PostgreSQL driver implementation and test adapter

- [ ] `QRY-17` Add `github.com/jackc/pgx/v5` to `examples/go.mod` only (verify root `go.mod` stays untouched) and implement `examples/builder/quarry/driver/postgres.go` using `pgxpool.Pool` with configurable SSLMode (`disable`, `prefer`, `require`, `verify-full`), connect timeout, `Ping`, `Query` (with `maxRows` cap and proper `pgtype` decoding for `jsonb`, `uuid`, `timestamptz`, `bytea`, arrays), and `Exec`. Proof: `go vet -p 1 ./examples/builder/quarry/...` exits 0 and `git diff go.mod` is empty.
- [ ] `QRY-18` Implement PostgreSQL schema introspection in `driver/postgres.go` querying `pg_namespace`, `pg_class`, `pg_attribute`, `pg_index`, `pg_constraint`, and `pg_views` to list schemas (excluding `pg_toast` by default), tables, views, materialized views, indexes, functions, estimated row counts (`reltuples`), and synthesized/fetched DDL. Proof: SQL queries compile and pass unit tests against the mock PG adapter plus opt-in `QUARRY_TEST_PG_DSN` integration test.
- [ ] `QRY-19` Implement PostgreSQL `EXPLAIN (FORMAT JSON)` and `EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` decoding in `driver/postgres.go` into the shared `ExplainPlan` tree (capturing node type, relation, startup/total cost, actual rows, actual time, and buffers). Proof: table-driven unit tests parse captured real Postgres JSON explain fixtures.
- [ ] `QRY-20` Implement a deterministic in-memory mock PostgreSQL driver (`examples/builder/quarry/driver/pgmock.go`, test/demo selectable) returning realistic multi-schema (`public`, `analytics`, `billing`) metadata, sample query results, `EXPLAIN` JSON plans, and live-looking telemetry snapshots so developers can test Postgres UI flows offline. Proof: `driver/pgmock_test.go` verifies every `Driver` method.

### 2.4 Schema explorer sidebar and structure inspector

- [ ] `QRY-21` Build the collapsible schema explorer in `examples/builder/quarry/ui/sidebar.go`: connection header with refresh button, instant filter input, expandable tree (`Schemas -> Tables / Views / Indexes / Functions`), row-count pills, and collapse/expand sidebar toggle. Proof: `TestSidebarTreeAndFilter` verifies expanding schemas, filtering tables by substring, and clicking a table to open it.
- [ ] `QRY-22` Add sidebar context menu actions (`Open Table Data`, `Inspect Structure / DDL`, `New Query for Table`, `Copy Qualified Name`, `Refresh Schema`) wired via `ContextMenuFor` and `MenuAction`. Proof: `TestSidebarContextMenu` exercises each action and asserts resulting tab or clipboard state (`ownframe.UseMemory(true)`).
- [ ] `QRY-23` Build the Structure / DDL inspector view in `examples/builder/quarry/ui/sidebar.go` (or `ui/structure.go`) displaying three sub-tabs: `Columns` (name, type, nullable, default, PK/FK badges), `Indexes & Keys` (unique flag, columns, referenced table), and `DDL` (syntax-highlighted `CREATE TABLE` statement with Copy DDL button). Proof: `TestStructureInspectorTab` verifies column, index, and DDL rendering on the seeded `orders` table.
- [ ] `QRY-24` Run Phase 02 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/quarry/...`, `go test -v ./examples/builder/quarry/...`, `make internal-import-check`, and verify root `go.mod` has no new driver dependencies. Proof: all commands exit 0 with results recorded beside this row.

---

## Dependencies

- `QRY-13` must land before `QRY-14` and `QRY-17`.
- `QRY-15` and `QRY-20` must land before `QRY-21` and `QRY-23` so the sidebar and structure inspector render real SQLite and mock-Postgres schemas.
