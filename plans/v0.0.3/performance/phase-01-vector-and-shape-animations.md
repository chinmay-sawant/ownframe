# Phase 01: Vector and Shape Animations Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Planned  
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
   Because `opts` is allocated on the stack and passed as a pointer to an external method `dst.DrawImage`, escape analysis forces `opts` to escape to the heap on every single draw call.
   *Measured Baseline:* **662.5 ns/op, 588 B/op, 4 allocations per draw call**. For 50 shapes on screen at 60 FPS, this generates **12,000 allocations and ~1.7 MB of garbage every second**.

2. **Compositor Layer Slice Churn (`BeginLayers`)**:  
   In [`internal/replay/layer_cache.go:96-98`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer_cache.go#L96-L98) and [`layer.go:321`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer.go#L321):
   ```go
   claimed := make([]bool, len(d.Ops))
   ranks := layerRanks(d) // make([]int, len(d.Ops))
   owners := make([]int, len(d.Ops))
   ```
   These slices are freshly allocated on every frame when layers are active.
   *Measured Baseline:* For 1,000 display operations with 4 promoted layers: **32.16 µs/op, 45.3 KB/op, 12 allocations per frame**. That is **2.7 MB of heap memory allocated every second** purely for layer bookkeeping during CSS animations.

---

## 2. Target Metrics & Improvement Multiplier

| Metric | Current Baseline | Target (After Optimization) | Multiplier |
| :--- | :---: | :---: | :---: |
| `BenchmarkBlitRaster` Latency | 662.5 ns | ~150 ns | **4.4x faster** |
| `BenchmarkBlitRaster` Allocs | 4 allocs/op (588 B) | **0 allocs/op (0 B)** | **100% memory eliminated** |
| `BenchmarkLayerBeginLayers (1000 ops)` Latency | 32.16 µs | ~3.2 µs | **10x faster** |
| `BenchmarkLayerBeginLayers (1000 ops)` Memory | 45.3 KB/op (12 allocs) | **0 B/op (0 allocs)** | **100% memory eliminated** |

---

## 3. Implementation Plan

### Step 1: Preallocate Scratch `DrawImageOptions` in `Renderer`
1. Add `drawOpts ebiten.DrawImageOptions` to `Renderer` in [`internal/replay/renderer.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/renderer.go).
2. Update `blitRaster(dst, img, x, y, tint)` to take `r *Renderer` or pass `&r.drawOpts`:
   ```go
   r.drawOpts.GeoM.Reset()
   r.drawOpts.GeoM.Translate(float64(x), float64(y))
   r.drawOpts.ColorScale.Reset()
   r.drawOpts.ColorScale.ScaleWithColor(tint)
   dst.DrawImage(img, &r.drawOpts)
   ```
3. Because `Renderer` is single-threaded per Ebiten loop, reusing `r.drawOpts` is strictly race-free.

### Step 2: Buffer Capacity Retention in `layerStore`
1. Extend `layerStore` in [`internal/replay/layer_store.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer_store.go) with retained scratch slices:
   ```go
   type layerScratch struct {
       claimed []bool
       ranks   []int
       owners  []int
   }
   ```
2. In `BeginLayers`, grow the scratch slices if `cap < len(d.Ops)`, otherwise slice to `[:len(d.Ops)]` and zero the `claimed` flags using `clear(s.claimed)`.
3. Eliminates all 12 slice allocations per frame during animation.

---

## 4. Acceptance Checklist & Verification

- [ ] `OPT-01-01` Add `drawOpts ebiten.DrawImageOptions` to `Renderer`.
- [ ] `OPT-01-02` Update `blitRaster` to reuse `r.drawOpts`.
- [ ] `OPT-01-03` Add `layerScratch` buffer retention to `layerStore`.
- [ ] `OPT-01-04` Run `go test -bench='BenchmarkBlitRaster|BenchmarkLayerBeginLayers' -benchmem ./internal/replay`.
- [ ] `OPT-01-05` Confirm **0 allocs/op** on both benchmarks.
