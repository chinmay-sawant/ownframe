# v0.0.8 - Phase 04: Formatting ribbon and conditional styles

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 14 open)
> **Estimated effort:** 3-4 days

---

## Overview

Phase 04 builds out the full Excel Home ribbon formatting controls, number formatting engine, Format Painter, and conditional formatting rules. Every toolbar control uses `components/chadcn` primitives (`Button`, `ButtonGroup`, `Toggle`, `Select`, `Popover`, `DropdownMenu`, `Separator`, `Tooltip`) styled with `--xl-*` tokens so both `ExcelLight` and `ExcelDark` look cohesive and crisp.

## Executive summary

- Implement the Home ribbon groups: Clipboard, Font, Alignment, Number, Styles (Conditional Formatting, Cell Styles), Cells (Insert/Delete/Format), and Editing (AutoSum, Clear, Sort & Filter, Find).
- Build a pure-Go Excel number formatter covering General, Number, Currency, Accounting, Percentage, Date, Time, Scientific, and decimal-place steppers.
- Add interactive Office swatch pickers for fill color and font color, border presets, merge & center, text wrap, Format Painter, and live Conditional Formatting rules (highlight rules, top/bottom N, data bars, 3-color scales).

---

## Phase 4: Formatting ribbon and conditional styles

### 4.1 Home ribbon groups and typography controls

- [ ] `EXL-45` Implement the `Home` ribbon layout in `examples/builder/excel/ribbon.go` organized into labeled vertical groups separated by hairlines (`Clipboard`, `Font`, `Alignment`, `Number`, `Styles`, `Cells`, `Editing`), reflecting the active cell's current style state on every toggle and dropdown. Proof: `TestHomeRibbonReflectsActiveStyle` selects a bold currency cell and asserts active toggle states.
- [ ] `EXL-46` Wire Font controls in `ribbon.go`: font family select (`Sans`, `Mono`, `Serif`), font size select (`9` through `36` px) plus Increase/Decrease Font Size buttons, and `Bold` (`Ctrl+B`), `Italic` (`Ctrl+I`), `Underline` (`Ctrl+U`), and `Strikethrough` toggles applied across the full current selection. Proof: `TestFontFormattingShortcuts` applies Bold/Italic/Underline to `A1:B2` and verifies style IDs and rendered text styles.
- [ ] `EXL-47` Build the reusable Office Color Picker popover in `examples/builder/excel/pickers.go` displaying the 10x5 Office theme swatch matrix, 10 standard colors, `No Fill` / `Automatic` reset button, and recent custom colors, wired to both Fill Color and Font Color split buttons. Proof: `TestColorPickerAppliesFillAndFont` opens the fill popover, selects `#c6efce`, and verifies cell background updates.
- [ ] `EXL-48` Implement Border presets dropdown in `pickers.go` and `grid.go`: `Bottom Border`, `Top Border`, `Left Border`, `Right Border`, `No Border`, `All Borders`, `Outside Borders`, `Thick Outside Borders`, and `Top and Double Bottom Border` (accounting total style). Proof: `TestBorderPresetsRender` applies `Outside Borders` and `All Borders` to `B2:D4` and verifies per-edge border bits on interior vs edge cells.

### 4.2 Alignment, wrap, and merge controls

- [ ] `EXL-49` Wire Horizontal Alignment (`Left`, `Center`, `Right`), Vertical Alignment (`Top`, `Middle`, `Bottom`), and Indent Increase/Decrease buttons in `ribbon.go`, preserving Excel's default General alignment (numbers/dates right-aligned, text left-aligned, booleans/errors centered) when alignment is unset. Proof: `TestDefaultAndExplicitAlignment` checks computed alignment classes for numbers, text, errors, and explicitly centered cells.
- [ ] `EXL-50` Implement `Wrap Text` toggle and `Merge & Center` / `Unmerge Cells` dropdown in `ribbon.go` and `workbook.go`, recalculating row height when wrapped text overflows and safely warning or preserving top-left value when merging non-empty ranges. Proof: `TestWrapAndMergeCells` merges `A1:C1`, verifies anchor value and span, and unmerges back to individual cells.

