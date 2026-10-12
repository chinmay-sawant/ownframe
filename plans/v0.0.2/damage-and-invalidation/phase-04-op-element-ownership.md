# Phase 04, operations carry their element

Goal: ownframe can answer "which operations belong to element X" without a
geometric search, so `Invalidate` maps to a tight op range and the id path does
not need `growBox`.

Where it lives: blinkless, small and additive. This is the one engine change
phases 1 to 3 may eventually want.

## Today

- `layout.Box` carries the DOM id, which is how hit testing and
  `Invalidate` find element geometry today.
- `layout.DisplayOp.ID` is an internal op sequence number set in the emitter
  (`internal/layout/layout.go:921`, `paintOp.ID = e.nextOpID`). It is op
  identity, not an element link.
- `internal/frame` finds operations by geometry and colour, and
  `page_invalidate.go` grows a box by scanning `p.boxes`.

Chrome has this link: a paint invalidation names a display item client, and
the paint chunks and display items carry which client produced them.

## Design

- blinkless: add an element reference on `Op` where the emitter already knows
  the current element, and expose it on `layout.DisplayOp`. Keep `Op.ID` as
  the op sequence.
- ownframe: build an op index by element at `Redraw`. `Invalidate(id)` maps to
  the element's op range and dirties its bounds; the automatic diff can drop
  `growBox` for the id path.
- Keep the geometry fallback for callers that only have a box, and for pages
  where the engine did not tag an op.

## Checklist

- [ ] blinkless: element reference on `Op`, set in the emitter, with a test
      that a nested span reports both elements on its text op.
- [ ] blinkless: push the branch, then bump the pin in `go.mod` and
      `examples/go.mod`. AGENTS.md: the blinkless require is the only engine
      pin and there is no replace.
- [ ] ownframe: op index per generation; `Invalidate` uses it; tests for
      nested and same-position elements.
- [ ] ownframe: measure the id path with a new benchmark before and after.
- [ ] Docs: `documentation/repaint.md` and the AGENTS.md pin note.

## Exit

- `Invalidate(id)` on two overlapping elements dirties only the right one and
  performs no box scan.
- The from-scratch `DisplayList` path and its tests are unchanged, and a page
  without tagged ops still invalidates correctly through geometry.
- Engine and ownframe tests pass with the new pin.

## Risks

- The engine pin move touches both modules. Land phases 1 to 3 first so the
  bump is not carrying unrelated fixes.
- An element reference per op adds a string per operation. Measure the alloc
  and memory cost on the large example before landing it.
