# v0.0.8 - Quarry (SQLite & PostgreSQL database client)

> **Parent:** [plans/README.md](../../README.md)
> **Target package:** `examples/builder/quarry`
> **Status:** Partial (63 complete, 5 partial, 0 open)
> **Estimated effort:** 6 phases (iterative implementation and verification)

---

## Overview

This directory is the phase-wise checklist for **Quarry**, a local-first desktop SQLite and PostgreSQL database client under `examples/builder/quarry/`. Quarry uses the ownframe builder pattern (`ui.Mount`, `*ui.Element` trees, and `components/chadcn` controls) with `ownframe.RunWithOptions`, `CustomHeader: true`, and zero imports of `ownframe/internal/`.

Quarry opens in **dark theme by default** (`QuarryDark`) and supports runtime switching to light theme (`QuarryLight`). It ships with a seeded local SQLite database so first launch is immediately interactive, connects to real SQLite files and PostgreSQL servers via pure-Go drivers, and includes a live monitoring dashboard for connection activity, throughput, cache hit ratio, locks, and slow queries.

No checklist row may be marked `[x]` without code, tests, or screenshot evidence recorded beside the row.

## Executive summary

- **Builder architecture:** Compose the custom title bar, connection rail, collapsible schema tree, multi-tab SQL editor, virtualized result grid, cell/JSON inspector, and live monitoring dashboard from Go `*ui.Element` builders and `components/chadcn` components.
- **Dark-first theming with light toggle:** Ship `QuarryDark` as the default startup theme and `QuarryLight` as a one-click toggle, layered on `chadcn.Dark` / `chadcn.Light` with `--qr-*` tokens covering chrome, grid cells, SQL syntax highlighting, staged-diff states, and monitoring severity levels.
- **Real SQLite and PostgreSQL support in pure Go:** Use `modernc.org/sqlite` (already in `examples/go.mod`) and `github.com/jackc/pgx/v5` (examples module only, never root `go.mod`) behind a unified `Driver` interface with context cancellation, query timeouts, schema introspection, `EXPLAIN` plans, and staged transactional edits.
- **Live database monitoring:** Poll PostgreSQL (`pg_stat_database`, `pg_stat_activity`, `pg_locks`) and SQLite (`PRAGMA` storage/WAL stats + session query latency) on a background collector, draining bounded ring-buffer snapshots in `SetTick` into sparklines, active-query tables, and slow-query logs.
- **Bounded paint cost:** Render query results and table data through `ui/virtual` so 100,000+ rows scroll with constant layout and replay cost.

---

## Color schemas and design tokens

Quarry defines `--qr-*` CSS custom properties layered on top of `chadcn.Dark` (default) and `chadcn.Light`. Both themes share identical geometry (36 px title bar, 48 px connection rail, 248 px default schema sidebar, 34 px tab bar, 26 px grid row height, 28 px sticky grid header, 24 px status bar).

### 1. Primary theme tokens (`QuarryDark` default and `QuarryLight`)

