# Live log viewer

A desktop ownframe example that follows local log files, or a seeded dummy
stream by default. One window shows a compact source sidebar, a monospaced
list of fixed-height summary rows, and an entry detail pane. Searches,
filters, paging, export, follow, and pause all run against a local SQLite
database with bounded work per interaction.

Desktop `ownframe.Run` only. There is no `-web` mode, no phone build, and no
extra module dependency.

## Run it

```sh
cd examples
go run ./live-log-viewer                 # dummy mode: 10,000 seeded entries, 30/s stream
go run ./live-log-viewer -file /var/log/app.log
go run ./live-log-viewer -file a.log -file b.log -from-end=false
go run ./live-log-viewer -dir /tmp/logs-demo
go run ./live-log-viewer -temp           # throwaway database, removed on close
go run ./live-log-viewer -perf           # redraw timings in the status bar
```

Flags:

| flag | meaning |
| --- | --- |
| `-dir` | data directory; empty uses the user config directory |
| `-temp` | fresh temporary directory, removed on close |
| `-file` | log file to follow; repeatable, and any file switches to real mode |
| `-from-end` | start `-file` sources at their current end (default true) |
| `-perf` | measure redraw time and coalesced requests in the footer |
| `-debug-addr` | serve window profiling on this loopback address (off by default; needs `-tags ownframe_debug`) |

Dummy mode seeds three generated sources (`api`, `worker`, `database`) and
keeps streaming about 30 entries a second. Seeding is idempotent: restarting
never duplicates the fixture. Real mode adds the paths to a `Local logs`
session and resumes each source from its committed checkpoint.

## Controls

- Click a source in the sidebar to filter; click it again to go back to all.
- Severity chips set a minimum level. The active chip is highlighted; click
  it to clear.
- The search box filters text case-insensitively. It debounces for 250 ms
  and runs on the worker, never on the UI loop.
- `follow` jumps to the newest page and keeps the viewport pinned to the
  bottom as entries arrive. `pause` freezes the display and counts unread
  entries; the count keeps rising while ingestion continues. `N new` jumps
  to the newest page and clears the count.
- `older` and `newer` move one 200-entry page at a time. A page browse
  freezes a high-water mark, so new entries do not shift the rows.
- Click a row to select it; click the selected row or press Enter to open
  the detail pane, which shows the full multiline message. Escape or
  `< Back` returns to the list at the same reading position.
- `export` writes a bounded CSV to the exports directory. With a range
  selected (Shift+Up/Down) it exports that range, capped at 5,000 entries;
  otherwise it exports the loaded page.
- Keys: Up/Down select, Shift+Up/Down extend the export range, PageUp and
  PageDown scroll, Home loads the oldest page, End jumps to the newest,
  `f` toggles follow, `space` toggles pause, `enter` opens detail, and
  Escape closes it. Shortcuts stay out of the way while the search box has
  focus.
- The theme button switches light and dark. The choice persists.

## Storage

`-dir` overrides the location. The default is
`os.UserConfigDir()/ownframe/live-log-viewer`, which is
`~/.config/ownframe/live-log-viewer` on Linux, `%AppData%\ownframe\live-log-viewer`
on Windows, and `~/Library/Application Support/ownframe/live-log-viewer` on
macOS.

The directory holds:

- `live-log-viewer.db`: sessions, sources, ordered entries, and committed
  ingestion checkpoints. WAL when the filesystem supports it, otherwise a
  rollback journal. One writer, one serialized connection, a busy timeout,
  and foreign keys on.
- `exports/`: CSV exports from the export button.
- `ui-settings.json`: theme, follow, severity, and source filters. The core
  schema has no settings table, so the adapter keeps these beside the
  database.

Retention prunes in bounded batches: 200,000 rows, 256 MiB, or 7 days,
whichever comes first. When a cursor falls out of the retained window the
status bar says so and the page lands on the nearest rows that remain.

## Recovery

- The data directory is not writable: the example logs the error and opens a
  temporary database instead, labeled in the log. Nothing in the configured
  directory is modified.
- Database from a newer build (`store.ErrSchemaNewer`): the file is left
  untouched and Open fails. Run the newer build or move the file aside.
- Foreign or unreadable file: Open fails with `ErrForeign` or `ErrCorrupt`.
  The file is not reset.
- Second instance: SQLite locks serialize writers; reads continue. The busy
  timeout bounds waits.
- Exports are written through a temporary file and renamed, so a failure
  midway leaves no partial export.

## Limits and platform notes

- Pages hold 200 entries. Tail polls fetch 64, run every 700 ms, and count
  unread entries with one query.
- One tick drains at most 8 worker results or 2 ms of wall time. Snapshot
  requests coalesce when the bounded request channel is full; the footer
  shows the coalesced count with `-perf`.
- Only visible rows plus 8 rows of overscan render. Rows are fixed at 22 px
  and a headless test asserts that height with long Unicode and unbroken
  text, because ownframe's row windowing does not do variable-height rows.
- The window minimum is 1000x400. Smaller frames are refused; the top bar
  fits its controls at the minimum.
- Shutdown cancels the UI worker and waits up to 2 s, then cancels each
  ingestor and joins it within 2 s before the database closes.
- Native Windows behavior (100/150/200 percent scaling, monitor moves,
  minimize/restore, focus loss, rotation under a live writer) is not tested
  here. The core reader opens Windows files with
  `FILE_SHARE_READ|WRITE|DELETE`, but the checklist items stay pending until
  a native session runs.
- There is no `-web` preview. `ownframe.Serve` does not tick and its PNG
  output ignores retained operation edits, so a still preview would need an
  explicit update mechanism. This example does not add one.

## Measured notes

Measured on Linux (headless layout, no GPU window) with 200 loaded rows at
1100x720: three full `Redraw` calls took 29.1 ms total, about 9.7 ms for the
last one. The test is `TestMeasureRedrawCost` in `ui/render_cost_test.go`.
With `-perf`, the footer shows the live redraw time.

Not measured yet, and therefore unchecked in the plan:

- Per-frame replay cost of an active tick on a GPU window (needs a real
  session).
- A 30-minute mixed-workload soak with bounded heap and goroutine counts.
- p99 frame time against the 16.67 ms target on a named reference machine.

## Tests

```sh
cd examples
go test -p 2 ./live-log-viewer/...
```

The `ui` package tests are headless: they drive the page, the worker, and
the real store adapter with a temporary database. They cover keyset paging,
stale result and generation discard, debounce timing, anchor preservation
across retention and resize, follow/pause unread behavior, export bounds,
fixed row height, and the retained-operation pulse after a redraw. No window
opens, and no test touches the desktop clipboard.
