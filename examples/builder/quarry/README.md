# Quarry

A local-first database client for SQLite and PostgreSQL, built with the ownframe builder UI (`ui.Mount`, `components/chadcn`) and no browser. It opens in the **QuarryDark** theme with a one-click **QuarryLight** toggle, and first launch opens a seeded demo SQLite database so you can query immediately.

## Run

```sh
go run ./examples/builder/quarry
go run ./examples/builder/quarry -db /path/to/file.sqlite
go run ./examples/builder/quarry -db postgres://user:pass@localhost:5432/app?sslmode=disable
go run ./examples/builder/quarry -db mock:postgres        # offline mock PostgreSQL
go run ./examples/builder/quarry -web                     # browser preview on 127.0.0.1:8150
```

| Flag | Meaning |
|---|---|
| `-web`, `-headless` | Serve the page instead of opening a window. |
| `-addr` | Address for `-web` (default `127.0.0.1:8150`). |
| `-db` | SQLite path, `postgres://` DSN, or `mock:postgres`. Empty opens the demo database. |
| `-data-dir` | Overrides the config directory (`state.db`, `quarry_demo.db`). |
| `-snapshot ADDR` | Window mode only: serve `GET /frame.png` of the live window on a loopback address, for `scripts/gui-drive.sh`. |
| `-dev`, `-perf`, `-debug-addr` | DevTools dock, perf counters, loopback profiling (needs `-tags ownframe_debug`). |

State (saved connections without passwords, history, snippets, theme, split sizes) lives in `os.UserConfigDir()/ownframe/quarry/state.db`.

## Shortcuts

| Keys | Action |
|---|---|
| **Ctrl+K** / **Ctrl+P** | Command palette |
| **Ctrl+Enter**, **F5** | Run statement at the caret (or the selection) |
| **Ctrl+Shift+Enter**, **F6** | Run the whole script |
| **F7** | Explain |
| **Ctrl+Space** | Autocomplete (Up/Down/Enter/Tab/Escape inside the popover) |
| **Ctrl+N**, **Ctrl+W**, **Ctrl+Tab** | New, close, cycle tabs |
| **Ctrl+Shift+T** | Toggle dark/light |
| **Escape** | Cancel a running query, close an overlay |

The page handlers receive letter and Enter keys without modifier names, so Quarry reads Ctrl and Shift from Ebiten (the `excel` example does the same).

## Using it

- Drop a `.db`, `.sqlite`, or `.sqlite3` file on the window to open it; drop a `.csv` to import it.
- Click a table to open its data tab: WHERE filter, header-click sort, page size, Prev/Next. Double-click a cell to edit; edits are staged, tinted, and written in one transaction from the changeset bar (Preview SQL, Discard, Commit).
- `Prod`-tagged connections open read-only behind a banner; unlocking asks for confirmation, and `DROP TABLE` / `TRUNCATE` require typing the table name.
- The Monitor tab polls the driver in the background (1s/2s/5s/Pause), drains samples on the frame tick, and shows KPIs, four live charts, storage breakdown, active queries and the slow-query log.

## Limits and notes

- The editor draws its own lines (no highlighted textarea exists), so it supports one caret, shift-selection, undo/redo, and no word wrap. Caret and selection are not exposed by the page, so the editor tracks them itself.
- The look is deliberately minimal: neutral near-black and near-white surfaces, 1px hairlines, flat secondary buttons that gain a border only on hover, and a single green accent for the primary action, the active tab marker and the selection. Shadows are a faint 3-4 px offset on floating overlays only (modals, palette, menus) and nowhere on in-page controls; gradients are avoided because the engine ignores border-radius on them. Radii follow one scale: 6 px controls, 8 px containers and cards, 10 px modals, and a segment inside a 2 px-padded segmented control is 6 px inside 8 px. Idle controls are fill-only (no 1 px strokes on rounded corners), which keeps the corners clean.
- Every scrollable area draws a thin themed scrollbar (`chadcn.Scrollbar` over a `chadcn.ScrollState`): the result grid, SQL editor, schema tree, monitor, palette, history and snippet lists, inspector, structure view, messages and plan. The engine does not scroll nested overflow, so each area is a translated pane or a virtual list with its own offset; the thumb tracks that offset, the track pages on click, the thumb drags, and the wheel scrolls the area under the pointer. Inside dialogs and the palette the bar is laid out in flow, because the engine does not paint absolutely positioned children of a fixed overlay.
- Icons are SVG documents registered with `Page.SetImage` and drawn with `<img>` (the builder has no `<svg>` tag), once per tone and theme. The Quarry mark sits in the title bar and the rail.
- Settings (rail gear) persist theme, default row limit, monitor poll interval, the DROP/TRUNCATE confirmation and the slow-query threshold in the config store.
- Text uses a small letter and word spacing to hide the engine's over-tight kerning; the editor and monospace blocks opt out so click-to-column stays exact.
- The engine's accessibility tree supports only main, navigation, group, heading, paragraph, img, link, button, textbox, checkbox, radio, combobox, list and listitem. Quarry maps grid to list, row to listitem, tab to button and cell/dialog/status to group, each with an explicit `aria-label`, so grid rows and cells, column headers, tabs, tree rows, buttons, inputs and KPI cards are all in the tree (`TestGridCellsAndTabsInAccessibilityTree`). The password field stays redacted.
- Atomic commits and CSV imports use `driver.TxExecer.ExecTx` (SQLite, PostgreSQL, and the mock implement it), with a `BEGIN ... COMMIT` script as the fallback for other drivers.
- Android and macOS were not exercised.
