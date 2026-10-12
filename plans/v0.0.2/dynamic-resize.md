# Dynamic resize

Recorded 2026-10-03 against `d6026e0`.

Dragging a window edge today changes the size of the picture and almost nothing
else. The placement is computed once at a clamped frame size, the boxes the
pointer hit tests against belong to the previous layout, and the stretch path
takes over once the window passes `MaxWidth` or `MaxHeight`. This plan makes the
inner layout follow the window.

## What is static today

1. `internal/window/resize.go:5` lays the page out again on every distinct
   clamped size, and `internal/page/page_draw.go:13` rebuilds everything from
   the template string: `html.Parse`, `css.Apply`, then `layout.DisplayList` or
   `layout.Lay`. `documentation/frames.md:46` measures that at about 130 ms on
   the audio player example and 400 ms on the Spotify example. A drag emits a
   size change per frame, so the drag costs a full page render per frame.
2. `internal/page/page_size.go:29` pulls every size into `[min, max]`, and
   `internal/page/page.go:14` defaults both maxes to 2560. Past the cap the
   layout stops changing: `internal/window/view.go:7` reports `stretched` and
   `documentation/window.md:33` describes the picture scaling up to the window.
   The inner layout is frozen at the cap. This is the behaviour to remove.
3. `internal/window/game.go:61` runs pointer before resize. The hover and press
   hit tests therefore use `p.boxes` from the layout that was on screen before
   the drag, not the one the new size produces.
4. `internal/page/pointer_state.go:53` redraws the whole page when the hover id
   changes, because `:hover` is matched during the cascade. There is no cheaper
   path, and `../../PHASES.md:32` records that the engine matches the exact id
   only, with no ancestor hover.
5. `css.Apply` recollects every `<style>` sheet and re-merges the font registry
   on every call (engine `css/css.go:146`), so a size change pays for CSS work
   that the size did not affect. The viewport the cascade sees comes from
   `WidthPx` and `HeightPx` (`internal/render/state.go:21`), so media queries
   and viewport units do follow the size. They follow the clamped size.
6. `internal/window/scrollbar.go:21` derives the content size from the canvas
   plus every box that overflows it. After a relayout the scroll clamp is stale,
   and `documentation/scrolling.md:26` records that nothing re-clamps the offset
   when the page or window gets smaller.

## Target behaviour

| Case | Today | After |
|---|---|---|
| Drag inside the max | Full parse and repaint per frame, layout correct | Relayout per committed size, no reparse |
| Drag past `MaxWidth` | Picture scales, layout frozen | Layout follows the window; the cap becomes opt in |
| Text column | Rewraps, at the cost of a full render | Rewraps, cheaper |
| `@media (min-width: 700px)` | Matches the clamped size | Matches the window size |
| `100vw` bar | Matches the clamped size | Matches the window size |
| Hover while dragging | Resolved against the previous boxes | Re-resolved after the relayout, same frame |
| Scroll offset after shrink | Can point past the content | Clamped to the new content |

## Phase 1: split the pipeline

The gate for every other phase. A size change must not reparse the HTML or
recollect the stylesheets.

- [ ] Add a benchmark in `internal/page` that times `Redraw` stage by stage:
      template execute, `html.Parse`, `css.Apply`, `layout.DisplayListOptions`,
      and `layout.LayOptions`. Record the numbers in this file before changing
      anything.
- [ ] In `gowkhtmltopdf`, add a relayout entry point that takes an already
      styled document plus a new viewport and state, and returns a new
      placement without recollecting sheets. Suggested shape:
      `layout.RelayoutOptions(ctx, styled, layout.Viewport{WidthPx, HeightPx,
      State})`.
- [ ] Cache the executed template output and the parsed HTML tree in
      `internal/page`. Key it on the template name plus the data value, or on a
      hash of the executed bytes when the data is not comparable.
- [ ] Cache the parsed stylesheets, including the `Config.Theme` sheet, on the
      same source hash. `SetTheme` and `Load` invalidate it.
- [ ] Route `Redraw` through the cache. A relayout at a new size must not call
      `html.Parse` or the sheet collector.
- [ ] Count parses and cascades on the page under a test-only hook, so a test
      can assert that a resize does zero of each.
- [ ] Run `make test`. Every existing test must pass without a behaviour change.

Exit: a test that calls `SetSize` and `Redraw` ten times with unchanged data
reports one parse and one cascade.

## Phase 2: lay out at the real window size

- [ ] Decide the cap semantics and write the decision down here. Recommended:
      `MaxWidth` and `MaxHeight` stop clamping the layout and start clamping the
      window, so the layout always tracks the frame. `New` keeps validating that
      min is below max.
- [ ] Change `internal/page/page_size.go` so `Clamp` and `SetSize` enforce the
      minimum only. Keep the maximum available as a window bound.
- [ ] Have `internal/window/resize.go` pass the outside size through and let the
      shell cap the window instead of the page.
- [ ] Delete the stretch branch that fires because of the cap. `stretched` in
      `internal/window/view.go` should only be true while a relayout is in
      flight, which is Phase 3.
- [ ] Update `documentation/window.md` for the new cap semantics and remove the
      paragraph that says the frame stays at the max.

Exit: a window at 2000x1500 lays out at 2000x1500 with `MaxWidth` 800, and
`stretched()` is false.

## Phase 3: relayout once per committed size

A drag must not cost one full render per mouse event.

- [ ] Give `internal/window/resize.go` three fields: the pending size, the last
      laid-out size, and the time of the last relayout.
