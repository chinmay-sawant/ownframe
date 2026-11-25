# v0.0.8 - Phase 02: Virtualized grid, selection, and navigation

> **Parent:** [README.md](README.md)
> **Status:** In progress (13 complete, 2 partial, 0 open)
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

- [x] `EXL-14` Implement `gridMetrics` in `examples/builder/excel/scroll.go` computing cumulative column X offsets and row Y offsets from `Sheet` overrides (default 84 px width, 22 px height, skipping hidden rows/columns) and returning the visible `[firstCol..lastCol]` and `[firstRow..lastRow]` window for any `(scrollX, scrollY, viewportW, viewportH)`. Proof: unit tests in `scroll_test.go` verify O(log N) or prefix-sum lookups with variable widths, hidden rows, and 10,000-row sheets. **Evidence:** `TestDGridWindowVariableAxes` checks column offsets are prefix sums of custom widths, hidden rows take no space, a 40 px row shifts the next one, and a window scrolled to the bottom of the 1,000-row sheet holds only the rows that fit. The functions are `layoutCols`, `layoutRows` and `maxScroll` in `scroll.go` (no `gridMetrics` type, no `scroll_test.go`); the walk is O(rows on screen) from the scroll index, not a prefix-sum table.
- [x] `EXL-15` Build the virtualized grid widget in `examples/builder/excel/grid.go` rendering only visible cells positioned inside the grid canvas, with optional overscan of 2 rows and 1 column so wheel scrolling stays smooth without exploding node count. Proof: `TestGridNodeBudget` asserts rendered cell element count stays under 400 for a 1440x960 viewport on a 1,000x52 sheet. **Evidence:** `TestGridNodeBudget` and `TestDGridNodeBudgetFullSheet` (every 7th row, every 3rd column filled, at the top and at row 900) keep the cell element count under 400. There is no overscan: a row or column that does not fit whole is not drawn.
- [x] `EXL-16` Render sticky column headers (`A..Z, AA..`), sticky row headers (`1..N`), and the top-left Select-All corner button in `grid.go`, keeping headers locked to their axis while scrolling the opposite axis and highlighting the header span that intersects the current selection. Proof: `TestStickyHeadersAndActiveHighlight` checks `page.Probe().Display()` boxes before and after scrolling by 400 px horizontally and vertically. **Evidence:** `TestStickyHeadersAndActiveHighlight` and `TestDStickyHeadersHorizontalAndVertical` scroll to row 60 and column 8 and assert the row headers keep their x, the column headers keep their y, the corner stays, and only the headers of the active cell carry `xl-hdr--on`. Scrolling moves whole rows and columns, not 400 px.
- [x] `EXL-17` Render merged cell regions in `grid.go` spanning their combined width and height, suppressing covered interior cells and routing clicks inside the merged area to the top-left anchor cell. Proof: `TestMergedCellLayoutAndHitTest` verifies merged box dimensions and click target resolution. **Evidence:** `TestDMergedCellLayoutAndHitTest`: a B2:C3 merge draws as one box of the combined width and height, covered cells have no element, a pointer click and an id click on a covered cell select the anchor, the active outline spans the merge, and a merge whose anchor row scrolled off still draws once.

### 2.2 Freeze panes and scrolling

- [x] `EXL-18` Implement freeze panes (`Freeze Top Row`, `Freeze First Column`, `Freeze Panes at Active Cell`, `Unfreeze`) in `grid.go` by splitting the viewport into up to four panes (top-left frozen corner, frozen top rows, frozen left columns, scrollable body) separated by a 1 px `--xl-gridline-frozen` divider. Proof: `TestFreezePanesStayPinned` scrolls the body by 500 px and verifies frozen row 1 and column A coordinates remain fixed on `page.Probe().Display()`. **Evidence:** `TestFreezePanesStayPinned` and `TestDFreezeCommands` (Freeze Top Row, First Column, Panes at the active cell, Unfreeze through the View ribbon actions) pass; the divider is 2 px wide, not 1 px.
- [x] `EXL-19` Wire `Wheel` in `app.go` / `scroll.go` for vertical scrolling, `Shift+Wheel` (or horizontal trackpad `dx`) for horizontal scrolling, and clamp offsets to exact sheet content bounds. Proof: `TestGridWheelScroll` drives vertical and horizontal wheel deltas and asserts scroll offsets and visible row/column indices. **Evidence:** `TestGridWheelScroll`: vertical wheel moves 3 rows per notch, Shift+wheel moves columns, offsets clamp at the freeze boundary and the sheet end, and a wheel over the title bar is not consumed. Steps are whole rows and columns.
- [x] `EXL-20` Render horizontal and vertical grid scrollbars with proportional thumbs that support click-on-track and `DragStart`/`DragMove`/`DragEnd` thumb dragging. Proof: `TestScrollbarDragUpdatesViewport` drags both thumbs and verifies scroll offset updates and redraws. **Evidence:** `TestScrollbarDragUpdatesViewport` drags the thumb; `TestDScrollbarTrackClickPages` clicks both tracks and checks a one-page jump.

### 2.3 Selection model and pointer interactions

