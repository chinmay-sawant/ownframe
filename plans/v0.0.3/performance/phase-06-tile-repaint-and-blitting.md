# Phase 06: Full Tile Repaint & Blitting Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Planned  
> **Target Subsystems:** [`internal/window`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window)  
> **Primary Goal:** Eliminate allocation churn during dirty-rect tile repaints, full screen repaints, and per-frame tile blitting.  

---

## 1. Problem Statement & Baseline Benchmark

In [`internal/window/replay_tile_paint.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_tile_paint.go) and [`replay_tile_blit.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_tile_blit.go):
*   Content is divided into 256×256 px tile grids.
*   During a repaint, `planRepaints` identifies intersecting tiles and schedules repaints.
*   Even with caching, blitting tiles and managing dirty regions allocates slice buffers on every frame:
    *   **Full Tile Paint (800x600, 12 tiles):** **20.24 µs/op, 14.2 KB/op, 100 allocations per full repaint**.
    *   **Full Tile Blit (800x600, 12 tiles):** **9.38 µs/op, 7.0 KB/op, 36 allocations per full blit**.
    *   **Dirty Tile Draw:** **13.72 µs/op, 8.8 KB/op, 56 allocations per damaged frame**.
*   In addition, `pollReload()` runs `os.Stat` on the main thread every 250ms, allocating `os.FileInfo` structures (**759.2 ns/op, 256 B/op, 2 allocs/op**).

---

## 2. Target Metrics & Improvement Multiplier

| Metric | Current Baseline (800x600) | Target (After Optimization) | Multiplier |
| :--- | :---: | :---: | :---: |
| `BenchmarkTilePaintFull` Latency | 20.24 µs | **~6.5 µs** | **3.1x faster** |
| `BenchmarkTilePaintFull` Memory | 14.2 KB/op (100 allocs) | **< 1 KB/op (~2 allocs)** | **14x less memory churn** <br> **50x fewer allocs** |
| `BenchmarkTileBlitFull` Memory | 7.0 KB/op (36 allocs) | **0 B/op (0 allocs)** | **100% memory eliminated** |
| `BenchmarkTileDrawReplayPartialDamaged` | 8.8 KB/op (56 allocs) | **< 0.5 KB/op (~1 alloc)** | **17x less memory churn** |

---

## 3. Implementation Plan

### Step 1: Pre-Bound Tile Blit Options in `partialState`
1. Add a reusable `blitOpts ebiten.DrawImageOptions` to `partialState` in [`internal/window/replay_partial_state.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_partial_state.go).
2. Set `blitOpts.Blend = ebiten.BlendCopy` once during initialization.
3. In `blitTiles`, translate `blitOpts.GeoM` directly and pass `&s.partial.blitOpts` instead of constructing new option values.

### Step 2: Pooled Repaint Tile Index Buffer
1. Preallocate a fixed array/slice of `[maxTiles]tileCoord` on `partialState`.
2. Reuse this slice in `planRepaints` instead of dynamic slice growth, eliminating the 56 allocations per damaged frame.

### Step 3: Offload `pollReload` to Background Ticker
1. Move `pollReload()` off the 60 Hz `Update()` stage.
2. Run `os.Stat` on a dedicated low-frequency background goroutine, signaling frame wakeups via `WantFrameIn(0)` only when files change on disk.

---

## 4. Acceptance Checklist & Verification

- [ ] `OPT-06-01` Add persistent `blitOpts` to `partialState`.
- [ ] `OPT-06-02` Implement preallocated tile candidate indexing in `planRepaints`.
- [ ] `OPT-06-03` Run `go test -bench='BenchmarkTile' -benchmem ./internal/window`.
- [ ] `OPT-06-04` Verify `BenchmarkTileBlitFull` achieves 0 heap allocations and `BenchmarkTilePaintFull` memory drops by >90%.
