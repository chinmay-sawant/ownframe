# Damage list and invalidation

Recorded 2026-10-12 against the working tree on `perf/departure-board`, after
the departure-board frame-dirty fix.

The predecessor is [incremental-repaint.md](../incremental-repaint.md), which
shipped the first partial path: `Invalidate`, the automatic display-list diff,
one dirty rect, and the persistent buffer. That plan closed with a click
repainting one box. This folder starts where that one stopped.

## Why now

The departure board fix made the tick opt into the frame-dirty path, and the
draw dropped from a 23 ms full replay of 631 to 640 operations every frame to
under a millisecond on most frames. Two costs remain, both measured with
`examples/departure-board/dump` at `-pace 16ms`:

- The window takes exactly one dirty rect per frame and the page unions every
  claim into it. When the refresh bar moves while a cascade band is active, the
  rect spans the bar down through the band and the frame replays up to 375
  operations. A bar update lands on a band frame about once in six.
- A rotation calls `Redraw`, 5.7 to 13.5 ms of page work once per 12 seconds,
  plus one full-content rebuild frame. That is the React-shaped path: re-render
  the page to change 17 cells.

The devtools dock inflates the numbers while the user reads them. The audit
estimated the Ops tab at 0.5 to 1.5 ms plus 200 to 400 KB per frame while open,
and several ms more with paint outlines on.

## What Chrome does

Chromium's [How cc Works](https://chromium.googlesource.com/chromium/src/+/main/docs/how_cc_works.md)
names three caches with three granularities:

1. Paint invalidation in Blink. Content is flagged during style and layout and
   the flags resolve in the pre-paint walk, per the
   [BlinkNG deep dive](https://developer.chrome.com/docs/chromium/blinkng).
2. Raster invalidation in cc. Painted content is cut into tiles, and only tiles
   covered by invalidated content are re-rastered.
3. Damage in the display compositor, which the doc defines as "the part of the
   screen that needs to be redrawn", tracked as a list and used for partial
   swap when the platform allows it.

Chrome also runs transform, opacity, and filter animations on the compositor
thread without touching page paint, and aborts frames that did no work.

[React's render and commit](https://react.dev/learn/render-and-commit) is the
application-layer version of the same idea: render computes the tree, commit
applies minimal DOM mutations, and the browser's invalidation does the rest.

## What ownframe already has

- Element invalidation by id: `Page.Invalidate` with the previous-rect
  bookkeeping in [page_invalidate.go](../../../internal/page/page_invalidate.go).
- Automatic invalidation: `diffDisplay` compares two display lists in
  [page_dirty_diff.go](../../../internal/page/page_dirty_diff.go) and falls back
  to the frame past 8 changed ops or a third of the area.
- One rect per frame, unioned in
  [page_dirty_box.go](../../../internal/page/page_dirty_box.go).
- The persistent buffer and the four frame plans in
  [replay_partial.go](../../../internal/window/replay_partial.go) and
  [replay_plan.go](../../../internal/window/replay_plan.go).
- The per-rect replay primitive, `replay.DrawRect`.

The engine side is a pure function by design. blinkless parses, cascades, lays
out, and returns `Display{Ops, Order, Boxes}` with no dirty state anywhere in
its non-test source. `layout.Box` carries the DOM id; `layout.DisplayOp.ID` is
an internal op sequence (`internal/layout/layout.go:921`), not an element link.
That is why the boundary below puts the damage work in ownframe.

## Phases

| Phase | Where | Job |
|---|---|---|
| [phase-01-damage-list.md](phase-01-damage-list.md) | ownframe | Replace the union rect with a capped list, cc's damage list in miniature |
| [phase-02-tick-content-updates.md](phase-02-tick-content-updates.md) | ownframe | A rotation retargets operations instead of calling `Redraw` |
| [phase-03-invalidation-cost.md](phase-03-invalidation-cost.md) | ownframe | Cheap marking for many elements, cheap hover, and a devtools dock that does not distort the reading |
| [phase-04-op-element-ownership.md](phase-04-op-element-ownership.md) | blinkless, small | Operations carry the element they came from, so invalidation needs no geometry |
| [phase-05-incremental-layout.md](phase-05-incremental-layout.md) | blinkless, large | Incremental style and layout, gated on a measurement that asks for it |
| [phase-06-docs-and-tracking.md](phase-06-docs-and-tracking.md) | docs | Repaint docs, performance docs, baseline numbers, `PHASES.md` |

Phases 1 to 3 need no engine change. Phase 4 is additive and small. Phase 5 is
the Blink-sized project and stays closed until phases 1 to 4 have numbers.

Order: phase 1 first, because every later phase measures against its numbers.
Phase 2 is independent of it. Phase 3 follows phase 1 so the dock and marking
numbers are read on the final shape. Phase 4 waits for phase 1. Phase 5 waits
for its gate. The evidence behind phase 3 is kept in
[audit-2026-10-12.md](audit-2026-10-12.md).

## Ground rules

- Every Go file stays at or under 2000 characters, counted with `wc -m`. Split
  inside the same package when a file would go over.
- `gofmt` every file you edit, then run `make test` from the repo root. Use
  `make test TEST_P=4` when a faster run is worth the load.
- No new modules. Engine work lands in the blinkless checkout first, is pushed,
  and then the pin moves in `go.mod` and `examples/go.mod`. AGENTS.md: the
  blinkless require is the only engine pin and there is no replace.
- No optimization lands without a number that asks for it
  ([performance.md](../../../documentation/performance.md), section 8). Every
  phase records a before and after with the same command.
- The dump is the measurement tool for this work:
  `go run ./examples/departure-board -dump -pace 16ms`. Compare like for like;
  the fast mode cannot see wall-clock driven updates.
- Keep `TakeDirty()` working. Examples call it, so the list arrives as a new
  method and the old one keeps returning the bounding box.
- Markdown has no character cap. Go does, so plan small files in
  `internal/page` and `internal/window`.

## Definition of done

- Two separated changes in one frame produce two rects, and the board's worst
  cascade frame replays under 60 approximate operations, measured by the dump.
- A board rotation leaves `Page.Generation()` unchanged, flips the changed
  cells, and causes no full-content rebuild frame.
- The idle board holds 60 FPS with the devtools dock open on the Ops tab, and
  the dock allocates nothing per frame while the generation is still.
- `Invalidate(id)` maps to ops without a box scan once phase 4 lands, and the
  engine's from-scratch `DisplayList` path is unchanged.
- `documentation/repaint.md`, `documentation/performance.md`, and
  `examples/departure-board/PERFORMANCE.md` describe the shipped behavior, and
  `PHASES.md` records it with commit hashes.
