# Phase 05: Window Resizing with Soft-Wrap Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Planned  
> **Target Subsystems:** [`examples/code-view`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view)  
> **Primary Goal:** Eliminate massive slice allocation churn during editor window resizing with soft-wrap enabled.  

---

## 1. Problem Statement & Baseline Benchmark

In [`examples/code-view/wrap.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go):
*   When a user drags to resize the window, or when the explorer/diff split ratio changes, the editor's visual width changes continuously.
*   [`buildWrapPlan()`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go) recalculates the soft-wrap breakpoints for every line in the file, allocating thousands of small sub-slices for visual row segments.
*   **Measured Baseline (1,000 lines):**
    *   **Latency:** **656.47 µs per resize step**.
    *   **Memory Overhead:** **1,815,885 B (1.82 MB)** allocated per step!
    *   **Allocations:** **6,350 heap allocations per step**.
*   A fluid window drag lasting 1 second (60 resize events) allocates over **100 MB of heap memory**, choking the garbage collector and causing window lag.

---

## 2. Target Metrics & Improvement Multiplier

| Metric | Current Baseline (1,000 lines) | Target (After Optimization) | Multiplier |
| :--- | :---: | :---: | :---: |
| `Editor_SoftWrap_Resize_1000` Latency | 656.47 µs | **~130 µs** | **5x faster** |
| `Editor_SoftWrap_Resize_1000` Memory | 1.82 MB / resize step | **~50 KB / resize step** | **36x less memory churn** |
| `Editor_SoftWrap_Resize_1000` Allocations | 6,350 allocs / resize step | **~10 allocs / resize step** | **635x fewer allocations** |

---

## 3. Implementation Plan

### Step 1: Flat Row Plan Arena & Capacity Retention
1. Avoid creating thousands of independent sub-slices (`[]wrapSpan`).
2. Store visual row breakpoints in a continuous flat integer array (`[]int` offsets) on a reusable `wrapArena` struct:
   ```go
   type wrapPlanArena struct {
       offsets []int // start character offset of each visual row
       lengths []int // length of each visual row
   }
   ```
3. When resizing, reset slice lengths to zero (`offsets = offsets[:0]`) and reuse pre-allocated capacity rather than creating 6,350 slice objects.

### Step 2: Viewport-Only Breakpoint Calculation
1. Calculate precise wrapping only for lines currently within the visible scroll window + overscan margin.
2. Estimate average row counts for offscreen lines, refining when scrolled into view.

---

## 4. Acceptance Checklist & Verification

- [ ] `OPT-05-01` Create flat `wrapPlanArena` in `examples/code-view/wrap.go`.
- [ ] `OPT-05-02` Retain pre-allocated slice capacities across window resize recalculations.
- [ ] `OPT-05-03` Run `go test -bench='BenchmarkEditor_SoftWrap_Resize' -benchmem ./examples/code-view`.
- [ ] `OPT-05-04` Verify memory allocation drops from 1.82 MB to < 50 KB per resize step.
