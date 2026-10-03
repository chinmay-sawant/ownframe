# Incremental repaint

Recorded 2026-10-03 against `d6026e0`.

Click the counter in `examples/platform` and the whole page is rebuilt to show
one new digit. `internal/page/page_click.go:40` ends every click with an
unconditional `Redraw`, and `Redraw` starts from the template string
(`internal/page/page_draw.go:13`). This plan makes a click repaint the box that
changed and leave the rest of the frame alone.

## What happens today

1. `Click`, `Hover`, `Press`, and `Release` each end in a full `Redraw`
   (`internal/page/page_click.go:40`, and the three at
   `internal/page/pointer_state.go:28`, `:43`, `:53`). The typing and editing
   handlers do the same.
2. `Redraw` re-executes the template, calls `html.Parse`, recollects every
   stylesheet and re-merges the font registry (engine `css/css.go:146`), then
   relays out the whole page. On the replay path it also builds a new display
   list, so every operation object a tick was holding is replaced.
3. `internal/window/draw.go:20` replays the entire list every frame regardless.
   The window keeps a retained buffer only on the stretched path
   (`internal/window/replay_scale.go:18`), so there is nowhere to blit a
   partial update from.
4. The engine has no element tag on an operation. `layout.DisplayOp.ID` is a
   `uint64` operation identity, not the DOM id (engine
   `layout/layout.go:439`). Mapping an element id to its operations today means
   the geometric search in `internal/frame`, which finds a fill by colour inside
   a box (`internal/frame/frame.go:22`).
5. The manual version of this feature already ships. `internal/frame` finds an
   operation and a caller sets `Text` on it, which is how the audio player moves
   a seek bar without a `Redraw` (`documentation/frames.md`). The work here is
   to make the library do it for content changes, not just for ticks.

## Three levels of partial, pick in this order

| Level | What it skips | Engine work | Ships in |
|---|---|---|---|
| A. Stage level | Parse and cascade when only state or data changed | One additive call | Phase 1 |
| B. Region level | Repainting every operation outside the changed box | None | Phase 2 to 4 |
| C. Subtree level | Laying out the untouched branches again | Engine work, hard | Stretch goal |

Level A is the cheapest win and unblocks the drag work in
[dynamic-resize.md](dynamic-resize.md). Level B is what the counter needs. Level
C is not worth it until A and B are measured and fast.

## Phase 1: stage-level caching on the click path

Same work as dynamic-resize Phase 1, from the click's side. A click that only
bumps a number must not reparse or recollect.

- [ ] Cache the executed template output and the parsed tree, keyed on the
      source bytes.
- [ ] Cache the parsed stylesheets and the theme sheet, keyed on the same
      bytes. `SetTheme`, `SetImage`, `Load`, `Back`, `Forward`, and the form
      rewrite in `syncForm` invalidate.
- [ ] On the click path, reuse the parse and the cascade, and go straight to
      the relayout call from dynamic-resize Phase 1.
- [ ] Add a counter for parses, cascades, layouts, and repaints on the page,
      behind a test-only file so the public API stays clean.
- [ ] Benchmark `examples/platform` click to visible frame before and after.

Exit: one click reports zero parses and one layout, and the click benchmark is
faster than the recorded baseline.

## Phase 2: declare what changed

- [ ] Add `Page.Invalidate(id string)`. A handler calls it for the ids it
      changed. An empty id, or no call at all, means the whole page, which keeps
      every existing example correct without an edit.
- [ ] Keep it a `Page` method. The root package re-exports page calls in a
      small file the way `fetch.go` and `ipc.go` do, so follow that only if the
      signature needs a package-level helper.
- [ ] Look the id up in `p.boxes`. A box that is not found dirties everything,
      because a missing box means the element moved or appeared.
- [ ] Grow the rect to cover the element's descendants, using the boxes that
      contain the id's box, so a counter with a nested span repaints whole.
- [ ] Union with the previous rect for the same id, so text that grew or shrank
      repaints the space it used and the space it now uses.
- [ ] Pad the rect by a few pixels for shadows, outlines, and antialiasing.
- [ ] Write the invalidation test before the rect maths. Every path that
      changes the source must invalidate.

Exit: `Page.Invalidate("count")` on the platform example produces a rect that
covers `#count` and nothing else.

## Phase 3: find the dirty rect without asking

Handlers will forget to declare, so the page derives the rect from the two
display lists it already holds.

- [ ] Diff the previous and current `layout.Display` operation by operation,
      matching on index and, when the count differs, on position.
- [ ] Treat a changed paint field, a changed geometry, a changed text, a
      changed font, and an added or removed operation as a change.
- [ ] Union the bounds of every changed operation. An operation removed from
      the list dirties its old bounds.
- [ ] If more than a third of the page changed, or the diff found more than a
      fixed number of separate rects, fall back to a full repaint. Merging many
      small rects is slower than redrawing.
- [ ] Only diff on the replay path. A fallback page has no display list.
- [ ] Compare the diffed rect against the declared one and use whichever is
      larger, so a wrong `Invalidate` cannot produce a wrong picture.

Exit: a click with no `Invalidate` call still repaints only the counter's box.

## Phase 4: replay only the dirty rect

- [ ] Add `replay.DrawRect(dst, display, rect, dx, dy)` beside the existing
      `replay.Draw`. It walks `display.Order` as `Draw` does and skips any op
      whose bounds do not intersect the rect.
- [ ] Clamp the rect to the canvas and convert CSS pixels to points with
      `display.PixelPerPoint` before the comparison. A text op carries its
      baseline in `Y`, so test the rect against the line box built from
      `op.Font.Ascent`, `op.Size`, and `op.InkDescent`, not against the centre.
      `internal/frame/text.go:22` already does the baseline variant of this.
