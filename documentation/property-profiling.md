# CSS Property Profiling & Dump Guide

This guide explains how to benchmark all 688 CSS properties in `examples/blinkless-all` individually and in combination, generate structured performance dumps, compare runs with `cmd/benchscore`, and attribute slow paths to `blinkless` or `ownframe` functions.

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

Run the same harness and fixture sources on both revisions. Use the same host, Go version, CPU settings, and sampling flags. When the renderer baseline predates measurement version 2, use the current benchmark and fixture sources with the baseline renderer.

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

`--reference` reuses the baseline's property and combination refinement choices, top-offender pairs, and seeded stacks. The harness checks for matching property samples and combination declarations. Keep the sampling flags, filter, and seed identical.

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
go run ./cmd/benchscore -ref temp/perf-dumps/props-wsl2/props-bench.txt -cand temp/perf-dumps/props-win/props-bench.txt
```

## Flags and environment variables

| Flag (`scripts/perf-props-dump.sh`) | Environment variable | Default | Purpose |
|---|---|---:|---|
| `--out DIR` | `OWNFRAME_PROP_DUMP_DIR` | `temp/perf-dumps/props-<ts>` | Output directory. The test skips when the environment variable is unset. |
| `--reference DIR` | `OWNFRAME_PROP_REFERENCE_DIR` | *(unset)* | Version 2 dump that supplies the combination selection and refinement choices |
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
2. **Timing and allocations.** Each timed loop disables garbage collection. Refined pairs share an explicit garbage collection boundary before timing. The harness samples allocations separately. `summary.md` reports the maximum drift of a fixed control measured every 100 properties. Use the full sweep to find candidates and repeat the relevant cases before attributing small timing changes.
3. **Effect evidence.** The harness audits every property in `props_gen.go`, including properties the visual catalog marks as declined. It overrides some passive samples, such as `mask-image: none`. A row sets `effective` when its paired fixture shows a display signature, motion, or PNG difference. `effect_evidence` records the observations. A row with no observed effect needs a stronger fixture or another way to verify the declaration.
4. **Measured operations.** Cold build times `NewWithOptions` and `Redraw`. Warm redraw runs after a settle tick and can include active motion. Hover times enter, tick, leave, and tick on a 64-node retained page. Resize cycles from 360 × 160 to 480 × 200 and back. The 64-node build repeats the property across uniform boxes. The scaling ratio compares the two measured deltas. A growth-rate claim needs more node counts.
5. **Tick and PNG.** The harness times `Tick` separately. It also times one fresh `PNGWithContext` per page, including any on-demand software paint and PNG encoding. It decodes every timed PNG after the timer stops. Invalid PNG samples record a phase error.
6. **Profiles.** The harness captures CPU and allocation profiles for selected slow scale, hover, and PNG cases. The PNG profile includes page setup and redraw. Use the phase time and call stack together when distinguishing paint from encoding. Use a separate replay benchmark to measure Ebiten rendering.

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

The dump flags `max_multiplier >= 1.5` as **`super_linear: true`**. Use the flag to select cases for repeated measurement. The multiplier is sensitive to small deltas in its denominator.

## Output artifacts

| File | Format | Use |
|---|---|---|
| `summary.md` | Markdown | Ranked tables, hover fallbacks, warm allocations, bitmap fallbacks, UI recipes, and profile summaries |
| `props-individual.jsonl` | JSON Lines | One record per property with controls, timing spread, effect evidence, and phase errors |
| `props-combos.jsonl` | JSON Lines | Combination declarations, phase timings, per-phase multipliers, and the `super_linear` flag |
| `props-bench.txt` | Go benchmark text | Direct input to `make bench-gate` (`cmd/benchscore`) for before/after regression gating |
| `profiles/*.pprof` | `pprof` protobuf | Interactive inspection via `go tool pprof -http=:0 <file>` |

## Triage and optimization loop

The `bottleneck_layer` tag is a starting point for profiling. Verify the attribution with a call stack and a targeted timing.

| `bottleneck_layer` | Dominant signal | Where to optimize |
|---|---|---|
| **`blinkless/css`** | `StyleTime` dominates cold/resize | Upstream `blinkless` CSS value parsing, cascade, or `var()` resolution |
| **`blinkless/layout`** | `LayoutTime` dominates cold/resize/scale64 | Upstream `blinkless` flex/grid/text layout or display-list emission |
| **`ownframe/page+motion`** | `HoverFallbacks > 0`, `WarmAllocs > 0`, or `OtherTime` dominates | `internal/page` (hover scope/patch, dirty diff, `KidIndex`) or `internal/motion` |
| **`ownframe/replay+bitmap`** | `BitmapFallback == true` or `RasterDeltaNS` dominates | `internal/bitmap` for software paint, `internal/page/draw` for PNG, or `internal/replay` for native draw calls |

Recommended loop:

1. Capture a baseline dump with `sh scripts/perf-props-dump.sh --out temp/perf-dumps/props-base`.
2. Open `summary.md`, pick one root cause shared by several top offenders, and inspect the attached `pprof` frames (or open `go tool pprof -http=:0 temp/perf-dumps/props-base/profiles/<name>.cpu.pprof`).
3. Land the fix in `ownframe` or `blinkless` with a focused regression test.
4. Capture the candidate with `--reference temp/perf-dumps/props-base --out temp/perf-dumps/props-after`. Compare the results with `make bench-gate`.
5. Keep `TestPropPerfHarnessSmoke` green in `make test` so the profiling harness itself never drifts.
