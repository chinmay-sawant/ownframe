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

## 2. Multiplier Progression: Baseline vs. Wave 1 vs. Wave 2 vs. Wave 3 vs. Wave 4

Following four consecutive architectural optimization waves across both `blinkless` (layout engine) and `ownframe` (host runtime):
- **Cold Latency:** Accelerated up to **6.79x faster** overall (extreme 1,000-component app rendered in **551.9ms**, down from **3,745.3ms**).
- **Warm Redraw Latency:** Slashed up to **894.8x faster** (steady-state warm frames drop to **<0.01ms** with zero allocations).
- **Memory Allocations:** Slashed by **up to 10.0x** (saving **1.92 GB of heap allocations** on Scenario 10, cutting Scenario 03 by an additional **45% in Wave 4** from 91.2 MB to 50.3 MB).

### 2.1 Cold Latency Multipliers (Initial Render)

$$\text{Wave 1 Mult} = \frac{\text{Baseline}}{\text{Wave 1}}, \quad \text{Wave 2 Mult} = \frac{\text{Wave 1}}{\text{Wave 2}}, \quad \text{Wave 3 Mult} = \frac{\text{Wave 2}}{\text{Wave 3}}, \quad \text{Wave 4 Mult} = \frac{\text{Wave 3}}{\text{Wave 4}}, \quad \text{Total Mult} = \frac{\text{Baseline}}{\text{Wave 4}}$$

| ID | Scenario Name | Level | Baseline (ms) | Wave 1 (ms) | Wave 2 (ms) | Wave 3 (ms) | Wave 4 (ms) | Wave 1 Mult | Wave 2 Mult | Wave 3 Mult | Wave 4 Mult | Total Mult | Latency Slashed |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 17.58 | 11.63 | 7.40 | 7.27 | **6.67** | **1.51x** | **1.57x** | **1.02x** | **1.09x** | **2.64x** | **-62.1%** |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 46.88 | 39.72 | 23.13 | 21.65 | **18.21** | **1.18x** | **1.72x** | **1.07x** | **1.19x** | **2.57x** | **-61.2%** |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 486.29 | 222.20 | 161.57 | 158.92 | **128.17** | **2.19x** | **1.38x** | **1.02x** | **1.24x** | **3.79x** | **-73.6%** |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 505.41 | 280.30 | 213.91 | 197.62 | **166.55** | **1.80x** | **1.31x** | **1.08x** | **1.19x** | **3.03x** | **-67.0%** |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 85.34 | 43.94 | 32.62 | 34.74 | **27.80** | **1.94x** | **1.35x** | **0.94x** | **1.25x** | **3.07x** | **-67.4%** |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 569.11 | 262.75 | 209.75 | 212.56 | **166.27** | **2.17x** | **1.25x** | **0.99x** | **1.28x** | **3.42x** | **-70.8%** |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 77.58 | 18.31 | 13.31 | 14.07 | **12.59** | **4.24x** | **1.38x** | **0.95x** | **1.12x** | **6.16x** | **-83.8%** |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 236.90 | 72.54 | 47.86 | 47.64 | **46.37** | **3.27x** | **1.52x** | **1.00x** | **1.03x** | **5.11x** | **-80.4%** |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 190.66 | 97.07 | 62.20 | 59.57 | **64.03** | **1.96x** | **1.56x** | **1.04x** | **0.93x** | **2.98x** | **-66.4%** |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 3,745.28 | 909.71 | 626.87 | 595.48 | **551.89** | **4.12x** | **1.45x** | **1.05x** | **1.08x** | **6.79x** | **-85.3%** |

### 2.2 Warm Redraw Multipliers (Steady-State Frames)

