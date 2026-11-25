# v0.0.8 - Excel example (builder pattern)

> **Parent:** [plans/README.md](../../README.md)
> **Target package:** `examples/builder/excel`
> **Status:** Planned (0 complete, 0 partial, 84 open)
> **Estimated effort:** 6 phases (iterative implementation and verification)

---

## Overview

This directory is the canonical phase-wise checklist for building a desktop spreadsheet replica modeled on Microsoft Excel under `examples/builder/excel/`. The application uses the ownframe builder pattern (`ui.Mount`, `*ui.Element` trees, and `components/chadcn` controls) with `ownframe.RunWithOptions`, `CustomHeader: true`, and zero imports of `ownframe/internal/`.

Every feature, visual token, interaction contract, and verification gate lives in one of the six phase files below. No checklist row may be marked `[x]` without code, tests, or screenshot evidence recorded beside the row.

## Executive summary

- **Builder architecture:** Compose all chrome, ribbon tabs, formula bar, virtualized grid, sheet tabs, status bar, dialogs, and popovers from Go `*ui.Element` builders and `components/chadcn` components, matching `examples/builder/inventory` and `examples/builder/loglens`.
- **Two first-class themes plus print/high-contrast tokens:** Ship Excel Office Green Light (`ExcelLight`) and Excel Dark (`ExcelDark`), both wired through `--of-*` ChadCN variables and `--xl-*` spreadsheet tokens, plus a fixed 10-column Office theme swatch matrix and formula reference highlight palette.
- **Real spreadsheet engine in pure Go:** Sparse workbook storage, dependency-graph formula recalculation, cycle detection, standard Excel error values, 35+ core functions, number formatting, conditional formatting, sort/filter, undo/redo, TSV clipboard interop, and CSV/JSON/XLSX import and export.
- **Bounded paint cost:** Render only the visible row and column window using `ui/virtual` and explicit scroll offsets, keeping layout and replay cost independent of total sheet capacity (1,000+ rows by 52+ columns).

---

## Color schemas and design tokens

The spreadsheet defines `--xl-*` CSS custom properties layered on top of `chadcn.Light` and `chadcn.Dark`. Both themes share identical geometry (22 px default row height, 84 px default column width, 40 px row header width, 24 px column header height, 36 px title bar, 88 px expanded ribbon, 28 px formula bar, 28 px sheet bar, 24 px status bar).

### 1. Primary theme tokens (`ExcelLight` and `ExcelDark`)

| Token | Role | **ExcelLight** | **ExcelDark** |
| --- | --- | --- | --- |
| `--xl-brand` | Excel title bar and active accent | **#107c41** | **#185c37** |
| `--xl-brand-hover` | Brand button hover fill | **#0c6634** | **#217346** |
| `--xl-brand-strong` | Active selection border and fill handle | **#107c41** | **#33c481** |
| `--xl-brand-tint` | Active ribbon tab pill and selected header tint | **#e6f4ea** | **#163826** |
| `--xl-titlebar-bg` | Custom window title bar background | **#107c41** | **#18181b** |
| `--xl-titlebar-fg` | Title bar text and window icons | **#ffffff** | **#f4f4f5** |
| `--xl-ribbon-bg` | Ribbon strip and toolbar surface | **#f3f2f1** | **#202024** |
| `--xl-ribbon-group-bg` | Ribbon card/group container | **#ffffff** | **#27272b** |
| `--xl-ribbon-border` | Ribbon divider and group hairline | **#e1dfdd** | **#35353b** |
| `--xl-formula-bg` | Name Box and formula input surface | **#ffffff** | **#18181b** |
| `--xl-sheet-bg` | Grid cell canvas background | **#ffffff** | **#141417** |
| `--xl-cell-text` | Default cell text | **#1f2328** | **#f4f4f5** |
| `--xl-cell-muted` | Placeholder and secondary hint text | **#605e5c** | **#a1a1aa** |
| `--xl-gridline` | Standard 1 px cell gridline | **#d4d4d4** | **#2d2d33** |
| `--xl-gridline-frozen` | Freeze-pane divider line | **#8a8886** | **#71717a** |
| `--xl-header-bg` | Row and column header fill | **#f3f2f1** | **#222226** |
| `--xl-header-fg` | Row and column header label | **#323130** | **#d4d4d8** |
| `--xl-header-active-bg` | Header fill for active selection span | **#d2e8db** | **#1d3b2b** |
| `--xl-header-active-fg` | Header label for active selection span | **#107c41** | **#4ade80** |
| `--xl-selection-fill` | Multi-cell range overlay fill | **#107c411f** | **#33c48129** |
| `--xl-selection-border` | Active cell 2 px outline | **#107c41** | **#33c481** |
| `--xl-cut-dash` | Marching/dashed cut-source border | **#107c41** | **#4ade80** |
| `--xl-tabbar-bg` | Bottom sheet tab strip | **#f3f2f1** | **#1c1c20** |
| `--xl-tab-active-bg` | Active sheet tab fill | **#ffffff** | **#27272b** |
| `--xl-statusbar-bg` | Bottom status bar background | **#f3f2f1** | **#18181b** |
| `--xl-statusbar-fg` | Status bar text and metrics | **#323130** | **#d4d4d8** |
| `--xl-error-fg` | Formula error values (`#DIV/0!`, `#REF!`) | **#c42b1c** | **#f87171** |
| `--xl-comment-corner` | Top-right comment triangle indicator | **#c42b1c** | **#fb7185** |
| `--xl-negative-num` | Accounting/red negative numbers | **#b91c1c** | **#f87171** |

