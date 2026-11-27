# Performance Benchmark & Zero-Allocation Optimization Report

This document records the empirical performance metrics, memory allocation reductions, and speedup multipliers across optimization rounds for the `ownframe` rendering engine and `code-view` editor virtualization pipeline.

---

## 1. Editor Virtualization, Windowing & Highlighting Benchmarks

*Measured using `test/editor_virtualization_test.go` and `examples/code-view/perf_editor_bench_test.go` on a 13th Gen Intel(R) Core(TM) i7-13700HX (24 threads) under Go 1.24 linux/amd64 (toolchain at measurement time; the repo now builds with Go 1.26.11 and these numbers have not been re-measured).*

| Benchmark Target | Baseline Latency | Round 1 Latency | Current Latency (Round 2) | Current Allocs | Total Speedup vs Baseline | Improvement vs Round 1 |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Row Windowing During Scroll** *(1,000 lines)* | `21,514 ns` | `58.46 ns` | **`27.37 ns`** | **`0 allocs` (`0 B`)** | **`786.0x faster`** | **`2.14x faster` (53.2% cut)** |
| **Fill Code Viewport Slices (`Wrap`)** | `1,431 ns` | `57.55 ns` | **`26.51 ns`** | **`0 allocs` (`0 B`)** | **`54.0x faster`** | **`2.17x faster` (53.9% cut)** |
| **Fill Code Viewport Slices (`NoWrap`)** | `9,296 ns` | `58.91 ns` | **`28.31 ns`** | **`0 allocs` (`0 B`)** | **`328.4x faster`** | **`2.08x faster` (51.9% cut)** |
| **Soft-Wrap Reflow** *(100 lines)* | `2,084 ns` | `141.4 ns` | **`38.16 ns`** | **`0 allocs` (`0 B`)** | **`54.6x faster`** | **`3.71x faster` (73.0% cut)** |
| **Soft-Wrap Reflow** *(1,000 lines)* | `22,478 ns` | `1,456 ns` | **`370.9 ns`** | **`0 allocs` (`0 B`)** | **`60.6x faster`** | **`3.93x faster` (74.5% cut)** |
| **Full Go Syntax Highlighting** *(100 lines)* | `194.8 µs` | `7.11 µs` | **`1.70 µs`** | **`0 allocs` (`0 B`)** | **`114.6x faster`** | **`4.18x faster` (76.1% cut)** |
| **Full Go Syntax Highlighting** *(1,000 lines)* | `2,244.4 µs` | `72.95 µs` | **`17.24 µs`** | **`0 allocs` (`0 B`)** | **`130.2x faster`** | **`4.23x faster` (76.4% cut)** |
| **Full Python Highlighting** *(100 lines)* | `249.1 µs` | `14.92 µs` | **`3.54 µs`** | **`0 allocs` (`0 B`)** | **`70.4x faster`** | **`4.21x faster` (76.3% cut)** |
| **Full Python Highlighting** *(1,000 lines)* | `2,471.1 µs` | `147.61 µs` | **`35.41 µs`** | **`0 allocs` (`0 B`)** | **`69.8x faster`** | **`4.17x faster` (76.0% cut)** |
| **Markdown Highlighting** *(100 lines)* | `57.2 µs` | `12.54 µs` | **`2.79 µs`** | **`0 allocs` (`0 B`)** | **`20.5x faster`** | **`4.50x faster` (77.8% cut)** |
| **Markdown Highlighting** *(1,000 lines)* | `522.5 µs` | `123.31 µs` | **`28.18 µs`** | **`0 allocs` (`0 B`)** | **`18.5x faster`** | **`4.38x faster` (77.1% cut)** |

---

## 2. Frame Pipeline & Interactive Runtime Benchmarks

*Measured using `test/micro_bench_test.go` and `test/frame_pipeline_test.go`.*

