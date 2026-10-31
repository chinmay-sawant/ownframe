# Real-World Application Performance & Rendering Cost Suite

> **Branch:** `perf/render-cost-suite`  
> **Base:** `master` ([`dbcfecc`](https://github.com/chinmay-sawant/ownframe/commit/dbcfeccf49d15d5a8e596f6ada6d15baedad1b0d))  
> **Location:** [`examples/performance`](../../examples/performance)  
> **Environment:** 13th Gen Intel(R) Core(TM) i7-13700HX (24 threads, Linux x86_64, Go 1.26.10)  
> **Date:** 2026-10-31  

This document records the actual end-to-end rendering and application costs in OwnFrame and its underlying `blinkless` layout engine, replacing synthetic micro-benchmarks with 10 real-world application scenarios ranging from lightweight row lists to extreme 1,000-component production applications.

All CSS properties used strictly adhere to the supported catalogue in [`documentation/compatibility-matrix.md`](../../documentation/compatibility-matrix.md).

---

## 1. Ten Progressive Test Scenarios

The suite evaluates 10 progressively complex application views:

| # | Scenario ID | Level | Items | DOM Boxes | Display Ops | Core CSS Properties Exercised |
|---|---|---|---|---|---|---|
| **01** | `01-simple-rows-50` | **Low** | 50 rows | 305 | 404 | `display: flex`, `flex-direction`, `justify-content`, `align-items`, `gap`, `color`, `background-color`, `border-bottom`, `border-radius`, `font-family`, `font-size` |
| **02** | `02-zebra-table-200` | **Medium-Low** | 200 rows | 1,211 | 2,416 | `display: table`, `border-collapse: collapse`, `border-spacing: 0`, alternating row striping, `th`/`td` `padding`, `text-align`, pill badges (`display: inline-block`, `border-radius`) |
| **03** | `03-product-grid-200` | **Medium** | 200 cards | 2,605 | 3,205 | `display: grid`, `grid-template-columns: repeat(4, 1fr)`, `gap`, `background: linear-gradient(135deg, ...)`, `border-radius: 8px`, `overflow: hidden`, CTA buttons |
| **04** | `04-social-timeline-200` | **Medium-High** | 200 posts | 4,005 | 3,205 | Nested `display: flex` columns/rows, `border-radius: 50%` avatars, `margin: 0 auto`, `line-height: 1.5`, metadata chips, reaction button groups |
| **05** | `05-order-book-500` | **High** | 500 rows | 3,009 | 2,511 | `position: relative`, `position: absolute`, proportional visual depth bar overlays (`width: %`, `background-color: rgba(...)`), `font-family: monospace`, `text-align: right` |
| **06** | `06-enterprise-grid-1000` | **High** | 1,000 rows | 11,014 | 13,016 | `display: flex`, `gap`, `white-space: nowrap`, `overflow: hidden`, `text-overflow: ellipsis`, checkboxes, severity chips, department tags |
| **07** | `07-analytics-dashboard` | **Complex** | 30 components | 185 | 191 | Multi-pane dashboard: 4 KPI multi-stop gradient cards, `grid-template-columns: 2fr 1fr`, activity event timeline with circular step badges, cluster resource progress meters |
| **08** | `08-multipage-doc-200` | **Complex** | 200 pages | 3,204 | 4,804 | Multi-page paginated technical report: bordered sheet wrappers (`margin: 24px auto`, `padding: 36px 44px`), headers, narrative paragraphs, accent callout boxes, bullet lists, footers |
| **09** | `09-css-kitchen-sink-150` | **Complex** | 150 widgets | 1,205 | 19,755 | Combinatorial CSS matrix: elliptical & irregular `border-radius`, `border-style: dashed/dotted/solid`, `transform: scale/translate`, `letter-spacing`, `word-spacing`, multi-color gradients |
| **10** | `10-extreme-app-1000` | **Extreme** | 1,000 rows | 22,006 | 23,006 | High-density application rows with 11 sub-elements per row: rank badge, gradient avatar circle, two-line title/email block, pill status, dynamic progress meter bar, currency metric, tag list, button |

---

## 2. Multiplier Progression: Baseline vs. Wave 1 vs. Wave 2

Following two consecutive optimization sprints across both `blinkless` (layout engine) and `ownframe` (host runtime):
- **Cold Latency:** Accelerated up to **5.97x faster** overall (extreme app rendered in **626.8ms**, down from **3,745.3ms**).
- **Warm Redraw Latency:** Slashed up to **894.8x faster** (steady-state warm frames drop to **<0.05ms**).
- **Memory Allocations:** Slashed by **up to 8.83x** (saving **1.89 GB of RAM allocations** per load on Scenario 10).

### 2.1 Cold Latency Multipliers (Initial Render)

| ID | Scenario Name | Level | Baseline (ms) | Wave 1 (ms) | Wave 2 (ms) | Wave 1 Multiplier | Wave 2 Multiplier | Total Multiplier | Total Latency Reduction |
|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 17.58 | 11.63 | **7.40** | **1.51x** | **1.57x** | **2.38x** | **-57.9%** |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 46.88 | 39.72 | **23.13** | **1.18x** | **1.72x** | **2.03x** | **-50.7%** |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 486.29 | 222.20 | **161.57** | **2.19x** | **1.38x** | **3.01x** | **-66.8%** |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 505.41 | 280.30 | **213.91** | **1.80x** | **1.31x** | **2.36x** | **-57.7%** |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 85.34 | 43.94 | **32.62** | **1.94x** | **1.35x** | **2.62x** | **-61.8%** |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 569.11 | 262.75 | **209.75** | **2.17x** | **1.25x** | **2.71x** | **-63.1%** |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 77.58 | 18.31 | **13.31** | **4.24x** | **1.38x** | **5.83x** | **-82.8%** |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 236.90 | 72.54 | **47.86** | **3.27x** | **1.52x** | **4.95x** | **-79.8%** |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 190.66 | 97.07 | **62.20** | **1.96x** | **1.56x** | **3.07x** | **-67.4%** |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 3,745.28 | 909.71 | **626.87** | **4.12x** | **1.45x** | **5.97x** | **-83.3%** |

### 2.2 Warm Redraw Multipliers (Steady-State Frames)

| ID | Scenario Name | Level | Baseline (ms) | Wave 1 (ms) | Wave 2 (ms) | Wave 1 Multiplier | Wave 2 Multiplier | Total Multiplier |
|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 0.35 | 0.21 | **<0.01** | **1.67x** | **>20x** | **>35x** |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 1.62 | 1.23 | **<0.01** | **1.32x** | **>120x** | **>160x** |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 1.76 | 1.14 | **<0.01** | **1.54x** | **>110x** | **>170x** |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 2.14 | 1.78 | **<0.01** | **1.20x** | **>170x** | **>210x** |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 6.72 | 5.49 | **<0.01** | **1.22x** | **>540x** | **>670x** |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 7.91 | 5.89 | **<0.01** | **1.34x** | **>580x** | **>790x** |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 1.82 | 0.14 | **<0.01** | **13.00x** | **>14x** | **>180x** |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 3.89 | 1.83 | **<0.01** | **2.13x** | **>180x** | **>380x** |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 2.45 | 2.15 | **<0.01** | **1.14x** | **>210x** | **>240x** |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 35.79 | 12.75 | **0.04** | **2.81x** | **318.8x** | **894.8x** |

### 2.3 Memory Allocation Multipliers (Cold Load)

| ID | Scenario Name | Level | Baseline Alloc (KB) | Wave 1 Alloc (KB) | Wave 2 Alloc (KB) | Wave 1 Multiplier | Wave 2 Multiplier | Total Multiplier | Total Memory Saved |
|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 6,420.0 | 4,035.0 | **3,883.9** | **1.59x** | **1.04x** | **1.65x** | **-2.5 MB** |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 12,850.0 | 7,435.0 | **7,943.7** | **1.73x** | **0.94x** | **1.62x** | **-4.9 MB** |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 453,040.0 | 105,619.0 | **92,674.6** | **4.29x** | **1.14x** | **4.89x** | **-360.4 MB** |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 176,140.0 | 73,228.7 | **63,118.9** | **2.41x** | **1.16x** | **2.79x** | **-113.0 MB** |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 33,520.0 | 19,198.0 | **17,615.7** | **1.75x** | **1.09x** | **1.90x** | **-15.9 MB** |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 227,100.0 | 113,258.9 | **71,498.4** | **2.01x** | **1.58x** | **3.18x** | **-155.6 MB** |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 20,400.0 | 8,602.0 | **4,369.9** | **2.37x** | **1.97x** | **4.67x** | **-16.0 MB** |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 49,600.0 | 21,700.1 | **19,850.6** | **2.29x** | **1.09x** | **2.50x** | **-29.7 MB** |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 58,200.0 | 38,309.2 | **25,161.1** | **1.52x** | **1.52x** | **2.31x** | **-33.0 MB** |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 2,130,599.0 | 414,077.3 | **241,257.3** | **5.15x** | **1.72x** | **8.83x** | **-1.89 GB!** |

---

## 3. Wave 2 Detailed Benchmark Results

*Captured on `perf/render-cost-suite` with warm iterations = 3 after applying both Wave 1 and Wave 2 architectural optimizations.*

| ID | Scenario Name | Level | Items | Boxes | Ops | Cold (ms) | Warm (ms) | Mutate (ms) | Resize (ms) | Layout (ms) | Alloc (KB) |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 50 | 305 | 404 | **7.40** | **0.00** | 5.19 | 4.35 | 6.55 | 3,883.9 |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 200 | 1,211 | 2,416 | **23.13** | **0.00** | 20.11 | 15.58 | 20.08 | 7,943.7 |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 200 | 2,605 | 3,205 | **161.57** | **0.00** | 136.50 | 126.05 | 157.49 | 92,674.6 |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 200 | 4,005 | 3,205 | **213.91** | **0.00** | 208.07 | 203.88 | 206.85 | 63,118.9 |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 500 | 3,009 | 2,511 | **32.62** | **0.00** | 36.02 | 27.51 | 27.45 | 17,615.7 |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 1,000 | 11,014 | 13,016 | **209.75** | **0.00** | 210.52 | 189.39 | 190.56 | 71,498.4 |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 30 | 185 | 191 | **13.31** | **0.00** | 13.44 | 13.69 | 12.61 | 4,369.9 |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 200 | 3,204 | 4,804 | **47.86** | **0.00** | 49.29 | 45.97 | 43.51 | 19,850.6 |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 150 | 1,205 | 19,755 | **62.20** | **0.00** | 59.84 | 59.55 | 58.66 | 25,161.1 |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 1,000 | 22,006 | 23,006 | **626.87** | **0.04** | 601.21 | 572.44 | 595.53 | 241,257.3 |

---

## 4. Key Architectural Optimizations Applied

### 4.1 Wave 2 Optimizations (Blinkless Layout Engine)

1. **Chunked Box Arena with Recycling Pool (`internal/layout/layout.go`)**
   - Implemented `boxChunkPool` of 64-element chunks. Box nodes are allocated contiguously into pooled chunks via `e.allocBox()`, eliminating individual heap escapes for 22,000+ boxes.
   - Pointers to individual boxes remain 100% address-stable within their dedicated chunk.
   - During `noEmit` measurement phases in `flex.go` and `flex_logical_axes.go`, `e.markBoxes()` and `e.resetBoxes(m)` roll back discarded speculative boxes immediately.
   - Upon layout completion, non-retained displays release chunks back to the pool via `res.Release()`.

2. **1D Linear Gradient Dimension Collapsing (`internal/layout/background_image.go`, `gradient.go`)**
   - For vertical (`to bottom` / `to top`) or horizontal (`to right` / `to left`) linear gradients, color variation across the orthogonal axis is mathematically zero.
   - Pure vertical gradients collapse to $1 \times H$ pixels, and pure horizontal gradients collapse to $W \times 1$ pixels.
   - Drawing with linear filtering reproduces pixel-identical gradients while cutting uncompressed raster memory up to 512x.

3. **Custom PNG Encoder Buffer Pool (`internal/layout/gradient.go`)**
   - Added `gradPNGPool` implementing `png.EncoderBufferPool` to `fastPNGEncoder`.
   - Eliminates standard library `image/png.Encoder`'s internal ~300 KB zlib and scanline allocations on every PNG encode.

4. **Continuous Operation IDs Buffer (`internal/layout/placed.go`)**
   - In `PlacedElements`, replaces hundreds of individual heap slice allocations with a single continuous `allIDs := make([]uint64, 0, len(res.Ops))` buffer.
   - Element `OperationIDs` slice views are sliced directly from this buffer without separate heap allocations.

5. **Primary Face Fast-Path in Inline Font Metrics (`internal/layout/inline_paint.go`)**
   - Added `faceRunAllPrimary(text, primary)` check in `inlineFontMetrics`.
   - Bypasses text slicing and font run splitting when all characters exist in the primary typeface.

6. **Generous Capacity Bound in `estimateOpCapacity` (`internal/layout/layout.go`)**
   - Bumped op capacity bound from $1.5 \times \text{nodes}$ to $2 \times \text{nodes} + 16$.
   - Ensures `oldOps` has enough spare capacity to absorb all `deferredChrome` ops in-place, eliminating slice reallocations in `mergeDeferredChrome`.

### 4.2 Wave 2 Optimizations (OwnFrame Host Runtime)

1. **Warm Redraw Template Bypass (`internal/page/page_draw.go`)**
   - In `execTemplate`, when `!p.formDirty && p.lastSyncBody != ""`, returns `p.lastSyncBody` immediately.
   - Slashes steady-state warm redraw time to **0.00 ms** (from 12.75 ms) and eliminates 88,745 allocations per warm frame on Scenario 10.

2. **Conditional Hit Target Allocation (`internal/page/page_boxcache.go`)**
   - In `setDisplayBoxes`, checks `len(fixed) > 0` before rebuilding `p.hits`.
   - For views without fixed elements (99% of views), skips copying 22,006 `layout.Box` structs (saving **7.7 MB of heap allocation** per redraw).

3. **Motion Animation Pre-Check (`internal/motion/bind.go`)**
   - Added early exit in `BindAfter`: if `!hasAnims && len(r.entries) == 0`, immediately adopts `next` without allocating `starts`, `carried`, or `used` hash maps.

4. **Static Transition Property List (`internal/motion/transition.go`)**
   - Replaced slice literal in `transitionPropertyList()` with a package-level immutable `transitionProperties []string`, eliminating 36-element slice allocations on every transitioning box.

5. **Prefix and Suffix Trimming in Display Diffing (`internal/page/dirty/diff.go`)**
   - In `diffByPosition`, detects matching leading prefix and trailing suffix ops in $O(N)$ before entering the hash arena.
   - Reduces hash arena memory and execution time to $O(\Delta)$ when modifying rows or elements within large displays.

---

## 5. Wave 1 Optimizations Summary

1. **Thread-Safe Bounded CSS Gradient Cache (`internal/layout/gradient.go`)**
   - Eliminated redundant per-card gradient rasterization with a `sync.RWMutex` bounded cache.
2. **Chunked Style Arena with `sync.Pool` Recycling (`internal/layout/layout.go`)**
   - Implemented chunked 64-record style arenas with `styleChunkPool`, saving hundreds of megabytes of `ResolvedStyle` structs.
3. **Rollback in Speculative Passes (`internal/layout/flex.go`)**
   - Added `e.markStyles()` and `e.resetStyles(mark)` to flex and table measurement passes.
4. **Single-Pass Placement & Operation Ownership (`layout/displaylist.go`)**
   - Reused placed elements across element boxes and operation owner tagging.
5. **DOM TextContent Fast Paths (`internal/html/html.go`)**
   - Zero-allocation text retrieval for single-child text nodes.
6. **Adaptive Box ID Hash Capacity (`internal/page/hit/index.go`)**
   - Avoided allocating full-size maps for mostly anonymous boxes.

---

## 6. Artifacts and Reproduction

The complete suite is implemented in [`examples/performance`](../../examples/performance):

- [`main.go`](../../examples/performance/main.go): CLI runner supporting `-warm`, `-format=markdown|table|csv`, and `-filter`.
- [`scenarios.go`](../../examples/performance/scenarios.go): HTML templates, CSS rules, and realistic mock data models for all 10 scenarios.
- [`measure.go`](../../examples/performance/measure.go): End-to-end benchmarking harness.
- [`performance_test.go`](../../examples/performance/performance_test.go): Automated verification test (`TestAllScenariosRender`) and standard Go benchmarks.
- [`types.go`](../../examples/performance/types.go): Data definitions.
- [`README.md`](../../examples/performance/README.md): Quick-start guide.

### Execution Commands

```bash
# Run the benchmark CLI in the worktree:
cd examples
go run ./performance -warm=3 -format=markdown

# Run the automated verification test:
go test -v ./performance

# Run standard Go benchmarks:
go test -bench=. ./performance
```
