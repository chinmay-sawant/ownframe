# v0.0.8 - Phase 05: Data tools, sheets, clipboard, and file I/O

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 14 open)
> **Estimated effort:** 4-5 days

---

## Overview

Phase 05 delivers the day-to-day spreadsheet workflow tools that make an Excel replica genuinely usable: multi-sheet management, undo/redo, clipboard with TSV interoperability and Paste Special, AutoFill drag-to-fill, row/column structural edits, Sort & AutoFilter, Find & Replace, right-click context menus, CSV/JSON/XLSX file import and export, and the live Status Bar.

## Executive summary

- Build the bottom sheet tab bar (add, switch, rename, duplicate, reorder, delete, tab color) and live Status Bar (Sum, Average, Count, Min, Max, zoom slider).
- Implement command-pattern Undo/Redo (`Ctrl+Z`/`Ctrl+Y`) covering edits, formatting, structural row/col mutations, sort, paste, and AutoFill.
- Wire OS clipboard (`ownframe.Read`/`Write`) with TSV serialization, Paste Special (Values, Formulas, Formats, Transpose), and AutoFill series detection.
- Add Sort (A-Z, Z-A, multi-column), AutoFilter header dropdowns, Find & Replace (`Ctrl+F`/`Ctrl+H`), `ContextMenuFor` right-click menus, and CSV/TSV/JSON/XLSX import and export (including drag-and-drop file open).

---

## Phase 5: Data tools, sheets, clipboard, and file I/O

### 5.1 Multi-sheet tab bar and structural row/column edits

- [ ] `EXL-59` Build `examples/builder/excel/sheetbar.go` with horizontal sheet tabs, active tab elevation (`--xl-tab-active-bg` + `--xl-brand-strong` bottom accent), `+` Add Sheet button, scroll arrows for many sheets, inline rename on double-click, and tab context menu (Rename, Duplicate, Delete, Move Left/Right, Tab Color). Proof: `TestSheetTabOperations` adds, renames, duplicates, recolors, reorders, and deletes a sheet while preserving cross-sheet formulas.
- [ ] `EXL-60` Implement Insert/Delete Rows and Columns, Hide/Unhide Rows and Columns, and Row/Column reference shifting in `workbook.go` and `depgraph.go` so inserting a row above `B5` shifts formulas referencing `B5` to `B6` and updates dependency edges cleanly. Proof: `TestInsertDeleteRowColShiftsFormulas` verifies formula rewriting, `#REF!` on deleted targets, and recalculation.

### 5.2 Undo/Redo command stack

- [ ] `EXL-61` Implement `examples/builder/excel/history.go` as a bounded command stack (default 200 entries) storing reversible deltas (cell raw/style diffs, row/col structural ops, sheet ops, sort permutations, conditional rule changes) wired to `Ctrl+Z`, `Ctrl+Y`/`Ctrl+Shift+Z`, and the Quick Access Toolbar Undo/Redo buttons. Proof: `history_test.go` verifies round-trip Undo/Redo across edits, formatting, paste, AutoFill, and row deletion.

### 5.3 Clipboard, Paste Special, and AutoFill

- [ ] `EXL-62` Implement `examples/builder/excel/clipboard.go` wired to `Ctrl+C`, `Ctrl+X`, and `Ctrl+V` using `ownframe.Write`/`ownframe.Read` for TSV external clipboard interop plus an internal rich buffer preserving formulas and styles, painting the `--xl-cut-dash` dashed border around the copied/cut source range. Proof: `TestClipboardCopyCutPaste` (with `ownframe.UseMemory(true)`) tests internal formula shifting, cut-and-move, and external TSV paste.
- [ ] `EXL-63` Implement the `Paste Special` dialog (`Ctrl+Alt+V`) in `dialogs.go` and `clipboard.go` supporting `All`, `Values only`, `Formulas only`, `Formats only`, `Transpose`, and `Skip blanks`. Proof: `TestPasteSpecialModes` verifies each mode including 2x3 -> 3x2 transpose.
- [ ] `EXL-64` Implement the AutoFill drag handle (`DragStart`/`DragMove`/`DragEnd` from the 6x6 px selection corner) in `selection.go` and `workbook.go`: copy single values, extrapolate numeric series (`1, 2 -> 3, 4, 5`), weekday/month names (`Jan, Feb -> Mar, Apr`), dates, and relative-shifted formulas. Proof: `TestAutoFillSeriesAndFormulas` drags down and right across numeric, date, month, and formula seeds.

