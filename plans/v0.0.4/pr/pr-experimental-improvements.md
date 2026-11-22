# PR: `v0.0.4` Experimental Performance, Rendering, Motion, and Property Harness v3 (`blinkless v0.1.5`)

## Summary

Bumps `ownframe` and `examples` to `github.com/chinmay-sawant/blinkless v0.1.5` and lands a broad wave of rendering, bitmap, motion, hover, text-shaping, and benchmark-harness improvements across 80+ files. Every single property across the 688-property CSS catalog now hovers with sub-0.35 ms delta, all 12 real-world 64-node UI recipes build in 0.59–1.65 ms and hover in 1.14–4.34 ms, and the two super-linear Tier-2 gradient combinations dropped by 15x–44x.

## Benchmark Comparison (Property Harness v3, 64-Node Uniform Grid)

| Benchmark | Before (`v0.1.4` / pre-cache) | After (`blinkless v0.1.5`) | Improvement |
|---|---:|---:|---:|
| **`bg_gradient + clip_shape` (`scale64`)** | 42.33 ms / 35.90 MB | **0.96 ms / 750.6 KB** | **44.1x faster, 47.8x less memory** |
| **`bg_gradient + clip_shape` (`hover`)** | 90.71 ms | **2.04 ms** | **44.5x faster** |
| **`bg_gradient + filter_fx` (`scale64`)** | 12.03 ms / 5.28 MB | **0.77 ms / 609.6 KB** | **15.7x faster, 8.7x less memory** |
| **`bg_gradient + filter_fx` (`hover`)** | 26.31 ms | **1.94 ms** | **13.6x faster** |
| **`width` (`scale64`)** | 1.73 ms / 3.23 MB (2,533 allocs) | **0.91 ms / 730.6 KB (1,428 allocs)** | **1.9x faster, -77% bytes, -44% allocs** |
| **`max-width` (`scale64`)** | 1.59 ms / 3.24 MB (2,610 allocs) | **0.93 ms / 739.3 KB (1,495 allocs)** | **1.7x faster, -77% bytes, -43% allocs** |
| **`column-width` (`scale64`)** | 1.26 ms / 1.87 MB (2,279 allocs) | **1.17 ms / 1.03 MB (1,895 allocs)** | **1.1x faster, -45% bytes, -17% allocs** |
| **`clip-path` (`resize` delta)** | 0.73 ms | **0.03 ms (≤noise)** | **24x faster** |
| **Layout-animation tick (64 / 256 boxes)** | 485 µs / 1.48 ms (686 / 1,867 allocs) | **328 µs / 1.16 ms (266 / 478 allocs)** | **-32% / -22% time, -61% / -74% allocs** |

### Real-World 64-Node UI Recipes (`blinkless v0.1.5`)

| Recipe (64 Nodes) | Scale-64 Build | Hover Cycle | Software Raster (PNG) | Display Ops |
|---|---:|---:|---:|---:|
| `glassmorphic-card` | **1.47 ms** | **3.57 ms** | 3.09 ms | 1,216 |
| `blend-mode-hero` | **0.98 ms** | **4.34 ms** | 1.29 ms | 1,024 |
| `gradient-badge-strip` | **1.65 ms** | **2.60 ms** | 0.72 ms | 832 |
| `mask-gradient-fade` | **1.19 ms** | **2.31 ms** | 0.69 ms | 448 |
| `fixed-modal-backdrop` | **1.00 ms** | **2.29 ms** | 1.45 ms | 768 |
| `transform-perspective-deck` | **0.83 ms** | **1.67 ms** | 1.65 ms | 512 |
| `sticky-table-hover` | **0.71 ms** | **1.40 ms** | 0.46 ms | 384 |
| `animated-hover-cards` | **0.72 ms** | **1.36 ms** | 0.54 ms | 192 |
| `dashboard-grid-shadow` | **0.64 ms** | **1.18 ms** | 1.01 ms | 448 |
| `outline-focus-form` | **0.73 ms** | **1.26 ms** | 0.69 ms | 576 |
| `scrollable-feed-clip` | **0.59 ms** | **1.16 ms** | 0.49 ms | 384 |
| `code-block-typography` | **0.74 ms** | **1.14 ms** | 0.50 ms | 448 |

- **Flags:** unchanged redraw allocations = **0 / 688**, bitmap fallbacks = **0 / 688**, failed rows = **0 / 688**, phase errors = **0 / 688**.

## Changes

### 1. `blinkless v0.1.5` Pin (`go.mod`, `examples/go.mod`)
- Bumped both root `go.mod` and `examples/go.mod` to `github.com/chinmay-sawant/blinkless v0.1.5` (PR #44):
  - Paint-only `background-color` restyle fast path, root margin splice escalation, and single-slab op payload cloning.
  - Grouped hover splice window stabilization (eliminating +2 ops/hover leak), layout-animation base style caching with DOM-edit invalidation, and explicit OpenType font kerning measurement.
  - Bounded process-wide image filter and clip-path mask PNG caches (plus direct-`Pix` `clip-path: inset(...)` fast path) and pointer-receiver `breakToken` with zero-copy ASCII word breaking.

### 2. Replay, Bitmap, and Text Rendering Optimizations
- **Nine-slice rounded fills and borders (`0814a184`):** Analytical corner masks in `internal/replay/shapes.go` and `internal/replay/raster_cache.go`.
- **Bitmap, masking, and PNG encoding (`506bad42`, `bb62c12f`):** Optimized masked groups, font caching, rotated text, paths, software bitmap paint, and PNG encoding.
- **Native quarter-turn text & font shaping (`b664137c`, `26389478`):** Native quarter-turn text replay (`internal/replay/quarter_turn_gpu_test.go`), OpenType font-feature and `text-autospace` replay (`internal/replay/text_shaping_test.go`, `internal/replay/text_shaping_gpu_test.go`), and skipping dead 2D function blends.

### 3. Page, Hover, and Motion Runtime Correctness & Speed
- **Page-owned box cache & motion frame reuse (`b664137c`, `bb62c12f`):** Faster unchanged hover checks and frame reuse during active motion.
- **Hover & transition preservation (`bf79b0c5`, `6f048c7e`, `08a7874a`):** Preserves live transitions across scoped patches, unhover after `PatchElements`, preserves sibling paint during retained redraws, and adds end-to-end grouped-hover regression coverage (`TestHoverGroupedCardKeepsOpCount`).
- **Transform base stability across consecutive rebinds (`89e4e536`):** Keeps the original paint base when a hover patch and a layout-animation tick rebind a paint player before the next sample (`internal/motion/bind.go`, `internal/motion/runtime_test.go`), preventing double-applied translation drift.

### 4. Property Benchmark Harness v3 (`d50e547d`)
- Upgraded `examples/blinkless-all` property profiling harness and `documentation/property-profiling.md` to measurement version 3: workload-only allocation profiles (`pprof -base`), control & anchor drift tracking, per-phase noise floors, stage-delta bottleneck attribution, active animation fixtures, and paired same-fixture combination multipliers.

## Verification

- `make test`: **PASS**
- `make lint`: **PASS**
- `make graphics-check`: **PASS**
- `GOMAXPROCS=1 sh scripts/perf-props-dump.sh`: **PASS** (688 properties + 186 combinations, 0 failed rows, 0 phase errors)
