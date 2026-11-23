# CSS Property Profiling & Dump Guide

This guide explains how to benchmark all 688 CSS properties in `examples/blinkless-all` individually and in combination, generate structured performance dumps, compare runs with `scripts/cmd/benchscore`, and attribute slow paths to `blinkless` or `ownframe` functions.

## Quick start

Run the full property and combination matrix with `pprof`:

```sh
sh scripts/perf-props-dump.sh
```

By default, output lands in `temp/perf-dumps/props-<YYYYMMDD-HHMMSS>/` (untracked scratch space under `temp/`).

For a targeted subset (for example, while optimizing filters, shadows, and masks):

```sh
sh scripts/perf-props-dump.sh --filter '^(filter|backdrop-filter|box-shadow|text-shadow|mask|clip-path)' --iters 7 --top-pairs 8 --fuzz 0 --pprof-top 6
```

## Compare two revisions

Run the same harness and fixture sources on both revisions. Use the same host, Go version, CPU settings, and sampling flags. Every dump writes `manifest.json` with its measurement version (currently **3**) and sampling flags. `--reference` refuses a dump with another version, no manifest (version 2 and older), or different flags. When the baseline revision predates version 3, copy the current benchmark and fixture sources into the baseline checkout and generate a new baseline there.

Generate the baseline in the baseline checkout. Generate the candidate in the candidate checkout. Use output paths accessible from both checkouts.

```sh
# Baseline checkout
GOMAXPROCS=1 sh scripts/perf-props-dump.sh --out /tmp/ownframe-props-before

# Candidate checkout
GOMAXPROCS=1 sh scripts/perf-props-dump.sh --out /tmp/ownframe-props-after \
  --reference /tmp/ownframe-props-before

make bench-gate BENCH_REF=/tmp/ownframe-props-before/props-bench.txt \
  BENCH_CAND=/tmp/ownframe-props-after/props-bench.txt
```

`--reference` reuses the baseline's property and combination refinement choices, top-offender pairs, and seeded stacks. The harness checks the manifest, the property samples, and the combination declarations. Keep the sampling flags, filter, and seed identical.

`props-bench.txt` holds only refined rows, so pass `--reference` on the candidate run. Both runs then refine the same rows, and `benchscore` fails a row the candidate lacks.

### Running on native Windows (PowerShell) and comparing with WSL2

Use directories visible to both systems and match the baseline settings. The comparison includes operating system differences such as font discovery and timer behavior.

```powershell
$stamp = Get-Date -Format "yyyyMMdd-HHmmss"
$env:OWNFRAME_PROP_DUMP_DIR = Join-Path (Get-Location).Path "temp/perf-dumps/props-win-$stamp"
$env:OWNFRAME_PROP_REFERENCE_DIR = Join-Path (Get-Location).Path "temp/perf-dumps/props-wsl2"
$env:GOMAXPROCS = "1"
$env:OWNFRAME_PROP_ITERS = "5"
$env:OWNFRAME_PROP_TOP_PAIRS = "10"
$env:OWNFRAME_PROP_FUZZ = "25"
$env:OWNFRAME_PROP_PPROF_TOP = "9"
go test -v -count=1 -timeout=20m ./examples/blinkless-all -run '^TestPropPerfDump$'
```

Then compare the WSL2 and Windows benchmark dumps (or before/after runs on either OS) with:

```sh
make bench-gate BENCH_REF=temp/perf-dumps/props-wsl2/props-bench.txt BENCH_CAND=temp/perf-dumps/props-win/props-bench.txt
```

## Flags and environment variables