- [ ] While the size keeps changing, keep drawing the previous frame scaled to
      the window. No `Redraw`.
- [ ] Relayout when the size has been stable for one update, or at most every
      100 ms during motion, so text rewraps while the drag runs.
- [ ] Commit the last size at drag end even if the throttle window closed.
- [ ] Keep `internal/window/sync.go` as the only place that rebuilds the
      Ebiten image, and make it skip the rebuild when the generation and the
      size both match.
- [ ] Guard with a test: thirty `Layout` calls in a row followed by one settled
      update produce one `Redraw`, and the drawn size is the final size.

Exit: `TestResizeCommitsOncePerSettledSize` in `internal/window`.

## Phase 4: re-resolve state after a relayout

- [ ] Move `resize` before `pointer` in `Update` in `internal/window/game.go`,
      so the hit test runs against the boxes that match the frame being drawn.
- [ ] After a relayout, re-run `boxIDAt` for the last cursor position and
      clear a hover id that no longer exists in the new layout.
- [ ] Clear `p.active` the same way when the pressed element moves or vanishes.
- [ ] Redraw only when the re-resolved id differs from the stored one, so a
      relayout that does not move the cursor under anything costs no extra
      cascade.
- [ ] Add a test that resizes so a control moves out from under a fixed cursor
      and asserts the hover clears.

Exit: hover paints at the post-relayout position, and a stale hover id can
never survive a layout change.

## Phase 5: viewport-dependent values recompute

Every value that depends on the frame must be recalculated, not carried over.

- [ ] Verify with a test page that `@media (width)`, `@media (height)`, and
      `@media (orientation)` flip at the window size, not the old clamped size.
- [ ] Verify `vw`, `vh`, `dvh`, `svh`, and `lvh` against the window size.
- [ ] Verify percentage widths and `height: 100%` resolve against the new
      containing block after a relayout.
- [ ] Verify a `var()` chain that reads a viewport-dependent custom property
      recomputes. Record any that do not in this file as engine work.
- [ ] Verify that a `Config.Theme` or `SetTheme` sheet containing a media query
      re-evaluates on relayout.
- [ ] Verify `:hover`, `:active`, `:focus`, `:focus-visible`, and `:checked`
      still resolve after a relayout, since they ride in the cascade options.
- [ ] Land any engine gaps found above in the sibling checkout and note them in
      `../../PHASES.md`.

Exit: `examples/resize` shows a `@media` block, a `100vw` bar, a rewrapping text
column, and a hover control, and all four follow a drag.

## Phase 6: scroll and content size after a relayout

- [ ] Recompute the scroll clamp on every relayout. `contentSize` in
      `internal/window/scrollbar.go` already reads the new boxes; call it and
      pull `scrollX` and `scrollY` back into range.
- [ ] Redraw the thumbs on every relayout, so the bar tracks the new content.
- [ ] Close the gap `documentation/scrolling.md:26` records for navigation too:
      `Load`, `Back`, and `Forward` should clamp the offset as well.
- [ ] Add a test that scrolls to the bottom, shrinks the window, and asserts
      the offset is back inside the content.

Exit: no empty space below the content after any size change.

## Phase 7: the other hosts

- [ ] wasm: `scripts/browser.sh` builds the canvas. Confirm `Layout` receives
      the canvas size on resize and that the same path runs. Add a js build
      check to the test plan.
- [ ] mobile: `BindMobile` must survive rotation and a split-screen resize.
      Check that a relayout during a rotation does not fight the tick.
- [ ] `Serve` has no window. It lays out at `Page.Size()` and must document
      that a resize has no effect there, since `documentation/web.md` describes
      the picture only.
- [ ] The `bitmap fallback` badge must not flicker during a drag now that the
      relayout happens on commit. Check `internal/window/badge.go`.

Exit: the same page behaves the same on desktop, wasm, and mobile, and `Serve`
says so in its guide.

## Phase 8: docs and the example

- [ ] New example `examples/resize`: a two column page, a `@media (min-width)`
      switch, a `100vw` bar, a long paragraph that rewraps, and a hover control.
- [ ] Wire it into `examples/readme.md`.
- [ ] Rewrite the Resize and Fit sections of `documentation/window.md`.
- [ ] Add a note to `documentation/theming.md` that a theme's media queries
      follow the window.
- [ ] Note in `documentation/frames.md` that a relayout replaces the display
      list, so a tick holding an operation pointer must find it again. It
      already says this for `Redraw`; check the wording still fits.
- [ ] Add the new files to the map in `../../AGENTS.md`.
- [ ] Record the shipped items in `../../PHASES.md`.

## Risks and limits

- Phase 1 changes the engine's public surface. Keep the new call additive and
  leave `css.Apply` and `layout.Lay` alone so nothing else in gowkhtmltopdf
  moves.
- The cache in Phase 1 must invalidate on every path that changes the source:
  `SetData`, `SetTheme`, `SetImage`, `Load`, `Back`, `Forward`, and the form
  rewrite in `syncForm`. A missed invalidation shows as a stale page, so write
  the invalidation test before the cache.
- Engine hover matching is exact id only, so `:hover` on a parent of the
  hovered element still does not match (`../../PHASES.md:32`). That is engine
  work and out of scope here unless it blocks Phase 4.
- A page with an unsupported operation keeps the bitmap path, and that path
  re-rasterizes the whole page. Phase 3 makes it lay out less often; it does
  not make it cheaper per relayout.
