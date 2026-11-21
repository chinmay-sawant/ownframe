# CSS Property Profiling & Dump Guide

This guide explains how to benchmark all 688 CSS properties in `examples/blinkless-all` individually and in combination, generate structured performance dumps, compare runs with `cmd/benchscore`, and attribute slow paths to `blinkless` or `ownframe` functions.

## Quick start

Run a full 688-property + Tier-2 combination + `pprof` dump:

```sh
sh scripts/perf-props-dump.sh
```

By default, output lands in `temp/perf-dumps/props-<YYYYMMDD-HHMMSS>/` (untracked scratch space under `temp/`).

For a targeted subset (for example, while optimizing filters, shadows, and masks):

```sh
sh scripts/perf-props-dump.sh --filter '^(filter|backdrop-filter|box-shadow|text-shadow|mask|clip-path)' --iters 7 --top-pairs 8 --fuzz 0 --pprof-top 6
```

To compare a candidate dump against a baseline dump using the shared benchmark scorecard:

```sh
make bench-gate BENCH_REF=temp/perf-dumps/props-before/props-bench.txt BENCH_CAND=temp/perf-dumps/props-after/props-bench.txt
```

### Running on native Windows (PowerShell) and comparing with WSL2

Run the same harness from a native Windows PowerShell terminal so you can compare WSL2 vs. Windows numbers directly (different font discovery, timer granularity, and allocator behavior):

```powershell
$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$out = "temp/perf-dumps/props-win-$stamp"
$env:OWNFRAME_PROP_DUMP_DIR = $out
$env:OWNFRAME_PROP_ITERS = "5"
$env:OWNFRAME_PROP_TOP_PAIRS = "15"
$env:OWNFRAME_PROP_FUZZ = "40"
$env:OWNFRAME_PROP_PPROF_TOP = "9"
go test -v -count=1 -timeout=20m ./examples/blinkless-all -run '^TestPropPerfDump$'
```

Then compare the WSL2 and Windows benchmark dumps (or before/after runs on either OS) with:

```sh
go run ./cmd/benchscore -ref temp/perf-dumps/props-wsl2/props-bench.txt -cand temp/perf-dumps/props-win/props-bench.txt
```

## Flags and environment variables

| Flag (`scripts/perf-props-dump.sh`) | Environment variable | Default | Purpose |
|---|---|---:|---|
| `--out DIR` | `OWNFRAME_PROP_DUMP_DIR` | `temp/perf-dumps/props-<ts>` | Output directory; unset skips `TestPropPerfDump` during normal `make test` |
| `--iters N` | `OWNFRAME_PROP_ITERS` | **5** | Timed iterations used during Pass-2 outlier refinement and shared baselines |
| `--refine-top R` | `OWNFRAME_PROP_REFINE_TOP` | **40** | Number of Pass-1 outliers re-measured with full GC-isolated precision |
| `--top-pairs M` | `OWNFRAME_PROP_TOP_PAIRS` | **10** | Auto-selected Tier-1 offenders paired all-against-all ($\binom{M}{2}$ pairs) |
| `--fuzz K` | `OWNFRAME_PROP_FUZZ` | **25** | Deterministic 5-property cross-group stacks (`OWNFRAME_PROP_SEED`, default `42`) |
| `--pprof-top P` | `OWNFRAME_PROP_PPROF_TOP` | **9** | Top offenders profiled with CPU + alloc `pprof` across `scale64`, `hover`, and `raster` |
| `--filter REGEX` | `OWNFRAME_PROP_FILTER` | *(all)* | Regular expression filter over property names |
| `--limit N` | `OWNFRAME_PROP_LIMIT` | **0** | Cap audited properties (`0` = all 688) |

## Methodology (why the numbers are trustworthy)

The harness lives in `examples/blinkless-all/perf_props_harness_test.go`, `perf_props_combos_test.go`, and `perf_props_dump_test.go`. Five design rules keep the dump free of common micro-benchmark artifacts:

1. **Paired control per property.** `specFor(r)` builds multi-node grids, tables, or inline SVGs for layout and SVG properties. Measuring `withStyle=true` minus `withStyle=false` on the exact same `DemoSpec.expand` scaffold cancels out DOM scaffold complexity.
2. **GC-isolated timing, decoupled allocations.** Latency runs with `debug.SetGCPercent(-1)` after a warmup and explicit `runtime.GC()`, never calling `runtime.ReadMemStats` inside the timed loop. Allocations are measured in a separate averaged pass. A fixed control anchor is re-checked every 100 properties and reported as `Control drift` in `summary.md`.
3. **Active sample overrides and `effective` tagging.** Every property in `props_gen.go` is audited (including properties marked `declined` in the visual catalog such as `filter`, `backdrop-filter`, `gap`, and `text-overflow`). Passive or incomplete demo samples (such as `mask-image: none` or `backdrop-filter: 12px`) are overridden with active CSS values, and the harness diffs `on` vs `off` displays to set `effective: true/false` so no-op samples cannot masquerade as fast paths.
4. **Six lifecycle phases including hover, scaling, and software raster.**
   - **Cold build:** `NewWithOptions(..., WithPerf(true))` + `Redraw` (records `StyleTime`, `LayoutTime`, `PaintTime`, `OtherTime`, `Ops`, `ExtraOps`, `BitmapFallback`).
   - **Warm unchanged redraw:** `Redraw` after a settle tick (flags any non-zero `warm_allocs`).
   - **Hover toggle (64-node `RetainState` container):** toggles `:hover` on `#h0` + `Tick`, recording scoped `HoverPatches`, `HoverFallbacks`, `HoverFailureReasons`, `DirtyRegions`, and `DirtyOps`.
   - **Resize relayout:** resizes `360x160 -> 480x200 -> 360x160` to measure cached `css.Relayout`.
   - **Scaled 64-node build:** repeats the property across 64 boxes and computes `scaling_ratio` (`scale64_delta_ns / max(1000, cold_delta_ns)`) to expose $O(N^2)$ layout or cascade scaling.
   - **Tick + uncached software raster:** measures `Tick` cost, CSS motion binding, and fresh `PNGWithContext` raster deltas.