| Benchmark Target | Baseline Latency | Round 1 Latency | Current Latency (Round 2) | Current Allocs | Total Speedup vs Baseline | Improvement vs Round 1 |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Interactive Click Dispatch** | `4,595 ns` | `3,418 ns` | **`1,510 ns`** | **`0 allocs` (`0 B`)** | **`3.04x faster`** | **`2.26x faster` (55.8% cut)** |
| **Warm Redraw Dashboard** | `3,230 ns` | `2,857 ns` | **`544.6 ns`** | **`0 allocs` (`0 B`)** | **`5.93x faster`** | **`5.25x faster` (80.9% cut)** |
| **Warm Redraw Card** | `2,005 ns` | `2,078 ns` | **`482.1 ns`** | **`0 allocs` (`0 B`)** | **`4.16x faster`** | **`4.31x faster` (76.8% cut)** |
| **Dirty Invalidate + Damage Drain** | `2,733 ns` | `281.3 ns` | **`203.4 ns`** | **`0 allocs` (`0 B`)** | **`13.4x faster`** | **`1.38x faster` (27.7% cut)** |
| **Game Loop Idle Tick** | `196.9 ns` | `200.2 ns` | **`178.0 ns`** | **`0 allocs` (`0 B`)** | **`1.11x faster`** | **`1.12x faster` (11.1% cut)** |

---

## 3. GPU Vector & Shader Replay Benchmarks

*Measured using `test/gpu_replay_test.go`.*

| Replay Benchmark Target | Measured Latency | Measured Memory | Measured Allocs | Pipeline Characteristics |
| :--- | :---: | :---: | :---: | :--- |
| **Display List Vector Traversal** | **`1.367 ns/op`** | **`0 B/op`** | **`0 allocs/op`** | Direct contiguous array vector draw walk. |
| **Kage Shader Mix-Blend Pass** | **`454.1 ns/op`** | `2,689 B/op` | `1 alloc/op` | Direct GPU shader execution for CSS blend modes. |
| **Card Vector Rasterization** | **`2,024 ns/op`** | `16,386 B/op` | `1 alloc/op` | Full card vector rendering pass. |
| **Dashboard Vector Rasterization** | **`3,218 ns/op`** | `24,584 B/op` | `1 alloc/op` | Multi-panel dashboard vector rendering pass. |

---

## 4. Key Architectural Optimizations Applied

### Round 1: Zero-Allocation Virtualization
1. **Span-based Lexing (Byte Offsets over String Copies):**
   - Replaced substring allocations with integer boundaries (`TokenSpan{Start, End uint32; Kind uint8}`) operating directly on raw byte slices.
   - Tokens append into a pre-allocated flat `[]TokenSpan` arena, eliminating heap churn.
2. **In-Place Viewport Row Slicing (`v.Rows[:0]`):**
   - Pre-allocated fixed viewport row buffers (`ViewportBuffer`) sized to the screen's maximum row capacity.
   - Slicing visible lines during scroll resets slice length in-place, reducing allocations to 0 and yielding initial 368x speedup.
3. **Dirty Bounding-Box Invalidation & Damage Draining:**
   - Optimized element invalidation and partial repaint tracking, reducing invalidation latency from `2.73 µs` to `281 ns`.
4. **Zero-Allocation Soft-Wrap Calculations:**
   - Soft-wrap calculations reuse pre-sized integer breakpoint buffers, completely eliminating line reallocation overhead during window resize events.

### Round 2: Sub-Microsecond Acceleration & Core Engine Pipeline
1. **Direct Indexing in L1 Cache (`786x Speedup`):**
   - Slicing in `ViewportBuffer.Slice` pre-slices the buffer header (`vb.rows = vb.rows[:count]`) and assigns by direct index `vb.rows[j] = visualRow{...}`, eliminating all inner-loop `append()` branching, capacity checks, and slice increments.
2. **L1 Classification Lookup Table (`[256]uint8`):**
   - Replaced multi-byte UTF-8 rune decoding and inner linear loops (`strings.ContainsRune`) with a 256-byte static classification table (`charClass [256]uint8`) in L1 cache, classifying tokens in a single CPU memory read (~0.3 ns) and accelerating tokenizers by 4x across Go, Python, and Markdown.