- [~] `EXL-21` Implement `SelectionState` in `examples/builder/excel/selection.go` tracking `Active CellRef`, `Anchor CellRef`, `Ranges []RangeRef` (supporting non-contiguous `Ctrl+Click` ranges), and selection mode (`Cells`, `Rows`, `Cols`, `All`). Proof: unit tests in `selection_test.go` cover single cell, rectangular drag, Shift extension, Ctrl multi-range, and full row/column selection. **Status:** Done: the app tracks `active`, `anchor` and one rectangular `sel` range with column, row and all-cells selection (`TestDSelectionModes`, `TestPointerSelectionFlows`). Open: there is no `SelectionState` type, no `selection_test.go`, no selection-mode enum, and no non-contiguous `Ctrl+Click` multi-range selection. Next gate: add `Ranges []Range` or drop multi-range from the row.
- [x] `EXL-22` Wire pointer selection in `app.go` and `selection.go`: single click sets active cell, `Shift+Click` extends range from anchor, `Press` + `DragMove` + `DragEnd` sweeps a live rectangular selection, clicking a column or row header selects the entire column/row, and clicking the top-left corner selects all cells. Proof: `TestPointerSelectionFlows` drives click, Shift+click, drag-sweep, header click, and corner click headlessly. **Evidence:** `TestPointerSelectionFlows` drives click, Shift+click, drag sweep (with the `3R x 3C` Name Box readout), column header, row header and corner clicks.
- [x] `EXL-23` Paint the active cell 2 px `--xl-selection-border` outline, translucent `--xl-selection-fill` range overlay, and the 6x6 px bottom-right AutoFill square handle in `grid.go`. Proof: `TestSelectionOverlayGeometry` asserts overlay box bounds align with the selected cell boxes in `page.Probe().Display()`. **Evidence:** `TestSelectionOverlayGeometry` aligns `xl-gov-sel` and `xl-gov-active` boxes with the cell boxes; the CSS gives the active cell a 2 px `--xl-selection-border` outline and the fill handle is 6x6 px.

### 2.4 Keyboard navigation and resizing

- [x] `EXL-24` Implement full Excel keyboard navigation in `selection.go` and `app.go`: `Arrow` keys move active cell (auto-scrolling viewport when leaving visible bounds), `Shift+Arrow` extends range, `Tab`/`Shift+Tab` moves horizontally (wrapping inside multi-cell selections), `Enter`/`Shift+Enter` moves vertically, `Home` jumps to column A, `Ctrl+Home` jumps to `A1`, `Ctrl+End` jumps to last populated cell, `PageUp`/`PageDown` scrolls one viewport height, and `Ctrl+Arrow` jumps to the next data region edge. Proof: `TestKeyboardNavigationMatrix` covers every shortcut in headless tests. **Evidence:** `TestKeyboardNavigationMatrix` and `TestDKeyboardWrapAndRegionJumps` cover arrows, Shift+arrows, Tab and Shift+Tab, Enter and Shift+Enter wrapping inside a multi-cell selection, Home, Ctrl+Home, Ctrl+End, Page Up and Page Down, Ctrl+Arrow region edges, and auto-scroll when the active cell leaves the window.
- [x] `EXL-25` Implement interactive column-width and row-height resizing via drag handles on header borders (`DragStart`/`DragMove`/`DragEnd`), clamping minimums (24 px width, 16 px height) and showing a live size tooltip (`Width: 112 px`). Proof: `TestHeaderResizeDrag` drags column B divider by +40 px and verifies updated width in `Sheet` and `page.Probe().Display()`. **Evidence:** `TestHeaderResizeDrag` (+40 px) and `TestDResizeClampTooltipAndDoubleClick` (live `Width: N px` tooltip, 24 px and 16 px minimums, tooltip removed after the drag) pass; `TestDZoomScalesGrid` checks the drag converts back to unzoomed pixels.
- [x] `EXL-26` Implement double-click on a column or row header border to auto-fit that column or row to its widest/tallest visible content plus padding. Proof: `TestHeaderAutoFit` populates a long string in column C, triggers auto-fit, and verifies column C expands to fit. **Evidence:** `TestHeaderAutoFit` and the double-click step in `TestDResizeClampTooltipAndDoubleClick` widen column E. Row auto-fit (`autoFitRow`) is reachable the same way but only the column path has a test.
- [x] `EXL-27` Wire the Name Box in `formula_bar.go` to display the current cell or range (`A1`, `B2:D8`, or `5R x 3C` during drag) and accept typed cell/range references (`G45`, `A1:C10`) on Enter to jump and select. Proof: `TestNameBoxJump` types `D25` + Enter and asserts active cell and scroll offset. **Evidence:** `TestNameBoxJump` types `D25` and Enter; `TestPointerSelectionFlows` checks the `A7:C9` and `3R x 3C` readouts.

### 2.5 Phase gate

- [~] `EXL-28` Run Phase 02 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/excel`, `go test -v ./examples/builder/excel/...`, `make lint`, and inspect a `-web` `/frame.png` screenshot of the grid in both Light and Dark themes. Proof: commands exit 0 and visual check confirms aligned headers, gridlines, and selection border. **Status:** Done: `gofmt` clean, `go vet -p 1` exit 0, `go test ./examples/builder/excel -count=1` exit 0, and `-web` `/frame.png` screenshots (Light) show aligned headers, gridlines, selection border, merged cells, filter chevrons and the autocomplete popover. Dark is covered by the pixel tests (`TestThemePixels`), not by an inspected `-web` screenshot. Open: `make lint` is not green; `golangci-lint run ./builder/excel/...` reports about 150 issues (mostly `mnd`, `contextcheck`, `gochecknoglobals`). Owner: lint pass. Next gate: clear or waive those linters for the example.

---

## Dependencies

- Requires Phase 01 (`EXL-02`, `EXL-04`, `EXL-05`, `EXL-10`, `EXL-11`).
- `EXL-14` (`gridMetrics`) is required by `EXL-15`, `EXL-16`, `EXL-18`, `EXL-19`, and `EXL-25`.
- `EXL-21` (`SelectionState`) is required by `EXL-22`, `EXL-23`, `EXL-24`, `EXL-27`, and all of Phase 03.
