# DevTools

Recorded 2026-10-03 against `5a428f4`.

Status note, 2026-10-05: the shipped inspector differs from the panel this plan
describes. Commits `c5ff5a6` through `140dc48` and `067ba1a` replaced the
bottom-left overlay panel with a resizable right dock whose Elements, Frame,
and Ops tabs carry the pinned JSON, the counters, and the operation list. The
guide is [documentation/devtools.md](../../documentation/devtools.md).

The page knows everything a debugger needs and shows none of it. `Page.Boxes`
returns the hit-test boxes, `Page.Display` the operations, and the window knows
the frame size, the scroll offset, and the fallback flag. To see any of it you
write a test or a throwaway tick. This plan adds an inspector the window can
toggle: an overlay that outlines the box under the cursor, pins one to read its
tag, id, action, text, and geometry, outlines the display operations, and
shows the frame numbers. `Serve` gets a JSON route with the same data. The
overlay is window chrome, the way the scrollbar thumbs and the fallback badge
are: it draws after the page and never enters `Page.PNG`.

## What happens today

1. Boxes are complete but invisible. `internal/page/page_types.go:36` keeps
   `boxes`, `Page.Boxes()` returns them, and the only reader outside tests is
   the click hit test at `internal/page/page_click.go:45`. A `layout.Box`
   carries `ID`, `Tag`, `Action`, `Text`, `X`, `Y`, `W`, and `H` (engine
   `layout/layout.go:19`), which is a full label already.
2. Window chrome draws over the page and does not touch it. `Draw` fills,
   draws the content, draws the scrollbars, and defers the fallback badge
   (`internal/window/draw.go:9`, `internal/window/scrollbar.go:100`,
   `internal/window/badge.go:33`). The badge is the pattern: an Ebiten
   `vector` rectangle plus one `basicfont.Face7x13` line, invisible to
   `Page.PNG`.
3. There is no element-to-operation mapping. A display operation carries an
   `ID`, but it is operation identity, not the DOM id (engine
   `layout/displaylist.go:36`). `internal/frame` is the closest thing to an
   inspector, and it finds an operation by colour and geometry from a manual
   call (`internal/frame/frame.go:22`).
4. Input has one consumer. Every frame sends hover and click to the page
   (`internal/window/pointer.go:11`), and every key goes to the page unless a
   window chord consumed it (`internal/window/keys.go:20`, chords in
   `internal/window/shortcut.go:40`). F12 and Ctrl+Shift+I are unused.
5. The work behind a frame is not measured. `Page.Generation` increments on
   each redraw (`internal/page/page_size.go:40`), and that is all. Parse,
   cascade, layout, and draw times are invisible in a release build. The
   counters in incremental-repaint Phase 1 are written for tests.
6. `Serve` returns a picture and an image map (`internal/web/server.go:46`).
   The browser sees pixels. It cannot see the boxes or the numbers behind
   them, and `documentation/web.md:59` says the shell is the picture only.
7. `documentation/features.md:75` lists DevTools as absent, and no guide
   describes an inspector.

## Target behaviour

| Case | Today | After |
|---|---|---|
| Toggle | none | F12 or Ctrl+Shift+I shows and hides the overlay. The page key handler never sees that press. `Config.DevTools` or `SetDevTools` starts it on. |
| Hover | page hover only | The overlay outlines the box under the cursor and labels it `tag#id  w x h`. The page picture stays still. |
| Pick | none | A click pins the box under the cursor. A click on the panel or on the same box clears the pin. |
| Detail | none | The panel lists tag, id, action, text, geometry, and the operation count inside the box. |
| Ops | none | Ops mode outlines every operation in the retained display, colour by kind, with a count per kind. |
| Stats | generation only | FPS, TPS, draw time, redraws, parse, cascade, layout, repaint, op count, page and window size, scroll, stretched, fallback. |
| PNG | n/a | With the overlay off, `Page.PNG` and `Page.Generation` are unchanged. The overlay never enters the display list. |
| Serve | picture only | `GET /debug/state` returns boxes, stats, and op counts as JSON. No other response changes. |

## Shape

The page reports, the shell draws. `internal/host` gets an optional `Inspector`
interface with `DevTools() bool`, `SetDevTools(bool)`, and `Stats() Stats`.
`*page.Page` implements it. The window type-asserts the screen it already
holds, so a custom screen without the interface gets no overlay and no compile
change. Picking, the modes, and the pinned box are window state, in
`internal/window`, because they are input and drawing concerns. The stats are
page state, because only `Redraw` can measure a parse.

The overlay draws into the screen image after the page, so it can never reach
`Page.PNG`, the display list, or the box list. A test renders the same page
with the inspector off and on and asserts the PNG bytes and the generation
match.

## Phase 1: a stats snapshot on the page