| Token | Role | **QuarryDark (default)** | **QuarryLight** |
| --- | --- | --- | --- |
| `--qr-brand` | Primary accent, active tab underline, primary CTA | **#3ecf8e** | **#1f9d6b** |
| `--qr-brand-hover` | Primary button hover fill | **#34b97c** | **#178a5c** |
| `--qr-brand-tint` | Active sidebar row and subtle selection tint | **#1f2c26** | **#e6f4ed** |
| `--qr-titlebar-bg` | Custom window title bar background | **#171717** | **#fafafa** |
| `--qr-titlebar-fg` | Title bar text and window control icons | **#ededed** | **#171717** |
| `--qr-rail-bg` | Left connection/activity icon rail | **#141414** | **#f4f4f4** |
| `--qr-sidebar-bg` | Schema explorer sidebar surface | **#171717** | **#fafafa** |
| `--qr-canvas-bg` | Main workspace and grid viewport background | **#1c1c1c** | **#ffffff** |
| `--qr-surface-bg` | Cards, popovers, inspector drawer, toolbar | **#232323** | **#fafafa** |
| `--qr-surface-elevated` | Modals, autocomplete menu, context menus | **#262626** | **#ffffff** |
| `--qr-border` | Standard 1 px pane divider and gridline | **#2e2e2e** | **#e5e5e5** |
| `--qr-border-strong` | Focused input ring and active split handle | **#3e3e3e** | **#d4d4d4** |
| `--qr-text` | Primary text and cell values | **#ededed** | **#171717** |
| `--qr-text-muted` | Secondary labels, types, row numbers | **#8f8f8f** | **#6b6b6b** |
| `--qr-row-alt` | Zebra row background in result grid | **#1f1f1f** | **#fafafa** |
| `--qr-row-hover` | Hovered grid row background | **#262626** | **#f4f4f4** |
| `--qr-cell-selected` | Selected cell outline | **#3ecf8e** | **#1f9d6b** |
| `--qr-null-bg` | `NULL` badge pill background | **#2a2a2a** | **#efefef** |
| `--qr-null-fg` | `NULL` badge pill text | **#7a7a7a** | **#737373** |
| `--qr-diff-edit-bg` | Staged cell update highlight | **#34301a** | **#fdf3d3** |
| `--qr-diff-insert-bg` | Staged row insert highlight | **#1b3326** | **#dff3e8** |
| `--qr-diff-delete-bg` | Staged row delete highlight | **#3a2023** | **#fbe4e4** |
| `--qr-status-ok` | Connected badge, healthy cache hit, index scan | **#3ecf8e** | **#1a9a68** |
| `--qr-status-warn` | Slow query, lock wait, seq scan warning | **#e5b340** | **#b7791f** |
| `--qr-status-err` | Query error, disconnected, Prod env badge | **#e5645f** | **#d1453b** |

Additional implementation tokens (not part of the original table): `--qr-on-brand` (text on the accent fill), `--qr-shadow-pop` / `--qr-shadow-edge` (soft shadows for floating surfaces and hairline edges), `--qr-menu-hover`, `--qr-btn-bg` / `--qr-btn-bg-hover`, `--qr-seg-bg`, `--qr-brand-soft` and `--qr-chart-2`, plus the chadcn scrollbar tokens `--of-scrollbar` (`#4a4a4a` dark, `#c4c4c4` light), `--of-scrollbar-width` (6 px) and `--of-scrollbar-inset` (2 px). `--qr-shadow-pop` is deliberately faint (alpha 0.30 dark, 0.10 light) and used only by floating overlays. The palette is neutral near-black (`#1c1c1c` canvas, `#2e2e2e` hairlines) and near-white, with one restrained green accent; there are no gradients, because the engine ignores border-radius on them.

### 2. SQL syntax highlight tokens

| Token | Role | **QuarryDark** | **QuarryLight** |
| --- | --- | --- | --- |
| `--qr-syn-kw` | SQL keywords (`SELECT`, `FROM`, `JOIN`, `EXPLAIN`) | **#8ab4f8** | **#2563a8** |
| `--qr-syn-fn` | Built-in functions (`COUNT`, `COALESCE`, `NOW`) | **#c3a6e8** | **#7c4fb3** |
| `--qr-syn-str` | String literals | **#7fc99a** | **#2f7d4f** |
| `--qr-syn-num` | Numeric literals | **#e0a875** | **#b45f1a** |
| `--qr-syn-type` | Data types (`BIGINT`, `JSONB`, `TIMESTAMPTZ`) | **#e59bb5** | **#b13e6a** |
| `--qr-syn-comment` | `--` and `/* ... */` comments | **#6b6b6b** | **#9a9a9a** |
| `--qr-syn-op` | Operators and punctuation | **#9a9a9a** | **#5a5a5a** |

### 3. Environment safety badges

| Tag | Meaning | **QuarryDark** | **QuarryLight** |
| --- | --- | --- | --- |
| `Local` | Local SQLite or dev Postgres | **#3ecf8e** | **#1a9a68** |
| `Staging` | Shared staging database | **#e5b340** | **#b7791f** |
| `Prod` | Production (defaults to read-only lock banner) | **#e5645f** | **#d1453b** |