| Flag (`scripts/perf-props-dump.sh`) | Environment variable | Default | Purpose |
|---|---|---:|---|
| `--out DIR` | `OWNFRAME_PROP_DUMP_DIR` | `temp/perf-dumps/props-<ts>` | Output directory. The test skips when the environment variable is unset. |
| `--reference DIR` | `OWNFRAME_PROP_REFERENCE_DIR` | *(unset)* | Version 3 dump with a matching `manifest.json` that supplies the combination selection and refinement choices |
| `--iters N` | `OWNFRAME_PROP_ITERS` | **5** | Timed iterations used during Pass-2 outlier refinement and shared baselines |
| `--refine-top R` | `OWNFRAME_PROP_REFINE_TOP` | **40** | Maximum number of sweep outliers to refine across phase rankings |
| `--top-pairs M` | `OWNFRAME_PROP_TOP_PAIRS` | **10** | Auto-selected Tier-1 offenders paired all-against-all ($\binom{M}{2}$ pairs) |
| `--fuzz K` | `OWNFRAME_PROP_FUZZ` | **25** | Deterministic 5-property cross-group stacks (`OWNFRAME_PROP_SEED`, default `42`) |
| `--pprof-top P` | `OWNFRAME_PROP_PPROF_TOP` | **9** | Top offenders profiled with CPU + alloc `pprof` across `scale64`, `hover`, and `raster` |
| `--filter REGEX` | `OWNFRAME_PROP_FILTER` | *(all)* | Regular expression filter over property names |
| `--limit N` | `OWNFRAME_PROP_LIMIT` | **0** | Cap audited properties (`0` = all 688) |

## Measurement method

The property sweep in `examples/blinkless-all` takes one sample per workload. Refinement selects rows across phase rankings, then takes repeated samples. Scale and raster refinement use half the requested iterations, rounded down to at least one.

1. **Paired controls.** Cold build, hover, resize, 64-node build, and PNG timing each use the same fixture with and without the property. The timer alternates which workload runs first. Combination measurements use shared controls.
2. **Timing and allocations.** Each timed loop disables garbage collection. Refined pairs share an explicit garbage collection boundary before timing. The harness samples allocations separately. Use the full sweep to find candidates and repeat the relevant cases before attributing small timing changes.
   - **Drift.** `summary.md` reports two drift figures. The control drift takes the median 64-node control of each window of 100 properties during the sweep and reports the largest change from the first window. The warning above 15% uses this figure, because the 64-node control is the same page for every property. The cold control figure is informational only: each cold control is that property's own page without the declaration, so its window medians change with the property groups. The anchor drift times one fixed page at the start and every 100 properties, each time with the same warm-up and garbage collection, and reports the median and maximum change from the start.
   - **Noise floor.** For each phase, the floor is the 95th percentile of the absolute delta over rows with no observed effect. Those rows add no work, so their spread is the measurement noise. Ranked tables mark a delta at or below the floor as `≤noise`.
3. **Effect evidence.** The harness audits every property in `props_gen.go`, including properties the visual catalog marks as declined. It overrides some passive samples, such as `mask-image: none`. A row sets `effective` when its paired fixture shows a display signature, motion, or PNG difference. `effect_evidence` records the observations. A row with no observed effect needs a stronger fixture or another way to verify the declaration.
4. **Measured operations.** Cold build times `NewWithOptions` and `Redraw`. The unchanged redraw (`warm_*` fields) repeats `Redraw` with the same input after a settle tick, with a paired control. A page with the same input skips the rebuild, including during active motion, so a nonzero allocation count here means the skip failed. Hover times enter, tick, leave, and tick on a 64-node retained page. Resize cycles from 360 × 160 to 480 × 200 and back. The 64-node build repeats the property across uniform boxes. The scaling ratio divides the 64-box delta by the delta of the same declaration on one uniform box, and it is empty when either delta is inside timer noise. A growth-rate claim needs more node counts.
5. **Tick and PNG.** The harness times `Tick` separately, and the summary flags rows whose tick allocates. It times the PNG two ways on fresh pages. `raster_full_*` times `Redraw` plus `PNGWithContext` and is comparable on every row. `raster_*` times `PNGWithContext` alone. A bitmap-fallback page has already painted inside `Redraw`, so the PNG-only delta is marked invalid (`raster_delta_valid: false`) when the control and the variant paint on different paths. The harness decodes every timed PNG after the timer stops. Invalid PNG samples record a phase error.
6. **Profiles.** The harness captures CPU and allocation profiles for selected slow scale, hover, and PNG cases. The runtime's allocation profile counts from process start, so the harness writes `<case>.mem-base.pprof` before the workload and `<case>.mem.pprof` after it. The summary's allocation frames come from `go tool pprof -base`. Open the pair with `go tool pprof -base <case>.mem-base.pprof <case>.mem.pprof`. The PNG profile includes page setup and redraw. Use the phase time and call stack together when distinguishing paint from encoding. Use a separate replay benchmark to measure Ebiten rendering.

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

