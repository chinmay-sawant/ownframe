# Spreadsheet

A desktop spreadsheet example: an editable grid with row and column
headers, a formula bar, rectangular selection, copy and paste, bounded
undo and redo, CSV import and export, and durable workbooks in SQLite.
The screen is an ownframe page; the core lives in `workbook/`, `formula/`,
and `storage/`.

```sh
go run ./examples/spreadsheet
```

## Flags

| Flag | Meaning |
|---|---|
| `-data <dir>` | Data directory override. Default is `os.UserConfigDir()/ownframe/spreadsheet`. |
| `-workbook <name>` | Open a seeded workbook by name. Default is the first one. |
| `-perf` | Record pipeline timing for the devtools Frame tab and `-debug-addr`. |
| `-debug-addr <addr>` | Serve window profiling on this loopback address (off by default; needs `-tags ownframe_debug`). |

## Dummy mode and real data

The first run seeds two workbooks through the storage layer:

- `Demo workbook`: three sheets (`Numbers`, `Text`, `Mixed`), 200 rows by
  20 columns, with numbers, text, formulas, Unicode, zeroes, and
  deliberate errors (`#DIV/0!`, `#NAME?`, `#REF!`, `#CYCLE!`).
- `Stress sparse`: 100,000 rows by 20 columns, sparse, with formulas every
  50 rows.

Seeding is idempotent, so a restart never duplicates fixtures or
overwrites edits. `-workbook "Stress sparse"` opens the stress sheet for
scrolling and profiling work.

The grid bounds are the model's `MaxRows` and `MaxCols` (1,048,576 by
18,278), so the scrollable area matches a real sheet. Ctrl+End jumps to
the used range.

## Storage location and recovery

The database is `<dir>/spreadsheet.db`, WAL mode on local writable
storage. One serialized store worker owns one connection. Edits are
acknowledged only after COMMIT; a failed save leaves the database at the
last commit, restores the editor with the typed text, and prints a status
line. Reopening recalculates every formula, so a stale calculated cache on
disk is never trusted.

Undo history is volatile. It is bounded to 200 steps in memory and does
not survive a restart; the cells do. A CSV import resets it.

## Controls

- Click a cell to select it; Shift+click or Shift+arrows extends the
  range; click a row or column header to select the whole row or column.
- Type to replace a cell, F2 or a second click to edit in place, Enter to
  commit and move down, Escape to cancel.
- Arrows move, Tab moves right, PageUp and PageDown page, Home and End
  move along the row, Ctrl+Home and Ctrl+End jump, Ctrl+arrows jump.
- Ctrl+C, Ctrl+X, Ctrl+V copy, cut, and paste tab-separated ranges.
- Ctrl+Z and Ctrl+Y undo and redo; Delete clears the selection.
- The toolbar has Undo, Redo, Import, Export, and the theme toggle.
- Import picks a CSV file, shows the first rows, and commits atomically
  with the current replace mode. Export writes the used range with
  displayed values to the dialog path, by default under the data
  directory.

The window is at least 480x320. Smaller windows still scroll, but the
chrome takes most of the height.

## Measured performance

Headless numbers on an Intel i7-13700HX (Linux), 1200x800 window, with
`go test -bench` on the UI package. The rendered window is the visible
cells only; the fetched window adds three rows and columns of overscan.
One wheel step repaints the chrome and redraws the page.

| Workload | Cost |
|---|---|
| Scroll step, window unchanged | ~5.9 ms |
| Scroll step crossing an overscan edge | ~4.2 ms |
| Dense window, every visible cell holds text | ~6.2 ms |
| Tick that applies a full-window fetch and redraws | ~8.0 ms |

The grid draws one fill per column and row boundary instead of a border on
every cell, and one rectangle per multi-cell selection. That change cut a
scroll step from 17.8 ms to under 6 ms. The acceptance target is p99 frame
work under 16.67 ms at 60 Hz; these headless redraw numbers leave room for
the replay and input work, but they are not a desktop GPU measurement.

The devtools Frame tab and `-perf` show the same stages live.

## Platform limits

Desktop `ownframe.Run` only. There is no `-web` mode, so no still-preview
path to mislabel, and no Android or iOS bind. The 100, 150, and 200
percent Windows scaling, monitor moves, minimize and restore, and the
native Windows host checks are untested on this machine and stay pending.
The example uses the examples module's existing dependencies only.

## Layout

| Path | Role |
|---|---|
| `ui/` | Screen: viewport windows, selection, editing, clipboard, CSV dialogs, worker, theme. |
| `ui/corebackend/` | Adapter from the core to the `ui.Backend` seam. |
| `workbook/` | Sparse sheets, edits, bounded undo, recalc, CSV parse and write. |
| `formula/` | Parser and evaluator with explicit limits. |
| `storage/` | SQLite store: one worker, transactions, prefs, backups, migrations. |
