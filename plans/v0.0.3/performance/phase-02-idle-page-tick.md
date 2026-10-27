# Phase 02: Idle Page Tick Optimization

> **Parent:** [`plans/v0.0.3/performance/README.md`](README.md)  
> **Status:** Planned  
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
1. `captureTickOps()` performs a full slice `memmove` copying all operations into a rented snapshot buffer.
2. `noteTickDamage()` executes `diffDisplay()`, iterating through all operations and running deep equality comparisons ([`dirty.SameOp`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/dirty/op.go#L14)).

When the tick callback does not alter the DOM (e.g. an audio tick polling position, a timer checking deadlines, or an idle page):
*   **Measured Baseline (2,000 ops, 0 modified):** **373.92 µs/op** and **812 B/op** wasted every frame comparing identical ops!
*   On a 120 Hz display (frame budget 8.33 ms), **~4.5% of the total frame budget is burned running pointless diff loops**.

---

## 2. Target Metrics & Improvement Multiplier

| Metric | Current Baseline (0 ops modified) | Target (After Optimization) | Multiplier |
| :--- | :---: | :---: | :---: |
| `Tick_CaptureAndDamage_2000Ops_Mod0` Latency | 373.92 µs | **< 1.0 µs** | **> 370x faster** |
| Idle Tick CPU Overhead | ~4.5% of 120 Hz frame | **< 0.01% of frame** | **Virtually 0% CPU consumption** |
| Memory Churn on Idle Frames | 812 B/op | **0 B/op** | **100% memory eliminated** |

---

## 3. Implementation Plan

### Step 1: Atomic Dirty Flag on `Page`
1. Add `tickMutated atomic.Bool` to `Page` in [`internal/page/page_types.go`](file:///home/chinmay/ChinmayPersonalProjects/ownframe/internal/page/page_types.go).
2. Set `p.tickMutated.Store(true)` inside any mutation method that patches retained ops (e.g., `p.PatchOp()`, `p.patchElement()`, `p.MarkRect()`).

### Step 2: Lazy Snapshot / Fast-Path Bypass in `Page.Tick`
1. In `page_tick.go`, before invoking `fn(ctx)`:
   - Clear `p.tickMutated.Store(false)`.
2. After `fn(ctx)` returns:
   - Check `p.tickMutated.Load()`.
   - If `!p.tickMutated.Load()` and caret state did not change: **bypass `captureTickOps()` and `noteTickDamage()` completely!**
3. This eliminates both the 500 KB memory copy and the 2,000-op iteration loop on every clean frame.

---

## 4. Acceptance Checklist & Verification

- [ ] `OPT-02-01` Add `tickMutated atomic.Bool` to `Page` struct.
- [ ] `OPT-02-02` Instrument mutating operations (`PatchOp`, caret modifications) to mark `tickMutated.Store(true)`.
- [ ] `OPT-02-03` Update `Page.Tick()` to check `tickMutated` and skip diffing when clean.
- [ ] `OPT-02-04` Run `go test -bench='BenchmarkTick_CaptureAndDamage_2000Ops_Mod0' -benchmem ./internal/page`.
- [ ] `OPT-02-05` Confirm latency drops from 374 µs to < 1 µs.
