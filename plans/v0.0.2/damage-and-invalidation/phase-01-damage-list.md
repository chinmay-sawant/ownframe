# Phase 01, damage list

Goal: one dirty rect becomes a small capped list, so two separated changes
repaint separately instead of unioning into their bounding box. This is cc's
damage list in miniature.

## Where it lives

All ownframe, no engine work:

- `internal/page/page_dirty.go`: `TakeDirty`, the accumulation, the counters.
- `internal/page/page_dirty_box.go`: `markRect` unions into `p.dirty`.
- `internal/page/page_invalidate.go`: the id path that feeds `markRect`.
- `internal/page/page_dirty_diff.go`: the automatic path with `maxDirtyOps`
  and `capDirty`.
- `internal/window/replay_partial.go`: `dirtyTaker`, one rect into
  `planRepaint`.
- `internal/window/replay_plan.go`: `planRepaint` and the four modes.
- `internal/window/replay_buffer.go`: `applyRepaint`, one rect into the buffer.
- `internal/window/replay_viewport.go`: the oversized path and its own single
  cached image.

## Design

- `p.dirty image.Rectangle` becomes `p.dirtyRects []image.Rectangle`, plus the
  existing `dirtyFull`.
- Insert with coalescing: merge a new rect into an existing one when they touch
  within a gap, the same spirit as `touches` in `page_dirty_box.go`.
- Cap the list at 8 rects, the same order as `maxDirtyOps`. Past the cap, or
  when total coverage passes a third of the content, `markFull`. Chrome caps
  its damage the same way.
- `TakeDirty()` keeps returning the bounding box, so the download-manager,
  perf-complex, and platform callers keep compiling. New:
  `TakeDirtyRects() ([]image.Rectangle, bool)`. The window prefers the list; a
  screen that only implements `TakeDirty` keeps the single-rect path.
- Plan per rect: any rect equal to the full content becomes one
  `repaintBuffer`, an empty list a `repaintBlit`, otherwise one `repaintRect`
  per rect against the same buffer.
- `replay.DrawRect` already repaints one rect; the buffer fills the sub-image
  and replays. Nothing new is needed in `internal/replay`.
- The oversized viewport path keeps its single cached image. Map the list
  through the visible clamp and write down in the plan what happens when a rect
  crosses the cached area.

## Checklist

- [ ] Write the failing test first: change two far-apart fields in one frame
      and assert two rects whose areas sum below the bounding box.
- [ ] `page`: list state, coalescing, cap, and fallback across
      `page_dirty.go`, `page_dirty_box.go`, and a new file for the cap if the
      2000 character limit asks for one.
- [ ] `page`: `TakeDirtyRects`, with `TakeDirty` as the bounding box wrapper.
- [ ] `window`: `dirtyTaker` gains the list method; `replay_partial.go`
      consumes it.
- [ ] `window`: `planRepaint` returns a list of plans or one plan with a list;
      `replay_buffer.go` loops with per-rect sub-images.
- [ ] `window`: `replay_viewport.go` decision for the list, recorded in a
      comment next to the code.
- [ ] Tests: page coalescing and cap; window plans for two rects, one rect,
      empty, and overflow; a per-rect `DrawRect` unit test in `internal/replay`.
- [ ] Dump: rerun `go run ./examples/departure-board -dump -pace 16ms` and
      record the worst-frame and mean numbers in
      `examples/departure-board/PERFORMANCE.md`.
- [ ] Docs: `documentation/repaint.md` gains the list shape and the cap rule.

## Exit

- Two separated changes produce two rects; the sum of their areas is below the
  bounding box; an empty list still blits.
- The board's worst cascade frame replays under 60 approximate operations,
  measured by the dump.
- `make test` passes, and the existing one-rect tests pass through `TakeDirty`
  unchanged.

## Risks

- Each rect costs one op-list scan in `DrawRect`, about 12 ns per op per
  `BenchmarkDrawRectScan`. Eight rects on a 2000 op page stays under 0.2 ms.
  Cap the count and fall back to full when the scan cost would beat a replay.
- Tests and examples that assert one rect must move to the list shape. Keep
  `TakeDirty` for them rather than editing every caller.
