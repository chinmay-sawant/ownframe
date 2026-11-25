# v0.0.8 - Phase 05: Live database monitoring and slow-query diagnostics

> **Parent:** [README.md](README.md)
> **Status:** Open (0 complete, 0 partial, 10 open)
> **Estimated effort:** 2-3 days

---

## Overview

Phase 05 adds the built-in **Monitor** tab to Quarry, giving developers real-time visibility into database health without switching to an external Grafana or CLI session. It polls PostgreSQL system views or SQLite pragmas/session stats on a background goroutine, stores a bounded rolling history in a ring buffer, and renders live KPI cards, sparkline charts, active queries/locks, and a slow-query log.

## Executive summary

- Follow the `system-monitor` collector pattern: background goroutine polls `Driver.CollectMetrics` on a configurable interval (`1s`, `2s`, `5s`, `Paused`) and sends snapshots over a bounded channel drained in `SetTick`.
- Support both engines meaningfully: PostgreSQL live server telemetry (`pg_stat_database`, `pg_stat_activity`, `pg_locks`) and SQLite storage/WAL/cache diagnostics plus per-session query latency percentiles.
- Include a synthetic Postgres telemetry stream in the mock PG driver so the full monitoring dashboard can be explored and tested offline.

---

## Phase 5: Live database monitoring and slow-query diagnostics

### 5.1 Background collector and ring-buffer state

- [ ] `QRY-49` Implement `examples/builder/quarry/monitor/ring.go` and `monitor/collector.go` with a bounded `MetricRing` (default 120 samples, zero unbounded growth), rate delta calculator (deriving TPS and tuples/sec from monotonic Postgres `xact_commit + xact_rollback` counters), and goroutine lifecycle tied to the active connection with clean cancellation on switch or close. Proof: `monitor/collector_test.go` verifies ring wrap-around, rate math, and goroutine exit with zero leaks.
- [ ] `QRY-50` Implement `CollectMetrics` in `driver/postgres.go` querying `pg_stat_database` (commits, rollbacks, `blks_hit` / `blks_read` cache hit ratio, `pg_database_size`), `pg_stat_activity` (active, idle, idle-in-transaction, waiting counts + running query rows), and `pg_locks` (blocked vs blocking PIDs), plus `CancelQuery` (`pg_cancel_backend`) and `TerminateQuery` (`pg_terminate_backend`). Proof: SQL query builders and row decoders pass unit tests against captured fixtures + mock PG driver.
- [ ] `QRY-51` Implement `CollectMetrics` in `driver/sqlite.go` querying `PRAGMA page_count`, `PRAGMA page_size`, `PRAGMA freelist_count`, `PRAGMA journal_mode`, `PRAGMA cache_size`, WAL file size (when present), per-table size estimates, and combining them with Quarry's in-process session query counter and p50/p95 latency tracker. Proof: `driver/sqlite_test.go` verifies non-zero DB size, page stats, and latency percentiles after running queries.

### 5.2 Monitor dashboard UI and live charts

- [ ] `QRY-52` Build the Monitor tab header and KPI strip in `examples/builder/quarry/ui/monitor_view.go`: poll-interval selector (`1s`, `2s`, `5s`, `Pause`), `Refresh Now` button, `Freeze Snapshot` toggle, and four KPI cards (`Active Connections` / `Journal Mode`, `Transactions / sec`, `Cache Hit Ratio %`, `Database Size & Latency p95`) styled with `--qr-status-ok` / `--qr-status-warn` / `--qr-status-err` thresholds. Proof: `TestMonitorKPICards` asserts KPI card values and threshold classes on dark and light themes.
- [ ] `QRY-53` Build live time-series charts in `examples/builder/quarry/ui/monitor_charts.go` using builder SVG/layout primitives (`components/chadcn` chart style): 2x2 grid of sparkline/area charts for `Throughput (TPS / QPS)`, `Connections (Active vs Idle vs Waiting)`, `Cache Hit Ratio (%)`, and `Query Latency (p50 / p95 ms)` with hover tooltips and min/max/current labels. Proof: `TestMonitorChartsRender` feeds 60 samples into `MetricRing` and asserts non-empty SVG/display ops with zero NaN coordinates.
- [ ] `QRY-54` Render a Storage & Table Breakdown panel in `ui/monitor_view.go` showing top tables by row count / disk footprint with horizontal proportion bars and a one-click button to jump to that table's Data or Structure tab. Proof: `TestMonitorStorageBreakdown` verifies sorted table bars on the seeded SQLite DB.

### 5.3 Active queries, lock waits, and slow-query log

- [ ] `QRY-55` Build the Active Queries & Locks table in `ui/monitor_view.go`: filter by state (`All`, `Active`, `Waiting`, `Idle in Tx`), display PID, user, duration pill, wait event, query preview, and `Cancel` / `Terminate` action buttons gated behind a confirmation dialog (disabled in read-only mode). Proof: `TestActiveQueriesAndCancelAction` tests filtering and cancel dispatch against the mock PG driver.
- [ ] `QRY-56` Build the Slow-Query Log panel in `ui/monitor_view.go`: automatically capture any query executed in Quarry (or spotted in `pg_stat_activity`) exceeding a configurable threshold (default `50 ms`), showing timestamp, duration, rows, SQL preview, and a one-click `Open in Editor & Explain` button. Proof: `TestSlowQueryCaptureAndOpenExplain` captures a slow query and opens a pre-filled Explain tab.
- [ ] `QRY-57` Ensure monitoring adds zero background CPU/repaint work when the Monitor tab is paused or not open (respecting `ownframe`'s idle frame parking when idle). Proof: `TestMonitorIdleWhenPaused` verifies `onTick` returns `false` when paused and no query completions are queued.
- [ ] `QRY-58` Run Phase 05 closure gates: `gofmt`, `go vet -p 1 ./examples/builder/quarry/...`, `go test -v ./examples/builder/quarry/...`, and `make internal-import-check`. Proof: all commands exit 0 with results recorded beside this row.

---

## Dependencies

- `QRY-14`, `QRY-17`, and `QRY-20` (Phase 02) must land before `QRY-50` and `QRY-51`.
- `QRY-34` (Phase 03) must land before `QRY-56` wires `Open in Editor & Explain`.
- Keep `SetTick` strictly for draining completed collector/query channels (document this reason in `examples/readme.md` Motion audit table in Phase 06).
