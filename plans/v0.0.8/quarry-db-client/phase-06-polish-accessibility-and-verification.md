# v0.0.8 - Phase 06: Command palette, accessibility, benchmarks, and verification

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 10 open)
> **Estimated effort:** 2 days

---

## Overview

Phase 06 polishes Quarry into a showcase-ready desktop tool: adds a fast `Ctrl+K` command palette, draggable split panes, full keyboard and semantic accessibility coverage, production read-only safety polish, visual pixel tests for dark and light themes, performance benchmarks, and repository registration.

## Executive summary

- Wire `Ctrl+K` / `Ctrl+P` fuzzy command palette for instant table jump, connection switch, theme toggle, new query, and monitor view.
- Prove layout stability, theme contrast, 100k-row grid paint speed, and monitor tick cost with automated tests and benchmarks.
- Register `examples/builder/quarry` in `examples/readme.md` (including the Motion audit table) and pass every repo quality gate.

---

## Phase 6: Command palette, accessibility, benchmarks, and verification

### 6.1 Command palette and split-pane ergonomics

- [ ] `QRY-59` Implement the fuzzy Command Palette in `examples/builder/quarry/ui/palette.go` triggered by `Ctrl+K` or `Ctrl+P` (or title-bar search pill), indexing all open connections, introspected tables/views, saved snippets, and built-in actions (`Toggle Dark/Light Theme`, `Open Live Monitor`, `New SQL Query`, `Export Current Results`, `Toggle Read-Only Mode`). Proof: `TestCommandPaletteSearchAndRun` filters tables and triggers theme toggle via keyboard.
- [ ] `QRY-60` Add draggable split-pane handles (`DragStart`/`DragMove`/`DragEnd`) for resizing the schema sidebar width (180 px to 440 px), the vertical split between SQL editor and results grid, and the Cell Inspector drawer width, persisting split ratios in `store/store.go`. Proof: `TestSplitPaneDragClampsAndPersists` verifies min/max clamping and persistence.
- [ ] `QRY-61` Polish `Prod` environment guardrails: display a clear top read-only banner (`Production connection - Read-Only Mode active`) on `Prod`-tagged connections, require an explicit unlock confirmation to enable writes, and require typing the table name before destructive actions (`TRUNCATE` / `DROP`). Proof: `TestProdReadOnlyGuardrail` verifies writes stay blocked until explicitly unlocked.

### 6.2 Accessibility and keyboard coverage

- [ ] `QRY-62` Audit and wire semantic labels (`aria-label`, roles, focusable tab order, visible focus outlines) across the title bar, connection rail, schema tree, SQL editor, virtualized grid, and monitor KPI cards; verify via `page.Access().Snapshot()`. Proof: `TestAccessibilitySnapshot` asserts all primary controls have non-empty names and valid bounds, and password fields in the connection modal remain redacted.
- [ ] `QRY-63` Verify full keyboard-only workflow (`Ctrl+K` jump to table -> navigate grid -> `Ctrl+N` new query -> type SQL -> `Ctrl+Enter` run -> `Ctrl+Shift+T` toggle theme -> `Ctrl+W` close tab). Proof: `TestKeyboardOnlyWorkflow` passes end-to-end in headless mode.

### 6.3 Visual pixel verification and performance benchmarks

- [ ] `QRY-64` Add headless pixel verification tests (`pixels_test.go`) rendering Quarry at 1440x960 and 1280x800 in both `QuarryDark` (default) and `QuarryLight` across the SQL Workbench, Table Data + Changeset, and Live Monitor views, asserting expected theme background/accent pixels and zero overflow. Proof: `go test -v ./examples/builder/quarry -run TestPixels` passes.
- [ ] `QRY-65` Add Go benchmarks (`bench_test.go`) measuring `BenchmarkGridScroll100kRows` (virtualized window scroll repaint), `BenchmarkSQLHighlightLargeScript` (500-line SQL script), and `BenchmarkMonitorTickDrain` (60-sample sparkline update), recording measured `ns/op` and `allocs/op`. Proof: `go test ./examples/builder/quarry -run '^$' -bench . -benchmem` completes cleanly with numbers recorded beside this row.
- [ ] `QRY-66` Drive Quarry live via `-web` (`127.0.0.1:8150`) and desktop `-snapshot` mode per `documentation/testing.md`, inspecting `/frame.png` screenshots for dark mode, light mode, SQL highlight readability, grid alignment, and monitor charts. Proof: inspected screenshots show zero clipped text or layout defects in both desktop and `-web`.

### 6.4 Documentation and final repo gates

- [ ] `QRY-67` Add `examples/builder/quarry/README.md` (usage, flags, shortcuts, SQLite/Postgres setup, dark/light theme screenshots, recovery/platform notes) and register `builder/quarry` in `examples/readme.md` (Feature examples table with port `8150` and Motion audit table explaining `SetTick` channel draining). Proof: `make docs-check` and `make examples-check` exit 0.
- [ ] `QRY-68` Run final repository closure gates: `gofmt -l examples/builder/quarry`, `make lint`, `make test`, `make graphics-check`, `make plans-check`, `make internal-import-check`, and `make repo-size-check`. Proof: every gate exits 0 with results recorded beside this row.

---

## Dependencies

- Phases 01 through 05 must be complete before `QRY-59`, `QRY-63`, `QRY-64`, and `QRY-68`.
- `QRY-67` must land before `QRY-68` so `make docs-check` and `make examples-check` validate the registered catalog entries.
