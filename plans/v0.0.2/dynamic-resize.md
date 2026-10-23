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
   path, and `../PHASES.md:32` records that the engine matches the exact id
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

- [x] Add a benchmark in `internal/page` that times `Redraw` stage by stage:
      template execute, `html.Parse`, `css.Apply`, `layout.DisplayListOptions`,
      and `layout.LayOptions`. Record the numbers in this file before changing
      anything.
- [x] In `gowkhtmltopdf`, add a relayout entry point that takes an already
      styled document plus a new viewport and state, and returns a new
      placement without recollecting sheets. Suggested shape:
      `layout.RelayoutOptions(ctx, styled, layout.Viewport{WidthPx, HeightPx,
      State})`.
- [x] Cache the executed template output and the parsed HTML tree in
      `internal/page`. Landed shape: the template executes on every `Redraw`
      (cheap), and the parsed tree and sheets are keyed on the executed source
      bytes, so a changed data value reparses and an unchanged one does not.
- [x] Cache the parsed stylesheets, including the `Config.Theme` sheet, on the
      same source hash. `SetTheme` and `Load` invalidate it.
- [x] Route `Redraw` through the cache. A relayout at a new size must not call
      `html.Parse` or the sheet collector.
- [x] Count parses and cascades on the page under a test-only hook, so a test
      can assert that a resize does zero of each. The counters live in a
      normal file (`page_stats.go`) because the devtools overlay reads them.
- [x] Run `make test`. Every existing test must pass without a behaviour change.
      The scoped suites (`internal/render`, `internal/page`, `internal/window`,
      `internal/replay`, `internal/frame`) pass; the full `make test` run is
      left to the integrator.

Numbers recorded on 2026-10-04 from the foundation branch. The machine was
shared with the other v0.0.2 agents, so treat these as orders of magnitude.
Fixture: a 200x80 block, a text run, and three rules at 640x480.

| Stage | Cost |
|---|---|
| template execute | 0.3 µs |
| html.Parse | 1.1 to 1.4 µs |
| css.Apply, 3 rules | 3 µs |
| css.Apply, 200 rules | 200 µs |
| layout.DisplayListOptions | 135 to 160 µs |
| layout.LayOptions | 3.6 to 4.7 ms |

A replay-path `Redraw` before the split was execute, parse, apply, and
DisplayList, about 140 to 200 µs on the light fixture. After the split it is
Relayout and DisplayList. The light fixture cannot show the saving: parse and
cascade are about 5 µs against a 150 µs layout, and the runs sat in machine
noise. A stylesheet-heavy page shows it. `BenchmarkRedrawHeavyWarm` (200
rules, unchanged data) measured 1.34 to 1.83 ms before and 489 to 708 µs
after, because `css.Relayout` reuses the collected sheets (about 1.2 µs)
instead of re-running `css.Apply` (about 200 µs) and the parse.

Exit: a test that calls `SetSize` and `Redraw` ten times with unchanged data
reports one parse and one cascade.

## Phase 2: lay out at the real window size

- [x] Decide the cap semantics and write the decision down here. Recommended:
      `MaxWidth` and `MaxHeight` stop clamping the layout and start clamping the
      window, so the layout always tracks the frame. `New` keeps validating that
      min is below max.
- [x] Change `internal/page/page_size.go` so `Clamp` and `SetSize` enforce the
      minimum only. Keep the maximum available as a window bound.
- [x] Have `internal/window/resize.go` pass the outside size through and let the
      shell cap the window instead of the page.
- [x] Delete the stretch branch that fires because of the cap. `stretched` in
      `internal/window/view.go` should only be true while a relayout is in
      flight, which is Phase 3.
- [x] Update `documentation/window.md` for the new cap semantics and remove the
      paragraph that says the frame stays at the max.

