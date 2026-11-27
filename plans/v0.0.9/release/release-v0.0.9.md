# OwnFrame v0.0.9

Release on top of `v0.0.8`, delivering retained DOM mutation APIs with Blinkless `v0.1.11`, 3.31x faster 1,000-row element patching, tile-surface retained motion layers (taking `examples/trade` on native Windows from 7.75 FPS to 144 FPS locked), in-place `border-radius` / color-alpha / individual-transform CSS animations, and a 688-property Ebiten game-loop profiler.

- Range: `v0.0.8` through `feat/retained-dom-mutation` (2026-11-28)
- Engine: pins `github.com/chinmay-sawant/blinkless` to `v0.1.11`

## Summary

| Area | What changed |
| :--- | :--- |
| **Retained DOM Mutations (`internal/render`, `internal/page/patch`, `internal/page`, `internal/page/a11y`)** | Extended `ElementEdit` / `Page.PatchElement` / `Page.PatchElements` with `RemoveAttribute`, `AddClass`, `RemoveClass`, `ToggleClass`, `Remove`, `InnerHTML`, and `AppendHTML`, backed by Blinkless `v0.1.11`. Added in-place `a11y.State` synchronization, form dirty tracking on structural commits, atomic rollback on apply errors, and `dirty.DiffDirect` / `DiffDirectWithBounds` to skip FNV region hashing on spliced equal-length patches. |
| **Tile-Surface Retained Layers & Replay Caches (`internal/window`, `internal/replay`)** | Composited top-order-safe retained motion layers over static tiles in `drawTileSurface` (matching `drawViewport`), wired anonymous `BoxIndex` promotion (`ActiveLayerBoxIndices` -> `SetPromotedLayerBoxes`), supported uniform `OverflowClip` retained layers, replaced the whole-page `hasGroups(d)` kill switch with per-layer blend-group checks, skipped offscreen dirty rects in `planRepaints`, and raised replay LRU entry caps (`imageCacheLimit` 32 -> 256, `textRunMaxEntries` 256 -> 1024, `rasterMaxEntries` 128 -> 512). |
| **CSS Motion Property Fast Paths (`internal/motion`)** | Moved uniform pixel `border-radius` keyframes to in-place `DisplayOp` mutation (`animFieldRadius` / `MaskRadius`), added `animFieldAlpha` / `MaskAlpha` so color animations and transitions interpolate `op.Alpha` in place, added individual `translate`, `rotate`, and `scale` properties on the 2D compositor path, kept planar `translate3d(x, y, 0)` / `scale3d(sx, sy, 1)` on 2D affine transforms, scoped anonymous color animations to owned ops, and suppressed background damage on promoted transitions. |
| **Benchmarks, Game-Loop Tests & Examples (`BENCHMARKS.md`, `examples/trade`, `examples/blinkless-all`, `examples/perf-benchmarks`)** | Added `BenchmarkRetainedMutations`, `BenchmarkMotionPropertyGaps`, `BenchmarkBlendSiblingLayerPromotion`, multi-resolution `TestDenseSingleScreenTileLayerCompositing` (1280×720, 1920×1080, 2560×1600), and `TestPropPerfGameLoopDump` (688-property Ebiten game-loop dump). Scoped per-screen CSS in `examples/trade/build.go` and defaulted `-dev`/`-perf` to `false` in `examples/trade/main.go`. |

## Measured Impact (vs. `v0.0.8`)

| Benchmark / Workload | `v0.0.8` | `v0.0.9` | Improvement |
| :--- | ---: | ---: | ---: |
| `ListPatchVsRebuild/PatchElement1Row` (1,000 rows) | `584 µs` / `208 KB` / `102 allocs` | **`176.5 µs` / `77.9 KB` / `75 allocs`** | **`3.31x faster`**, **`-62.5%` bytes** (**`86.6x`** vs. full rebuild) |
| `ListPatchVsRebuild/PatchElements10Rows` (1,000 rows) | `815 µs` / `275 KB` / `624 allocs` | **`452.6 µs` / `249.7 KB` / `540 allocs`** | **`1.80x faster`**, **`-9.2%` bytes** |
| `MotionPropertyGaps/BorderRadiusAnim` | `46.9 µs` / `89.2 KB` / `178 allocs` | **`239.0 ns` / `0 B` / `0 allocs`** | **`196x faster`**, **`0` relayouts** |
| `BlendSiblingLayerPromotion` | `0 promoted/op` | **`1.000 promoted/op`** | **Unrelated blend groups no longer demote layers** |
| `trade.exe` live native Windows (1600×864, 8 s) | `7.75 FPS` / `55.7 ms` avg / `120` long / `262 MB/s` | **`143.96 FPS` / `3.12 ms` avg / `0` long / `8.89 MB/s`** | **`18.6x FPS`**, **`29.5x` lower alloc rate** |

## Breaking changes

None. All new `ElementEdit` fields are zero-valued by default; existing callers keep prior behavior.

## Validation

- `make lint` (all 10 checks)
- `make graphics-check` (all 20 GPU tests)
- `go test -p 1 ./internal/page/... ./internal/replay/... ./internal/motion/... ./internal/window/... .`
