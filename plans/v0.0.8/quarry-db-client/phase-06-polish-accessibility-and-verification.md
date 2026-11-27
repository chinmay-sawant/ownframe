# v0.0.8 - Phase 06: Command palette, accessibility, benchmarks, and verification

> **Parent:** [README.md](README.md)
> **Status:** Partial (8 complete, 2 partial, 0 open)
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

- [x] `QRY-59` Implement the fuzzy Command Palette in `examples/builder/quarry/ui/palette.go` triggered by `Ctrl+K` or `Ctrl+P` (or title-bar search pill), indexing all open connections, introspected tables/views, saved snippets, and built-in actions (`Toggle Dark/Light Theme`, `Open Live Monitor`, `New SQL Query`, `Export Current Results`, `Toggle Read-Only Mode`). Proof: `TestCommandPaletteSearchAndRun` filters tables and triggers theme toggle via keyboard. **UI:** `ui/palette.go`; `TestCommandPaletteSearchAndRun` passes. **Reopened and re-closed:** panel sized to its rows with an inner window and wheel scroll, outside click and Escape close, fixed badge width, visible hover tint. `TestPaletteHoverLayoutAndDismiss`, `TestPaletteInputTakesFocusOnOpen` pass.
- [x] `QRY-60` Add draggable split-pane handles (`DragStart`/`DragMove`/`DragEnd`) for resizing the schema sidebar width (180 px to 440 px), the vertical split between SQL editor and results grid, and the Cell Inspector drawer width, persisting split ratios in `store/store.go`. Proof: `TestSplitPaneDragClampsAndPersists` verifies min/max clamping and persistence. **UI:** `ui/drag.go`; `TestSplitPaneDragClampsAndPersists` passes.
- [x] `QRY-61` Polish `Prod` environment guardrails: display a clear top read-only banner (`Production connection - Read-Only Mode active`) on `Prod`-tagged connections, require an explicit unlock confirmation to enable writes, and require typing the table name before destructive actions (`TRUNCATE` / `DROP`). Proof: `TestProdReadOnlyGuardrail` verifies writes stay blocked until explicitly unlocked. **UI:** `TestProdReadOnlyGuardrail` passes (banner, blocked write, confirmed unlock, typed-name DROP).

### 6.2 Accessibility and keyboard coverage

- [x] `QRY-62` Audit and wire semantic labels (`aria-label`, roles, focusable tab order, visible focus outlines) across the title bar, connection rail, schema tree, SQL editor, virtualized grid, and monitor KPI cards; verify via `page.Access().Snapshot()`. Proof: `TestAccessibilitySnapshot` asserts all primary controls have non-empty names and valid bounds, and password fields in the connection modal remain redacted. **UI:** `TestAccessibilitySnapshot` and `TestGridCellsAndTabsInAccessibilityTree` pass. The engine's accessibility tree supports only main, navigation, group, heading, paragraph, img, link, button, textbox, checkbox, radio, combobox, list and listitem, so `role=grid/row/gridcell/columnheader/tab/tree/dialog` are ignored; Quarry maps them to list, listitem, group and button with explicit `aria-label`s, which puts grid rows and cells, column headers, tabs and tree rows in the tree. Password fields stay redacted.
- [x] `QRY-63` Verify full keyboard-only workflow (`Ctrl+K` jump to table -> navigate grid -> `Ctrl+N` new query -> type SQL -> `Ctrl+Enter` run -> `Ctrl+Shift+T` toggle theme -> `Ctrl+W` close tab). Proof: `TestKeyboardOnlyWorkflow` passes end-to-end in headless mode. **UI:** `TestKeyboardOnlyWorkflow` passes.

### 6.3 Visual pixel verification and performance benchmarks