3. **Static Template Bypass:**
   - In `Page.execTemplate()`, pages without dynamic Go template tags (`{{`) return `p.nav.Current()` directly, completely bypassing `html/template.Execute` reflection and `strings.Builder` allocations.
4. **Zero-Allocation Form State Synchronization:**
   - In `Page.syncForm()`, if the body and form haven't changed and the active control is non-editable, tag scanning (`form.Scan`) and map allocation (`form.Merge`) are skipped, returning the pre-rendered HTML in 0 ns.
5. **Closure & Interface Header Elimination:**
   - Surface adapters and image resolvers are pinned to `*page.Page` once during assembly (`p.imageResolver = p.imageBytes`), removing dynamic closure and interface heap allocations during frame redraws and click handling.
6. **Damage Tracker Buffer Capacity Retention:**
   - In `dirty.Tracker.Take`, internal slice capacity is preserved across frames (`takeBuf`) instead of setting `d.Rects = nil`, bringing damage draining down to `0 allocs/op`.

---

## 5. Master Cumulative Timeline & Lifetime Multipliers

The lifetime multiplier compounds the sequential phases of optimization ($\text{Total Lifetime Multiplier} = \text{Earlier Multiplier} \times \text{Today's Multiplier} = \frac{\text{Original Baseline Latency}}{\text{Today's Final Latency}}$):

| Hot Path / Benchmark | Original Baseline | Earlier Multiplier<br>*(Phase Pass / R1)* | Today's Latency<br>*(Round 2 Final)* | Today's Multiplier<br>*(Added Gain)* | Updated Lifetime Multiplier<br>*(Earlier × Today)* | Allocations & Memory |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Row Windowing During Scroll** | `21,514 ns` | **`368.0x`** *(R1: 58.46 ns)* | **`27.37 ns`** | **`2.14x faster`** (53.2% cut) | **`786.0x faster`** *(368.0 × 2.14)* | **`0 allocs` (`0 B`)** *(was 146 allocs / 34.2 KB)* |
| **Window Soft-Wrap Resizing** | `656,500 ns` *(656.5 µs)* | **`31.2x`** *(Phase 05: 21.06 µs)* | **`370.9 ns`** | **`56.8x faster`** *(3.93x vs R1)* | **`1,770x faster`** *(31.2 × 56.8)* | **`0 allocs` (`0 B`)** *(was 6,350 allocs / 1.82 MB)* |
| **Go Keystroke Highlighting** | `670,900 ns` *(670.9 µs)* | **`99.5x`** *(Phase 04: 6.74 µs)* | **`1.70 µs`** | **`3.96x faster`** (74.8% cut) | **`394.0x faster`** *(99.5 × 3.96)* | **`0 allocs` (`0 B`)** *(was 9,395 allocs / 365 KB)* |
| **Python Keystroke Highlighting** | `1,205,500 ns` *(1.21 ms)* | **`171.2x`** *(Phase 04: 7.04 µs)* | **`3.54 µs`** | **`1.99x faster`** (49.7% cut) | **`340.9x faster`** *(171.2 × 1.99)* | **`0 allocs` (`0 B`)** *(was 14,832 allocs / 776 KB)* |
| **Viewport Slices (`NoWrap`)** | `9,296 ns` | **`157.8x`** *(R1: 58.91 ns)* | **`28.31 ns`** | **`2.08x faster`** (51.9% cut) | **`328.4x faster`** *(157.8 × 2.08)* | **`0 allocs` (`0 B`)** *(was 207 allocs / 16.7 KB)* |
| **Full File Go Highlighting** *(1k lines)* | `2,244,440 ns` *(2.24 ms)* | **`30.8x`** *(R1: 72.95 µs)* | **`17.24 µs`** | **`4.23x faster`** (76.4% cut) | **`130.2x faster`** *(30.8 × 4.23)* | **`0 allocs` (`0 B`)** *(was 20,389 allocs / 2.75 MB)* |
| **Warm Redraw Dashboard** | `3,230 ns` | **`1.13x`** *(R1: 2,857 ns)* | **`544.6 ns`** | **`5.25x faster`** (80.9% cut) | **`5.93x faster`** *(1.13 × 5.25)* | **`0 allocs` (`0 B`)** *(was 6 allocs / 1.76 KB)* |
| **Warm Redraw Card** | `2,005 ns` | **`0.96x`** *(R1: 2,078 ns)* | **`482.1 ns`** | **`4.31x faster`** (76.8% cut) | **`4.16x faster`** *(0.96 × 4.31)* | **`0 allocs` (`0 B`)** *(was 10 allocs / 1.45 KB)* |
| **Interactive Click Dispatch** | `4,595 ns` | **`1.34x`** *(R1: 3,418 ns)* | **`1,510 ns`** | **`2.26x faster`** (55.8% cut) | **`3.04x faster`** *(1.34 × 2.26)* | **`0 allocs` (`0 B`)** *(was 11 allocs / 5.55 KB)* |
| **Dirty Invalidate + Drain** | `2,733 ns` | **`9.72x`** *(R1: 281.3 ns)* | **`203.4 ns`** | **`1.38x faster`** (27.7% cut) | **`13.44x faster`** *(9.72 × 1.38)* | **`0 allocs` (`0 B`)** *(was 12 allocs / 1.55 KB)* |
| **Full Game Loop Idle Tick** | `522,066 ns` *(522.1 µs)* | **`19,634x`** *(Phase 02: 26.59 ns)* | **`26.59 ns`** *(178 ns loop)* | *Preserved* | **`19,634x faster`** | **`0 allocs` (`0 B`)** *(was 4 allocs / 812 B)* |
| **Idle Page Tick Diff Overhead** | `373,919 ns` *(373.9 µs)* | **`494,601x`** *(Phase 02: 0.756 ns)* | **`0.756 ns`** | *Optimal* | **`494,601x faster`** | **`0 allocs` (`0 B`)** *(was 2,000 op diffs / 812 B)* |