---

## Planned package layout (`examples/builder/quarry/`)

Divide the application into cohesive Go subpackages by ownership boundary (matching `examples/download-manager` and `examples/system-monitor`), keeping domain logic free of UI/Ebiten dependencies and keeping all CSS in separate `.css` files embedded via `//go:embed`:

```text
examples/builder/quarry/
├── main.go                  # CLI flags, signal handling, ownframe.RunWithOptions (CustomHeader)
├── driver/                  # Unified Driver interface, SQLite, PostgreSQL, mock PG, seeded demo DB
│   ├── driver.go            # Driver interface, CellValue, SchemaNode, QueryResult, read-only guard
│   ├── sqlite.go            # modernc.org/sqlite driver, PRAGMA introspection, EXPLAIN QUERY PLAN
│   ├── postgres.go          # pgx/v5 pool driver, pg_catalog introspection, EXPLAIN JSON, pg_stat_*
│   ├── pgmock.go            # Deterministic offline mock Postgres driver for headless/CI tests
│   └── seed.go              # Seeded local SQLite demo DB (users, products, orders, audit_events)
├── sqlkit/                  # Pure-Go SQL tooling (zero UI dependency)
│   ├── highlight.go         # Dialect-aware SQLite/Postgres SQL tokenizer and syntax spans
│   ├── statement.go         # Statement-at-cursor and multi-statement script splitter
│   ├── autocomplete.go      # Schema- and alias-aware completion engine
│   ├── changeset.go         # Staged UPDATE/INSERT/DELETE tracker and dialect SQL generator
│   └── io.go                # CSV/JSON/NDJSON/SQL export and CSV type-inferred import
├── store/                   # Local SQLite config database (os.UserConfigDir()/ownframe/quarry/)
│   └── store.go             # Saved connections, query history, snippets, UI/theme preferences
├── monitor/                 # Live telemetry collector (zero UI dependency)
│   ├── ring.go              # Bounded MetricRing buffer and p50/p95 latency tracker
│   └── collector.go         # Background polling goroutine, TPS/rate deltas, slow-query capture
└── ui/                      # ownframe builder UI (ui.Mount + components/chadcn)
    ├── styles/              # Separate CSS files embedded via //go:embed
    │   ├── base.css         # Shared layout geometry, grid, editor, tabs, monitor styles
    │   ├── dark.css         # QuarryDark (--qr-* tokens, default startup theme)
    │   └── light.css        # QuarryLight (--qr-* tokens)
    ├── theme.go             # //go:embed styles/*.css loader composing chadcn + Quarry themes
    ├── app.go               # App state, ui.Mount, ownframe.Handlers, SetTick channel draining
    ├── shell.go             # CustomHeader title bar, activity rail, 100vh workbench layout
    ├── sidebar.go           # Collapsible schema tree, filter input, Structure/DDL inspector
    ├── tabs.go              # Multi-tab strip (Query, Table, Structure, Monitor)
    ├── editor.go            # SQL editor pane, gutter, toolbar, autocomplete popover
    ├── explain.go           # Visual EXPLAIN plan tree and Messages sub-tab
    ├── grid.go              # ui/virtual result grid, typed cells, filter/sort, changeset bar
    ├── inspector.go         # Right-hand cell detail and pretty JSON inspector drawer
    ├── monitor_view.go      # Monitor KPI cards, storage breakdown, active queries, slow-query log
    ├── monitor_charts.go    # Live SVG sparkline/area charts (TPS, conns, cache hit, latency)
    ├── dialogs.go           # Connection manager, export/import, destructive-action confirm modals
    ├── palette.go           # Fuzzy Ctrl+K command palette
    └── statusbar.go         # Bottom connection, latency, row-count, and theme status bar
```

---

## Sample code blueprints

These sketches fix the key interfaces and wiring before Phase 01 starts.

### 1. Unified database driver interface (`driver/driver.go`)

