# Phase 03: Typing & Letter-Spacing Calculation Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Planned  
> **Target Subsystems:** [`internal/page/dirty`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty), [`internal/replay`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay)  
> **Primary Goal:** Eliminate allocation storms during DOM content updates and optimize letter-spaced glyph rendering.  

---

## 1. Problem Statement & Baseline Benchmark

### 1. `diffByPosition` Allocation Surge During Typing & Form Edits
In [`internal/page/dirty/diff.go:41-47`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty/diff.go#L41-L47):
```go
func diffByPosition(prev, next *layout.Display) (Set, bool, int) {
    queue := make(map[opKey][]int, len(next.Ops))
    for i := range next.Ops {
        k := keyOf(&next.Ops[i])
        queue[k] = append(queue[k], i)
    }
    seen := make([]bool, len(next.Ops))
```
Whenever an operation count changes (e.g. typing a character into an `<input>`, appending a message in chat, expanding an accordion):
*   A new Go `map[opKey][]int` is allocated on the heap.
*   Because each operation in UI layouts typically has a distinct `(x, y, w, h)`, `queue[k] = append(queue[k], i)` creates **thousands of independent heap slices**.
*   **Measured Baseline (2,000 ops):** **954.47 µs (~0.95 ms), 313.1 KB/op, 2,011 allocations per diff!**
*   Typing at 60 keystrokes/minute produces massive GC pressure and latency spikes.

### 2. Letter-Spacing Glyph Run Allocation Overhead
In [`internal/replay/text.go:42`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/text.go#L42):
```go
for _, r := range op.Text {
    str := string(r)
    adv := text.Measure(str, face, 0)
    ...
}
```
Letter-spaced text converts every rune into a heap string `string(r)` and runs individual layout measurement, completely bypassing the shaped glyph cache.
*   **Measured Baseline:** **79.31 µs/op, 28.5 KB/op, 393 allocations per string** (vs 32.38 µs / 109 allocs without letter spacing).

---

## 2. Target Metrics & Improvement Multiplier

| Metric | Current Baseline | Target (After Optimization) | Multiplier |
| :--- | :---: | :---: | :---: |
| `Tick_DiffByPosition_2000Ops` Latency | 954.47 µs | **~95 µs** | **10x faster** |
| `Tick_DiffByPosition_2000Ops` Allocations | 2,011 allocs/op (313.1 KB) | **0 allocs/op (< 1 KB)** | **> 300x less memory churn** |
| `Text_LetterSpacing_Cold` Latency | 79.31 µs | **~15 µs** | **5.3x faster** |
| `Text_LetterSpacing_Cold` Allocations | 393 allocs/op (28.5 KB) | **~5 allocs/op (~2 KB)** | **78x fewer allocations** |

---

## 3. Implementation Plan

### Step 1: Flat Open-Addressing Hash Arena for `diffByPosition`
1. Define a reusable `diffArena` using flat integer slices instead of `map[opKey][]int`:
   ```go
   type diffArena struct {
       head []int // hash bucket heads (-1 for empty)
       next []int // linked list of next operation indices with matching hash
       seen []bool
   }
   ```
2. Pool `diffArena` instances using `sync.Pool`.
3. In `diffByPosition`, rent the arena, clear heads with `-1`, and build the index chain using flat array indexing.
4. Drops allocations from **2,011 to 0**.

### Step 2: Shaped Letter-Spaced Text Caching
1. Instead of rune-by-rune `string(r)` conversions, shape the full string once through `go-text/typesetting`.
2. Apply letter spacing directly to the output glyph advance array before blitting.
3. Cache the shaped glyph run in the text raster cache so subsequent frames execute in < 1 µs.

---

## 4. Acceptance Checklist & Verification

- [ ] `OPT-03-01` Create pooled `diffArena` in `internal/page/dirty/diff.go`.
- [ ] `OPT-03-02` Refactor `diffByPosition` to eliminate `map[opKey][]int` allocations.
- [ ] `OPT-03-03` Update `internal/replay/text.go` letter-spacing logic to avoid per-rune string allocations.
- [ ] `OPT-03-04` Run `go test -bench='BenchmarkTick_DiffByPosition|BenchmarkText_LetterSpacing' -benchmem ./internal/page ./internal/replay`.
- [ ] `OPT-03-05` Verify `diffByPosition` achieves 0 heap allocations and letter-spacing allocations drop by >90%.
