# v0.0.8 - Phase 06: Charts, validation, accessibility, and verification

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 12 open)
> **Estimated effort:** 3-4 days

---

## Overview

Phase 06 completes the Excel experience with range-bound embedded charts, Data Validation rules, Cell Comments/Notes, View ribbon toggles (`Gridlines`, `Headings`, `Formula Bar`, `Show Formulas Ctrl+~`), accessibility labels, performance benchmarks at scale, and full visual + automated verification gates.

## Executive summary

- Embed movable, live-updating charts (Column, Bar, Line, Area, Pie/Donut) bound to sheet cell ranges using `components/chadcn` chart components and builder layout patterns from `examples/builder/charts/`.
- Add Data Validation (List dropdowns, Whole Number / Decimal bounds, Custom formula) and Cell Comments/Notes with hover popovers.
- Wire View ribbon toggles (`Gridlines`, `Headings`, `Formula Bar`, `Show Formulas`, `Freeze Panes`, `Theme`) and `Review` ribbon comment navigation.
- Prove correctness, accessibility, frame-time performance, and visual quality across Light and Dark themes via headless tests, paint benchmarks, `-web` `/frame.png` captures, and `make lint` / `make test` / `make graphics-check`.

---

## Phase 6: Charts, validation, accessibility, and verification

### 6.1 Embedded live charts (`Insert` ribbon)

- [ ] `EXL-73` Define `ChartSpec` in `examples/builder/excel/charts.go` storing chart ID, sheet name, chart type (`Column`, `Bar`, `Line`, `Area`, `Pie`), title, source `RangeRef`, position `(X, Y, W, H)`, and theme palette bindings, updating chart series automatically whenever any source cell recalculates. Proof: `charts_test.go` edits a source cell in `Sales Q1-Q4` and verifies chart series values update on the next render.
- [ ] `EXL-74` Render embedded floating chart cards over the grid canvas in `charts.go` using `components/chadcn` chart/area-chart/bar primitives, supporting click-to-select, header drag-to-move (`DragStart`/`DragMove`/`DragEnd`), delete (`Delete` key or close icon), and `Insert` ribbon buttons to create a chart from the current selection. Proof: `TestInsertAndMoveChart` selects `A1:C6`, inserts a Column chart, drags it +80 px, and verifies bounds on `page.Probe().Display()`.

### 6.2 Data validation, comments, and View/Review ribbons

- [ ] `EXL-75` Implement Data Validation in `workbook.go`, `dialogs.go`, and `grid.go`: `List` (inline dropdown chevron on the active cell), `Whole Number` / `Decimal` min-max bounds, and `Text Length`, rejecting invalid commits with a `chadcn.Dialog` or `chadcn.Toast` error alert. Proof: `TestDataValidationListAndBounds` picks a list option from the cell dropdown and verifies an out-of-range number is rejected.
- [ ] `EXL-76` Implement Cell Comments/Notes (`Shift+F2` or `Review -> New Comment`) painting the `--xl-comment-corner` top-right red triangle on commented cells and showing a `chadcn.HoverCard` / `Popover` on hover or focus, with `Previous`/`Next`/`Delete Comment` buttons on the `Review` ribbon. Proof: `TestCellCommentsAddHoverDelete` attaches a comment to `C4`, verifies the corner indicator and popover text, and deletes it.
- [ ] `EXL-77` Wire `View` ribbon toggles in `ribbon.go` and `grid.go`: `Gridlines` on/off, `Headings` (row/col headers) on/off, `Formula Bar` on/off, `Show Formulas` (`Ctrl+~` displaying raw formulas instead of values), and `Zoom to Selection` / `Reset 100%`. Proof: `TestViewRibbonToggles` toggles each flag and checks `page.Probe().Display()` changes.

### 6.3 Accessibility and keyboard discoverability

- [ ] `EXL-78` Audit and add ARIA roles, labels, and `aria-selected` / `aria-expanded` attributes across the shell, ribbon buttons, Name Box, formula bar, grid cells (`role="gridcell"`, `aria-label="B4: $1,250.00"`), sheet tabs, and dialogs so `page.Access()` produces a clean accessibility tree. Proof: `TestAccessibilitySnapshot` inspects `page.Access()` for active cell, ribbon tabs, and open dialog labels.
- [ ] `EXL-79` Add a Keyboard Shortcuts quick-reference dialog (`?` or `Ctrl+/`) in `dialogs.go` using `chadcn.Kbd` chips grouped by Navigation, Selection, Editing, Formatting, and Data. Proof: `TestShortcutsDialogOpensAndCloses` opens with `Ctrl+/` and closes with `Escape`.

### 6.4 Performance budgets, benchmarks, and visual verification

- [ ] `EXL-80` Add `examples/builder/excel/paint_bench_test.go` benchmarking: (1) full initial mount of a 1,000-row x 26-column populated sheet, (2) single-cell edit + dependency recalc + redraw, (3) 1-row wheel scroll redraw, and (4) Light/Dark theme switch. Record exact `ns/op` and `allocs/op` on `go test -bench=. -benchmem ./examples/builder/excel`. Proof: benchmark output recorded beside this row with steady-state scroll/edit under target budget (<16 ms/frame).
- [ ] `EXL-81` Add pixel regression tests (`examples/builder/excel/theme_pixels_test.go`) asserting Light and Dark theme colors on title bar, active ribbon tab, grid canvas, active cell selection border, and conditional formatting data bars. Proof: `go test -v -run TestThemePixels ./examples/builder/excel` exits 0.
- [ ] `EXL-82` Drive the running app with `-web` and inspect `/frame.png` captures across key states (default Light `Sales Q1-Q4`, Dark theme, active formula editing with colored reference highlights, Conditional Formatting + AutoFilter popover open, embedded Column chart on `Summary`). Compare description, painted picture, and readout per `documentation/testing.md`. Proof: visual review notes recorded beside this row with zero overlapping chrome or clipped headers.
- [ ] `EXL-83` Update `examples/readme.md` and `examples/builder/excel/README.md` with run commands (`go run ./examples/builder/excel`, `-web`, `-dev`, `-perf`), feature summary, and keyboard shortcuts. Proof: `make docs-check` exits 0.

### 6.5 Final release gate

- [ ] `EXL-84` Run the full repository verification gates: `gofmt`, `make build`, `make lint`, `make test`, `make test-race`, `make graphics-check`, `make docs-check plans-check examples-check tools-check repo-size-check`. Proof: all commands exit 0 and every phase table count in `plans/v0.0.8/excel-example/README.md` is synchronized.

---

## Dependencies

- Requires Phase 04 (`EXL-45` through `EXL-57`) for ribbon chrome, formatting, and theme palettes.
- Requires Phase 05 (`EXL-59` through `EXL-71`) for multi-sheet navigation, Undo/Redo, AutoFilter, and Status Bar.
- `EXL-80`, `EXL-81`, and `EXL-82` must pass before closing `EXL-84`.