```go
package driver

import (
	"context"
	"time"
)

type EngineKind string

const (
	EngineSQLite   EngineKind = "sqlite"
	EnginePostgres EngineKind = "postgres"
)

type ConnectionConfig struct {
	ID       string
	Name     string
	Engine   EngineKind
	EnvTag   string // "Local", "Staging", "Prod"
	ReadOnly bool
	FilePath string // SQLite path (or ":memory:")
	Host     string // Postgres fields
	Port     int
	Database string
	User     string
	Password string // Stored in memory only unless user opts in
	SSLMode  string
}

type ColumnMeta struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	DefaultVal string
}

type QueryResult struct {
	Columns      []ColumnMeta
	Rows         [][]CellValue
	RowsAffected int64
	Duration     time.Duration
	Truncated    bool
}

type MetricSnapshot struct {
	CapturedAt     time.Time
	ActiveConns    int
	IdleConns      int
	WaitingLocks   int
	TPS            float64
	CacheHitRatio  float64
	DBSizeBytes    int64
	ActiveQueries  []ActiveQuery
}

type Driver interface {
	Engine() EngineKind
	Ping(ctx context.Context) error
	Close() error
	ListSchemas(ctx context.Context) ([]SchemaNode, error)
	TableColumns(ctx context.Context, schema, table string) ([]ColumnMeta, error)
	TableDDL(ctx context.Context, schema, table string) (string, error)
	Query(ctx context.Context, sqlText string, maxRows int) (QueryResult, error)
	Exec(ctx context.Context, sqlText string) (int64, error)
	Explain(ctx context.Context, sqlText string, analyze bool) (ExplainPlan, error)
	CollectMetrics(ctx context.Context) (MetricSnapshot, error)
	CancelQuery(ctx context.Context, pid int) error
}
```

### 2. Separate embedded CSS themes and `ui.Mount` shell (`ui/theme.go`, `ui/app.go`)

```go
package ui

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

//go:embed styles/base.css
var baseCSS string

//go:embed styles/dark.css
var darkCSS string

//go:embed styles/light.css
var lightCSS string

func ThemeCSS(dark bool) string {
	if dark {
		return chadcn.Dark + "\n" + chadcn.Stylesheet + "\n" + darkCSS + "\n" + baseCSS
	}
	return chadcn.Light + "\n" + chadcn.Stylesheet + "\n" + lightCSS + "\n" + baseCSS
}

func NewApp(opts LaunchOptions) (*App, error) {
	a := &App{
		dark:       true, // Dark theme by default
		activeView: viewWorkspace,
		pollEvery:  2 * time.Second,
	}
	if err := a.initDemoOrFlagDB(opts.DBPath); err != nil {
		return nil, err
	}
	page, err := ui.Mount(a.render, ui.MountOptions{
		Width:       1440,
		Height:      960,
		Theme:       ThemeCSS(a.dark),
		LockView:    true,
		RetainState: true,
		Scrollbar:   ownframe.ScrollbarHidden,
		Handlers:    a.handlers(),
	})
	if err != nil {
		return nil, err
	}
	a.page = page
	a.page.SetTick(a.onTick)
	return a, nil
}

func (a *App) toggleTheme() {
	a.dark = !a.dark
	a.page.SetTheme(ThemeCSS(a.dark))
	a.redraw()
}
```

### 3. Virtualized result grid row window (`ui/grid.go`)

```go
func (a *app) renderResultGrid(res *QueryResult, scrollY, viewH int) *ui.Element {
	const rowH = 26
	win := virtual.Window(len(res.Rows), rowH, scrollY, viewH, 4)

	rows := make([]*ui.Element, 0, win.End-win.Start+2)
	if win.TopSpacer > 0 {
		rows = append(rows, ui.Div().Style("height", fmt.Sprintf("%dpx", win.TopSpacer)))
	}
	for idx := win.Start; idx < win.End; idx++ {
		rows = append(rows, a.renderGridRow(idx, res.Columns, res.Rows[idx]))
	}
	if win.BottomSpacer > 0 {
		rows = append(rows, ui.Div().Style("height", fmt.Sprintf("%dpx", win.BottomSpacer)))
	}
	return ui.Div().Class("qr-grid-body").Children(rows...)
}
```

