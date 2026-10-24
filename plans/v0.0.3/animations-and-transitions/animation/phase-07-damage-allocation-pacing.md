# Phase 07 - damage, allocation, pacing

Goal: a frame with N running animations costs time proportional to N, allocates
nothing, damages only what moved, and asks the window for exactly the frames it
needs.

## Where it lives

- `internal/animation/sample.go:18` allocates a stop slice per property per
  effect per target. `compose.go:18-22` samples all six fields before it checks
  whether the keyframes carry them, so the allocation runs six times per target.
- `sample.go:27-32` prepends and appends to the stop slice, which can allocate
  again.
- `sample.go:49` re-parses the easing string on every sample.
- `compose.go:79-86` walks every keyframe once per field to decide presence.
- `player.go:58-60` writes every target on every step while `touched` is set,
  even after the value stopped changing.
- `internal/page/page_tick.go:53-62` snapshots every op and diffs it when a user
  callback exists. The nil-callback path already avoids this.
- `internal/page/dirty` caps at 8 rects and collapses past a third of the frame
  (`dirty/dirty.go:7-19`, `take.go:10-38`).
- `internal/window/pacing/policy.go:73-143` treats a declared `WantFrameAt` as
  the page owning the schedule.
- `documentation/performance-baseline.md:95` measures `internal/frame` at 176 ns
  and 0 allocs for a patched frame. The CSS path has no matching number yet.

## Design

1. Sample per effect, apply per target. The current loop is
   `for target { for effect { sample six fields } }`. Invert it:
   `for effect { sample once into a small value set }`, then
   `for target { apply the set through the mask }`. The sample depends on the
   effect and the time, not on the target. This alone removes most of the
   repeated keyframe walks.

2. Precompute at bind, per effect and per field:
   - a flat, sorted `[]stop{at, value}` with explicit stops at 0 and 1 when the
     keyframes omit them,
   - the segment index for the last sampled time, so a monotonic clock does a
     forward scan instead of a binary search from zero,
   - the parsed easing function, one per segment,
   - the presence mask per field.

   Store these in the bound entry. A rebind rebuilds them. Nothing in `Step`
   allocates.

3. Zero-allocation step. Use fixed arrays for the value set and the damage set,
   sized by the field count and the existing eight-rect cap. Run
   `go test -bench` with `-benchmem` and treat any `B/op` above zero on a live
   sample as a bug.

4. Damage per element, in page space. For each changed target, union the old and
   new painted box. For a matrix change use the transformed bounds. Union the
   targets of one element into one rect and mark it once. Do not mark the union
   of all elements; that is what the animations example does today and it
   collapses the dirty set to a full repaint on a wide page.

5. Marking order. The driver marks its rects before the user callback snapshot,
   so `diffDisplay` never sees animation changes. If the user callback is nil,
   the existing fast path stays and the page's tracker receives only driver
   damage plus any caret redraw.

6. Skip unchanged writes. Keep the last written value per target and field.
   A finished player, a paused player, and a frame where the curve is flat write
   nothing and damage nothing. A static hold at the end of a fill must not keep
   the window awake.

7. Pacing:
   - any active sample: `WantFrameIn(0)`,
   - end-delay hold: `WantFrameIn(remaining)`,
   - paused or finished: do not ask,
   - scroll or view timeline: ask on the scroll that changed the input, not on a
     timer.
   A page with an infinite animation asks every frame by design. A page with a
   one-shot animation asks until it finishes and then parks.

8. 120 Hz. Sampling runs in `Update`, which Ebiten defaults to 60 TPS while
   `Draw` can run at panel rate. Add `WindowOptions.TPS` (zero means the Ebiten
   default) so a host can ask for 120. The pace timer already captures
   `ebiten.TPS()` as its base and restores it after idle
   (`internal/window/idle_game.go:48`, `idle_frame.go:62`), so the option is a
   one-line pass-through. Document that a 120 Hz sample needs the option, not a
   custom tick loop.

9. Counters, not logs: sampled frames, sampled players, field writes, skipped
   writes, damage rects, dropped rects, rebinds, and the per-frame sample cost.
   Expose them through `Probe()` in phase 09.

## Checklist

### A. Sampling

- [ ] Invert the loop to sample per effect, apply per target.
- [ ] Precompute stops with explicit 0 and 1 sentinels.
- [ ] Precompute segment easing functions.
- [ ] Keep a forward cursor per effect for a monotonic clock.
- [ ] Presence mask per field, no `hasValue` walk per frame.
- [ ] Remove the prepend and append allocations.

### B. Damage

- [ ] Old ∪ new painted box per target.
- [ ] Transformed bounds for a matrix change.
- [ ] Union per element, mark once.
- [ ] Never mark the union of all elements.
- [ ] Count dropped rects when the eight-rect cap overflows.
- [ ] Test a full-width 43-element page: damage stays under a third.

### C. Writes

- [ ] Last-value cache per target and field.
- [ ] Skip unchanged fields.
- [ ] Skip a finished or paused player.
- [ ] Do not call `SetXform` when the matrix is unchanged.

### D. Pacing

- [ ] `WantFrameIn(0)` while active.
- [ ] `WantFrameIn(remaining)` during an end-delay hold.
- [ ] No request when paused or finished.
- [ ] No request for a static scroll timeline.
- [ ] Add `WindowOptions.TPS` and test that the idle round trip restores it.

### E. Measurement

- [ ] Benchmark 1 animation, 10 animations, and the 43-card catalog.
- [ ] Record ns/frame, allocs, and damage rects in
      `documentation/performance-baseline.md`.
- [ ] Compare the CSS path against the 176 ns `internal/frame` number.
- [ ] Profile one catalog frame and attach the top three costs to the phase note.

Gate: `-benchmem` shows 0 allocs on a live sample, the catalog frame allocates
nothing in the sample path, and a 43-element page does not collapse damage to a
full repaint.

## Risks

- The forward cursor assumes time moves forward. A reverse direction still moves
  forward in raw progress, so this is safe; a paused-then-resumed player is not,
  so reset the cursor on resume.
- Per-element damage can still exceed the eight-rect cap on a page with many
  independent animations. That is a tracker policy question, not a driver bug;
  measure before changing the cap.
- A 120 Hz option changes the idle base. Test the park-and-wake round trip, not
  just the fast path.
