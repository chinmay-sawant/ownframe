# v0.0.8 - Phase 01: Shell, dark/light themes, and connection manager

> **Parent:** [README.md](README.md)
> **Status:** Complete (12 complete, 0 partial, 0 open)
> **Estimated effort:** 2-3 days

---

## Overview

Phase 01 creates the runnable Quarry application under `examples/builder/quarry/`, sets up `QuarryDark` as the default theme with a runtime `QuarryLight` toggle, builds the multi-pane `CustomHeader` window shell, persists saved connections in a local SQLite config store, and seeds a realistic local SQLite demo database so the app opens ready to query.

## Executive summary

- Create the modular package tree (`main.go`, `ui/`, `ui/styles/`, `store/`, and `driver/seed.go`).
- Put all CSS in separate `.css` files (`ui/styles/dark.css`, `ui/styles/light.css`, `ui/styles/base.css`) embedded via `//go:embed` in `ui/theme.go`, defaulting startup to `QuarryDark` (`#1c1c1c` canvas, `#3ecf8e` accent) with instant runtime switching to `QuarryLight`.
- Wire the Connection Manager modal supporting both SQLite file paths (including drag-and-drop) and PostgreSQL connection settings, tagged as `Local`, `Staging`, or `Prod`.

---

## Phase 1: Shell, dark/light themes, and connection manager

### 1.1 Application scaffold and custom window chrome

- [x] `QRY-01` Create `examples/builder/quarry/main.go` with `-web`, `-headless`, `-addr` (default `127.0.0.1:8150`), `-db` (optional SQLite path or Postgres DSN), `-data-dir` (override config dir), `-dev`, `-perf`, and `-debug-addr` flags, wiring `ownframe.RunWithOptions` with `CustomHeader: true`, `PlacementID: "quarry"`, and a `Draggable` guard that yields when any modal, popover, or menu is open. Proof: `go vet -p 1 ./examples/builder/quarry/...` exits 0 and `TestDragGuardYieldsToOverlays` passes. **UI:** `main.go` wires flags, `CustomHeader`, `PlacementID: "quarry"` and `App.Draggable`; `go vet -p 1 ./examples/builder/quarry/...` exit 0; `TestDragGuardYieldsToOverlays` passes (`ui/shell_test.go`).
- [x] `QRY-02` Create `examples/builder/quarry/ui/app.go` defining the `App` struct, `NewApp()`, `ui.Mount` with `LockView: true`, `RetainState: true`, `Scrollbar: ownframe.ScrollbarHidden`, and `ownframe.Handlers` (`Click`, `Press`, `DragStart`, `DragMove`, `DragEnd`, `Wheel`, `Hover`, `KeyDown`, `Change`, `Type`, `Backspace`, `Copy`, `Paste`, `Undo`, `Redo`, `Drop`). Proof: `TestNewAppMounts` mounts cleanly at 1440x960 and 1280x800 with zero errors. **UI:** `ui/app.go` mounts with `LockView`, `RetainState`, `ScrollbarHidden` and the full handler set; `TestNewAppMounts` passes at 1440x960 and 1280x800.
- [x] `QRY-03` Build the Quarry custom title bar in `examples/builder/quarry/ui/shell.go` with the Quarry badge, active connection selector pill (with engine icon and `Local`/`Staging`/`Prod` color dot), read-only lock indicator, command palette trigger (`Ctrl+K`), dark/light theme toggle button, and desktop window controls (`win-min`, `win-max`, `win-close` via `data-action`). Proof: `TestTitleBarActions` verifies window-control `data-action` attributes and theme toggle redraw. **UI:** `ui/shell.go`; `TestTitleBarActions` checks `win-min`/`win-max`/`win-close` `data-action` and theme-toggle repaint. **Reopened and re-closed:** glyph text replaced by SVG icons (logo, sun/moon theme toggle, minimize/maximize/restore/close); `TestNoTextGlyphsStandInForIcons`, `TestWindowControlsAndResize` (Restore label follows window state, resize re-lays out shell and grid) pass. Real maximize/restore needs a desktop window and is untested here.
- [x] `QRY-04` Build the top-level 100vh shell layout in `ui/shell.go` with fixed slots for title bar (36 px), left activity rail (48 px: Connections, Explorer, Monitor, History, Settings), collapsible schema sidebar (248 px), main workspace (`flex: 1; min-height: 0`), and bottom status bar (24 px). Proof: `TestShellFitsViewport` asserts on `page.Probe().Display()` that all panes fit inside 1440x960 and 1280x800 with zero overflow. **UI:** `ui/shell.go`; `TestShellFitsViewport` asserts title 36, rail 48, sidebar 248, tab bar 34, status 24 and no overflow at both sizes from `Probe().Display()` boxes.

### 1.2 Dark-first separate CSS schemas and runtime theme toggle

