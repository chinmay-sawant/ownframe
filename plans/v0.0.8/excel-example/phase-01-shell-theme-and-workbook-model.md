# v0.0.8 - Phase 01: Shell, color schemas, and workbook model

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 13 open)
> **Estimated effort:** 2-3 days

---

## Overview

Phase 01 establishes the runnable builder application under `examples/builder/excel/`, the two complete Excel color schemas (`ExcelLight` and `ExcelDark`), the `CustomHeader` window chrome, and the pure-Go workbook data model. At the end of this phase, the app boots in both desktop (`ownframe.RunWithOptions`) and `-web` preview modes, toggles themes cleanly, and holds a realistic multi-sheet starter workbook.

## Executive summary

- Create `main.go`, `app.go`, `styles.go`, `shell.go`, `workbook.go`, and `cell.go` using the `ui` builder API and `components/chadcn` controls.
- Implement `ExcelLight` (Office Green `#107c41` chrome, white grid) and `ExcelDark` (`#185c37` brand accent, `#141417` grid canvas) with `--xl-*` tokens layered over `chadcn.Light` / `chadcn.Dark`.
- Build the sparse workbook model supporting cell coordinates, ranges, sheet metadata, default column widths and row heights, and a seeded multi-sheet demo dataset (Budget, Sales Q1-Q4, Inventory) so every later phase has realistic data to test against.

---

## Phase 1: Shell, color schemas, and workbook model

### 1.1 Application scaffold and custom window chrome

- [ ] `EXL-01` Create `examples/builder/excel/main.go` with `-web`, `-headless`, `-addr` (default `127.0.0.1:8142`), `-dev`, `-perf`, and `-debug-addr` flags, wiring `ownframe.RunWithOptions` with `CustomHeader: true`, `PlacementID: "excel"`, and a `Draggable` guard that yields when any popover, dialog, or menu is open. Proof: `go vet -p 1 ./examples/builder/excel` exits 0.
- [ ] `EXL-02` Create `examples/builder/excel/app.go` defining the `app` struct, `newApp()`, `ui.Mount` with `LockView: true`, `RetainState: true`, `Scrollbar: ownframe.ScrollbarHidden`, and `ownframe.Handlers` (`Click`, `Press`, `DragStart`, `DragMove`, `DragEnd`, `Wheel`, `Hover`, `KeyDown`, `KeyUp`, `Change`, `Type`). Proof: `TestNewAppMounts` mounts cleanly at 1440x960 and 1280x800 without error.
- [ ] `EXL-03` Build the Excel title bar in `examples/builder/excel/shell.go` with the app badge, Quick Access Toolbar (Save, Undo, Redo), editable/clickable workbook title pill (`Quarterly_Model.xlsx`), global search box trigger (`Ctrl+P` / command palette), theme toggle button, and custom window controls (`win-min`, `win-max`, `win-close` via `data-action`). Proof: `TestTitleBarActions` verifies window-control `data-action` attributes and theme toggle redraw.
- [ ] `EXL-04` Build the top-level 100vh vertical flex shell in `shell.go` with fixed slots for title bar (36 px), ribbon bar (88 px expanded / 32 px collapsed), formula bar (28 px), main grid viewport (`flex: 1; min-height: 0`), sheet tab strip (28 px), and status bar (24 px). Proof: `TestShellFitsViewport` asserts on `page.Probe().Display()` that all six bands fit inside 1440x960 with zero vertical overflow.

### 1.2 Color schemas and token sheets

- [ ] `EXL-05` Define `ExcelLight` and `ExcelDark` stylesheets in `examples/builder/excel/styles.go` implementing every `--xl-*` token specified in `README.md`, layered on top of `chadcn.Light + chadcn.Stylesheet` and `chadcn.Dark + chadcn.Stylesheet`. Proof: `TestThemeTokensPresent` checks both theme strings parse cleanly through `ui.Mount` with zero missing-color fallbacks.
- [ ] `EXL-06` Define the 6-color formula reference highlight palette (`--xl-ref-1` through `--xl-ref-6` border and fill tokens), conditional formatting preset tokens (red, yellow, green fills and text, data bars, 3-color scale), and the 10-column Office color swatch matrix in `styles.go` / `pickers.go`. Proof: `TestSwatchMatrixGeometry` verifies 10 theme columns x 5 rows + 10 standard colors with valid hex values.
- [ ] `EXL-07` Implement runtime light/dark theme switching via `app.setTheme(ctx, dark)` updating `page` theme and re-rendering without losing workbook state or active selection. Proof: `TestThemeTogglePixels` asserts distinct background pixels on the title bar and grid canvas between `ExcelLight` and `ExcelDark`.