- [x] Add `page.Stats` with `Redraws`, `Parses`, `Cascades`, `Layouts`,
      `Repaints`, `Boxes`, `Ops`, a `time.Duration` for the last redraw, and
      one for the last draw. The counters live in a normal file, not a
      test-only one: the overlay reads them. If incremental-repaint Phase 1
      lands first, reuse its counters instead of adding a second set.
- [x] Time the stages in `Redraw` (`internal/page/page_draw.go:13`): template
      execute, then `render.DisplayListState` or `render.PaintState`. Two
      `time.Now` calls per stage is noise next to a 130 ms render.
- [x] Define `host.Stats` in `internal/host` with the fields the overlay and
      the JSON route both need, so neither imports `internal/page`.
- [x] Add `host.Inspector` beside `Ticker` (`internal/host/screen.go:14`) and
      a compile-time assertion that `*page.Page` satisfies it, the way
      `internal/page/page_click.go:61` asserts `host.Screen`.
- [x] Add `Config.DevTools` and `Page.SetDevTools` / `Page.DevTools` in a new
      `internal/page/page_devtools.go`. The config flag is off by default;
      the toggle key is the usual way in.
- [x] Re-export `Stats` in `api.go` only if a caller has to name the type.
      `*Page` is an alias, so the methods need no re-export.
- [x] Test that two redraws report two redraws and a nonzero parse count, and
      that a click at a known box reports its id.

Exit: `page.Stats` reads on a live page, and both `*page.Page` and a test
double satisfy `host.Inspector`.

## Phase 2: the overlay frame

- [x] `shell.Draw` calls `s.drawDevTools(screen)` after `drawScrollbars`, so
      the overlay sits on top of the page and the thumbs. It draws only when
      the inspector flag is on.
- [x] New files in `internal/window`: `devtools.go` for state and toggle,
      `devtools_draw.go` for the outline, label, and panel, and
      `devtools_stats.go` for the line formatter. Every file stays under 2000
      characters; split again when one goes over.
- [x] Panel geometry: a fixed-width box in the bottom-left corner, above the
      horizontal scrollbar strip, 260 CSS pixels wide, height by line count.
      It must not cover the badge in the top-right.
- [x] Convert between box space and screen space in one function, and test
      it. Not stretched: subtract `scrollX` and `scrollY`. Stretched: scale
      by `screenW/frameW` and `screenH/frameH`, the same factors
      `drawReplayScaled` uses (`internal/window/replay_scale.go:32`).
- [x] Outline the hovered and pinned boxes with `vector.StrokeRect`, one
      colour each, so the pinned box and the box under the cursor stay apart.
      Put the label above the box and flip it below when it would clip the
      top edge.
- [x] Nothing in the draw path calls `Redraw`, `Hover`, or any other page
      method. The overlay shows the last picture until the page changes on
      its own.

Exit: a test draws the overlay into an offscreen Ebiten image at 1920x1080 in
normal and stretched modes, and finds the outline at the box's computed
screen coordinates. On a page with no boxes, the panel still draws.

## Phase 3: picking and input gating

- [x] Add the toggle to the window's key handling before the page sees the
      key. The check runs before `keyEvents` in `keys()`
      (`internal/window/keys.go:9`), not inside `shortcutChord`, which runs
      after the page already received the press. Use the same two-frame key
      watch the chords use (`internal/window/key_watch.go`), so one press
      toggles once and the matching release is swallowed too.
- [x] While the overlay is on, the pointer belongs to the overlay.
      `pointer()` returns before `s.app.Hover` when the cursor is over the
      panel, and a click in the content pins the box under the cursor
      instead of reaching `Click`. Alt+click forwards to the page, so a
      control can still be exercised while inspecting.
- [x] While the overlay is on, hover does not reach the page at all. The
      picture stays still while the cursor moves, which keeps the pixels the
      stats describe stable. Closing the overlay sends one hover for the
      current cursor position, so a stale `:hover` does not stick.
- [x] Panel clicks are hit-tested by the shell against its own rects, never
      against `Boxes()`. The panel is not an element.
- [x] Clear the pinned and hovered boxes when the overlay closes, and drop
      either one when the generation changes and its id is gone from
      `Boxes()`.

Exit: `TestDevToolsToggleSwallowsF12` asserts the page's key handler saw
nothing, and `TestDevToolsPickSkipsClick` asserts a click selects an id
without calling the page's click handler.

## Phase 4: the operations view

- [x] Add an ops mode, toggled from the panel or the `o` key while the
      overlay is on. It outlines every operation in `s.display` in paint
      order (`s.display.Order`, `internal/window/sync.go:39`).
- [x] Colour by kind: fills, strokes, text, images, grid runs, each with a
      fixed colour. The panel prints the count per kind and the total.