- [x] `QRY-05` Create separate CSS files `ui/styles/base.css`, `ui/styles/dark.css` (`QuarryDark`, default), and `ui/styles/light.css` (`QuarryLight`), embedded via `//go:embed` in `examples/builder/quarry/ui/theme.go`, implementing every `--qr-*` chrome, grid, diff, status, and `--qr-syn-*` SQL highlight token specified in `README.md`, layered on `chadcn.Dark + chadcn.Stylesheet` and `chadcn.Light + chadcn.Stylesheet`. Proof: `TestThemeTokensPresent` verifies every documented token exists as a valid hex color in both CSS files and mounts cleanly. **UI:** `ui/styles/{base,dark,light}.css` embedded in `ui/theme.go`; `TestThemeTokensPresent` passes.
- [x] `QRY-06` Ensure Quarry boots into `QuarryDark` by default and persists theme preference in `store`, with runtime switching via `App.toggleTheme()` preserving active tabs, query text, and selection. Proof: `TestDefaultDarkAndThemeToggle` asserts default dark canvas pixels, distinct light canvas pixels after toggle, and preserved state across two toggles. **UI:** `TestDefaultDarkAndThemeToggle` asserts canvas pixels #161a24 dark / #ffffff light, preserved tabs and text, and the stored `theme` pref. **Re-verified:** `TestFreshStoreStartsDark` (fresh data dir paints QuarryDark; a saved light choice restores). I could not reproduce a light first launch with a fresh `-data-dir`; a leftover `theme=light` in a state.db produces exactly that.

### 1.3 Local config store, connection manager, and seeded demo DB

- [x] `QRY-07` Implement `examples/builder/quarry/store/store.go` backed by SQLite under `os.UserConfigDir()/ownframe/quarry/state.db` (overridable via `-data-dir`, `:memory:` in tests) storing saved connection profiles (excluding raw passwords unless explicitly opted in), query history, saved snippets, and UI preferences. Proof: `store/store_test.go` verifies CRUD, persistence across reopen, and `:memory:` isolation. **Evidence:** `store/store.go` (connections without passwords unless opted in, history capped at 2000, snippets, prefs, theme default dark, `:memory:` and `<dir>/state.db`); `TestConnectionCRUD`, `TestRecentSQLiteFile`, `TestHistorySnippetsPrefs`, `TestHistoryPrunes`, `TestPersistenceAcrossReopenAndMemoryIsolation` pass (`go test -count=1 ./examples/builder/quarry/store/` exit 0, `-race` exit 0).
- [x] `QRY-08` Build the Connection Manager modal in `examples/builder/quarry/ui/dialogs.go` with engine switcher (`SQLite` vs `PostgreSQL`), SQLite file picker + drag-and-drop `.db`/`.sqlite` support, PostgreSQL host/port/database/user/password/SSLMode form, environment tag (`Local`/`Staging`/`Prod`), read-only default for `Prod`, and a `Test Connection` button with inline status readout. Proof: `TestConnectionModalFlow` validates switching engines, saving a profile, and rejecting empty required fields. **UI:** `ui/dialogs.go`; `TestConnectionModalFlow` passes (engine switch, save, empty path rejected, Test Connection, Prod read-only default).
- [x] `QRY-09` Wire OS file drop (`Handlers.Drop`) so dropping a `.db`, `.sqlite`, or `.sqlite3` file onto the Quarry window opens it immediately as a SQLite connection and adds it to recent connections. Proof: `TestDropSQLiteFileOpensConnection` passes in headless test. **UI:** `ui/dialogs.go` `onDrop`; `TestDropSQLiteFileOpensConnection` passes.
- [x] `QRY-10` Create `examples/builder/quarry/driver/seed.go` generating a realistic local SQLite demo database (`quarry_demo.db`) with `users` (500 rows), `products` (120 rows), `orders` (2,500 rows), `order_items` (6,000 rows), `audit_events` (5,000 rows with JSON payloads), foreign keys, secondary indexes, and two summary views (`v_revenue_by_day`, `v_top_customers`). Proof: `TestSeedDemoDB` creates the DB in a temp dir and verifies table counts, views, and foreign keys. **Evidence:** `driver/seed.go` (`SeedDemoDB`, `EnsureDemoDB`, deterministic, idempotent); `TestSeedDemoDB` in `driver/sqlite_test.go` verifies 500/120/2500/6000/5000 rows, both views, and order_items foreign keys (`go test -count=1 ./examples/builder/quarry/driver/` exit 0). Test lives in the driver package, not `ui`.

### 1.4 Status bar skeleton and phase gate

- [x] `QRY-11` Render the bottom status bar in `examples/builder/quarry/ui/statusbar.go` showing connection status dot, engine version string, active database/schema, read-only badge, last query latency + row count, and current theme indicator. Proof: `TestStatusBarReadout` checks status bar elements via `page.Probe().Display()`. **UI:** `ui/statusbar.go`; `TestStatusBarReadout` passes.
- [x] `QRY-12` Run Phase 01 closure gates: `gofmt -w examples/builder/quarry`, `go vet -p 1 ./examples/builder/quarry/...`, `go test -v ./examples/builder/quarry/...`, `make internal-import-check`, and `make examples-check`. Proof: all commands exit 0 with results recorded beside this row. **UI:** `gofmt -l examples/builder/quarry` clean; `go vet -p 1 ./examples/builder/quarry/...` and `go test -p 1 ./examples/builder/quarry/...` pass; `make internal-import-check` and `make examples-check` exit 0.

---

## Dependencies

- `QRY-05` must land before `QRY-06` so pixel checks compare real `QuarryDark` and `QuarryLight` tokens.
- `QRY-07` and `QRY-10` must land before `QRY-08` so the Connection Manager can save profiles and auto-load the seeded demo database on first launch.
- Do not import any package under `github.com/chinmay-sawant/ownframe/internal/`.