5. **Automated `pprof` stack attribution.** For the top offenders in `scale64`, `hover`, and `raster`, the harness captures `.cpu.pprof` and `.mem.pprof` profiles under `<out>/profiles/` and embeds the top 5 non-runtime cumulative stack frames in `summary.md`.

## Tier-2 combination matrix

Testing all $\binom{688}{2} \approx 236{,}000$ pairs is slow and mostly tests orthogonal properties. Instead, `runComboMatrix` runs four focused layers:

- **Subsystem-bucket pairwise:** 15 engine mechanism buckets (`flex`, `grid`, `position`, `overflow_clip`, `border_radius`, `bg_gradient`, `shadow`, `filter_backdrop`, `filter_fx`, `blend_opacity`, `opacity`, `transform`, `typography`, `clip_shape`, `transition`) paired all-against-all.
- **Top-offender pairwise:** all pairs among the Top `M` highest-cost effective Tier-1 properties.
- **Real-world UI recipes:** 12 multi-property 64-node component stacks (glassmorphic cards, sticky table rows, dashboard grids, scrollable clip feeds, fixed modal backdrops, code blocks, animated hover cards, gradient badges, blend heroes, focus forms, mask fades, and perspective decks).
- **Seeded random stacks:** `K` deterministic 5-property cross-group stacks (`math/rand/v2` PCG).

For paired combinations, the harness computes the **interaction multiplier**:

```text
multiplier = delta(A + B) / (delta(A) + delta(B))
```

Any combination with `max_multiplier >= 1.5` is flagged as **`super_linear: true`** in `props-combos.jsonl` and ranked in `summary.md`.

## Output artifacts

| File | Format | Use |
|---|---|---|
| `summary.md` | Markdown | First stop for humans and agents: ranked Top-25 tables, hover fallbacks, warm-alloc leaks, bitmap fallbacks, super-linear combos, UI recipes, and `pprof` hot frames |
| `props-individual.jsonl` | JSON Lines | One structured record per property across all 6 phases (easy to query with `jq` or Python) |
| `props-combos.jsonl` | JSON Lines | All Tier-2 combination records with `interaction_multiplier` and `super_linear` flags |
| `props-bench.txt` | Go benchmark text | Direct input to `make bench-gate` (`cmd/benchscore`) for before/after regression gating |
| `profiles/*.pprof` | `pprof` protobuf | Interactive inspection via `go tool pprof -http=:0 <file>` |

## Triage and optimization loop

Every row in `props-individual.jsonl` and `summary.md` carries a `bottleneck_layer` tag:

| `bottleneck_layer` | Dominant signal | Where to optimize |
|---|---|---|
| **`blinkless/css`** | `StyleTime` dominates cold/resize | Upstream `blinkless` CSS value parsing, cascade, or `var()` resolution |
| **`blinkless/layout`** | `LayoutTime` dominates cold/resize/scale64 | Upstream `blinkless` flex/grid/text layout or display-list emission |
| **`ownframe/page+motion`** | `HoverFallbacks > 0`, `WarmAllocs > 0`, or `OtherTime` dominates | `internal/page` (hover scope/patch, dirty diff, `KidIndex`) or `internal/motion` |
| **`ownframe/replay+bitmap`** | `BitmapFallback == true` or `RasterDeltaNS` dominates | `internal/replay` (raster cache, Kage layers, fallback chooser) or `internal/bitmap` |

Recommended loop:

1. Capture a baseline dump with `sh scripts/perf-props-dump.sh --out temp/perf-dumps/props-base`.
2. Open `summary.md`, pick one root cause shared by several top offenders, and inspect the attached `pprof` frames (or open `go tool pprof -http=:0 temp/perf-dumps/props-base/profiles/<name>.cpu.pprof`).
3. Land the fix in `ownframe` or `blinkless` with a focused regression test.
4. Capture a candidate dump (`--out temp/perf-dumps/props-after`) and diff with `make bench-gate`.
5. Keep `TestPropPerfHarnessSmoke` green in `make test` so the profiling harness itself never drifts.
