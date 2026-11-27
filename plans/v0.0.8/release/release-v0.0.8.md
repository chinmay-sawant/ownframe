# OwnFrame v0.0.8

Release on top of `v0.0.7`, delivering pure-Go pipeline optimizations with Blinkless `v0.1.10`, opt-in `Config.Render` and `Config.CullWindow` fast paths, the Builder Excel spreadsheet and Quarry SQLite/PostgreSQL examples, wide-display adaptation for Wispr Flow and Spotify, and updated benchmark baselines.

- Range: `v0.0.7` through `perf/pipeline-improvement` (2026-11-27)
- Engine: pins `github.com/chinmay-sawant/blinkless` to `v0.1.10`

## Summary

| Area | What changed |
| :--- | :--- |
| **Pipeline & Animation Performance (`internal/page`, Blinkless `v0.1.10`)** | Short-circuited unchanged box slices in `setBoxes` on 3D restamp frames, replaced 4-corner `math.Inf`/`Min`/`Max` loops in `opBoundsXform` with direct affine box math, and adopted Blinkless `v0.1.10` (`flowCBState` zero-boxing, lazy `Result`/`Display` box indices, pooled `covered` mask, stack `[16]string` transform args, partitioned text-width maps, 4-slot `stylePtr` cache, and `flexMeas.style`). |
| **Opt-In Fast Paths (`Config.Render`, `Config.CullWindow`)** | Added `Config.Render func(w io.Writer, data any) error` to bypass `html/template` reflection on hot screens (`2.40 ms` → `64.5 µs` on 1,000 rows, `37.2×` faster) and `Config.CullWindow bool` with overscan windowing on scroll plus the plain HTML/CSS `examples/cull-window` demo (`-dev`, `-perf`, `-cull`, `-debug-addr`). |
| **Builder Excel & Quarry Database Client (PR #113, PR #114)** | Added `examples/builder/excel` with click-to-edit cells, whole-UI zoom, Quarterly/Charts sheets, and pie/donut/stacked charts, plus `examples/builder/quarry` SQLite/PostgreSQL client with grid editing, word wrap, column filters, and upgraded `pgx` for security. |
| **Stability, Security & Wide Displays (PR #114, PR #116)** | Removed overflow-prone slice capacities, fixed bitmap `text-autospace` gaps at script boundaries, fixed motion transition capture indexes, stabilized the transition-reversal test clock, and adapted Wispr Flow and Spotify examples to wide displays. |

## Measured Impact (vs. `v0.0.7`)

| Benchmark | `v0.0.7` | `v0.0.8` | Improvement |
| :--- | ---: | ---: | ---: |
| `BenchmarkAnimationTick` (534-op CSS gallery) | `98.87 µs` / `38 allocs` / `16.5 KB` | **`52.36 µs` / `20 allocs` / `11.8 KB`** | **`1.89× faster`**, **`−47.4%` allocs** |
| `ScaleStages/Rows1000/DisplayList` | `11.56 ms` / `4,316 allocs` | **`10.09 ms` / `318 allocs`** | **`−12.7%` latency**, **`−92.6%` allocs** |
| `ListWorkloads/PerfOff/Resize` (1,000 rows) | `11.67 ms` / `4,295 allocs` | **`10.21 ms` / `290 allocs`** | **`−12.5%` latency**, **`−93.2%` allocs** |
| `ListWorkloads/PerfOn/Resize` (1,000 rows) | `11.91 ms` / `4,299 allocs` | **`10.17 ms` / `292 allocs`** | **`−14.6%` latency**, **`−93.2%` allocs** |
| `ListWorkloads/PerfOff/TextUpdate` (1,000 rows) | `16.74 ms` / `35,329 allocs` | **`15.05 ms` / `31,324 allocs`** | **`−10.1%` latency**, **`−11.3%` allocs** |
| `ListPatchVsRebuild/FullRebuild1Row` | `16.83 ms` / `44,267 allocs` | **`15.49 ms` / `40,278 allocs`** | **`−8.0%` latency**, **`−9.0%` allocs** |
| Opt-in `Config.Render` vs `TemplateExecute` (1,000 rows) | `2,400 µs` / `26,730 allocs` | **`64.45 µs` / `991 allocs`** | **`37.2× faster`**, **`−99.4%` bytes** |
| Opt-in `CullWindow` 600 px vs full `DisplayList` (1,000 rows) | `9.73 ms` / `2.88 MB` | **`1.24 ms` / `1.82 MB`** | **`7.83× faster`**, **`−36.7%` bytes** |

## Breaking changes

None. `Config.Render` and `Config.CullWindow` are opt-in fields on `ownframe.Config`; existing callers keep default behavior.

## Validation

- `make lint` (Go library + examples, docs, plans, tracked file size, examples inventory, internal imports, ESLint)
- `make test`
