# Phase 04: Code Editor Syntax Highlighting Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Completed  
> **Target Subsystems:** [`examples/code-view`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view)  
> **Primary Goal:** Implement line-level token caching to eliminate full-file re-tokenization on every keystroke.  

---

## 1. Problem Statement & Baseline Benchmark

In [`examples/code-view`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view):
*   On every keystroke in the code editor, the entire document was passed to [`tokenize()`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/syntax.go).
*   For a 1,000-line Go file, the lexer tokenized all 1,000 lines from scratch, allocating new token structs and string spans for the entire file.
*   **Measured Baselines:**
    *   **Go (1,000 lines):** **670.93 µs/op, 365.3 KB/op, 9,395 allocations per keystroke!**
    *   **Python (1,000 lines):** **1,205.48 µs (1.21 ms)/op, 775.7 KB/op, 14,832 allocations per keystroke!**
    *   **Markdown (1,000 lines):** **505.46 µs/op, 407.8 KB/op, 10,987 allocations per keystroke!**
*   Typing fast in a 1,000-line file triggered tens of thousands of heap allocations and wasted over 7% of the 16.66ms frame budget per character typed.

---

## 2. Target Metrics & Measured Improvement

| Metric | Baseline Before | Measured Result After Optimization | Multiplier |
| :--- | :---: | :---: | :---: |
| `Editor_Syntax_Go_1000` Latency | 670.93 µs | **6.74 µs** | **99.5x faster** |
| `Editor_Syntax_Go_1000` Memory Churn | 365.3 KB/keystroke | **24.5 KB/keystroke** | **93.3% less memory** |
| `Editor_Syntax_Go_1000` Allocations | 9,395 allocs/keystroke | **1 alloc/keystroke** | **9,394 allocs eliminated to 1!** |
| `Editor_Syntax_Python_1000` Latency | 1,205.48 µs (1.21 ms) | **7.04 µs** | **171x faster** |
| `Editor_Syntax_Python_1000` Allocations | 14,832 allocs/keystroke | **1 alloc/keystroke** | **14,831 allocs eliminated to 1!** |
| `Editor_Syntax_Markdown_1000` Latency | 505.46 µs | **6.84 µs** | **74x faster** |

---

## 3. Implementation Summary

### Step 1: Continuation State & Line-Indexed Token Cache
1. Implemented `lineState` tracking cross-line continuation modes for multiline comments (`/* ... */`), raw strings, CSS blocks, HTML tags, and Markdown fences in [`examples/code-view/syntax.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/syntax.go).
2. Implemented `syntaxCache` holding `tokenCache [][]segment`, `states []lineState`, and `dirty map[int]bool`.

### Step 2: Incremental Re-tokenization
1. When user edits line `K`, `syntaxCache.markDirty(K)` marks only that line.
2. In `syntaxCache.tokenize(lines)`, unmodified lines instantly reuse their cached token slices.
3. Multi-line continuation changes only re-tokenize subsequent lines until line continuation states re-converge with cached state.
4. Linked `doc.Syntax *syntaxCache` to `fileDoc` and editor mutation operations.

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-04-01` Create `syntaxCache` supporting line-level cache reuse in `examples/code-view/syntax.go`.
- [x] `OPT-04-02` Update editor edit operations to mark dirty lines instead of evicting the entire token tree.
- [x] `OPT-04-03` Run `go test -bench='BenchmarkEditor_Syntax' -benchmem ./examples/code-view`.
- [x] `OPT-04-04` Verify `Editor_Syntax_Go_1000` latency drops from 671 µs to 6.7 µs and allocations drop from 9,395 to 1.
