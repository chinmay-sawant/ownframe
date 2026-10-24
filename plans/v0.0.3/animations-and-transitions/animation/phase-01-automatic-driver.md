# Phase 01 - automatic driver

Goal: a page that declares `@keyframes` animates when it opens, without any
application code. This phase creates the runtime package and the page hook that
every later phase extends.

## Where it lives

Baseline, cited from the tree on 2026-10-24:

- `internal/cssanimation/bind.go:12-44` builds an `*animation.Player` from a
  display, a box, operation indices, base paints, and a start time.
- `internal/cssmotion/ownership.go:7` resolves owned operations through
  `box.OperationIDs` with a geometric fallback.
- `internal/page/page_tick.go:34-65` is the only per-frame callback. With a nil
  function it runs the caret blink and returns; the O(all ops) diff at
  `page_tick.go:75-90` only runs when a user callback exists.
- `internal/page/page_draw.go:174-185` (`commitDisplay`),
  `page_draw.go:206-224` (`paintFallback`), and
  `internal/page/patch_wire.go:127-136` (`spliceElement`) are the three places a
  display is published and `p.generation` moves.
- `internal/page/page_tick.go:122-152` is the existing single-shot frame
  deadline: `WantFrameAt`, `WantFrameIn`, `NextFrameAt`.
- `examples/animations/page.go:231-272` is the hand-written bind, `:199-218` the
  hand-written step loop, `:220-222` the hand-written damage.
- `examples/css-preview/bind.go:14-45` and `player.go:8-65` are the second
  hand-written driver, with a queue and generation rebind.

## Design

1. Add `internal/motion`, a sibling of `internal/page/dirty` and
   `internal/page/patch`. It imports `cssanimation`, `animation`, and
   `cssmotion`. It does not import `internal/page`, `internal/window`, or
   `internal/replay`. One package serves animations now and transitions in the
   transition track.

   ```go
   type Runtime struct {
       gen     uint64
       entries []entry
       stats   Stats
   }

   func New() *Runtime
   func (r *Runtime) Bind(d *Display) error
   func (r *Runtime) Step(d *Display, now time.Time) Damage
   func (r *Runtime) Active() bool
   func (r *Runtime) Stats() Stats
   ```

   `entry` holds the element id, the player, the captured target ops with a
   field mask, the original start time, and the animation record. `Damage` is a
   fixed-size rect set, mirroring `internal/page/dirty` rather than allocating a
   slice.

2. Bind at every publish, not only at `New`. `commitDisplay`, `paintFallback`,
   and `spliceElement` call one page helper, `p.bindMotion(display)`. A fallback
   page has no patchable list, so the helper clears the runtime there. The helper
   is the only bind point; `New`, `Load`, `PatchElement`, `SetTheme`, resize, and
   hot reload all reach it through those three functions.

3. Target selection moves out of examples. `cssmotion.OperationIndices(d, box)`
   and `cssmotion.Capture(d, box)` run at bind time. Store the op index and a
   field mask, not an untagged pointer. For this phase the mask is `all`, which
   preserves today's behaviour; phase 03 narrows it.

4. Honor the declared play state. `cssanimation.Bind` already starts only
   effects whose trigger is empty, but every example rewrites `PlayState` to
   `running` first (`examples/animations/page.go:275-289`,
   `examples/css-preview/bind.go:50-61`). Delete that rewrite. A paused sample
   stays paused until a class change or a trigger event says otherwise.

5. Generation guard belongs inside the runtime, matching `internal/frame`'s
   `Bind`/`stale` pair (`internal/frame/anim_gen.go:11-26`) and not the reverse.
   On the first `Step` after a generation change, rebind every entry and keep its
   original `start`. A caret blink calls `Redraw` (`internal/page/form_wire.go:362-380`),
   which bumps the generation, so this path runs on ordinary typing.

6. Tick order. `runTick` steps the runtime first, then handles the user
   callback:

   ```go
   gen := p.Generation()
   damage := p.motion.Step(p.draw.Display, time.Now())   // marks its own rects
   if fn == nil {
       err := p.caretBlink(ctx)
       p.flagTickDirty(gen)
       return err
   }
   snap := p.captureTickOps()                            // after the driver
   ...
   ```

   The snapshot moves after the driver so the O(all ops) diff never covers
   animation changes. The existing nil-function fast path stays as it is; the
   caret already marks its own damage by redrawing.

