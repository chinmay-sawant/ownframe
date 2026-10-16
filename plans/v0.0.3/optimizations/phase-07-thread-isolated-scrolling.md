# Phase 07 - Thread-isolated scrolling

> **Parent:** `plans/v0.0.3/optimizations/README.md` - Chromium optimization adoption plan
> **Status:** Not started. Optional for ownframe. No committed implementation evidence.
> **Estimated effort:** Large, only justified after a measured input-to-paint stall.

---

## Overview

This phase studies Chrome's thread-isolated scrolling against ownframe's Ebiten
contract. The goal is not to copy Chromium's compositor thread blindly. The
goal is to define an input path that can move a scroll viewport without
waiting for the template, layout, and paint pipeline to finish, while staying
compatible with Ebiten's callback model and the page's transactional state.

Chrome can scroll on the compositor thread because a property tree carries the
scroll offset, the compositor records stable tiles, and the main thread has
already committed layout. ownframe currently has none of those layers. The
page owns the offset, the window drives the Ebiten callbacks, and the frame
reads from the game goroutine.

## Executive summary

The current path is simple and correct but synchronous:

- `internal/window/update.go` calls `Update()` on the game loop and stages
  pointer, touch, keys, and wheel.
- `internal/host/scrollwindow.go` defines the window-owned offset contract.
- `internal/page/page_windowing.go` stores `ScrollOffset` and
  `SetScrollOffset`.
- `internal/window/scroll_window.go` applies the offset at replay time.

There is no fast scroll path. There is no compositor thread. There is no
non-passive listener concept, because there is no JavaScript.

What this phase can still deliver is narrower and safer: a scroll owner that
updates one viewport offset without holding engine work, plus a pending or
active split for tiles when the display list is unchanged. Any motion beyond
that should keep the normal frame path.

## Why this optimization is required

Chrome needs thread-isolated scrolling because input arrives faster than the
main thread. The user expectation is 120 Hz touch motion, while a document
parse or a long paint can take more than 8 ms. Scrolling from the main
thread causes dropped frames.

ownframe has the same class of risk, but with different evidence requirements:

- Touch and trackpad paths feed continuous motion.
- Ebiten still calls `Update` and `Draw` on the game goroutine.
- A page that repaints or re-layouts during wheel events can stall motion.
- The devtools dock, animation, and form caret can widen an already costly
  frame.
- A tile-based replay path can replay much earlier, but only if the offset
  changes independently from content.

The proposed value is not speculative background threading. It is a scoped
contract where a scroll viewport changes the tile origin while layout and
paint stay on the next frame. That is useful even on one goroutine.

## Current state

### Thread model

Evidence from the current tree:

- `internal/window/idle_game.go:51` starts the only non-host goroutine,
  `go pace.timer(ctx)`, and its job is scheduling, not rendering.
- `internal/window/update.go` performs all input staging on the game
  goroutine.
- `internal/render/engine.go:6-7` protects process-wide engine readers with
  `engineMu`.
- `internal/ipc/ipc.go:1-7` states that IPC is in-process and has no second
  process.

Therefore any new thread must not call the engine outside the existing lock.
Recommendation: keep all document work on the game goroutine and experiment
only with viewport and tile state.

### Scroll ownership

- `internal/host/scrollwindow.go` states that the window owns the offset.
- `internal/page/page_windowing.go` stores and exposes the scroll offset.
- `internal/window/scroll_window.go` consumes the offset at replay time.

This is already the right boundary. ownframe does not have a Chrome
compositor offset; it has a window-owned offset applied to replay.

### Replay tile state

The partial replay path gives a candidate acceleration layer:

- `internal/window/replay_tile.go:14` sets `tileSize = 256`.
- `internal/window/replay_tile_paint.go` repaints only touched tiles.
- `internal/window/replay_tile_copy.go` blits tile images.
- `internal/window/replay_partial_state.go` tracks `scroll image.Point`
  and `last` rects.

If scroll motion can be expressed as a change to the tile origin, the frame
can skip both layout and paint. If tiles must be repainted, no isolation
benefit remains.

## Design

### Phase 07A: measure before changing behavior

No input path change lands without a baseline. Measure:

1. Input-to-visible scroll latency during:
   - plain content
   - a form-heavy page
   - an active animation
   - the devtools dock
2. Wheel event cadence against frame cadence.
3. Whether `Update` already finishes before `Draw`.
4. Which operations, if any, scale with scroll offset.

Output must be a markdown table in
`plans/v0.0.3/optimizations` or another ledger, with the command, commit,
and hardware or package details.

### Phase 07B: scroll input cache

Introduce an explicit viewport scroll delta in `internal/window` that can be
updated without touching `internal/page`.

Proposed ownership:

