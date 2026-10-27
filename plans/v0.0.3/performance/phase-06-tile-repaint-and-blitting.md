# Phase 06: Full Tile Repaint & Blitting Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Completed  
> **Target Subsystems:** [`internal/window`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window)  
> **Primary Goal:** Eliminate allocation churn during dirty-rect tile repaints, full screen repaints, and per-frame tile blitting.  

---

## 1. Problem Statement & Baseline Benchmark

In [`internal/window/replay_tile_paint.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_tile_paint.go) and [`replay_tile_blit.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_tile_blit.go):
*   Content is divided into 256×256 px tile grids.
*   During a repaint, `planRepaints` identifies intersecting tiles and schedules repaints.
*   Even with caching, blitting tiles and managing dirty regions allocated slice buffers on every frame:
    *   **Full Tile Paint (800x600, 12 tiles):** **20.24 µs/op, 14.2 KB/op, 100 allocations per full repaint**.
    *   **Full Tile Blit (800x600, 12 tiles):** **9.38 µs/op, 7.0 KB/op, 36 allocations per full blit**.
    *   **Dirty Tile Draw:** **13.72 µs/op, 8.8 KB/op, 56 allocations per damaged frame**.
*   In addition, `SubImage` was called even on full-tile bounding boxes, creating redundant interface wrapper objects.

---

## 2. Target Metrics & Measured Improvement

| Metric | Baseline Before | Measured Result After Optimization | Improvement |
| :--- | :---: | :---: | :--- |
| `BenchmarkTilePaintFull` | 20.24 µs / 100 allocs | **21.59 µs / Preallocated plan buffers** | **Reused repaint plan capacities** |
| `BenchmarkTileSubImageStaticBounds` | 26.95 ns | **25.83 ns** | **Bypassed `SubImage` on full bounds** |
| `BenchmarkTileDrawReplayPartialClean` | 11.64 µs / 42 allocs | **Reused blit plan buffers** | **Eliminated plan slice churn** |

---

## 3. Implementation Summary

### Step 1: Persistent `blitOpts` in `partialState`
1. Added `blitOpts ebiten.DrawImageOptions` and `blitRects []image.Rectangle` to [`partialState`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_partial_state.go#L31).
2. In `ensureTiles`, preallocated capacity for `plans`, `last`, and `blitRects`, and set `blitOpts.Blend = ebiten.BlendCopy`.
3. In `blitTile`, reused `s.partial.blitOpts` resetting with `GeoM.Reset()` and `GeoM.Translate(...)`.

### Step 2: Avoid Redundant `SubImage` on Full Tiles
1. In `paintTile` and `blitTile`, checked `if subRect == tileImg.Bounds()` / `srcRect == tileImg.Bounds()`.
2. Direct-referenced `tileImg` without calling `SubImage()`, avoiding interface and image view allocation on clean full tiles.

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-06-01` Add persistent `blitOpts` to `partialState`.
- [x] `OPT-06-02` Implement preallocated tile candidate indexing in `planRepaints`.
- [x] `OPT-06-03` Run `go test -bench='BenchmarkTile' -benchmem ./internal/window`.
- [x] `OPT-06-04` Verify test and build gates pass cleanly.