Each combination times its own control, the same fixture without the declarations, in the same loop. The denominator sums the positive deltas of each single declaration, measured on the combination's own fixtures at the same sample count. A sum at or below 50 µs is inside timer noise, so the multiplier stays 0 and cannot set the flag. The dump flags `max_multiplier >= 1.5` as **`super_linear: true`**. Use the flag to select cases for repeated measurement. Entries marked `refined` were re-measured at `iters` samples, and only those reach `props-bench.txt`.

## Output artifacts

| File | Format | Use |
|---|---|---|
| `summary.md` | Markdown | Ranked tables, hover fallbacks, warm allocations, bitmap fallbacks, UI recipes, and profile summaries |
| `props-individual.jsonl` | JSON Lines | One record per property with controls, timing spread, effect evidence, and phase errors |
| `props-combos.jsonl` | JSON Lines | Combination declarations, phase timings, per-phase multipliers, and the `super_linear` flag |
| `props-bench.txt` | Go benchmark text | Refined rows only, with their real iteration counts, for `make bench-gate` (`scripts/cmd/benchscore`) |
| `manifest.json` | JSON | Measurement version and sampling flags that `--reference` checks |
| `profiles/*.pprof` | `pprof` protobuf | CPU profiles and before/after allocation snapshots; inspect with `go tool pprof -http=:0 <file>` (add `-base` for allocations) |

## Triage and optimization loop

The `bottleneck_layer` tag is a starting point for profiling. Verify the attribution with a call stack and a targeted timing. A bitmap fallback or a hover full-layout fallback sets the layer directly. Otherwise the harness subtracts the control page's style, layout, paint, and other times from the variant's and picks the largest positive stage delta, counting a valid raster delta as paint. A row with no stage delta above 20 µs is **`noise`**.

| `bottleneck_layer` | Dominant signal | Where to optimize |
|---|---|---|
| **`blinkless/css`** | Largest added `StyleTime` | Upstream `blinkless` CSS value parsing, cascade, or `var()` resolution |
| **`blinkless/layout`** | Largest added `LayoutTime` | Upstream `blinkless` flex/grid/text layout or display-list emission |
| **`ownframe/page+motion`** | `HoverFallbacks > 0`, or largest added `OtherTime` | `internal/page` (hover scope/patch, dirty diff, `KidIndex`) or `internal/motion` |
| **`ownframe/replay+bitmap`** | `BitmapFallback == true`, largest added `PaintTime`, or a valid raster delta | `internal/bitmap` for software paint, `internal/page/draw` for PNG, or `internal/replay` for native draw calls |
| **`noise`** | No stage adds more than 20 µs | Nothing to optimize; strengthen the fixture if the property should cost something |

Recommended loop:

1. Capture a baseline dump with `sh scripts/perf-props-dump.sh --out temp/perf-dumps/props-base`.
2. Open `summary.md`, pick one root cause shared by several top offenders, and inspect the attached `pprof` frames (or open `go tool pprof -http=:0 temp/perf-dumps/props-base/profiles/<name>.cpu.pprof`).
3. Land the fix in `ownframe` or `blinkless` with a focused regression test.
4. Capture the candidate with `--reference temp/perf-dumps/props-base --out temp/perf-dumps/props-after`. Compare the results with `make bench-gate`.
5. Keep `TestPropPerfHarnessSmoke` green in `make test` so the profiling harness itself never drifts.
