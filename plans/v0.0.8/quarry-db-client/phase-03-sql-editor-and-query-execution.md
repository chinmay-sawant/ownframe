# v0.0.8 - Phase 03: SQL editor, autocomplete, async runner, and EXPLAIN

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 12 open)
> **Estimated effort:** 2-3 days

---

## Overview

Phase 03 turns Quarry into a real SQL workbench: a tabbed workspace, a dialect-aware syntax-highlighted SQL editor, schema-aware autocomplete, non-blocking async query execution with cancellation, persistent query history and snippets, and a visual `EXPLAIN` plan tree.

## Executive summary

- Support multiple open tabs (SQL Query, Table Data, Structure, Monitor) with close, reorder/switch, and unsaved dirty dots.
- Build a fast, pure-Go SQL tokenizer supporting both SQLite and PostgreSQL keywords, functions, dollar-quoted strings (`$$...$$`), comments, and parameter placeholders.
- Execute queries on background goroutines so long-running queries never stall the Ebiten frame loop and can be cancelled mid-flight.

---

## Phase 3: SQL editor, autocomplete, async runner, and EXPLAIN

### 3.1 Multi-tab workspace

- [ ] `QRY-25` Implement the workspace tab bar in `examples/builder/quarry/ui/tabs.go` supporting `QueryTab`, `TableTab`, `StructureTab`, and `MonitorTab` with `+ New Query` button, per-tab close button (`Ctrl+W`), tab cycling (`Ctrl+Tab`), active tab accent underline, and dirty indicator on edited SQL tabs. Proof: `TestWorkspaceTabLifecycle` creates, switches, and closes multiple tab types cleanly.

### 3.2 SQL tokenizer, syntax highlighter, and editor

- [ ] `QRY-26` Implement a dialect-aware SQL lexer/highlighter in `examples/builder/quarry/sqlkit/highlight.go` classifying keywords, built-in functions, types, single/double/dollar-quoted strings, numbers, `--` and `/* ... */` comments, and operators mapped to the `--qr-syn-*` tokens in `ui/styles/`. Proof: table-driven unit tests in `sqlkit/highlight_test.go` verify SQLite and Postgres edge cases (nested comments, `$$` blocks, escaped quotes, `::jsonb` casts).
- [ ] `QRY-27` Build the SQL editor pane in `examples/builder/quarry/ui/editor.go` with line-number gutter, syntax-highlighted display, active-line highlight, statement formatting/indent helper (`Format SQL`), and editor toolbar (`Run`, `Run Selection`, `Explain`, `Explain Analyze`, `Cancel`, `Save Snippet`, row-limit selector `100 / 1,000 / 10,000 / Unlimited`). Proof: `TestEditorToolbarAndShortcuts` verifies `Ctrl+Enter`, `Ctrl+Shift+Enter`, and toolbar actions.
- [ ] `QRY-28` Implement statement-at-cursor extraction in `examples/builder/quarry/sqlkit/statement.go` so pressing `Ctrl+Enter` runs only the semicolon-delimited statement containing the caret (or the active text selection if non-empty), while `Ctrl+Shift+Enter` runs the full script sequentially. Proof: `sqlkit/statement_test.go` covers multi-statement buffers, semicolons inside string literals, and comments.

### 3.3 Schema-aware autocomplete

- [ ] `QRY-29` Implement context-aware completion engine in `examples/builder/quarry/sqlkit/autocomplete.go` suggesting SQL keywords, schema/table names after `FROM` / `JOIN` / `UPDATE` / `INTO`, and column names after `SELECT` / `WHERE` / `ORDER BY` / `<alias>.` using the active connection's introspected schema. Proof: `sqlkit/autocomplete_test.go` tests prefix filtering, alias resolution (`SELECT o.| FROM orders o`), and ranking.
- [ ] `QRY-30` Render the floating autocomplete popover in `ui/editor.go` triggered by `Ctrl+Space` or identifier typing, navigable with `Up`/`Down`/`Enter`/`Tab`/`Escape`, showing item kind badges (`TBL`, `COL`, `KW`, `FN`) and column type hints. Proof: `TestAutocompletePopoverSelection` inserts the selected completion into the editor buffer.

### 3.4 Async query runner, history, and snippets

- [ ] `QRY-31` Wire non-blocking async query execution in `examples/builder/quarry/ui/app.go`: dispatch `Driver.Query` / `Driver.Exec` on a goroutine with a cancellable `context.Context`, show an inline running spinner + elapsed timer, drain completion messages in `SetTick`, and wire the `Cancel` button (`Escape` or click) to cancel the context (and call `CancelQuery` on Postgres). Proof: `TestAsyncQueryAndCancel` runs a slow query, cancels it mid-flight, and verifies the UI stays responsive and reports cancellation.
- [ ] `QRY-32` Render structured query error banners in the results pane showing SQLSTATE / error code (when available), message, line/column hint, and a quick button to jump the caret to the offending line. Proof: `TestQueryErrorBanner` runs invalid SQL against SQLite and verifies error text and display layout.
- [ ] `QRY-33` Build the Query History and Saved Snippets drawer backed by `store/store.go`: automatically log every executed query with timestamp, connection name, duration ms, row count, and success/error status; allow search filtering, starring into Snippets, and one-click loading into a new or current tab. Proof: `TestQueryHistoryAndSnippets` verifies recording, searching, and restoring a query.

### 3.5 Visual EXPLAIN plan viewer

- [ ] `QRY-34` Build the `EXPLAIN` / Query Plan sub-tab in `examples/builder/quarry/ui/explain.go` rendering the parsed `ExplainPlan` as an indented tree of cards showing operation type, target table/index, estimated/actual cost and rows, execution time bar, and warning pills (`Seq Scan` in `--qr-status-warn`, `Index Scan` in `--qr-status-ok`), plus a toggle to view raw text/JSON plan output. Proof: `TestExplainPlanView` renders both SQLite `EXPLAIN QUERY PLAN` and Postgres JSON explain fixtures cleanly.
- [ ] `QRY-35` Wire a `Messages` sub-tab alongside `Results` and `Explain` showing multi-statement execution summaries (`Statement 1: 42 rows in 1.8 ms`, `Statement 2: 3 rows affected`, warnings/notices). Proof: `TestMessagesSubTab` runs a 2-statement script and verifies both log entries.
- [ ] `QRY-36` Run Phase 03 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/quarry/...`, `go test -v ./examples/builder/quarry/...`, and `make internal-import-check`. Proof: all commands exit 0 with results recorded beside this row.

---

## Dependencies

- `QRY-14` and `QRY-15` (Phase 02) must land before `QRY-29` and `QRY-31`.
- `QRY-26` must land before `QRY-27` and `QRY-34` (raw DDL/plan highlighting).
- `QRY-31` must land before Phase 04 binds results to the virtualized grid and before Phase 05 links slow queries to "Open in Editor & Explain".