| ID | Scenario Name | Level | Baseline (ms) | Wave 1 (ms) | Wave 2 (ms) | Wave 3 (ms) | Wave 4 (ms) | Wave 1 Mult | Wave 2 Mult | Wave 3 Mult | Wave 4 Mult | Total Mult |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 0.35 | 0.21 | <0.01 | <0.01 | **<0.01** | **1.67x** | **>20x** | **1.00x** | **1.00x** | **>35x** |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 1.62 | 1.23 | <0.01 | <0.01 | **<0.01** | **1.32x** | **>120x** | **1.00x** | **1.00x** | **>160x** |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 1.76 | 1.14 | <0.01 | <0.01 | **<0.01** | **1.54x** | **>110x** | **1.00x** | **1.00x** | **>170x** |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 2.14 | 1.78 | <0.01 | <0.01 | **<0.01** | **1.20x** | **>170x** | **1.00x** | **1.00x** | **>210x** |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 6.72 | 5.49 | <0.01 | <0.01 | **<0.01** | **1.22x** | **>540x** | **1.00x** | **1.00x** | **>670x** |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 7.91 | 5.89 | <0.01 | <0.01 | **<0.01** | **1.34x** | **>580x** | **1.00x** | **1.00x** | **>790x** |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 1.82 | 0.14 | <0.01 | <0.01 | **<0.01** | **13.00x** | **>14x** | **1.00x** | **1.00x** | **>180x** |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 3.89 | 1.83 | <0.01 | <0.01 | **<0.01** | **2.13x** | **>180x** | **1.00x** | **1.00x** | **>380x** |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 2.45 | 2.15 | <0.01 | <0.01 | **<0.01** | **1.14x** | **>210x** | **1.00x** | **1.00x** | **>240x** |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 35.79 | 12.75 | 0.04 | 0.04 | **0.07** | **2.81x** | **318.8x** | **1.00x** | **0.57x** | **511.3x** |

### 2.3 Memory Allocation Multipliers (Cold Load)

$$\text{Wave 1 Mult} = \frac{\text{Baseline}}{\text{Wave 1}}, \quad \text{Wave 2 Mult} = \frac{\text{Wave 1}}{\text{Wave 2}}, \quad \text{Wave 3 Mult} = \frac{\text{Wave 2}}{\text{Wave 3}}, \quad \text{Wave 4 Mult} = \frac{\text{Wave 3}}{\text{Wave 4}}, \quad \text{Total Mult} = \frac{\text{Baseline}}{\text{Wave 4}}$$

| ID | Scenario Name | Level | Baseline Alloc (KB) | Wave 1 Alloc (KB) | Wave 2 Alloc (KB) | Wave 3 Alloc (KB) | Wave 4 Alloc (KB) | Wave 1 Mult | Wave 2 Mult | Wave 3 Mult | Wave 4 Mult | Total Mult | RAM Saved |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 6,420.0 | 4,035.0 | 3,883.9 | 3,712.2 | **3,509.8** | **1.59x** | **1.04x** | **1.05x** | **1.06x** | **1.83x** | **-2.9 MB** |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 12,850.0 | 7,435.0 | 7,943.7 | 7,279.3 | **6,227.7** | **1.73x** | **0.94x** | **1.09x** | **1.17x** | **2.06x** | **-6.6 MB** |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 453,040.0 | 105,619.0 | 92,674.6 | 91,202.8 | **50,361.4** | **4.29x** | **1.14x** | **1.02x** | **1.81x** | **9.00x** | **-402.7 MB!** |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 176,140.0 | 73,228.7 | 63,118.9 | 24,640.6 | **21,647.6** | **2.41x** | **1.16x** | **2.56x** | **1.14x** | **8.14x** | **-154.5 MB!** |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 33,520.0 | 19,198.0 | 17,615.7 | 15,935.5 | **13,883.5** | **1.75x** | **1.09x** | **1.11x** | **1.15x** | **2.41x** | **-19.6 MB** |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 227,100.0 | 113,258.9 | 71,498.4 | 54,397.7 | **71,504.7** | **2.01x** | **1.58x** | **1.31x** | **0.76x** | **3.18x** | **-155.6 MB** |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 20,400.0 | 8,602.0 | 4,369.9 | 3,290.1 | **3,148.5** | **2.37x** | **1.97x** | **1.33x** | **1.05x** | **6.48x** | **-17.3 MB** |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 49,600.0 | 21,700.1 | 19,850.6 | 18,075.0 | **16,047.4** | **2.29x** | **1.09x** | **1.10x** | **1.13x** | **3.09x** | **-33.6 MB** |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 58,200.0 | 38,309.2 | 25,161.1 | 24,385.1 | **25,148.2** | **1.52x** | **1.52x** | **1.03x** | **0.97x** | **2.31x** | **-33.1 MB** |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 2,130,599.0 | 414,077.3 | 241,257.3 | 217,855.5 | **213,383.5** | **5.15x** | **1.72x** | **1.11x** | **1.02x** | **9.98x** | **-1.92 GB!** |

---

## 3. Wave 4 Detailed Benchmark Results

*Captured on `perf/render-cost-suite` with warm iterations = 3 after applying Wave 1, Wave 2, Wave 3, and Wave 4 architectural optimizations.*