7. Pacing is the driver's job. While any player is active, call
   `p.WantFrameIn(0)`. When the last player finishes, do not ask. A waiting end
   delay asks with its remaining time. This replaces the `WantFrameIn` calls in
   the examples. The idle policy already treats a declared deadline as the page
   owning the schedule (`internal/window/pacing/policy.go:89-114`).

8. Default on. Add `Page.Window().SetAnimations(bool)`; zero value means on, so
   `Config{}` and every existing example behave correctly without a call. Store
   the switch beside perf and devtools on the inner page. When turned off, the
   runtime stops stepping and releases its players; the next bind does nothing.

9. Counters on `Probe()`: bound players, active players, samples this frame,
   rebinds, skipped properties, and last error. No logs on the hot path. A bind
   error is kept in the counter block once, not returned from the tick.

10. Do not touch `internal/frame`. It stays the explicit helper for Go-driven
    motion; this phase only documents the boundary in `documentation/frames.md`.

## Checklist

### A. Package and model

- [ ] Create `internal/motion/runtime.go`, `bind.go`, `step.go`, `target.go`,
      `damage.go`, `stats.go`.
- [ ] Define `Damage` as a fixed `[8]image.Rectangle` set with an overflow flag,
      matching `internal/page/dirty`.
- [ ] Define the target field mask as a `uint16` with bits for X, Y, matrix,
      opacity, RGB, and future fields.
- [ ] Resolve op indices at bind, not per frame.
- [ ] Keep a pointer plus generation check for the hot path, and re-resolve by
      index after a rebind.
- [ ] Unit test bind, step, damage, and rebind with a hand-built display.

### B. Page integration

- [ ] Add `bindMotion(display *Display)` and call it from `commitDisplay`,
      `paintFallback`, and `spliceElement`.
- [ ] Step the runtime in `runTick` before the user callback.
- [ ] Move `captureTickOps` after the driver step.
- [ ] Keep `flagTickDirty` behaviour: a driver that marks nothing and a caret
      that does nothing leave the page clean.
- [ ] Ask `WantFrameIn(0)` while active, `WantFrameIn(remaining)` while an end
      delay holds, and nothing when idle.
- [ ] Add `SetAnimations(bool)` and `Animations() bool` to the `Window` door.
- [ ] Add the counters to the `Probe` door.

### C. Behaviour

- [ ] Delete `running()` from both examples; a paused CSS sample stays paused.
- [ ] Rebind after `PatchElement`, hover, press, focus, resize, and reload with
      the original start time.
- [ ] A caret blink during an animation does not restart it.
- [ ] Turning animations off stops the tick request and releases players.
- [ ] A page with no animations pays one nil check per tick.

### D. Migration proof

- [ ] Rewrite `examples/animations/page.go` to `New`, `SetData`, `Redraw`,
      `Handle` only. No `SetTick`, no `Bind`, no `MarkRect`, no `WantFrameIn`.
- [ ] Compare frames before and after at the same timestamp with a dump, not by
      eye.
- [ ] Remove the three allowlist rows for `examples/animations/page.go`.
- [ ] Keep `examples/css-preview` on the manual path for this phase; the
      transition track and phase 09 migrate it after the auto path has parity.

### E. Tests and docs

- [ ] Test that a paused animation does not request a frame.
- [ ] Test that an animation with a finish time stops requesting frames.
- [ ] Test rebind-on-redraw keeps the phase (compare sampled paint before and
      after a caret redraw).
- [ ] Test `SetAnimations(false)` leaves the base paint untouched.
- [ ] Add the automatic behaviour to `documentation/css-motion.md` in phase 09;
      this phase only adds a TODO note in the file header.

Gate: the animations example has no motion code in Go, `make test` passes, and a
dump of one frame matches the previous driver at the same time.

## Risks

- Binding on every publish can thrash when a page patches often. Measure bind
  cost with the catalog page before landing; if it dominates, keep players whose
  element id and keyframe set did not change and rebind only the rest.
- The runtime must not hold pointers across a `spliceElement` that changes op
  count. The generation check is the guard; do not skip it for speed.
- `SetAnimations(false)` must not leave a player writing through a stale pointer
  after the switch. Clear the entries on the switch, not on the next tick.
