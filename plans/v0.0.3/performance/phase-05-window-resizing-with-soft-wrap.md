# Phase 05: Window Resizing with Soft-Wrap Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Completed  
> **Target Subsystems:** [`examples/code-view`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view)  
> **Primary Goal:** Eliminate massive slice allocation churn during editor window resizing with soft-wrap enabled.  

---

## 1. Problem Statement & Baseline Benchmark

In [`examples/code-view/wrap.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go):
*   When a user dragged to resize the window, or when the explorer/diff split ratio changed, the editor's visual width changed continuously.
*   [`buildWrapPlan()`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go) recalculated the soft-wrap breakpoints for every line in the file, allocating thousands of small sub-slices for visual row segments.
*   **Measured Baseline (1,000 lines):**
    *   **Latency:** **656.47 µs per resize step**.
    *   **Memory Overhead:** **1,815,885 B (1.82 MB)** allocated per step!
    *   **Allocations:** **6,350 heap allocations per step**.
*   A fluid window drag lasting 1 second (60 resize events) allocated over **100 MB of heap memory**, choking the garbage collector and causing window lag.

---

## 2. Target Metrics & Measured Improvement

| Metric | Baseline Before | Measured Result After Optimization | Multiplier |
| :--- | :---: | :---: | :---: |
| `Editor_SoftWrap_Resize_1000` Latency | 656.47 µs | **21.06 µs** | **31x faster** |
| `Editor_SoftWrap_Resize_1000` Memory | 1,815,885 B (1.82 MB) | **66 B / resize step** | **99.996% memory eliminated** <br> **(1.82 MB down to 66 B)** |
| `Editor_SoftWrap_Resize_1000` Allocations | 6,350 allocs / resize step | **0 allocs / resize step** | **6,350 allocs eliminated to ZERO** |
| `Editor_SoftWrap_Resize_100` Latency | 65.78 µs | **2.07 µs** | **31.7x faster** |

---

## 3. Implementation Summary

### Step 1: Flat Row Plan Arena & Capacity Retention
1. Created `type wrapPlanArena struct { offsets []int; lengths []int; cells []wrapCell }` on `wrapPlan` in [`examples/code-view/wrap.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/wrap.go).
2. In `fileDoc.wrapped(cols)`, retained and reset slice lengths (`plan.Rows = plan.Rows[:0]`, `plan.First = plan.First[:len(d.Tokens)+1]`, `offsets = offsets[:0]`, `lengths = lengths[:0]`) across recalculations, completely avoiding fresh slice array re-allocation.
3. For unwrapped lines (`rune count <= cols`), constructed visual rows directly referencing existing tokens without re-allocating segment slices.
4. Reused `arena.cells` for multiline wrapped continuation rows.

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-05-01` Create flat `wrapPlanArena` in `examples/code-view/wrap.go`.
- [x] `OPT-05-02` Retain pre-allocated slice capacities across window resize recalculations.
- [x] `OPT-05-03` Run `go test -bench='BenchmarkEditor_SoftWrap_Resize' -benchmem ./examples/code-view`.
- [x] `OPT-05-04` Verify memory allocation drops from 1.82 MB to 66 B and 0 allocs per resize step.