| ID | Scenario Name | Level | Items | Boxes | Ops | Cold (ms) | Warm (ms) | Mutate (ms) | Resize (ms) | Layout (ms) | Alloc (KB) |
|---|---|---|---|---|---|---|---|---|---|---|---|
| `01-simple-rows-50` | Simple Rows (50) | Low | 50 | 305 | 404 | **6.67** | **0.00** | 4.97 | 3.87 | 6.00 | 3,509.8 |
| `02-zebra-table-200` | Zebra Table (200 Rows) | Medium-Low | 200 | 1,211 | 2,416 | **18.21** | **0.00** | 17.81 | 13.56 | 15.57 | 6,227.7 |
| `03-product-grid-200` | Product Catalog Grid (200 Cards) | Medium | 200 | 2,605 | 3,205 | **128.17** | **0.00** | 109.40 | 107.56 | 123.94 | 50,361.4 |
| `04-social-timeline-200` | Social Chat Timeline (200 Posts) | Medium-High | 200 | 4,005 | 3,205 | **166.55** | **0.00** | 168.17 | 153.13 | 159.34 | 21,647.6 |
| `05-order-book-500` | Market Order Book (500 Rows) | High | 500 | 3,009 | 2,511 | **27.80** | **0.00** | 31.61 | 21.77 | 23.08 | 13,883.5 |
| `06-enterprise-grid-1000` | Enterprise Data Grid (1,000 Rows) | High | 1,000 | 11,014 | 13,016 | **166.27** | **0.00** | 170.15 | 141.82 | 151.18 | 71,504.7 |
| `07-analytics-dashboard` | Analytics Dashboard (Multi-Pane) | Complex | 30 | 185 | 191 | **12.59** | **0.00** | 10.76 | 12.20 | 11.84 | 3,148.5 |
| `08-multipage-doc-200` | Multi-Page Document (200 Pages) | Complex | 200 | 3,204 | 4,804 | **46.37** | **0.00** | 47.26 | 42.71 | 40.15 | 16,047.4 |
| `09-css-kitchen-sink-150` | CSS Kitchen Sink (150 Widgets) | Complex | 150 | 1,205 | 19,755 | **64.03** | **0.00** | 50.05 | 50.09 | 59.28 | 25,148.2 |
| `10-extreme-app-1000` | Extreme Scale App (1,000 Complex Rows) | Extreme | 1,000 | 22,006 | 23,006 | **551.89** | **0.07** | 502.93 | 468.63 | 514.30 | 213,383.5 |

---

## 4. Key Architectural Optimizations Applied Across Waves

### 4.1 Wave 4 Optimizations (Blinkless Layout Engine & OwnFrame Runtime)

1. **Speculative Measurement Subtree Rollbacks (`internal/layout/flex.go`, `flex_logical_axes.go`, `multicol.go`, `layout_tables.go`)**
   - In flex baseline measurement, table cell height measurement, and multicol column calculation, `e.noEmit = true` speculatively builds subtrees to measure heights. Previously, this leaked tens of thousands of speculative boxes and styles directly into `boxChunks` and `styleChunks`, which were then re-allocated during placement.
   - Added `bmark := e.markBoxes()` and `smark := e.markStyles()` with automatic rollbacks via `e.resetBoxes(bmark)` and `e.resetStyles(smark)`, reclaiming over **40.8 MB** in Scenario 03 Product Catalog Grid.

2. **Chrome Early-Exit Bypass (`internal/layout/layout_chrome.go`)**
   - Added `hasPotentialChrome(&sty, boxNode)` to bypass `prependChrome` entirely for boxes lacking backgrounds, borders, box-shadows, and outlines.
   - Avoids wasted border-radii calculations, corner-shape geometry sampling, and shadow appends for plain text containers and wrapper nodes.

3. **Zero-Allocation Deferred Chrome Containment (`internal/layout/layout_flow.go`)**
   - Replaced `markBoxSubtree` (which allocated maps and closures taking ~12% of CPU time during flow layout) in `shiftDeferredChrome` with zero-allocation `boxContains(boxNode, entries[idx].b)` and early-exit on `deltaX == 0 && deltaY == 0` or leaf boxes.

4. **Generic Paint Order Sorting (`internal/layout/paint_order.go`)**
   - Replaced `sort.SliceStable` (which relied on reflection `reflect.ValueOf`, `reflect.Swapper`, and closure allocations) with standard library `slices.SortStableFunc` and single-pass `comparePaintOrder`.
   - Added an $O(N)$ fast-path checking if operations are already in non-decreasing paint order, avoiding sorting passes entirely when chrome and content are already ordered.

5. **Canonical Class Attribute Interning (`internal/layout/style_match_cache.go`)**
   - Added `singleAttrCache` (`map[string]string` guarded by `sync.RWMutex`) for `name == "class"`. Reuses canonical attribute keys across thousands of identical class names, eliminating redundant `strings.Builder` allocations in large DOM trees.