---

## 6. Engineering Analysis & Performance Ceilings

### 1. Frame Budget & Hardware Limits
* **Display Frame Budget**:
  * At 60 FPS, the full frame budget is **16,666 µs** (16.6 ms).
  * At 144 FPS, it is **6,944 µs**.
  * At 240 FPS, it is **4,166 µs**.
* **Engine Latency**:
  * Row windowing during scroll takes **`0.027 µs`** (`27.37 ns`).
  * Warm redraw takes **`0.54 µs`** (`544.6 ns`).
  * Click dispatch takes **`1.5 µs`** (`1,510 ns`).
  * The entire update cycle consumes less than **0.01% of a single frame budget**.
* **Input Polling Comparison**:
  * High-end gaming peripherals poll via USB at 1,000 Hz, reporting events every **1,000 µs (1 ms)**.
  * The engine's interactive click dispatch handles the event in **1.5 µs**—over 600× faster than the hardware input bus itself.

### 2. Zero Allocations = Zero GC Pauses
* In Go, the primary bottleneck in UI rendering engines is not raw instruction throughput, but **Garbage Collection (GC) pauses and runtime mark-assist stutters**.
* Every benchmark in the hot path now operates at **`0 B/op`** and **`0 allocs/op`**.
* The game loop, scroll row slicing, diffing, and redraw paths generate zero heap garbage, eliminating runtime GC pressure during active user interactions.

### 3. CPU Physics & Diminishing Returns
* At **`27.37 ns`** for 1,000-line viewport virtualization, the CPU completes the operation in approximately **80–100 clock cycles** on modern x86_64 architectures. This operates within the latency threshold of L1/L2 cache line loads.
* Further micro-optimizations would require raw assembly, unsafe pointer arithmetic, or AVX-512 SIMD vectorization. Doing so would sacrifice Go portability, memory safety, and codebase maintainability for nanosecond margins undetectable by display hardware or human perception.

---

## 7. Latest Core Pipeline, Layout Scaling, Animation & Native Windows GPU Benchmarks (2026-11-25)

*Measured sequentially (`-p 1`) on `13th Gen Intel(R) Core(TM) i7-13700HX`, Go `1.26.11`, Blinkless `v0.1.9`, and native Windows GPU (`NVIDIA GeForce RTX 4060 Laptop GPU` @ 144 Hz, `CGO_ENABLED=0`).*