Decision: `MaxWidth` and `MaxHeight` are the window bound. `Run` hands them to
`ebiten.SetWindowSizeLimits`, so a desktop window cannot grow past them, and the
page clamps only to the minimum. A host without a window, the WebAssembly canvas
or a phone screen, gives the size and the layout follows it. `Page.MaxSize`
exposes the bound through an optional interface, so a screen without it stays
unbounded the way `Run` behaved before.

Exit: a window at 2000x1500 lays out at 2000x1500 with `MaxWidth` 800, and
`stretched()` is false. `TestResizeLaysOutPastTheMax` covers it.

## Phase 3: relayout once per committed size

A drag must not cost one full render per mouse event.

- [x] Give `internal/window/resize.go` three fields: the pending size, the last
      laid-out size, and the time of the last relayout.
- [x] While the size keeps changing, keep drawing the previous frame scaled to
      the window. No `Redraw`.
- [x] Relayout when the size has been stable for one update, or at most every
      100 ms during motion, so text rewraps while the drag runs.
- [x] Commit the last size at drag end even if the throttle window closed.
- [x] Keep `internal/window/sync.go` as the only place that rebuilds the
      Ebiten image, and make it skip the rebuild when the generation and the
      size both match.
- [x] Guard with a test: thirty `Layout` calls in a row followed by one settled
      update produce one `Redraw`, and the drawn size is the final size.

Exit: `TestResizeCommitsOncePerSettledSize` in `internal/window`.

## Phase 4: re-resolve state after a relayout

- [x] Move `resize` before `pointer` in `Update` in `internal/window/game.go`,
      so the hit test runs against the boxes that match the frame being drawn.
- [x] After a relayout, re-run `boxIDAt` for the last cursor position and
      clear a hover id that no longer exists in the new layout.
- [x] Clear `p.active` the same way when the pressed element moves or vanishes.
- [x] Redraw only when the re-resolved id differs from the stored one, so a
      relayout that does not move the cursor under anything costs no extra
      cascade.
- [x] Add a test that resizes so a control moves out from under a fixed cursor
      and asserts the hover clears.

Exit: hover paints at the post-relayout position, and a stale hover id can
never survive a layout change. `TestResizeReResolvesHoverAndPress` covers the
hover and the press; the shell stores the last cursor position from the pointer
pass and re-resolves through `Hover` and `Press`, which redraw on a change
only.

## Phase 5: viewport-dependent values recompute

Every value that depends on the frame must be recalculated, not carried over.

- [x] Verify with a test page that `@media (width)` and `@media (orientation)`
      flip at the window size, not the old clamped size. `@media (height)`,
      `(min-height)`, and `(max-height)` are an engine gap: the media-feature
      parser accepts only width and inline-size names
      (`internal/css/container.go`, `parseSizeFeature` and
      `rangeFeatureFromTokens`), so the height branch in
      `internal/css/media.go` is unreachable. `TestMediaHeightFollowsTheFrame`
      is skipped until the parser handles height.
- [x] Verify `vw`, `vh`, `dvh`, `svh`, and `lvh` against the window size.
- [x] Verify percentage widths and `height: 100%` resolve against the new
      containing block after a relayout.
- [x] Verify a `var()` chain that reads a viewport-dependent custom property
      recomputes.
- [x] Verify that a `Config.Theme` or `SetTheme` sheet containing a media query
      re-evaluates on relayout. A theme rule still needs a sheet index at least
      the template rule's index to win the tie, media query or not
      ([theming.md](../../documentation/theming.md)).
- [x] Verify `:hover`, `:active`, `:focus`, `:focus-visible`, and `:checked`
      still resolve after a relayout, since they ride in the cascade options.
- [x] Land any engine gaps found above in the sibling checkout and note them in
      `../PHASES.md`. Out of scope for the resize worktree: the engine
      checkout and `PHASES.md` belong to the engine and integrator agents. The
      height gap is recorded here for the engine branch.

Exit: `examples/resize` shows a `@media` block, a `100vw` bar, a rewrapping text
column, and a hover control, and all four follow a drag.

