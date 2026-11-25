# v0.0.8 - Phase 02: Virtualized grid, selection, and navigation

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 15 open)
> **Estimated effort:** 3-4 days

---

## Overview

Phase 02 builds the interactive spreadsheet grid surface. A spreadsheet with 1,000 rows and 52 columns has 52,000 cells, so rendering the entire sheet as DOM nodes is out of the question. Instead, the grid renders only the visible row and column window plus a small overscan, keeps sticky row and column headers aligned to pixel boundaries, supports freeze panes, and delivers Excel-grade pointer and keyboard selection.

## Executive summary

- Virtualize both axes (rows and columns) inside the main grid viewport, computing exact pixel offsets from variable column widths and row heights.
- Render sticky column headers (`A`, `B`, `C`...), sticky row headers (`1`, `2`, `3`...), the top-left Select-All corner triangle, and active-header highlighting.
- Implement active cell focus, single and multi-range selection, drag-selection, Shift+click extension, Name Box jump, and full keyboard navigation (Arrows, Tab, Enter, Home/End, Ctrl+Arrows, PageUp/PageDown).
- Support interactive column-width and row-height dragging, double-click auto-fit, wheel scrolling (vertical + Shift+wheel horizontal), and custom scrollbars.

---

## Phase 2: Virtualized grid, selection, and navigation

### 2.1 Two-axis virtualized grid rendering

- [ ] `EXL-14` Implement `gridMetrics` in `examples/builder/excel/scroll.go` computing cumulative column X offsets and row Y offsets from `Sheet` overrides (default 84 px width, 22 px height, skipping hidden rows/columns) and returning the visible `[firstCol..lastCol]` and `[firstRow..lastRow]` window for any `(scrollX, scrollY, viewportW, viewportH)`. Proof: unit tests in `scroll_test.go` verify O(log N) or prefix-sum lookups with variable widths, hidden rows, and 10,000-row sheets.
- [ ] `EXL-15` Build the virtualized grid widget in `examples/builder/excel/grid.go` rendering only visible cells positioned inside the grid canvas, with optional overscan of 2 rows and 1 column so wheel scrolling stays smooth without exploding node count. Proof: `TestGridNodeBudget` asserts rendered cell element count stays under 400 for a 1440x960 viewport on a 1,000x52 sheet.
- [ ] `EXL-16` Render sticky column headers (`A..Z, AA..`), sticky row headers (`1..N`), and the top-left Select-All corner button in `grid.go`, keeping headers locked to their axis while scrolling the opposite axis and highlighting the header span that intersects the current selection. Proof: `TestStickyHeadersAndActiveHighlight` checks `page.Probe().Display()` boxes before and after scrolling by 400 px horizontally and vertically.
- [ ] `EXL-17` Render merged cell regions in `grid.go` spanning their combined width and height, suppressing covered interior cells and routing clicks inside the merged area to the top-left anchor cell. Proof: `TestMergedCellLayoutAndHitTest` verifies merged box dimensions and click target resolution.

### 2.2 Freeze panes and scrolling

- [ ] `EXL-18` Implement freeze panes (`Freeze Top Row`, `Freeze First Column`, `Freeze Panes at Active Cell`, `Unfreeze`) in `grid.go` by splitting the viewport into up to four panes (top-left frozen corner, frozen top rows, frozen left columns, scrollable body) separated by a 1 px `--xl-gridline-frozen` divider. Proof: `TestFreezePanesStayPinned` scrolls the body by 500 px and verifies frozen row 1 and column A coordinates remain fixed on `page.Probe().Display()`.
- [ ] `EXL-19` Wire `Wheel` in `app.go` / `scroll.go` for vertical scrolling, `Shift+Wheel` (or horizontal trackpad `dx`) for horizontal scrolling, and clamp offsets to exact sheet content bounds. Proof: `TestGridWheelScroll` drives vertical and horizontal wheel deltas and asserts scroll offsets and visible row/column indices.
- [ ] `EXL-20` Render horizontal and vertical grid scrollbars with proportional thumbs that support click-on-track and `DragStart`/`DragMove`/`DragEnd` thumb dragging. Proof: `TestScrollbarDragUpdatesViewport` drags both thumbs and verifies scroll offset updates and redraws.