### 4.3 Number formatting engine

- [ ] `EXL-51` Implement `examples/builder/excel/format.go` supporting Excel number formats: `General`, `Number` (`#,##0.00`), `Currency` (`$#,##0.00`), `Accounting` (`$ #,##0.00` with parentheses/red option for negatives), `Percentage` (`0.0%`), `Short Date` (`yyyy-mm-dd`), `Long Date` (`Mon, Nov 25, 2026`), `Time` (`hh:mm:ss`), `Scientific` (`0.00E+00`), and `Text` (`@`). Proof: table-driven unit tests in `format_test.go` covering positive, negative, zero, large numbers, and date serials.
- [ ] `EXL-52` Wire Number ribbon controls in `ribbon.go`: format dropdown (`General`, `Number`, `Currency`, `Accounting`, `Short Date`, `Percentage`, `Scientific`, `Text`), quick `$`, `%`, `,` buttons, and Increase/Decrease Decimal buttons (adjusting decimal places between 0 and 8). Proof: `TestNumberFormatButtons` cycles currency, percent, and decimal precision on `B2:B5` and verifies formatted display strings.
- [ ] `EXL-53` Implement `Format Painter` (single-click for one-shot apply, double-click for sticky mode, `Escape` to exit) and `Clear` dropdown (`Clear All`, `Clear Formats`, `Clear Contents`, `Clear Comments`) in `app.go` and `ribbon.go`. Proof: `TestFormatPainterAndClear` copies style from `A1` onto `C3:D4` and clears formats without losing values.

### 4.4 Conditional formatting and cell styles

- [ ] `EXL-54` Implement `examples/builder/excel/condfmt.go` storing per-sheet ordered `CondRule` entries over `RangeRef` targets: `GreaterThan`, `LessThan`, `Between`, `EqualTo`, `TextContains`, `DuplicateValues`, `TopN`, `BottomN`, `AboveAverage`, `DataBar`, and `ColorScale3`. Proof: unit tests in `condfmt_test.go` evaluating each rule type against mixed numeric/string/empty ranges.
- [ ] `EXL-55` Render conditional formatting overlays in `grid.go`: style overrides (red/yellow/green presets or custom fill/text), proportional horizontal `DataBar` fills behind cell text, and interpolated 3-color scale backgrounds (`#f8696b` -> `#ffeb84` -> `#63be7b`). Proof: `TestConditionalFormattingVisualOverlay` checks computed cell styles and data-bar width percentages in `page.Probe().Display()`.
- [ ] `EXL-56` Build the Conditional Formatting menu and Manage Rules dialog in `dialogs.go` / `menus.go` so users can add, reorder, toggle, and delete rules on the current selection, plus a Cell Styles gallery (`Good`, `Bad`, `Neutral`, `Heading 1-3`, `Total`, `Accent 1-4`). Proof: `TestCondFormatDialogAddAndDelete` opens the dialog, adds a `GreaterThan 1000` rule, and verifies live grid highlighting.
- [ ] `EXL-57` Build the `Format Cells` modal dialog (`Ctrl+1`) in `dialogs.go` with tabs for `Number`, `Alignment`, `Font`, `Border`, and `Fill` so power users can configure all properties in one place. Proof: `TestFormatCellsDialogCommitAndCancel` modifies font + fill + number format and verifies cancel leaves cells unchanged while OK applies atomically.

### 4.5 Phase gate

- [ ] `EXL-58` Run Phase 04 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/excel`, `go test -race -v ./examples/builder/excel/...`, `make lint`, and visual `-web` `/frame.png` inspection of the Home ribbon, swatch picker, accounting borders, and conditional data bars in Light and Dark themes. Proof: outputs recorded beside this row.

---

## Dependencies

- Requires Phase 03 (`EXL-31`, `EXL-37`, `EXL-38`) so formatters and conditional rules read evaluated typed `Value`s.
- `EXL-47` (Office color picker) is reused by Phase 05 sheet tab colors (`EXL-59`) and Phase 06 chart themes.
- `EXL-51` (`format.go`) is reused by Phase 05 status bar live metrics (`EXL-71`) and CSV export (`EXL-70`).