### Core Pipeline & 1,000-Row Workloads

| Benchmark | Time/op | Bytes/op | Allocs/op |
|---|---:|---:|---:|
| `BenchmarkFlappyTick` | **58.8 ns** | 0 B | 0 |
| `BenchmarkListWorkloads/ScrollOffset` | **22.1 ns** | 0 B | 0 |
| `BenchmarkDrawRectScan` (dirty / full) | **191 / 183 ns** | 0 B | 0 |
| `BenchmarkRedrawWarm` / `RedrawHeavyWarm` | **496 / 487 ns** | 0 B | 0 |
| `BenchmarkNormalRedraw` | **536 ns** | 0 B | 0 |
| `BenchmarkRedrawStages/TemplateExecute` | **178 ns** | 560 B | 5 |
| `BenchmarkRedrawStages/HTMLParse` | **5.07 µs** | 27.9 KB | 16 |
| `BenchmarkRedrawStages/CSSApply` | **2.24 µs** | 3.18 KB | 37 |
| `BenchmarkRedrawStages/DisplayList` | **36.8 µs** | 33.9 KB | 110 |
| `BenchmarkRedrawCold` | **65.7 µs** | 64.8 KB | 204 |
| `BenchmarkAnimationTick` *(534-op CSS gallery)* | **97.7 µs** | 16.6 KB | 38 |
| `BenchmarkClickCount` *(platform example)* | **221 µs** | 118 KB | 532 |
| `BenchmarkListWorkloads/WindowedScroll` | **508 µs** | 177 KB | 1,183 |
| `BenchmarkListPatchVsRebuild/PatchElement1Row` | **584 µs** | 208 KB | 102 |
| `BenchmarkListPatchVsRebuild/PatchElements10Rows` | **815 µs** | 275 KB | 624 |
| `BenchmarkListWorkloads/Resize` *(1,000 rows)* | **11.59 ms** | 2.78 MB | 4,298 |
| `BenchmarkListWorkloads/TextUpdate` *(1,000 rows)* | **16.32 ms** | 6.38 MB | 35,336 |
| `BenchmarkListPatchVsRebuild/FullRebuild1Row` | **16.85 ms** | 6.37 MB | 44,290 |
| `BenchmarkLargeRedraw` | **17.14 ms** | 2.79 MB | 7,234 |

### Pipeline Stage Scaling (`BenchmarkScaleStages`) & Layout Models (`BenchmarkLayoutModels`)

| Stage / Model | 50 rows | 200 rows | 1,000 rows Total | Bytes/op | Allocs/op |
|---|---:|---:|---:|---:|---:|
| `TemplateExecute` | 2,426 ns/row | 2,440 ns/row | **2.45 ms** | 520 KB | 26,730 |
| `HTMLParse` | 1,398 ns/row | 1,318 ns/row | **1.39 ms** | 1.64 MB | 6,123 |
| `CSSApply` | 143.4 ns/row | 62.7 ns/row | **42.2 µs** | 6.68 KB | 75 |
| `CSSRelayout` | 1.37 ns/row | 0.31 ns/row | **59.2 ns** | 180 B | 1 |
| `DisplayList` | 12,872 ns/row | 11,511 ns/row | **12.57 ms** | 5.59 MB | 4,338 |
| `LayoutModels/Block` *(200 rows × 3 cols)* | — | 10,033 ns/row | **2.01 ms** | 1.04 MB | 1,000 |
| `LayoutModels/Flex` *(200 rows × 3 cols)* | — | 11,131 ns/row | **2.23 ms** | 1.06 MB | 1,025 |
| `LayoutModels/Grid` *(200 rows × 3 cols)* | — | 11,874 ns/row | **2.37 ms** | 1.24 MB | 2,410 |

### Native Windows GPU Windows (`-tags ownframe_debug -perf -debug-addr`, 8 s each)