### 2.3 Selection model and pointer interactions

- [ ] `EXL-21` Implement `SelectionState` in `examples/builder/excel/selection.go` tracking `Active CellRef`, `Anchor CellRef`, `Ranges []RangeRef` (supporting non-contiguous `Ctrl+Click` ranges), and selection mode (`Cells`, `Rows`, `Cols`, `All`). Proof: unit tests in `selection_test.go` cover single cell, rectangular drag, Shift extension, Ctrl multi-range, and full row/column selection.
- [ ] `EXL-22` Wire pointer selection in `app.go` and `selection.go`: single click sets active cell, `Shift+Click` extends range from anchor, `Press` + `DragMove` + `DragEnd` sweeps a live rectangular selection, clicking a column or row header selects the entire column/row, and clicking the top-left corner selects all cells. Proof: `TestPointerSelectionFlows` drives click, Shift+click, drag-sweep, header click, and corner click headlessly.
- [ ] `EXL-23` Paint the active cell 2 px `--xl-selection-border` outline, translucent `--xl-selection-fill` range overlay, and the 6x6 px bottom-right AutoFill square handle in `grid.go`. Proof: `TestSelectionOverlayGeometry` asserts overlay box bounds align with the selected cell boxes in `page.Probe().Display()`.

### 2.4 Keyboard navigation and resizing

- [ ] `EXL-24` Implement full Excel keyboard navigation in `selection.go` and `app.go`: `Arrow` keys move active cell (auto-scrolling viewport when leaving visible bounds), `Shift+Arrow` extends range, `Tab`/`Shift+Tab` moves horizontally (wrapping inside multi-cell selections), `Enter`/`Shift+Enter` moves vertically, `Home` jumps to column A, `Ctrl+Home` jumps to `A1`, `Ctrl+End` jumps to last populated cell, `PageUp`/`PageDown` scrolls one viewport height, and `Ctrl+Arrow` jumps to the next data region edge. Proof: `TestKeyboardNavigationMatrix` covers every shortcut in headless tests.
- [ ] `EXL-25` Implement interactive column-width and row-height resizing via drag handles on header borders (`DragStart`/`DragMove`/`DragEnd`), clamping minimums (24 px width, 16 px height) and showing a live size tooltip (`Width: 112 px`). Proof: `TestHeaderResizeDrag` drags column B divider by +40 px and verifies updated width in `Sheet` and `page.Probe().Display()`.
- [ ] `EXL-26` Implement double-click on a column or row header border to auto-fit that column or row to its widest/tallest visible content plus padding. Proof: `TestHeaderAutoFit` populates a long string in column C, triggers auto-fit, and verifies column C expands to fit.
- [ ] `EXL-27` Wire the Name Box in `formula_bar.go` to display the current cell or range (`A1`, `B2:D8`, or `5R x 3C` during drag) and accept typed cell/range references (`G45`, `A1:C10`) on Enter to jump and select. Proof: `TestNameBoxJump` types `D25` + Enter and asserts active cell and scroll offset.

### 2.5 Phase gate

- [ ] `EXL-28` Run Phase 02 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/excel`, `go test -v ./examples/builder/excel/...`, `make lint`, and inspect a `-web` `/frame.png` screenshot of the grid in both Light and Dark themes. Proof: commands exit 0 and visual check confirms aligned headers, gridlines, and selection border.

---

## Dependencies

- Requires Phase 01 (`EXL-02`, `EXL-04`, `EXL-05`, `EXL-10`, `EXL-11`).
- `EXL-14` (`gridMetrics`) is required by `EXL-15`, `EXL-16`, `EXL-18`, `EXL-19`, and `EXL-25`.
- `EXL-21` (`SelectionState`) is required by `EXL-22`, `EXL-23`, `EXL-24`, `EXL-27`, and all of Phase 03.