### 4. Bounded monitor collector draining on `SetTick` (`ui/app.go`, `monitor/collector.go`)

```go
func (a *app) onTick(dt float64) bool {
	dirty := false
	// Drain completed async query results without blocking the render loop.
	for i := 0; i < 8; i++ {
		select {
		case res := <-a.queryDoneCh:
			a.applyQueryResult(res)
			dirty = true
		case snap := <-a.metricsCh:
			a.metricsRing.Push(snap) // Fixed capacity (e.g. 120 samples)
			dirty = true
		default:
			goto done
		}
	}
done:
	if dirty {
		a.redraw()
	}
	return dirty
}
```

---

## Phase progress

| Phase | File | Complete | Partial | Open |
| --- | --- | ---: | ---: | ---: |
| Phase 01: Shell, dark/light themes, and connection manager | [phase-01-shell-theme-and-connections.md](phase-01-shell-theme-and-connections.md) | 12 | 0 | **0** |
| Phase 02: SQLite & PostgreSQL drivers and schema explorer | [phase-02-drivers-and-schema-explorer.md](phase-02-drivers-and-schema-explorer.md) | 10 | 2 | **0** |
| Phase 03: SQL editor, autocomplete, async runner, and EXPLAIN | [phase-03-sql-editor-and-query-execution.md](phase-03-sql-editor-and-query-execution.md) | 12 | 0 | **0** |
| Phase 04: Virtualized data grid, staged edits, and import/export | [phase-04-data-grid-editing-and-io.md](phase-04-data-grid-editing-and-io.md) | 12 | 0 | **0** |
| Phase 05: Live database monitoring and slow-query diagnostics | [phase-05-live-monitoring-and-diagnostics.md](phase-05-live-monitoring-and-diagnostics.md) | 9 | 1 | **0** |
| Phase 06: Command palette, accessibility, benchmarks, and verification | [phase-06-polish-accessibility-and-verification.md](phase-06-polish-accessibility-and-verification.md) | 8 | 2 | **0** |
| **Total** | | **63** | **5** | **0** |

---

## Execution and evidence rules

- Use three statuses consistently across all phase files:
  - `[ ]` not started or not proven;
  - `[x]` implemented and validated with current evidence (commit hash, command, exit status, and measured result);
  - `[~]` intentionally deferred or partial, with reason, owner boundary, and next gate.
- Keep `examples/builder/quarry` free of `ownframe/internal/` imports (`make internal-import-check` must stay green).
- Keep `github.com/jackc/pgx/v5` and `modernc.org/sqlite` strictly inside `examples/go.mod`, never in the root `go.mod`.
- Use `ownframe.UseMemory(true)` in every test that touches the clipboard.
- All unit and headless tests must pass with zero external PostgreSQL server running (use seeded SQLite + mock/stub Postgres driver for CI, plus an opt-in `QUARRY_TEST_PG_DSN` integration test).
- Run `gofmt`, `make lint`, and `make test` before closing any code phase. Run `make graphics-check` when visual or pixel tests are added.

## Dependencies

- **Phase 01 before Phase 02:** The dark-first theme sheets, `ui.Mount` shell, and connection config store must compile and render before live drivers populate the schema tree.
- **Phase 02 before Phase 03 and Phase 04:** The unified `Driver` interface, schema metadata, and DDL/query execution are prerequisites for autocomplete, `EXPLAIN`, virtualized grids, and staged table edits.
- **Phase 02 and Phase 03 before Phase 05:** Live monitoring reuses the active `Driver`, background goroutine channel draining in `SetTick`, and the SQL editor for "Open in Editor & Explain".
- **Phase 01 through Phase 05 before Phase 06:** Command palette actions, accessibility labels, pixel theme checks, and virtual-grid / monitor-tick benchmarks validate the integrated app.
