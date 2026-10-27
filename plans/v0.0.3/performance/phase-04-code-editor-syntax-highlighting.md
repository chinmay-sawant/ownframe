# Phase 04: Code Editor Syntax Highlighting Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Planned  
> **Target Subsystems:** [`examples/code-view`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view)  
> **Primary Goal:** Implement line-level token caching to eliminate full-file re-tokenization on every keystroke.  

---

## 1. Problem Statement & Baseline Benchmark

In [`examples/code-view`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view):
*   On every keystroke in the code editor, the entire document is passed to [`tokenize()`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/examples/code-view/syntax.go).
*   For a 1,000-line Go file, the lexer tokenizes all 1,000 lines from scratch, allocating new token structs and string spans for the entire file.
*   **Measured Baselines:**
    *   **Go (1,000 lines):** **670.93 µs/op, 365.3 KB/op, 9,395 allocations per keystroke!**
    *   **Python (1,000 lines):** **1,205.48 µs (1.21 ms)/op, 775.7 KB/op, 14,832 allocations per keystroke!**
    *   **Markdown (1,000 lines):** **505.46 µs/op, 407.8 KB/op, 10,987 allocations per keystroke!**
*   Typing fast in a 1,000-line file triggers tens of thousands of heap allocations and wastes over 7% of the 16.66ms frame budget per character typed.

---

## 2. Target Metrics & Improvement Multiplier

| Metric | Current Baseline (1,000 lines) | Target (After Incremental Caching) | Multiplier |
| :--- | :---: | :---: | :---: |
| `Editor_Syntax_Go_1000` Latency | 670.93 µs | **~12 µs** | **55x faster** |
| `Editor_Syntax_Go_1000` Memory Churn | 365.3 KB/keystroke | **~2 KB/keystroke** | **180x less memory churn** |
| `Editor_Syntax_Go_1000` Allocations | 9,395 allocs/keystroke | **~30 allocs/keystroke** | **313x fewer allocations** |
| `Editor_Syntax_Python_1000` Latency | 1,205.48 µs | **~18 µs** | **66x faster** |

---

## 3. Implementation Plan

### Step 1: Line-Indexed Token Cache
1. Introduce a line-level cache structure on the editor model:
   ```go
   type syntaxCache struct {
       lines      []string
       tokenCache [][]syntaxSpan
       dirty      map[int]bool
   }
   ```
2. When the user types on line `K`:
   - Mark line `K` dirty.
   - For single-line edits not opening a multi-line comment (`/* ... */`) or raw string (`` `...` ``), re-tokenize **only line `K`**.
   - Reuse existing cached spans for the other 999 unmodified lines.

### Step 2: Multi-line Context Preservation
1. Track whether a line terminates with an open multi-line comment or multi-line string state (`stateInComment`, `stateInString`).
2. Only invalidate subsequent lines if line `K` alters the multiline continuation state.

---

## 4. Acceptance Checklist & Verification

- [ ] `OPT-04-01` Create `syntaxCache` supporting line-level cache reuse in `examples/code-view/syntax.go`.
- [ ] `OPT-04-02` Update editor edit operations to mark dirty lines instead of evicting the entire token tree.
- [ ] `OPT-04-03` Run `go test -bench='BenchmarkEditor_Syntax' -benchmem ./examples/code-view`.
- [ ] `OPT-04-04` Verify `Editor_Syntax_Go_1000` latency drops from 671 µs to < 15 µs and allocations drop from 9,395 to < 50.