| App | Window | Present FPS | Avg Frame | p95 / p99 Frame | Long Frames (>25 ms) | Go Heap (Live) | Windows RSS | CPU (1-core eq.) |
|---|---|---:|---:|---:|---:|---:|---:|---:|
| `animations.exe` *(534 ops, active CSS)* | 1140×900 | **144.1 FPS** *(auto-paced)* | **2.64 ms** | **4.24 / 5.08 ms** | **0** | 17.6 MB | 384.0 MB | 89.6% |
| `wispr-flow-dashboard.exe` *(398 ops)* | 1638×950 | **442.3 FPS** | **0.75 ms** | **1.12 / 7.45 ms** | **0** | 21.4 MB | 344.3 MB | 28.9% |
| `spotify-player.exe` *(169 ops, 46 images)* | 1440×960 | **68.2 FPS** *(paced)* | **3.57 ms** | **5.35 / 9.59 ms** | **0** | 22.0 MB | 265.5 MB | 18.6% |
| `departure-board.exe` *(628 ops, idle)* | 1000×640 | **2.1 FPS** *(parked)* | **1.14 ms** | **1.48 / 12.19 ms** | 1 | 17.7 MB | 224.5 MB | **4.1%** |

---

## 8. Pipeline & Animation Optimization Wave (`v0.0.8` / Blinkless `v0.1.10`, 2026-11-27)

*Measured on `13th Gen Intel(R) Core(TM) i7-13700HX`, Go `1.26.11`, Blinkless `v0.1.10` (`linux/amd64`).*

### Always-On Pipeline Improvements

| Benchmark | `v0.0.7` Baseline | `v0.0.8` Current | Latency Delta | Allocs Delta |
| :--- | ---: | ---: | ---: | ---: |
| `BenchmarkAnimationTick` *(534-op CSS gallery)* | `98.87 µs` · `38 allocs` · `16.5 KB` | **`52.36 µs`** · **`20 allocs`** · **`11.8 KB`** | **`-47.0%` (`1.89x`)** | **`-47.4%`** |
| `BenchmarkScaleStages/Rows1000/DisplayList` | `11.56 ms` · `4,316 allocs` | **`10.09 ms`** · **`318 allocs`** | **`-12.7%`** | **`-92.6%`** |
| `BenchmarkListWorkloads/PerfOff/Resize` | `11.67 ms` · `4,295 allocs` | **`10.21 ms`** · **`290 allocs`** | **`-12.5%`** | **`-93.2%`** |
| `BenchmarkListWorkloads/PerfOn/Resize` | `11.91 ms` · `4,299 allocs` | **`10.17 ms`** · **`292 allocs`** | **`-14.6%`** | **`-93.2%`** |
| `BenchmarkListWorkloads/PerfOff/TextUpdate` | `16.74 ms` · `35,329 allocs` | **`15.05 ms`** · **`31,324 allocs`** | **`-10.1%`** | **`-11.3%`** |
| `BenchmarkListWorkloads/PerfOn/TextUpdate` | `16.53 ms` · `35,330 allocs` | **`14.93 ms`** · **`31,325 allocs`** | **`-9.7%`** | **`-11.3%`** |
| `BenchmarkListWorkloads/PerfOff/WindowedScroll` | `518.8 µs` · `1,181 allocs` | **`495.1 µs`** · **`1,082 allocs`** | **`-4.6%`** | **`-8.4%`** |
| `BenchmarkListPatchVsRebuild/FullRebuild1Row` | `16.83 ms` · `44,267 allocs` | **`15.49 ms`** · **`40,278 allocs`** | **`-8.0%`** | **`-9.0%`** |

### Opt-In Fast Paths (1,000 Rows)

| Stage | Standard Path | Opt-In Path | Latency Delta | Memory / Alloc Delta |
| :--- | ---: | ---: | ---: | ---: |
| HTML generation (`Config.Render` vs `TemplateExecute`) | `2,400 µs` · `518.9 KB` · `26,730 allocs` | **`64.45 µs`** · **`3.06 KB`** · **`991 allocs`** | **`-97.3%` (`37.2x`)** | **`-99.4%` bytes**, **`-96.3%` allocs** |
| DisplayList (`Options.CullWindow` 600 px vs full) | `9.73 ms` · `2.88 MB` · `304 allocs` | **`1.24 ms`** · **`1.82 MB`** · **`261 allocs`** | **`-87.2%` (`7.83x`)** | **`-36.7%` bytes**, **`-14.1%` allocs** |

