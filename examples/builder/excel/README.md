# Excel example (builder pattern)

A spreadsheet in the style of Excel, built with the ownframe `ui` builder API and `components/chadcn` primitives. There is no web runtime or JavaScript. The formula engine, number formats, conditional formats, history, and file formats are plain Go with the standard library only.

## Run

```sh
go run ./examples/builder/excel                 # desktop window
go run ./examples/builder/excel -web            # browser preview on 127.0.0.1:8149
go run ./examples/builder/excel -web-op         # same as -web
go run ./examples/builder/excel -headless -addr 127.0.0.1:8149
go run ./examples/builder/excel -performance    # perf sampling; with -web also enables pprof
go run ./examples/builder/excel -dev            # DevTools dock at startup
```

| Flag | Meaning |
| --- | --- |
| `-web`, `-web-op`, `-headless` | Serve the app in a browser instead of opening a window. |
| `-addr` | Listen address for the preview. Default `127.0.0.1:8149`. |
| `-performance`, `-perf` | Turn on `Page.Window().SetPerf` and, in preview mode, the pprof endpoints. Off by default. |
| `-dev` | Open the DevTools dock at startup. |
| `-debug-addr` | Loopback profiling endpoint (needs `-tags ownframe_debug`). |

In the preview, read `/frame.png` to see the picture. See `documentation/testing.md` for driving it with `/click` and `/wheel`.

## What is in it

- **Ribbon** with Home, Insert, Page Layout, Formulas, Data, Review, and View tabs, and a chevron at the right of the tab strip that collapses it to the tab row (32 px instead of 88 px). Home has Clipboard, Font, Alignment, Number, Styles, Cells, and Editing groups. Controls reflect the active cell: font family (Default, Sans, Mono, Serif), size, Bold, Italic, Underline, Strikethrough, horizontal and vertical alignment, indent, wrap, merge, number formats (General, Number, Currency, Accounting, Short Date `yyyy-mm-dd`, Long Date, Time, Percentage, Scientific, Text), decimals, per-edge borders (bottom, top, left, right, none, all, outside, thick outside, top and double bottom), and fill and font color pickers with the Office theme matrix, standard colors, and recent colors. A row grows to fit wrapped text until you set its height by hand.
- **Format Painter** (click for one use, click again for sticky, `Escape` to stop) and a **Clear** menu.
- **Conditional formatting** from the Styles group: highlight rules, duplicates, top and bottom N, above and below average, data bars, and 3-color scales, plus a Manage Rules dialog with reorder, enable, and delete. A Cell Styles gallery sits next to it.
- **Dialogs:** Find and Replace, Paste Special, Format Cells (Number, Alignment, Font, Border, Fill), Data Validation, Sort (two keys), Filter (search box, value checkboxes, Sort A to Z and Z to A; the grid header chevron opens it), Comments, Open, Save or Export, and Keyboard Shortcuts.
- **Charts** from the Insert tab: Column, Bar, Line, Area, and Pie. A chart is bound to a cell range, reads the cells on every render, and moves by dragging its header. `Delete` removes the selected chart.
- **Context menus** for cells, row and column headers, and sheet tabs.
- **Starter workbook:** `Sheet1` holds a regional sales table with formulas, and `Summary` rolls it up with cross-sheet `SUM`, `AVERAGE`, `MAX`, and `IF` plus a quarter table to chart.
- **File menu:** New, Open, and Save or Export for workbook JSON, `.xlsx`, and CSV. The `.xlsx` writer and reader keep values, formulas, number formats, fonts, fills, per-edge borders, alignment and indent, widths and heights, hidden rows and columns, merges, freeze panes, tab colors, conditional formats, data validation, and comments. Dropped `.json`, `.xlsx`, and `.csv` files open through the same path.
- **Zoom:** the whole UI (title bar, ribbon, formula bar, grid, sheet tabs, status bar, dialogs) scales as one, from 50% to 200%. Use Ctrl+wheel, Ctrl+= / Ctrl++, Ctrl+-, Ctrl+0 (reset to 100%), or the View ribbon. The status bar shows the level. Until you choose one, the default follows the window width: 100% below 1800 px, 125% from 1800 px, 150% from 2400 px. The engine has no CSS `zoom`, so every px length in the theme sheet is rescaled and re-applied with `SetTheme` (`zoom.go`). The web preview has no modifier state, so use the ribbon there.
- **View:** gridlines, headings, formula bar, Show Formulas, zoom, freeze panes, and the Light and Dark themes.

## Keyboard shortcuts

| Shortcut | Action |
| --- | --- |
| `Ctrl+B`, `Ctrl+I`, `Ctrl+U` | Bold, italic, underline |
| `Ctrl+1` | Format Cells |
| `Ctrl+F`, `Ctrl+H` | Find and Replace |
| `Ctrl+Alt+V` | Paste Special |
| `Ctrl+Shift+L` | Toggle filter |
| `Shift+F2` | New comment |
| `Ctrl+/` | Keyboard shortcuts |
| `Ctrl+Z`, `Ctrl+Y` | Undo, redo |
| `Escape` | Close a menu or dialog, stop Format Painter |

## Known limits

- The grid must call `applyStyleExtras` (`style_css.go`) for font family, per-edge borders, vertical alignment, and indent to show in cells; the engine and ribbon store them already.
- A cut moves formulas that point at a single cell or a fully cut range. A range that is only partly cut stays where it is, as in Excel. `Ctrl+V` after `Ctrl+X` still pastes text, so only the ribbon Paste button moves dependents until the keyboard path calls `pasteCut`.
- Conditional-format rule changes, data validation, and comment edits are not undoable. Adding, deleting, and renaming a sheet, and row and column inserts and deletes, are.
- XLSX conditional formats keep the rule, range, and preset color; custom dxf styles from other apps map to the closest of the three presets. Disabled rules are written to JSON but not to `.xlsx`, and validation other than "between" bounds or inline lists is skipped on import.
- The Pie chart is a share-of-total strip with a legend, because the renderer has no conic gradient.

## Plan and phase-wise checklist

Track implementation and verification in [plans/v0.0.8/excel-example/README.md](../../../plans/v0.0.8/excel-example/README.md):

1. [Phase 01: Shell, color schemas, and workbook model](../../../plans/v0.0.8/excel-example/phase-01-shell-theme-and-workbook-model.md)
2. [Phase 02: Virtualized grid, selection, and navigation](../../../plans/v0.0.8/excel-example/phase-02-virtualized-grid-selection-and-navigation.md)
3. [Phase 03: Cell editing, formula bar, and formula engine](../../../plans/v0.0.8/excel-example/phase-03-cell-editing-formula-bar-and-engine.md)
4. [Phase 04: Formatting ribbon and conditional styles](../../../plans/v0.0.8/excel-example/phase-04-formatting-ribbon-and-conditional-styles.md)
5. [Phase 05: Data tools, sheets, clipboard, and file I/O](../../../plans/v0.0.8/excel-example/phase-05-data-tools-sheets-and-io.md)
6. [Phase 06: Charts, validation, accessibility, and verification](../../../plans/v0.0.8/excel-example/phase-06-charts-freeze-accessibility-and-verification.md)