- [x] Bounds come from `op.X`, `op.Y`, `op.W`, and `op.H` in points,
      converted with `display.PixelPerPoint` and then to screen space. A text
      op carries its baseline in `Y`, so build the line box from the face
      metrics before the conversion; `internal/frame/text.go:20` records the
      rule and the per-kind caveats are in the engine's `DisplayOp` comment
      (`layout/displaylist.go:13`).
- [x] Skip `DisplayOpNoop` and `DisplayOpUnknown` boundaries. They paint
      nothing, so an outline for them is wrong.
- [x] A fallback page has no display list. Ops mode prints `bitmap fallback`
      where the counts would go and outlines nothing.
- [x] Per-operation highlight on hover is a stretch. Keep it out until the
      outline view is measured on a 400-op page.

Exit: `examples/devtools` shows one coloured outline per operation, and every
kind in the page's display has a count in the panel.

## Phase 5: the stats panel

- [x] FPS and TPS from `ebiten.ActualFPS` and `ebiten.ActualTPS`. Frame time
      as the wall time between two `Draw` calls, and draw time around
      `drawContent` (`internal/window/draw.go:20`). Report both, because the
      display lookup in `sync.go` can dominate.
- [x] Print the frame state the window already holds: `screenW`, `screenH`,
      `app.Size()`, `scrollX`, `scrollY`, `stretched()`
      (`internal/window/view.go:7`), `fallback`, and `s.seq`.
- [x] Print the page stats from `host.Inspector.Stats()`: redraws, parses,
      cascades, layouts, repaints, last redraw, boxes, ops.
- [x] When dynamic-resize Phase 3 lands, add the committed and skipped
      relayout counts it stores. When hot-reload.md lands, add reloads and
      the last reload error. One line each, same panel.
- [x] Format with the same `badgeFace` and a tabular layout. Measure with
      `text.Measure` and size the panel from the widest line, so a growing
      number does not overflow it.

Exit: the panel updates every frame while the page is idle, and a click that
forces a redraw moves the redraw and generation lines in the same frame.

## Phase 6: the web route

- [x] Register `GET /debug/state` in `internal/web/server.go:47`, next to the
      existing routes. It takes the same mutex as the others
      (`internal/web/server.go:28`), so it never reads a page mid-redraw.
- [x] Return JSON: size, generation, fallback, boxes, stats when the screen
      implements `host.Inspector`, and an op-kind count map when a display
      list exists. Use the engine's kind constants as strings.
- [x] Do not change `GET /`, `GET /frame.png`, or the shell HTML. The picture
      page stays the picture page, and `documentation/web.md` gets a route
      row.
- [x] Wasm and mobile have the overlay through the shared window loop, but
      the browser may eat F12 and Ctrl+Shift+I. `Config.DevTools` and
      `SetDevTools` are the way in there; the key is a desktop convenience.

Exit: `curl /debug/state` on a running `-web` page returns the page's boxes,
and two calls around a click show the generation difference.

## Phase 7: example and docs

- [x] New `examples/devtools`: a page with a counter, a nested label, a
      rounded fill, an image through `SetImage`, a text run, and a table row,
      so every operation kind the replay draws appears once. `-web` on port
      8122.
- [x] `examples/readme.md` row.
- [x] New `documentation/devtools.md`: the toggle keys, what each panel line
      means, the ops colour key, `Config.DevTools` and `SetDevTools`, the
      `host.Inspector` interface for a custom screen, and the rule that the
      overlay never enters the picture.
- [x] `documentation/README.md` table row.
- [x] `documentation/features.md`: move DevTools out of Still absent and add
      a section.
- [x] `documentation/keys.md`: F12 and Ctrl+Shift+I belong to the window
      while a page has an inspector. `documentation/pointer.md`: the overlay
      pauses hover and click, and Alt+click forwards.
- [x] `documentation/web.md`: the `/debug/state` route.
- [x] Add the new files to the map in `../../AGENTS.md`, and record what
      shipped in `../PHASES.md`.

## Risks and limits

- The overlay must never change the page. No path in `internal/window` calls
  `Redraw` to show a tooltip, no path writes into `s.display`, and toggling
  the panel does not bump the generation. The byte test in Shape is the
  guard.
- Page hover and press are host state, so suppressing pointer input freezes
  them at their last value. A pinned `:hover` looks a little odd, and it is
  the honest choice: the alternative is a full cascade redraw per mouse move
  while inspecting.
- There are no computed styles to show. The engine styles a document
  internally and publishes no per-element reader, so the picker shows
  geometry and attributes, not `font-size` or `margin`. A computed-style call
  for one element id is the follow-up; record it in `../PHASES.md` as
  engine work, not here.
- The panel uses `basicfont.Face7x13`, not the page's fonts. It is chrome and
  should not depend on the page it inspects.
- F12 is a browser key in wasm builds and a phone has no key. The
  programmatic switch is the cross-platform entry.
- An ops view on a fallback page is empty by definition. The
  `bitmap fallback` line is the answer, not an outline.