### 1.3 Workbook, sheet, and sparse cell model

- [ ] `EXL-08` Implement cell coordinates and range helpers in `examples/builder/excel/cell.go`: `CellRef{Col, Row}`, `RangeRef{Start, End}`, `ColToLabel(0) == "A"`, `ColToLabel(26) == "AA"`, `ParseCellRef("B12")`, `ParseRangeRef("A1:D20")`, absolute/relative flags (`$A$1`, `$A1`, `A$1`), and `Normalized()` bounds. Proof: table-driven unit tests in `cell_test.go` covering single letters, multi-letter columns (`Z`, `AA`, `AZ`, `ZZ`, `XFD`), and invalid inputs.
- [ ] `EXL-09` Implement typed cell values in `cell.go`: `ValueKind` (`Empty`, `Number`, `String`, `Bool`, `Error`, `Date`), standard Excel error codes (`#DIV/0!`, `#VALUE!`, `#REF!`, `#NAME?`, `#N/A`, `#NUM!`, `#CIRC!`), raw input string (`Raw`), cached evaluated `Value`, `StyleID`, `Comment`, and `ValidationID`. Proof: unit tests in `cell_test.go` verifying zero-value empty semantics and error formatting.
- [ ] `EXL-10` Implement `Workbook` and `Sheet` in `examples/builder/excel/workbook.go` with sparse cell storage (`map[CellRef]*Cell`), deduplicated `StyleTable` (mapping `StyleID` to `CellStyle`), per-column width overrides, per-row height overrides, hidden row/column sets, merged cell list, freeze-pane coordinates (`FreezeRows`, `FreezeCols`), and sheet tab color. Proof: `workbook_test.go` tests sparse read/write, empty-cell deletion, and style deduplication.
- [ ] `EXL-11` Seed a realistic multi-sheet starter workbook (`seedWorkbook()`) containing three sheets: `Summary` (executive KPIs with cross-sheet `SUM`/`AVERAGE`/`IF` formulas), `Sales Q1-Q4` (40+ rows of regional revenue, cost, margin, YoY growth, and currency/percent formatting), and `Expenses` (categorized line items with dates and conditional formatting rules). Proof: `TestSeedWorkbookValid` verifies all seeded sheets, styles, and raw formulas load without error.

### 1.4 Ribbon strip skeleton and phase gate

- [ ] `EXL-12` Render the ribbon tab header bar (`File`, `Home`, `Insert`, `Page Layout`, `Formulas`, `Data`, `Review`, `View`) and collapse/expand chevron in `examples/builder/excel/ribbon.go` using `components/chadcn` tabs/buttons so switching tabs swaps the active ribbon body and collapsing reclaims vertical space for the grid. Proof: `TestRibbonTabSwitchAndCollapse` asserts active tab state and grid viewport height delta on `page.Probe().Display()`.
- [ ] `EXL-13` Run Phase 01 closure gates: `gofmt -w examples/builder/excel`, `go vet -p 1 ./examples/builder/excel`, `go test -v ./examples/builder/excel/...`, `make internal-import-check`, and `make examples-check`. Proof: all commands exit 0 with results recorded beside this row.

---

## Dependencies

- `EXL-05` and `EXL-06` must land before `EXL-07` so light/dark pixel assertions compare real theme tokens.
- `EXL-08`, `EXL-09`, and `EXL-10` must land before `EXL-11` and before Phase 02 binds the virtualized grid to `Sheet`.
- Do not import any package under `github.com/chinmay-sawant/ownframe/internal/` in any file created in this phase.