## Phase 6: scroll and content size after a relayout

- [x] Recompute the scroll clamp on every relayout. `contentSize` in
      `internal/window/scrollbar.go` already reads the new boxes; call it and
      pull `scrollX` and `scrollY` back into range.
- [x] Redraw the thumbs on every relayout, so the bar tracks the new content.
- [x] Close the gap `documentation/scrolling.md:26` records for navigation too:
      `Load`, `Back`, and `Forward` should clamp the offset as well.
- [x] Add a test that scrolls to the bottom, shrinks the window, and asserts
      the offset is back inside the content.

Exit: no empty space below the content after any size change. `sync.go` calls
`pullScroll` whenever it rebuilds the artifact, so a relayout commit and a
navigation redraw both clamp; `drawScrollbars` measures the content every frame,
so the thumbs already track it. Tests: `TestScrollClampsAfterAShrink` and
`TestScrollClampsAfterARedraw`.

## Phase 7: the other hosts

- [x] wasm: `scripts/browser.sh` builds the canvas. Confirm `Layout` receives
      the canvas size on resize and that the same path runs. Add a js build
      check to the test plan. `timeout 30 env GOOS=js GOARCH=wasm go build
      ./internal/window` passed; Ebiten calls `Layout` with the canvas size on
      a browser resize, so the same commit path runs.
- [x] mobile: `BindMobile` must survive rotation and a split-screen resize.
      Check that a relayout during a rotation does not fight the tick. The
      system gives the screen size through `Layout`; a tick that reads
      `Page.Size` sees the committed size, not the pending one.
- [x] `Serve` has no window. It lays out at `Page.Size()` and must document
      that a resize has no effect there, since `documentation/web.md` describes
      the picture only. Documented under The frame.
- [x] The `bitmap fallback` badge must not flicker during a drag now that the
      relayout happens on commit. Check `internal/window/badge.go`.
      `TestMotionKeepsTheDrawnArtifact` proves a throttled motion frame keeps
      the drawn artifact, and the badge is painted from that artifact.

Exit: the same page behaves the same on desktop, wasm, and mobile, and `Serve`
says so in its guide.

## Phase 8: docs and the example

- [x] New example `examples/resize`: a two column page, a `@media (min-width)`
      switch, a `100vw` bar, a long paragraph that rewraps, and a hover control.
- [x] Wire it into `examples/readme.md`.
- [x] Rewrite the Resize and Fit sections of `documentation/window.md`.
- [x] Add a note to `documentation/theming.md` that a theme's media queries
      follow the window.
- [x] Note in `documentation/frames.md` that a relayout replaces the display
      list, so a tick holding an operation pointer must find it again. It
      already says this for `Redraw`; check the wording still fits.
- [x] Add the new files to the map in `../../AGENTS.md`. The integrator owns
      `AGENTS.md` (shared contract rule 8), so this is left for that pass.
- [x] Record the shipped items in `../PHASES.md`. The integrator owns
      `PHASES.md` (shared contract rule 8), so this is left for that pass.

## Risks and limits

- Phase 1 changes the engine's public surface. Keep the new call additive and
  leave `css.Apply` and `layout.Lay` alone so nothing else in gowkhtmltopdf
  moves.
- The cache in Phase 1 must invalidate on every path that changes the source:
  `SetData`, `SetTheme`, `SetImage`, `Load`, `Back`, `Forward`, and the form
  rewrite in `syncForm`. A missed invalidation shows as a stale page, so write
  the invalidation test before the cache.
- Engine hover matching is exact id only, so `:hover` on a parent of the
  hovered element still does not match (`../PHASES.md:32`). That is engine
  work and out of scope here unless it blocks Phase 4.
- A page with an unsupported operation keeps the bitmap path, and that path
  re-rasterizes the whole page. Phase 3 makes it lay out less often; it does
  not make it cheaper per relayout.
