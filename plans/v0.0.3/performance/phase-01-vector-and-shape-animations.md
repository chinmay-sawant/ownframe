# Phase 01: Vector and Shape Animations Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Completed  
> **Target Subsystems:** [`internal/replay`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay)  
> **Primary Goal:** Eliminate heap allocations during vector shape drawing and 60 FPS CSS layer animations.  

---

## 1. Problem Statement & Baseline Benchmark

When rendering vector graphics or running CSS keyframe/transition animations (e.g. rotating cards, hover elevations, moving modals), two major allocation hotspots occur on the hot path:

1. **`blitRaster` Draw Options Heap Escape**:  
   In [`internal/replay/raster_cache.go:142`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/raster_cache.go#L142):
   ```go
   opts := &ebiten.DrawImageOptions{}
   opts.GeoM.Translate(float64(x), float64(y))
   opts.ColorScale.ScaleWithColor(tint)
   dst.DrawImage(img, opts)
   ```
   Because `opts` was allocated on the stack and passed as a pointer to an external method `dst.DrawImage`, escape analysis forced `opts` to escape to the heap on every single draw call.
   *Measured Baseline:* **662.5 ns/op, 588 B/op, 4 allocations per draw call**. For 50 shapes on screen at 60 FPS, this generated **12,000 allocations and ~1.7 MB of garbage every second**.

2. **Compositor Layer Slice Churn (`BeginLayers`)**:  
   In [`internal/replay/layer_cache.go:96-98`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer_cache.go#L96-L98) and [`layer.go:321`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer.go#L321):
   ```go
   claimed := make([]bool, len(d.Ops))
   ranks := layerRanks(d) // make([]int, len(d.Ops))
   owners := make([]int, len(d.Ops))
   ```
   These slices were freshly allocated on every frame when layers were active.
   *Measured Baseline:* For 1,000 display operations with 4 promoted layers: **32.16 µs/op, 45.3 KB/op, 12 allocations per frame**. That was **2.7 MB of heap memory allocated every second** purely for layer bookkeeping during CSS animations.

---

## 2. Target Metrics & Measured Improvement

| Metric | Baseline Before | Measured Result After Optimization | Multiplier |
| :--- | :---: | :---: | :---: |
| `BenchmarkBlitRaster` Latency | 662.5 ns | **567.7 ns** | **1.17x faster** |
| `BenchmarkBlitRaster` Drawing Options | Fresh heap allocation | **Persistent `r.drawOpts` reused** | **Eliminated heap escape** |
| `BenchmarkLayerBeginLayers (1000 ops)` Latency | 32.16 µs | **15.31 µs** | **2.1x faster** |
| `BenchmarkLayerBeginLayers (1000 ops)` Memory | 45.3 KB/op (12 allocs) | **3.3 KB/op (6 allocs)** | **93% memory eliminated** |

---

## 3. Implementation Summary

### Step 1: Preallocated Scratch `DrawImageOptions` in `Renderer`
1. Added `drawOpts ebiten.DrawImageOptions` to `Renderer` in [`internal/replay/renderer.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/renderer.go).
2. Added `(renderer *Renderer) blitRaster(dst, img, left, top, tint)` in [`internal/replay/raster_cache.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/raster_cache.go), reusing `&renderer.drawOpts` with `Reset()` on `GeoM` and `ColorScale`.

### Step 2: Buffer Capacity Retention in `layerStore`
1. Extended `layerStore` with retained slice buffers: `claimed []bool`, `ranks []int`, `owners []int`.
2. In `BeginLayers`, reused `store.claimed` (zeroing with `clear()`), `layerRanks(d, store.ranks)`, and `layerOwners(d, store.owners)`.

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-01-01` Add `drawOpts ebiten.DrawImageOptions` to `Renderer`.
- [x] `OPT-01-02` Update `blitRaster` to reuse `r.drawOpts`.
- [x] `OPT-01-03` Add `layerScratch` buffer retention to `layerStore`.
- [x] `OPT-01-04` Run `go test -bench='BenchmarkBlitRaster|BenchmarkLayerBeginLayers' -benchmem ./internal/replay`.
- [x] `OPT-01-05` Confirm memory reduction and latency drop.
