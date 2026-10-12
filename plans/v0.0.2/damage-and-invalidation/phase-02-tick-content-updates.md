# Phase 02, tick content updates

Goal: a board rotation retargets operations in the tick instead of calling
`Redraw`. This is the rAF-mutates-a-node pattern, not the React re-render shape.

## Today

`Advance` in `examples/departure-board/board/data.go` does `SetData` plus
`Redraw`: 5.7 to 13.5 ms of template execute, layout, and display-list build
once per 12 seconds. The diff then marks a full frame because about 17
operations changed, past `maxDirtyOps`, so one frame rebuilds the whole buffer.
The cascade flaps from the new operations after that.

Chrome's equivalent of this path is a script mutating a single text node inside
`requestAnimationFrame`: the mutation is known, no tree re-render happens, and
paint invalidation is local.

## Design

- The flipper already binds by cell id and retargets on a generation change.
  Extend it to accept a new rows slice without a generation change: set each
  flap's `want` from `rowsAt(feed)`, start the row band, invalidate the cells.
- Keep a resync rule: the next real `Redraw`, for a theme, night, or resize,
  rebuilds the display list from the view, so list and view agree again.
- `View` keeps the current rows for tests and for that resync.
- No engine work.

## Checklist

- [x] `data`: expose the target rows for one feed without a whole view rebuild.
- [x] `flip_rebind`: a retarget path that takes rows and marks changed cells,
      sharing the band calculation with `rebind`.
- [x] `tick`: rotation calls retarget and never `Advance`; `Advance` stays for
      tests and the dump scenario shape, or moves behind a helper.
- [x] Tests: rotation leaves `Page.Generation()` unchanged, flips the changed
      cells, and reports one band per row; `TestCascadeDirtyRectsStaySmall`
      covers the shape.
- [x] Dump: the `cascade` scenario reports no `full-buffer` frame after the
      rotation.
- [x] Docs: `documentation/frames.md` documents the pattern with the board as
      the example, next to the `UseFrameDirty` rule.

## Exit

- At real cadence, the board's only full-content frames are theme, night, and
  resize. Measured by the dump, `RedrawNS` per rotation is zero and the paced
  cascade row shows no `full-buffer` kind.
- `make test` passes.

## Risks

- The display list and the view can drift if retargeting runs for a long time
  without a `Redraw`. Cover it with a test that a `Redraw` after several
  rotations still matches the view.
- A resize during a cascade replaces the list. The generation check in the
  tick already rebinds; keep the target table from being overwritten by the
  resize.