- `internal/window/scroll_input.go` stores pending wheel and gesture deltas.
- `internal/window/update.go` queues deltas only.
- `internal/window/scroll_window.go` applies the newest delta in a clamp.
- `internal/page` is informed after the frame, not before.

This single change separates scroll input from early layout gates.

### Phase 07C: offset-driven replay

Make `internal/window/replay_partial_state.go` decide that scroll motion
alone can replay tiles from the retained buffer while the page offset advances.

Rules:

- A scroll-only frame changes only `scroll`.
- The content display list and generation remain unchanged.
- `paints` and `paintPx` stay zero for the pure-scroll case.
- `blits` records moved tiles.
- If a scroll delta exceeds tile coverage, fall back to the normal plan.

### Phase 07D: change detection on the page

Add `internal/page` metadata that reports whether a candidate update can skip
layout:

- template content changed
- data changed
- focused or hovered ids changed
- style state changed
- animation tick changed
- scroll offset only

`Page` already has the machinery for patch invalidation. Reuse those
signals rather than add a second invalidation vocabulary.

### Phase 07E: optional fast path in the game loop

Only if phases 07A through 07D show a miss, call `Redraw` skip a step at a
time. The checked version might look like:

- wheel produces one queued delta
- `Update` calls a scroll-only replay
- `Draw` blits to the screen at the new origin
- the page commits the offset later without another paint

This is still not Chromium-style compositor scrolling, but it is the
highest-value, lowest-risk version for a single-goroutine engine.

## Checklist

### Measurement

- [ ] Record wheel and touch baselines on three representative pages.
- [ ] Measure input latency with devtools open and with it closed.
- [ ] Measure frame cost for pure scroll after a tile repaint versus blit.
- [ ] Record hardware, PGO, environment, and command lines beside results.

### Contract

- [ ] Define whether a scroll-only update can bypass `Redraw`.
- [ ] Document constraints in `internal/host/scrollwindow.go` and
  `internal/page/page_windowing.go`.
- [ ] Define what `Page.Generation()` means during scroll.
- [ ] Define how forms, focus, caret, and IME react to an offset change.

### Input and engine

- [ ] Add `internal/window/scroll_input.go`.
- [ ] Keep `internal/render/engineMu` usage unchanged.
- [ ] Do not call `css.Relayout` on a scroll-only frame.
- [ ] Do not call `DisplayListDocument` on a scroll-only frame.

### Tile replay

- [ ] Add scroll-only detection to `replay_partial_state.go`.
- [ ] Add a tile-origin replay branch with tests.
- [ ] Keep viewport and overscan behavior documented in `replay_viewport.go`.
- [ ] Keep devtools dock bytes excluded from the content tiles.

### Page integration

- [ ] Add a read-only `Page` API for scroll-only eligibility.
- [ ] Patch the scroll offset without changing content generation.
- [ ] Add tests for IME composition during scrolling.
- [ ] Add tests for focus changes during scrolling.

### Compatibility

- [ ] Keep Examples compiling unchanged.
- [ ] Keep `TakeDirty` semantics unchanged for content updates.
- [ ] Keep WebAssembly behavior identical with single-threaded input.
- [ ] Keep Android and desktop viewport pinch behavior intact.

## Exit criteria

- A wheel burst on a 2000 operation card list maintains the current FPS
  target.
- `Update` input-to-Draw latency stays below the measured baseline.
- A scroll-only frame records zero `paints`.
- `Page.Generation()` does not change on a scroll-only frame.
- Non-scroll content updates still use the existing invalidation path.
- `make test` passes and `go vet ./...` passes.

## Risks

- Ebiten guarantees where callbacks run. The plan assumes the current
  `Update` and `Draw` model; moving work elsewhere requires rechecking that.
- Scroll-only detection must not hide real content changes.
- Fast scrolling can reveal tile gaps if overscan is too small.
- IME preedit and focus need strict ordering; `SetIMEComposition` must keep
  its current contract.
- A stale offset between game loop stages can produce flicker if only
  partially consumed.

## Dependencies

- Depends on phase 02 because tile reuse is the only acceleration available.
- Depends on phase 05 because damage tracking provides the fallback when
  scrolling must dirty content.
- Depends on phase 03 only if the scroll path needs a second buffer for
  double input capture.

## Open questions

- Can a scroll-only frame avoid all page calls, or must the page still commit
  a viewport offset?
- Should input deltas live in `internal/window` or `internal/page`?
- Will a tile-only scroll expose gaps with 256 px tiles on high-DPI panels?
- Can `internal/window/viewport` unify pure scroll with regioned painting
  without duplicating state?

Do not implement this phase before its measurements are recorded. If the
measurement shows no stall, close the phase as deferred with the same reason
and no code diff.