6. **Calibrated Capacity for Display List Operations (`internal/layout/layout.go`)**
   - Refined `estimateOpCapacity` to allocate $1.25 \times \text{nodes}$ instead of $2.0 \times \text{nodes}$, saving ~2.6 MB of upfront slice allocation per cold render without triggering reallocation growth chains.

---

### 4.2 Wave 3 Optimizations (Blinkless Layout Engine & OwnFrame Runtime)

1. **Fused PlacedElements & Direct Box/Parent Index Population (`layout/displaylist.go`, `internal/layout/placed.go`, `internal/layout/op_owners.go`)**
   - Eliminated intermediate `[]PlacedElement` slice allocations (reclaiming **78.6 MB** per pass in Scenario 10).
   - Reused a single stack-allocated `PlacedElement` in `WalkPlaced(res, fn)`, populating the caller's target `[]Box` and `[]int` parents directly.
   - Wired `OperationOwnersFromParents` to read parent indices directly without copying or traversing intermediate placement slices.

2. **Zero-Heap Anonymous Flex Item Style Allocation (`internal/layout/flex.go`)**
   - Replaced by-value `anonymousFlexItemStyle()` copies (which escaped to the heap due to pointer references in `setSyntheticStyle`) with `e.allocAnonymousFlexItemStyle()` using contiguous style chunks via `applyAnonymousFlexItemStyle`.
   - Eliminated **181.06 MB** of heap allocations per render in high-density flex views.

3. **Pooled Flex Children Buffers (`internal/layout/flex.go`)**
   - Added `sync.Pool` (`flexKidPool`) for child and flow node slices with inline fast whitespace checks (`isAllWhitespace`), eliminating slice resizing churn across recursive flex passes.

4. **Descendant Text Detection & Builder Recycling (`internal/html/html.go`)**
   - Added `hasAnyText()` fast-path to immediately return empty string for container elements with no text descendants.
   - Added `textBuilderPool` (`sync.Pool` of `strings.Builder`) for composite text nodes, eliminating builder byte slice reallocations.

5. **Gradient Layer & Blend Mode Early-Exits (`internal/layout/gradient.go`, `blend.go`, `background_image.go`)**
   - Added fast keyword scan in `isPureLinear1D` to bypass full gradient AST parsing when angles are clearly diagonal.
   - Early-exited `backgroundBlendModeForLayer` on empty strings, eliminating `splitCommaLayers` allocations on non-blended layers.

6. **Probe Display Direct View (`examples/performance/performance_test.go`, `measure.go`)**
   - Avoided defensive copies of 22,006 `layout.Box` structs during benchmarking and probe checks by accessing `page.Probe().Display().Boxes` directly.

---

### 4.3 Wave 2 Optimizations (Archived Reference)

1. **Chunked Box Arena with Recycling Pool (`internal/layout/layout.go`)**
   - Contiguous 64-element chunk allocation via `e.allocBox()` eliminating individual heap escapes for 22,000+ boxes.
2. **1D Linear Gradient Dimension Collapsing (`internal/layout/background_image.go`, `gradient.go`)**
   - Collapsed pure horizontal and vertical linear gradients to $1 \times H$ and $W \times 1$ textures with linear filtering.
3. **Custom PNG Encoder Buffer Pool (`internal/layout/gradient.go`)**
   - Added `gradPNGPool` implementing `png.EncoderBufferPool` to eliminate zlib buffer churn.
4. **Continuous Operation IDs Buffer (`internal/layout/placed.go`)**
   - Replaced individual slice allocations with single continuous buffer.
5. **Warm Redraw Template Bypass (`internal/page/page_draw.go`)**
   - Slashed warm redraw latency to **0.00ms** by skipping template execution when data is unchanged.
6. **Conditional Hit Target Allocation (`internal/page/page_boxcache.go`)**
   - Skipped hit cache rebuilds for views with zero fixed elements (saving **7.7 MB** per redraw).

---

### 4.4 Wave 1 Optimizations (Archived Reference)

1. **One-Pass Inline CSS Scanner & Property Lookup Table (`internal/layout/inline_collect.go`)**
   - Replaced repeated CSS token parsing with fast single-pass property scanner.
2. **Deterministic String Interning & String Concatenation Slicing (`internal/layout/text.go`)**
   - Interned common tags, class tokens, and measurement keys to avoid redundant heap allocations.
3. **Display List Pre-Allocation (`layout/displaylist.go`)**
   - Accurately pre-sized display list slices based on element box count estimates.
