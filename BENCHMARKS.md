# Performance Benchmark & Zero-Allocation Optimization Report

This document records the empirical before-and-after performance metrics, memory allocation reductions, and speedup multipliers for the `ownframe` rendering engine and `code-view` editor virtualization pipeline.

---

## 1. Editor Virtualization, Windowing & Highlighting Benchmarks

*Measured using `test/editor_virtualization_test.go` and `examples/code-view/perf_editor_bench_test.go` on a 13th Gen Intel Core i7-13700HX.*

| Benchmark Target | Baseline Latency | Optimized Latency | Baseline Allocs | Optimized Allocs | Speedup Multiplier | Memory Eliminated |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Row Windowing During Scroll** *(1,000 lines)* | `21,514 ns` | **`58.46 ns`** | `146 allocs` (`34.2 KB`) | **`0 allocs` (`0 B`)** | **`368.0x faster`** | **`100%`** |
| **Fill Code Viewport Slices (`NoWrap`)** | `9,296 ns` | **`58.91 ns`** | `207 allocs` (`16.7 KB`) | **`0 allocs` (`0 B`)** | **`157.8x faster`** | **`100%`** |
| **Fill Code Viewport Slices (`Wrap`)** | `1,431 ns` | **`57.55 ns`** | `1 alloc` (`6.5 KB`) | **`0 allocs` (`0 B`)** | **`24.8x faster`** | **`100%`** |
| **Full Go Syntax Highlighting** *(1,000 lines)* | `2,244.4 µs` | **`72.95 µs`** | `20,389 allocs` (`2.7 MB`) | **`0 allocs` (`0 B`)** | **`30.8x faster`** | **`100%`** |
| **Full Go Syntax Highlighting** *(100 lines)* | `194.8 µs` | **`7.11 µs`** | `2,028 allocs` (`275 KB`) | **`0 allocs` (`0 B`)** | **`27.4x faster`** | **`100%`** |
| **Full Python Highlighting** *(1,000 lines)* | `2,471.1 µs` | **`147.61 µs`** | `25,826 allocs` (`3.5 MB`) | **`0 allocs` (`0 B`)** | **`16.7x faster`** | **`100%`** |
| **Full Python Highlighting** *(100 lines)* | `249.1 µs` | **`14.92 µs`** | `2,551 allocs` (`357 KB`) | **`0 allocs` (`0 B`)** | **`16.7x faster`** | **`100%`** |
| **Soft-Wrap Reflow** *(1,000 lines)* | `22,478 ns` | **`1,456 ns`** | `0 allocs` (`11 B`) | **`0 allocs` (`0 B`)** | **`15.4x faster`** | **`100%`** |
| **Soft-Wrap Reflow** *(100 lines)* | `2,084 ns` | **`141.4 ns`** | `0 allocs` (`0 B`) | **`0 allocs` (`0 B`)** | **`14.7x faster`** | **`100%`** |
| **Markdown Highlighting** *(100 lines)* | `57.2 µs` | **`12.54 µs`** | `1,132 allocs` (`55 KB`) | **`0 allocs` (`0 B`)** | **`4.6x faster`** | **`100%`** |
| **Markdown Highlighting** *(1,000 lines)* | `522.5 µs` | **`123.31 µs`** | `10,992 allocs` (`538 KB`) | **`0 allocs` (`0 B`)** | **`4.2x faster`** | **`100%`** |

---

## 2. Frame Pipeline & Interactive Runtime Benchmarks

*Measured using `test/micro_bench_test.go` and `test/frame_pipeline_test.go`.*

| Benchmark Target | Baseline Latency | Optimized Latency | Baseline Allocs | Optimized Allocs | Speedup Multiplier | Memory Reduced |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **Dirty Invalidate + Damage Drain** | `2,733 ns` | **`281.3 ns`** | `12 allocs` (`1,552 B`) | **`2 allocs` (`96 B`)** | **`9.7x faster`** | **`93.8%`** |
| **Interactive Click Dispatch** | `4,595 ns` | **`3,418 ns`** | `11 allocs` (`5,552 B`) | **`10 allocs` (`1,456 B`)** | **`1.35x faster`** | **`73.8%`** |
| **Game Loop Idle Tick** | `196.9 ns` | **`200.2 ns`** | `0 allocs` (`0 B`) | **`0 allocs` (`0 B`)** | **`1.0x` *(Parity)*** | **`0 B (Optimal)`** |
| **Warm Redraw Dashboard** | `3,230 ns` | **`2,857 ns`** | `6 allocs` (`1,760 B`) | `6 allocs` (`1,760 B`) | **`1.13x faster`** | Baseline |
| **Warm Redraw Card** | `2,005 ns` | **`2,078 ns`** | `10 allocs` (`1,456 B`) | `10 allocs` (`1,456 B`) | **`1.0x` *(Parity)*** | Baseline |

---

## 3. GPU Vector & Shader Replay Benchmarks

*Measured using `test/gpu_replay_test.go`.*

| Replay Benchmark Target | Measured Latency | Measured Memory | Measured Allocs | Pipeline Characteristics |
| :--- | :---: | :---: | :---: | :--- |
| **Display List Vector Traversal** | **`1.496 ns/op`** | **`0 B/op`** | **`0 allocs/op`** | Linear array walk over vector draw operations. |
| **Kage Shader Mix-Blend Pass** | **`535.8 ns/op`** | `2,705 B/op` | `2 allocs/op` | Direct GPU shader execution for CSS blend modes. |
| **Card Vector Rasterization** | **`2,383 ns/op`** | `16,403 B/op` | `2 allocs/op` | Full card vector rendering pass. |
| **Dashboard Vector Rasterization** | **`4,021 ns/op`** | `24,602 B/op` | `2 allocs/op` | Multi-panel dashboard vector rendering pass. |

---

## 4. Key Architectural Optimizations Applied

1. **Span-based Lexing (Byte Offsets over String Copies):**
   - Replaced substring allocations (`segment{Text: string}`) with integer boundaries (`TokenSpan{Start, End uint32; Kind uint8}`) operating directly on raw byte slices.
   - Tokens are appended into a pre-allocated flat `[]TokenSpan` arena, eliminating up to 25,000 heap allocations per 1,000 lines of code.

2. **In-Place Viewport Row Slicing (`v.Rows[:0]`):**
   - Pre-allocated a fixed viewport row buffer (`ViewportBuffer`) sized to the screen's maximum row capacity.
   - Slicing visible lines during scroll resets slice length in-place, achieving **0 allocations** and a **368x speedup**.

3. **Dirty Bounding-Box Invalidation & Damage Draining:**
   - Optimized element invalidation and partial repaint tracking, reducing invalidation latency from `2.73 µs` to `281 ns` with a **93.8% allocation reduction**.

4. **Zero-Allocation Soft-Wrap Calculations:**
   - Soft-wrap calculations reuse pre-sized integer breakpoint buffers, completely eliminating line reallocation overhead during window resize events.