- [x] `QRY-64` Add headless pixel verification tests (`pixels_test.go`) rendering Quarry at 1440x960 and 1280x800 in both `QuarryDark` (default) and `QuarryLight` across the SQL Workbench, Table Data + Changeset, and Live Monitor views, asserting expected theme background/accent pixels and zero overflow. Proof: `go test -v ./examples/builder/quarry -run TestPixels` passes. **UI:** `go test ./examples/builder/quarry -run TestPixels` passes (2 sizes x 2 themes x 3 views).
- [x] `QRY-65` Add Go benchmarks (`bench_test.go`) measuring `BenchmarkGridScroll100kRows` (virtualized window scroll repaint), `BenchmarkSQLHighlightLargeScript` (500-line SQL script), and `BenchmarkMonitorTickDrain` (60-sample sparkline update), recording measured `ns/op` and `allocs/op`. Proof: `go test ./examples/builder/quarry -run '^$' -bench . -benchmem` completes cleanly with numbers recorded beside this row. **Evidence:** `go test ./examples/builder/quarry -run '^$' -bench . -benchmem` exit 0 on i7-13700HX: `BenchmarkGridScroll100kRows` 5.5-7.9 ms/op, 12.6k allocs/op; `BenchmarkSQLHighlightLargeScript` 1.35-1.59 ms/op, 7,023 allocs/op; `BenchmarkMonitorTickDrain` 10.3-13.9 ms/op, 21.5k allocs/op (two runs). **Re-measured after the editor, grid, monitor and forms work (main session, quiet machine, load 0.2, `-count=5`):** `BenchmarkGridScroll100kRows` 8.0-8.4 ms/op (16.5k allocs), `BenchmarkSQLHighlightLargeScript` 1.25-1.29 ms/op (7,023 allocs), `BenchmarkMonitorTickDrain` 17.5-18.4 ms/op (27.5k allocs), `BenchmarkGridHorizontalScroll100kRows` 0.31-0.36 ms/op (2.8k allocs), `BenchmarkGridClientSort10kRows` 10.2-14.4 ms/op (16.4k allocs), sqlkit `BenchmarkTokenize` 0.42-0.58 ms/op. An earlier run on a loaded machine read 17-19 ms for the grid and 31-34 ms for the monitor, so those were load noise. Versus the first quiet run the grid is about flat to +5% and the monitor about +30%, consistent with the larger page (~490 elements). The cost is whole-page relayout and selector matching, not the visible-window row build.
- [~] `QRY-66` Drive Quarry live via `-web` (`127.0.0.1:8150`) and desktop `-snapshot` mode per `documentation/testing.md`, inspecting `/frame.png` screenshots for dark mode, light mode, SQL highlight readability, grid alignment, and monitor charts. Proof: inspected screenshots show zero clipped text or layout defects in both desktop and `-web`. **UI:** Added `-snapshot ADDR` (`snapshot.go`); the real WSLg window answered `GET /frame.png` and, being maximized, drew the Restore icon. Pointer and key injection through `scripts/gui-drive.sh` did not reach the window here (xdotool 3.2016 reports a stale window geometry; clicks and `--window` keys left the page unchanged), so desktop maximize/restore, hover, click, wheel and palette are **untested**. The same flows were driven through `-web` (`/click id=`, `/hover`, `/wheel`): grid scroll, inspector, monitor, palette and settings in dark and light, screenshots inspected. Next gate: a session where xdotool input reaches the window.

### 6.4 Documentation and final repo gates

- [x] `QRY-67` Add `examples/builder/quarry/README.md` (usage, flags, shortcuts, SQLite/Postgres setup, dark/light theme screenshots, recovery/platform notes) and register `builder/quarry` in `examples/readme.md` (Feature examples table with port `8150` and Motion audit table explaining `SetTick` channel draining). Proof: `make docs-check` and `make examples-check` exit 0. **UI:** `examples/builder/quarry/README.md` and the `examples/readme.md` Feature and Motion audit rows added; `make docs-check` and `make examples-check` exit 0.
- [x] `QRY-68` Run final repository closure gates: `gofmt -l examples/builder/quarry`, `make lint`, `make test`, `make graphics-check`, `make plans-check`, `make internal-import-check`, and `make repo-size-check`. Proof: every gate exits 0 with results recorded beside this row. **Evidence:** `gofmt -l`, `make lint`, `make test`, `make graphics-check`, `make plans-check`, `make docs-check`, `make examples-check`, `make internal-import-check` and `make repo-size-check` all exit 0 (final run after the lint pass and `ExecTx`).

---

## Dependencies

- Phases 01 through 05 must be complete before `QRY-59`, `QRY-63`, `QRY-64`, and `QRY-68`.
- `QRY-67` must land before `QRY-68` so `make docs-check` and `make examples-check` validate the registered catalog entries.