- [ ] Give the shell a persistent frame buffer for the replay path, the one
      `replay_scale.go` builds on the stretched path, and blit the dirty rect
      from it into the screen each frame.
- [ ] Repaint into that buffer when the generation or the scroll offset changes,
      not only when the dirty rect changes.
- [ ] Keep the scrollbar thumbs and the `bitmap fallback` badge outside the
      partial path. They draw on the screen, not in the page buffer, and a
      repaint must not wipe them.
- [ ] Fall back to the current full replay whenever the rect is empty, the page
      changed size, or the page fell back to the bitmap path.

Exit: `TestClickRepaintsOneBox` walks the platform page, clicks `#inc`, and
asserts that every operation outside `#count` is unchanged.

## Phase 5: the bitmap fallback

- [ ] Decide and record here whether the fallback gets a partial path. The
      engine's `imageout.RenderLayout` paints the whole canvas, so a partial
      bitmap repaint needs a new engine call.
- [ ] If yes: add `render.PaintRegion(ctx, source, width, height, rect,
      state)` in the sibling checkout, backed by an `imageout` call that clips
      to a device-space rect. It must return a picture of the rect size, and its
      content must match the full paint byte for byte.
- [ ] If no: a fallback page keeps a full repaint, and this is documented as a
      limit next to the other replay fallbacks in `../../PHASES.md`.
- [ ] Whatever the answer, `render.Replayable` keeps its current meaning. Do not
      let a page switch paths because it took the fast route.

Exit: the fallback decision is written down, and either the engine call exists
with a test or the limit is documented.

## Phase 6: wire the examples

- [ ] `examples/platform`: call `page.Invalidate("count")` from `onClick` in
      `examples/platform/platform/platform_new.go`. Keep the file under 2000
      characters, so the call may belong in a new file in that package.
- [ ] Extend the platform page with a hover control and a checkbox so one
      example shows the content path, the hover path, and the `:checked` path.
- [ ] Add a test in that package that clicks `#inc` twice and asserts the repaint
      rect grew only inside `#count`.
- [ ] Check `examples/states`, `examples/forms`, and `examples/code-editor` for
      handlers that change one field and mark it, and leave the rest alone.
- [ ] Note the new call in `examples/readme.md`.

Exit: the counter repaints locally on screen and in the test.

## Phase 7: state changes and edits

- [ ] A hover or press change dirties the union of the old and new element
      boxes, taken from `p.boxes` before and after. The cascade still runs, but
      only those two rects repaint.
- [ ] A focus change dirties the field box plus the element that had focus, and
      the same for a blur.
- [ ] Typing dirties the field and the caret run. A field whose text grew
      dirties the old and new boxes, so a wrapped line repaints fully.
- [ ] A checkbox or radio toggle dirties the control and its label.
- [ ] `SetTheme` dirties everything. A theme can restyle any element, so there
      is no partial path for it.
- [ ] A `Load`, `Back`, or `Forward` dirties everything and clamps the scroll
      offset (dynamic-resize Phase 6).
- [ ] A tick that changes operations directly stays as it is. It already skips
      the cascade, and it must not start dirtying rects.

Exit: hover and typing repaint locally, and a theme change still repaints the
whole page.

## Phase 8: prove the two paths agree

- [ ] Render each example twice, once with incremental repaints and once with a
      full `Redraw`, and compare the PNG bytes. Pages: login, platform, states,
      forms, scrolling, theme.
- [ ] Fail the test when they differ. This is the guard that keeps a partial
      repaint from being a subtly wrong picture.
- [ ] Add a test that a repaint rect outside the canvas is clamped, not dropped.
- [ ] Add a test that an op removed from the list dirties its old bounds, so a
      vanishing element does not leave a ghost.
- [ ] Add a test for the repaint-count budget: one counter click causes one
      layout and one repaint of one rect.

Exit: the byte comparison passes on every example, and the counter click is
one rect.

## Phase 9: docs

- [ ] `documentation/frames.md`: put the automatic partial path next to the
      manual `internal/frame` path and say which one a page should reach for.
- [ ] `documentation/features.md`: the click no longer always redraws the whole
      page, and the fallback page still does.
- [ ] `documentation/forms.md` and `documentation/editing.md`: what typing and a
      toggle dirty.
- [ ] `documentation/pointer.md`: a hover change dirties two boxes.
- [ ] New `documentation/repaint.md` with the call shapes and the measured
      numbers, linked from `documentation/README.md`.
- [ ] Record what shipped in `../../PHASES.md` and add the new files to the map
      in `../../AGENTS.md`.

## Risks and limits

- An op-level element tag is what level C needs, and the engine does not have
  one. If the geometric mapping in Phase 3 turns out to be wrong for a nested
  or absolutely positioned element, ask the engine for an `ElementID` on each op
  rather than growing a bigger heuristic here.
- Partial repaint is only safe where the page is deterministic. A page that
  fetches during layout, or an example with a running clock, can change pixels
  the diff does not know about. Keep `Invalidate()` with no argument as the
  escape hatch and document it.
- `Page.PNG` paints from `p.source` on demand (`internal/page/page_image.go:42`),
  so `GET /frame.png` stays correct and simply costs a full paint per request.
  Do not try to make it incremental in this version.
- Every new file in `internal/page` and `internal/replay` has to stay at or
  under 2000 characters, so plan for small files: `page_invalidate.go`,
  `page_dirty.go`, `replay/draw_rect.go`.
