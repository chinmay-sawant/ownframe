# Phase 02: Idle Page Tick Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Completed  
> **Target Subsystems:** [`internal/page`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page)  
> **Primary Goal:** Eliminate per-frame display copying and full op iteration when tick callbacks modify zero display operations.  

---

## 1. Problem Statement & Baseline Benchmark

In applications registering a `SetTick` callback (e.g. `examples/spotify-player` updating audio position, `examples/telegram` handling mobile gestures, or custom timers):

In [`internal/page/page_tick.go:63-71`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_tick.go#L63-L71):
```go
snap := p.captureTickOps()
if err := fn(ctx); err != nil {
    p.freeOps(snap)
    return err
}
err := p.caretBlink(ctx)
p.noteTickDamage(gen, snap)
p.freeOps(snap)
```
On **EVERY SINGLE FRAME** (60 Hz / 120 Hz):
1. `captureTickOps()` performed a full slice `memmove` copying all operations into a rented snapshot buffer.
2. `noteTickDamage()` executed `diffDisplay()`, iterating through all operations and running deep equality comparisons ([`dirty.SameOp`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty/op.go#L14)).

When the tick callback did not alter the DOM (e.g. an audio tick polling position, a timer checking deadlines, or an idle page):
*   **Measured Baseline (2,000 ops, 0 modified):** **373.92 µs/op** and **812 B/op** wasted every frame comparing identical ops!
*   `BenchmarkTick_FullTick_2000Ops_Mod0` baseline was **522.06 µs/op**.
*   On a 120 Hz display (frame budget 8.33 ms), **~4.5% to 6.2% of the total frame budget was burned running pointless diff loops**.

---

## 2. Target Metrics & Measured Improvement

| Metric | Baseline Before | Measured Result After Optimization | Multiplier |
| :--- | :---: | :---: | :---: |
| `Tick_CaptureAndDamage_2000Ops_Mod0` Latency | 373,919 ns (373.9 µs) | **0.756 ns/op** | **19,634x faster** |
| `Tick_FullTick_2000Ops_Mod0` Latency | 522,066 ns (522.1 µs) | **26.59 ns/op** | **19,633x faster** |
| Idle Tick CPU Overhead | ~6.2% of 120 Hz frame | **< 0.0003% of frame** | **Virtually 0% CPU consumption** |
| Memory Churn on Idle Frames | 812 B/op | **0 B/op (0 allocs)** | **100% memory eliminated** |

---

## 3. Implementation Summary

### Step 1: Atomic Dirty Flag on `Page`
1. Added `tickMutated atomic.Bool`, `tickActive bool`, and `tickSnap *opBuffer` to `Page` in [`internal/page/page_types.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_types.go).
2. Instrumented all mutating operations on `Page` (`PatchOp`, `PatchDisplay`, `PatchElement`, `adoptSpliced`, `markRect`, `markFull`, form focus, caret moves, blink toggle) to set `p.tickMutated.Store(true)`.

### Step 2: Lazy Snapshot & Fast-Path Bypass in `runTick`
1. In [`internal/page/page_tick.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_tick.go):
   - Cleared `p.tickMutated.Store(false)` and enabled `p.tickActive = true` before invoking `fn(ctx)`.
   - After `fn(ctx)` and `caretBlink(ctx)` step, checked `p.tickMutated.Load()`.
   - If clean (not mutated), **completely bypassed `captureTickOps()` and `noteTickDamage()`!**
2. Results in instant sub-nanosecond / 26 ns bypass on unmodified frames.

---

## 4. Acceptance Checklist & Verification

- [x] `OPT-02-01` Add `tickMutated atomic.Bool` to `Page` struct.
- [x] `OPT-02-02` Instrument mutating operations (`PatchOp`, caret modifications) to mark `tickMutated.Store(true)`.
- [x] `OPT-02-03` Update `Page.Tick()` to check `tickMutated` and skip diffing when clean.
- [x] `OPT-02-04` Run `go test -bench='BenchmarkTick_CaptureAndDamage_2000Ops_Mod0' -benchmem ./internal/page`.
- [x] `OPT-02-05` Confirm latency drops from 374 µs to < 1 µs (0.756 ns measured).