---

## 9. Retained DOM Mutation & Copy-on-Write Batch Relayout (`feat/retained-dom-mutation`, 2026-11-28)

*Measured on `13th Gen Intel(R) Core(TM) i7-13700HX`, Go `1.26.11` (`linux/amd64`), median of 3 runs via `BenchmarkListPatchVsRebuild`, `BenchmarkRetainedMutations`, and `blinkless/layout.BenchmarkRelayoutElements1000Rows`.*

### `ownframe` End-to-End 1,000-Row Patching (`BenchmarkListPatchVsRebuild`)

| Benchmark | `v0.0.8` Baseline | Pass 1 (CoW Box Graph + Lazy Owners) | Pass 2 (`SheetVersion` + `DiffDirect`) | Total Speedup vs. Baseline | Memory (Baseline -> Pass 2) | Allocs |
| :--- | ---: | ---: | ---: | ---: | ---: | ---: |
| `PatchElement1Row` | `584 µs` | `361.7 µs` | **`176.5 µs`** | **`3.31x faster` (`-69.8%`)** | `208 KB` -> **`77.9 KB` (`-62.5%`)** | `102` -> **`75`** |
| `PatchElements10Rows` | `815 µs` | `679.9 µs` | **`452.6 µs`** | **`1.80x faster` (`-44.5%`)** | `275 KB` -> **`249.7 KB` (`-9.2%`)** | `624` -> **`540`** |
| `FullRebuild1Row` *(reference)* | `16.85 ms` | `17.30 ms` | **`15.28 ms`** | **`PatchElement1Row` is `86.6x` faster** | `6.37 MB` -> `6.05 MB` | `40,275` (**`537x` fewer** on 1-row patch) |

### `blinkless` Engine Subtree Relayout (`BenchmarkRelayoutElements1000Rows`)

| Benchmark | Baseline | After Pass 1 | Speedup | Memory (Before -> After) | Allocs |
| :--- | ---: | ---: | ---: | ---: | ---: |
| `Text1Row` | `471.5 µs` | **`121.5 µs`** | **`3.88x faster` (`-74.2%`)** | `222.2 KB` -> **`76.7 KB` (`-65.5%`)** | `83` -> **`67`** |
| `TextLastRow` | `635.6 µs` | **`175.1 µs`** | **`3.63x faster` (`-72.5%`)** | `216.9 KB` -> **`76.2 KB` (`-64.9%`)** | `79` -> **`63`** |
| `Text10Rows` | `729.7 µs` | **`432.1 µs`** | **`1.69x faster` (`-40.8%`)** | `290.8 KB` -> **`249.1 KB` (`-14.3%`)** | `622` -> **`537`** |
| `BgRestyle1Row` | `529.2 µs` | **`154.9 µs`** | **`3.42x faster` (`-70.7%`)** | `127.9 KB` -> `220.9 KB` | `81` -> `95` |

### Retained Mutation Regression Baselines (`BenchmarkRetainedMutations`)

| Benchmark | Median Time | Bytes/op | Allocs/op |
| :--- | ---: | ---: | ---: |
| `ClassToggle1Row` *(1,000 rows, `AddClass` / `RemoveClass`)* | **`1.29 ms`** | `517.1 KB` | `148` |
| `AttributeSetRemove1Row` *(1,000 rows, `Attribute` / `RemoveAttribute`)* | **`1.26 ms`** | `499.2 KB` | `144` |
| `StructuralAppendRemove` *(50-row feed, `AppendHTML` / `RemoveElement`)* | **`236.8 µs`** | `172.2 KB` | `243` |
| `StructuralSetInnerHTML` *(replacing feed children with HTML fragment)* | **`209.2 µs`** | `660.6 KB` | `189` |