### 5.4 Sort, AutoFilter, and Find & Replace

- [ ] `EXL-65` Implement `Data` ribbon Sort actions (`Sort A-Z`, `Sort Z-A`, and multi-column `Custom Sort` dialog with `My data has headers` checkbox) in `dialogs.go` and `workbook.go`, keeping row styles and relative formulas coherent. Proof: `TestSortSingleAndMultiColumn` sorts `Sales Q1-Q4` by Region ascending then Revenue descending and verifies row integrity.
- [ ] `EXL-66` Implement `AutoFilter` toggle (`Ctrl+Shift+L`) rendering filter chevron buttons on the header row of the active table; clicking a chevron opens a popover with distinct value checkboxes, search box, `Select All`, `Clear Filter`, and `Sort A-Z / Z-A`, hiding non-matching rows in `Sheet`. Proof: `TestAutoFilterHidesAndRestoresRows` filters `Expenses` to one category, checks visible row count, and clears the filter.
- [ ] `EXL-67` Implement the `Find & Replace` dialog (`Ctrl+F`, `Ctrl+H`) in `dialogs.go` supporting `Find Next`, `Find Previous`, `Find All` results list, `Replace`, `Replace All`, `Match case`, `Match entire cell contents`, and `Search in Formulas vs Values` across current sheet or entire workbook. Proof: `TestFindAndReplaceFlows` verifies navigation, single replace, replace-all count, and Undo reversal.

### 5.5 Context menu, File I/O, and Status Bar

- [ ] `EXL-68` Wire `Page.ContextMenuFor` and `MenuAction` (plus `components/chadcn` `ContextMenu` fallback) in `examples/builder/excel/menus.go` with context-sensitive right-click menus for cells (Cut, Copy, Paste, Paste Special, Insert, Delete, Clear Contents, Format Cells, Insert Comment), row/column headers (Insert, Delete, Hide, Unhide, Resize/Auto-Fit), and sheet tabs. Proof: `TestContextMenuActions` invokes cell and header menu actions via `MenuAction`.
- [ ] `EXL-69` Implement `examples/builder/excel/io.go` for CSV/TSV import and export (RFC 4180 quoting, formula/value export option) and native JSON workbook save/load preserving all sheets, formulas, styles, column widths, merges, freeze panes, and conditional rules. Proof: `io_test.go` round-trips the seeded workbook through JSON and CSV without data loss.
- [ ] `EXL-70` Add `File` backstage menu (`New`, `Open...`, `Save`, `Export CSV`, `Export XLSX` or minimal OpenXML `.xlsx` writer/reader without bloating dependencies) and drag-and-drop CSV/JSON open via `ownframe.Drop`. Proof: `TestFileBackstageAndDropOpen` loads a sample CSV via drop handler and verifies populated cells.
- [ ] `EXL-71` Build `examples/builder/excel/statusbar.go` showing edit mode indicator (`Ready`, `Enter`, `Edit`, `Point`), live selection aggregations (`Average`, `Count`, `Numerical Count`, `Min`, `Max`, `Sum` formatted with `format.go`), View mode shortcuts (`Normal`, `Page Break`, `Formula View`), and Zoom controls (`-`, slider/label `100%`, `+`). Proof: `TestStatusBarLiveAggregations` selects `B2:B6` and asserts exact `Sum`, `Average`, and `Count` text in the status bar.

### 5.6 Phase gate

- [ ] `EXL-72` Run Phase 05 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/excel`, `go test -race -v ./examples/builder/excel/...` (with `ownframe.UseMemory(true)`), and `make lint`. Proof: all commands exit 0 with results recorded beside this row.

---

## Dependencies

- Requires Phase 03 (`EXL-38`, `EXL-39`) for formula shifting and dependency graph updates during insert/delete, paste, and AutoFill.
- Requires Phase 04 (`EXL-47`, `EXL-51`) for tab color swatches, formatted status metrics, and Paste Special Formats.
- `EXL-61` (Undo/Redo) should wrap every mutating command in `EXL-59` through `EXL-70`.
