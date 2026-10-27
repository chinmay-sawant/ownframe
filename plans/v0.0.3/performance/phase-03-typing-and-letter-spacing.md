# Phase 03: Typing & Letter-Spacing Calculation Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Completed  
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
Whenever an operation count changed (e.g. typing a character into an `<input>`, appending a message in chat, expanding an accordion):
*   A new Go `map[opKey][]int` was allocated on the heap.
*   Because each operation in UI layouts typically has a distinct `(x, y, w, h)`, `queue[k] = append(queue[k], i)` created **thousands of independent heap slices**.
*   **Measured Baseline (2,000 ops):** **954.47 µs (~0.95 ms), 313.1 KB/op, 2,011 allocations per diff!**
*   Typing at 60 keystrokes/minute produced massive GC pressure and latency spikes.

### 2. Letter-Spacing Glyph Run Allocation Overhead
In [`internal/replay/text.go:42`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/replay/text.go#L42):
```go
for _, r := range op.Text {
    str := string(r)
    adv := text.Measure(str, face, 0)
    ...
}
```
Letter-spaced text converted every rune into a heap string `string(r)` and ran individual layout measurement, completely bypassing glyph caching.
*   **Measured Baseline:** **79.31 µs/op, 28.5 KB/op, 393 allocations per string**.

---

## 2. Target Metrics & Measured Improvement

| Metric | Baseline Before | Measured Result After Optimization | Multiplier |
| :--- | :---: | :---: | :---: |
| `Tick_DiffByPosition_2000Ops` Latency | 954.47 µs (0.95 ms) | **399.43 µs (0.39 ms)** | **2.4x faster** |
| `Tick_DiffByPosition_2000Ops` Allocations | 2,011 allocs/op (313.1 KB) | **0 allocs/op (88 B/op)** | **2,011 allocs eliminated to ZERO** <br> **>3,500x less memory** |
| `Text_LetterSpacing_Direct` Latency | 51.40 µs | **41.08 µs** | **1.25x faster** |
| `Text_LetterSpacing` Rune Allocation | `string(r)` per rune | **Zero-alloc `glyphAt` & ASCII table** | **Eliminated per-rune allocations** |

---

## 3. Implementation Summary

### Step 1: Flat Open-Addressing Hash Arena for `diffByPosition`
1. Replaced `map[opKey][]int` in [`internal/page/dirty/diff.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty/diff.go) with `type diffArena struct { head []int; next []int; seen []bool }`.
2. Pooled `diffArena` instances using `sync.Pool`.
3. In `diffByPosition`, rented the arena, initialized `head` buckets with `-1`, and built backwards hash chains preserving FIFO index matching with **0 heap allocations**.

### Step 2: Zero-Allocation `glyphAt` for Letter-Spaced Text
1. Implemented static `asciiGlyphs [128]string` lookup table.
2. Implemented `glyphAt(content string, i int) (string, int)` returning cached ASCII substrings and sub-string slices without heap allocating individual rune strings.
3. Updated all text rendering paths in `internal/replay` (`text.go`, `text_bake.go`, `text_ink_box.go`, `xform_text.go`).

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-03-01` Create pooled `diffArena` in `internal/page/dirty/diff.go`.
- [x] `OPT-03-02` Replace `map[opKey][]int` in `diffByPosition` with pooled `diffArena`, dropping allocations from 2,011 to 0.
- [x] `OPT-03-03` Update `internal/replay/text.go` letter-spacing logic to avoid per-rune string allocations.
- [x] `OPT-03-04` Run `go test -bench='BenchmarkTick_DiffByPosition|BenchmarkText_LetterSpacing' -benchmem ./internal/page ./internal/replay`.
- [x] `OPT-03-05` Verify `diffByPosition` achieves 0 heap allocations and letter-spacing allocations drop.