### 2. Formula reference highlight palette

When editing a formula, each referenced cell or range gets a distinct border and translucent fill in order, cycling modulo 6:

| Index | Border color (Light) | Fill color (Light) | Border color (Dark) | Fill color (Dark) |
| --- | --- | --- | --- | --- |
| 1 | **#2563eb** | `#2563eb1a` | **#60a5fa** | `#60a5fa24` |
| 2 | **#dc2626** | `#dc26261a` | **#f87171** | `#f8717124` |
| 3 | **#7c3aed** | `#7c3aed1a` | **#a78bfa** | `#a78bfa24` |
| 4 | **#0891b2** | `#0891b21a` | **#22d3ee** | `#22d3ee24` |
| 5 | **#d97706** | `#d977061a` | **#fbbf24** | `#fbbf2424` |
| 6 | **#db2777** | `#db27771a` | **#f472b6** | `#f472b624` |

### 3. Conditional formatting presets and Office swatch palette

- **Red fill with dark red text:** background `#ffc7ce`, text `#9c0006` (Dark: `#4c1d24`, `#fca5a5`).
- **Yellow fill with dark yellow text:** background `#ffeb9c`, text `#9c6500` (Dark: `#4a3712`, `#fcd34d`).
- **Green fill with dark green text:** background `#c6efce`, text `#006100` (Dark: `#143d24`, `#86efac`).
- **Data bars:** blue `#638ec6`, green `#63c384`, red `#ff555a`, amber `#ffb628`.
- **3-color scale (Green-Yellow-Red):** low `#f8696b`, mid `#ffeb84`, high `#63be7b`.
- **Font and fill color picker swatches:** 10 Office theme base colors (`#ffffff`, `#000000`, `#e7e6e6`, `#44546a`, `#4472c4`, `#ed7d31`, `#a5a5a5`, `#ffc000`, `#5b9bd5`, `#70ad47`) with 4 tint/shade rows, plus 10 standard colors (`#c00000`, `#ff0000`, `#ffc000`, `#ffff00`, `#92d050`, `#00b050`, `#00b0f0`, `#0070c0`, `#002060`, `#7030a0`).

---

## Planned package layout (`examples/builder/excel/`)

Keep files grouped by ownership boundary, matching `examples/builder/inventory` and `examples/builder/loglens`:

| File(s) | Responsibility |
| --- | --- |
| `main.go` | CLI flags (`-web`, `-headless`, `-addr`, `-dev`, `-perf`, `-debug-addr`), signal handling, `ownframe.RunWithOptions` with `CustomHeader` |
| `app.go` | `app` state struct, `ui.Mount`, `ownframe.Handlers` dispatch, overlay dismissal, keyboard shortcuts |
| `styles.go` | `ExcelLight`, `ExcelDark`, `--xl-*` token sheets, grid/ribbon/formula/status CSS |
| `workbook.go`, `cell.go` | Workbook, sheets, sparse cell store, cell address parsing/formatting (`A1`, `R1C1`, ranges), sample starter workbook |
| `formula_lexer.go`, `formula_parser.go`, `formula_eval.go`, `formula_funcs.go`, `depgraph.go` | Lexer, AST parser, dependency graph, topological recalc, cycle detection, Excel function library |
| `format.go`, `condfmt.go` | Number formatters (General, Currency, Accounting, Percent, Date, Scientific) and conditional formatting evaluator |
| `shell.go`, `ribbon.go`, `formula_bar.go`, `sheetbar.go`, `statusbar.go` | Builder components for title bar, Quick Access Toolbar, Ribbon tabs, Name Box + formula bar, sheet tabs, status bar |
| `grid.go`, `selection.go`, `scroll.go` | Virtualized grid builder, sticky headers, freeze panes, active cell, multi-range selection, fill handle, resize drags |
| `dialogs.go`, `menus.go`, `pickers.go` | Find & Replace, Paste Special, Format Cells, Data Validation, Sort/Filter popovers, context menu, color swatch pickers |
| `history.go`, `clipboard.go`, `io.go` | Command-based Undo/Redo stack, TSV clipboard copy/cut/paste, CSV/JSON/XLSX import & export |
| `charts.go` | Embedded range-bound charts (bar, column, line, area, pie) using `components/chadcn` / builder SVG/layout primitives |
| `*_test.go` | Engine unit tests, headless interaction tests, viewport/scroll tests, pixel/theme tests, paint benchmarks |

---

## Phase progress

| Phase | File | Complete | Partial | Open |
| --- | --- | ---: | ---: | ---: |
| Phase 01: Shell, color schemas, and workbook model | [phase-01-shell-theme-and-workbook-model.md](phase-01-shell-theme-and-workbook-model.md) | 0 | 0 | **13** |
| Phase 02: Virtualized grid, selection, and navigation | [phase-02-virtualized-grid-selection-and-navigation.md](phase-02-virtualized-grid-selection-and-navigation.md) | 0 | 0 | **15** |
| Phase 03: Cell editing, formula bar, and formula engine | [phase-03-cell-editing-formula-bar-and-engine.md](phase-03-cell-editing-formula-bar-and-engine.md) | 0 | 0 | **16** |
| Phase 04: Formatting ribbon and conditional styles | [phase-04-formatting-ribbon-and-conditional-styles.md](phase-04-formatting-ribbon-and-conditional-styles.md) | 0 | 0 | **14** |
| Phase 05: Data tools, sheets, clipboard, and file I/O | [phase-05-data-tools-sheets-and-io.md](phase-05-data-tools-sheets-and-io.md) | 0 | 0 | **14** |
| Phase 06: Charts, validation, accessibility, and verification | [phase-06-charts-freeze-accessibility-and-verification.md](phase-06-charts-freeze-accessibility-and-verification.md) | 0 | 0 | **12** |
| **Total** | | **0** | **0** | **84** |

---

## Phase 1: Execution and evidence rules

### 1.1 Checklist conventions

- Use three statuses consistently across all phase files:
  - `[ ]` not started or not proven;
  - `[x]` implemented and validated with current evidence (commit hash, command, exit status, and measured result);
  - `[~]` intentionally deferred or partial, with reason, owner boundary, and next gate.
- Keep `examples/builder/excel` free of `ownframe/internal/` imports (`make internal-import-check` must stay green).
- Use `ownframe.UseMemory(true)` in every test that touches the clipboard.
- Run `gofmt`, `make lint`, and `make test` before closing any code phase. Run `make graphics-check` when visual or pixel tests are added.
- Register `examples/builder/excel` under `examples/readme.md` once Phase 01 compiles cleanly (`builder` is already listed in `examples/maintenance.json`).

## Dependencies

- **Phase 01 before Phase 02:** The workbook data model, theme sheets, and `ui.Mount` shell must compile and render before virtualized grid windowing and selection overlays are wired.
- **Phase 02 before Phase 03:** Active cell coordinates, range selection, and keyboard focus navigation are prerequisites for inline editing, formula bar synchronization, and clicking cells to insert references.
- **Phase 03 before Phase 04 and Phase 05:** Typed cell values (number, string, bool, date, error) and evaluated results feed number formatters, conditional formatting rules, sorting, filtering, and status bar aggregations.
- **Phase 05 before Phase 06:** Multi-sheet ranges, clipboard/undo safety, and stable layout geometry must hold before binding live charts, validation dropdowns, and running the final visual and benchmark gates.
