# v0.0.3 Performance Engineering & Zero-Allocation Optimization Plan

> **Directory:** `plans/v0.0.3/performance/`  
> **Status:** Active Execution Plan  
> **Baseline Commit:** `253705c1` (v0.0.3 cycle)  
> **Target Hardware:** 13th Gen Intel Core i7-13700HX (24 threads), Linux 6.6, x86_64  

---

## 1. Executive Summary & Improvement Multipliers

Following the initial stabilization and correctness pass, this performance initiative targets the root causes of heap allocations, garbage collection pauses, and frame drops across the graphics pipeline, game loop, diffing engine, and real-world editor applications.

### 🚀 Overall Performance Improvement Multiplier

*   **Memory Allocations (GC Pressure): ~98% reduction (50x to 100x less garbage)**  
    *Currently, hot paths generate 2–5 MB of heap garbage every second during active typing or CSS animation. Target optimizations drop rendering hot paths to near **0 B/sec**, completely eliminating garbage collection micro-stutters.*
*   **Idle Frame Overhead: ~374x faster** (from 374 µs down to < 1 µs).
*   **Keystroke & Typing Response: ~55x faster** (from 671 µs down to ~12 µs).
*   **Animation Layer Math: ~10x faster** (from 32 µs down to ~3.2 µs).
*   **DOM Diffing Overhead: ~10x faster, >300x less memory** (from 954 µs / 313 KB down to 95 µs / <1 KB).

---

## 2. Master Performance & Savings Multiplier Table

The baseline numbers below were measured directly on the live codebase using `-benchmem -benchtime=200ms`:

| Subsystem & Focus Area | Current Baseline (Live Data) | Target (This will get you up to this) | Improvement Multiplier (Speed & Memory) | Owning Phase File |
| :--- | :--- | :--- | :---: | :--- |
| **Vector & Shape Drawing**<br>([`raster_cache.go:142`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/raster_cache.go#L142)) | 662.5 ns <br> • 588 B/op <br> • 4 allocs/op | **~150 ns** <br> • **0 B/op** <br> • **0 allocs/op** | **4.4x faster** <br> **100% memory eliminated (Zero allocs)** | [`phase-01-vector-and-shape-animations.md`](phase-01-vector-and-shape-animations.md) |
| **Compositor Animation Layers**<br>([`layer_cache.go:96`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer_cache.go#L96)) | 32.16 µs <br> • 45.3 KB/op <br> • 12 allocs/op (2.7 MB/s) | **~3.2 µs** <br> • **0 B/op** <br> • **0 allocs/op** | **10x faster** <br> **100% memory eliminated (Zero allocs)** | [`phase-01-vector-and-shape-animations.md`](phase-01-vector-and-shape-animations.md) |
| **Idle Page Tick (No changes)**<br>([`page_tick.go:63`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_tick.go#L63)) | 373.9 µs wasted <br> • 812 B/op <br> • 2,000 op diffs | **< 1 µs** <br> • **0 B/op** <br> • **0 op diffs** | **374x faster** <br> **Instant bypass of unmodified ticks** | [`phase-02-idle-page-tick.md`](phase-02-idle-page-tick.md) |
| **Typing & DOM Content Changes**<br>([`diff.go:41`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty/diff.go#L41)) | 954.5 µs (~1 ms) <br> • 313.1 KB/op <br> • 2,011 allocs/op | **~95 µs** <br> • **< 1 KB/op** <br> • **0 allocs/op** | **10x faster** <br> **>300x less memory churn** | [`phase-03-typing-and-letter-spacing.md`](phase-03-typing-and-letter-spacing.md) |
| **Text Letter-Spacing Calculation**<br>([`text.go:42`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/text.go#L42)) | 79.31 µs <br> • 28.5 KB/op <br> • 393 allocs/op | **~15 µs** <br> • **~2 KB/op** <br> • **~5 allocs/op** | **5.3x faster** <br> **78x fewer allocations** | [`phase-03-typing-and-letter-spacing.md`](phase-03-typing-and-letter-spacing.md) |
| **Code Editor Syntax Highlighting**<br>([`syntax.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/syntax.go)) | 670.9 µs <br> • 365.3 KB/op <br> • 9,395 allocs/keystroke | **~12 µs** <br> • **~2 KB/op** <br> • **~30 allocs/op** | **55x faster** <br> **180x less memory** <br> **313x fewer allocs** | [`phase-04-code-editor-syntax-highlighting.md`](phase-04-code-editor-syntax-highlighting.md) |
| **Window Resizing with Soft-Wrap**<br>([`wrap.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go)) | 656.5 µs <br> • 1.82 MB/op <br> • 6,350 allocs/resize | **~130 µs** <br> • **~50 KB/op** <br> • **~10 allocs/op** | **5x faster** <br> **36x less memory** <br> **635x fewer allocs** | [`phase-05-window-resizing-with-soft-wrap.md`](phase-05-window-resizing-with-soft-wrap.md) |
| **Full Tile Repaint (800x600 screen)**<br>([`replay_tile_paint.go:48`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_tile_paint.go#L48)) | 20.24 µs <br> • 14.2 KB/op <br> • 100 allocs/op | **~6.5 µs** <br> • **< 1 KB/op** <br> • **~2 allocs/op** | **3.1x faster** <br> **14x less memory** <br> **50x fewer allocs** | [`phase-06-tile-repaint-and-blitting.md`](phase-06-tile-repaint-and-blitting.md) |

---

## 3. Phase Index & Navigation

The optimization plan is divided into 6 distinct, self-contained implementation phases:

1. **[Phase 01: Vector and Shape Animations](phase-01-vector-and-shape-animations.md)**
   - Eliminating `DrawImageOptions` heap escape in `blitRaster()`.
   - Preallocated buffer pooling for `BeginLayers` (`claimed`, `ranks`, `owners`).
2. **[Phase 02: Idle Page Tick](phase-02-idle-page-tick.md)**
   - Atomic mutation tracking (`tickMutated`) to bypass display copying and comparison on unmodified ticks.
3. **[Phase 03: Typing and Content Letter Spacing Calculation](phase-03-typing-and-letter-spacing.md)**
   - Flat open-addressing hash table arena in `diffByPosition` (eliminating 2,011 allocs/diff).
   - Shaped text run caching for letter-spaced glyph rendering.
4. **[Phase 04: Code Editor Syntax Highlighting](phase-04-code-editor-syntax-highlighting.md)**
   - Incremental line-level tokenization cache for Go, Python, and Markdown files in `code-view`.
5. **[Phase 05: Window Resizing with Soft-Wrap](phase-05-window-resizing-with-soft-wrap.md)**
   - Reusable wrap row buffer arena eliminating 1.82 MB slice churn on window resize.
6. **[Phase 06: Full Tile Repaint & Blitting](phase-06-tile-repaint-and-blitting.md)**
   - Static scratch draw options and pooled dirty-tile index slices in `internal/window`.

---

## 4. Master Phase-Wise Checklist

### Phase 01: Vector and Shape Animations
- [ ] `OPT-01-01` Add persistent `drawOpts ebiten.DrawImageOptions` to `Renderer` struct in [`internal/replay/renderer.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/renderer.go).
- [ ] `OPT-01-02` Update `blitRaster` to reset and reuse `r.drawOpts` instead of allocating on the heap.
- [ ] `OPT-01-03` Pre-allocate slice buffers (`claimed []bool`, `ranks []int`, `owners []int`) on `layerStore` in [`internal/replay/layer_cache.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/layer_cache.go) to remove 45.3 KB per-frame allocations during CSS animations.
- [ ] `OPT-01-04` Verify zero allocations via `BenchmarkBlitRaster` and `BenchmarkLayerBeginLayers`.

### Phase 02: Idle Page Tick
- [ ] `OPT-02-01` Add `tickMutated atomic.Bool` and `tickOpVersion uint64` to `Page` in [`internal/page/page_types.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_types.go).
- [ ] `OPT-02-02` Instrument `PatchOp` and op mutating methods to flag `tickMutated = true`.
- [ ] `OPT-02-03` In [`internal/page/page_tick.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_tick.go), skip `captureTickOps()` and `noteTickDamage()` entirely when `tickMutated` is false.
- [ ] `OPT-02-04` Verify `BenchmarkTick_CaptureAndDamage_2000Ops_Mod0` drops from 374 µs to < 1 µs.

### Phase 03: Typing & Letter-Spacing Calculation
- [ ] `OPT-03-01` Design `type diffArena struct` with reusable flat integer slices in [`internal/page/dirty/diff.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty/diff.go).
- [ ] `OPT-03-02` Replace `map[opKey][]int` in `diffByPosition` with pooled `diffArena`, dropping allocations from 2,011 to 0.
- [ ] `OPT-03-03` In [`internal/replay/text.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/text.go), cache shaped letter-spaced glyph runs to avoid per-rune string conversion and individual `text.Measure()` calls.
- [ ] `OPT-03-04` Verify `BenchmarkTick_DiffByPosition_2000Ops` and `BenchmarkText_LetterSpacing_Cold` improvements.

### Phase 04: Code Editor Syntax Highlighting
- [ ] `OPT-04-01` Introduce `type lineTokenCache struct` in [`examples/code-view/syntax.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/syntax.go) storing tokenized spans per line index.
- [ ] `OPT-04-02` On edits, re-tokenize only edited and adjacent multiline comment/string lines rather than full 1,000-line files.
- [ ] `OPT-04-03` Verify `BenchmarkEditor_Syntax_Go_1000` drops from 671 µs / 9,395 allocs to ~12 µs / ~30 allocs.

### Phase 05: Window Resizing with Soft-Wrap
- [ ] `OPT-05-01` Introduce a reusable row plan buffer in [`examples/code-view/wrap.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go).
- [ ] `OPT-05-02` Reuse visual row index capacities across resize recalculations instead of constructing 6,350 slice objects.
- [ ] `OPT-05-03` Verify `BenchmarkEditor_SoftWrap_Resize_1000` drops from 1.82 MB to < 50 KB.

### Phase 06: Full Tile Repaint & Blitting
- [ ] `OPT-06-01` Add persistent `blitOpts ebiten.DrawImageOptions` to `shell.partial` in [`internal/window/replay_tile.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/window/replay_tile.go).
- [ ] `OPT-06-02` Reuse tile repaint candidate slices in `planRepaints` to avoid allocating 56 slices on damaged frames.
- [ ] `OPT-06-03` Verify `BenchmarkTilePaintFull` and `BenchmarkTileDrawReplayPartialDamaged`.
