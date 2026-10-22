# Input and interaction

Recorded 2026-10-03 against `5a428f4`.

The pointer and the keyboard cover the browser basics: hover, press, click,
taps, typing, and the editing chords. The layer under them is missing. There is
no caret and no selection range, so typing can only append. Focus comes from a
click and Tab does nothing (`documentation/keys.md:44`). There is no drag while
the button is held, no context menu, no cursor shape, no touch scroll or pinch,
no fullscreen, and no IME. This plan covers the input and interaction pieces a
caller has to build by hand today.

## What happens today

1. Editing is append-only. `Control` holds a value and nothing else
   (`internal/page/form_types.go:5`), `formState` holds `focusID` and one
   `selected` bool (`internal/page/form_types.go:33`), and `insertValue`
   appends to the end (`internal/page/form_key.go:81`). There is no caret
   index, no anchor, and no way to type into the middle of a value.
2. Focus is click-only. `activate` sets `focusID` on a click
   (`internal/page/form_activate.go:27`), and `documentation/keys.md:44` says
   Tab, Escape, and the arrows reach the handler like any other key; the
   window does not move focus. `:focus` styling already exists, so the styling
   side is done.
3. The pointer has four events and no move between them. `pointer()` sends
   hover every frame, press and click on the down edge, and release on the up
   edge (`internal/window/pointer.go:11`). Touch sends one tap per fresh touch
   (`internal/window/pointer_touch.go:7`). A drag has nothing to ride on.
4. Selection is select-all only. `SelectAll` flips one bool
   (`internal/page/page_edit.go:24`), and the window chords in
   `internal/window/shortcut.go:40` call it. There is no range, so every other
   selection interaction is absent.
5. The still-absent list carries the rest: fullscreen, cursor shape, touch
   scrolling and pinch, and programmatic scrolling (`documentation/features.md:73`),
   plus IME (`documentation/features.md:75`, `documentation/keys.md:45`).
6. Two of those are free from the toolkit. Ebiten v2.10.4 exposes
   `ebiten.SetCursorShape` with `CursorShapeText` and `CursorShapePointer`, and
   `ebiten.SetFullscreen` / `ebiten.IsFullscreen`. The window calls none of
   them (`internal/window/` has no cursor or fullscreen file).
7. The window owns the scroll offset and the page does not. `scrollX` and
   `scrollY` live on the shell (`internal/window/game.go:50`), the wheel sets
   them (`internal/window/view.go:21`), and no page call can move them.

## Target behaviour

| Case | Today | After |
|---|---|---|
| Click in a field | Focus, caret at the end | Focus, caret at the clicked glyph |
| Type | Appends | Inserts at the caret and replaces a selection |
| Arrows, Home, End | Reach the handler, move nothing | Move the caret. Shift extends. Ctrl jumps by word, Ctrl+Home/End to the ends |
| Tab / Shift+Tab | Reach the handler | Move focus in document order, skip disabled and `tabindex="-1"`, wrap |
| Mouse drag | Hover only | Extends the selection. Double-click selects a word, triple-click selects the line |
| Right click | Nothing | A shell menu with cut, copy, paste, select all, undo, and redo |
| Hover cursor | Default arrow | Text over a field, hand over a link or button, resize on a scrollbar thumb |
| Wheel | Scrolls | Unchanged, plus touch drag scrolls and pinch zooms |
| F11 | Nothing | Toggles fullscreen |
| Page scroll API | None | `Page.ScrollTo` and `Page.ScrollBy` clamp to the content |
| IME | Composition never arrives | Ebiten's `exp/textinput` is the route; the design is recorded, nothing lands in 0.0.2 |

## Shape

The caret and the selection range belong on the page, because that is where the
field value lives. The form state grows a caret rune offset and an anchor, and
`selected` becomes the derived full range. The window keeps the events it
already owns: focus keys, drag edges, the context menu, cursor shape, touch
gestures, fullscreen, and the scroll offset. The page gains optional methods on
`host.Screen` for the events it can consume, so a game page without fields
keeps every key and every click exactly as it has them today.

New files, kept under the 2000-character limit the way the rest of the package
is:

| File | Owns |
|---|---|
| `internal/page/form_caret.go` | Caret offset, anchor, clamp, word and line bounds |
| `internal/page/form_select.go` | Range selection, select-all, click, drag, double, triple |
| `internal/page/form_focus.go` | Focus order, next, previous, programmatic focus |
| `internal/page/page_scroll.go` | The scroll request the window consumes |
| `internal/textrun/measure.go` | Point to rune offset in a shaped run |
| `internal/window/cursor.go` | Hover target to `ebiten.CursorShape` |
| `internal/window/context_menu.go` | Right-click detection, items, actions |
| `internal/window/menu_draw.go` | The menu rectangle and rows |
| `internal/window/touch_gesture.go` | Touch drag and a pinch scale |
| `internal/window/fullscreen.go` | The F11 toggle |

## Phase 1: the caret and selection model

The gate for every other phase. Without a caret, the rest has nowhere to land.

- [x] Give `formState` a caret rune offset and an anchor offset. Normalize to
      `[start, end)` and keep the caret at the moving end.
- [x] Route `insertValue`, `backspaceValue`, and `deleteWordValue`
      (`internal/page/form_key.go:81`) through the range: replace the selection
      or edit at the caret.
- [x] `SelectAll` sets the full range instead of the bool
      (`internal/page/page_edit.go:24`).
- [x] Carry the caret and anchor through the redraw merge the way values
      already survive `Redraw` (`internal/page/form_merge.go`). A redraw must
      not move the caret.
- [x] Clear the range on blur and on a new click in a different control.
- [x] A caret move is not an edit. It must not call `BeforeEdit` or `Change`
      (`internal/page/page_click.go:40` and `form_key.go:28` fire those for
      value changes only).
- [x] Test: set a value, put the caret in the middle, insert, backspace, and
      delete a word. Assert the value and the caret after each step.

Exit: typing inserts at the caret and the value is correct for every edit in
the test. No Change callback fires for a caret-only move.

## Phase 2: caret placement from a click

The honestly hard part: the page needs a glyph offset from a point. The engine
shapes the text, so the offset is not in `layout.Box`, and the page cannot
import `internal/frame` because that package imports the root `ownframe` package
(`internal/frame/frame.go:9`) and would cycle.

- [x] Add a measure helper that turns a point inside a field's text run into a
      rune offset. The page already holds the display list (`Page.Display()`),
      so walk its text operation for the box and measure the run. Put the
      helper in `internal/render` or a new `internal/textrun`, both of which
      the page can import.
- [x] `SelectAt(x, y)` sets the anchor and the caret from the point: the field
      under the point gets focus and the offset at the click.
- [x] If the offset cannot be measured for a page shape, land at the end of
      the value and record that shape as a limit. Do not fake a caret at a
      wrong offset.
- [x] Test: click at three x positions in a known run and assert three offsets.

Exit: a click places the caret where the glyph is.

## Phase 3: caret keys and focus traversal

- [x] Add `host.Focuser` with `FocusNext`, `FocusPrev`, `Focus(id)`, and
      `FocusID`, and wire `Page` through it. Keep it optional, so a custom
      screen without fields is unaffected.
- [x] Arrow Left and Right move by rune, Home and End move to the ends,
      Ctrl/Alt with Left and Right jump by word.
- [x] Shift with any of those extends the range from the anchor.
- [x] Up and Down in a textarea are a stretch; the page keeps a single-line
      value today, so record the limit rather than faking it.
- [x] `Page.FocusNext` and `Page.FocusPrev` walk `form.order`, skip disabled
      controls and `tabindex="-1"`, and wrap. A form control list already
      exists (`internal/page/form_types.go:33`); the box list gives the
      document order (`internal/page/page_click.go:45`).
- [x] The window consumes Tab and Shift+Tab only when the screen implements
      `host.Focuser` and at least one field exists. Otherwise the keys reach
      the page handler exactly as they do today (`documentation/keys.md:44`).
      This is a behavior change and the guide has to say so.
- [x] Keyboard focus sets the same `:focus-visible` host state a click sets, so
      the focus ring draws; Escape clears focus and returns the keys to the
      page.
- [x] Test: two fields in a page, Tab twice, Shift+Tab once, assert the focused
      ids, the skipped disabled field, and the wraparound.

Exit: Tab reaches every enabled control in order and Shift+Tab reverses it,
with the focus ring visible at each stop.

## Phase 4: mouse drag, double, and triple

- [x] `host.Drag(ctx, x, y)` reaches the page only between a press in a field
      and the matching release. A drag over plain content keeps the old
      behavior. Each drag moves the far end of the range from Phase 2.
- [x] Double-click selects the word under the point, triple-click selects the
      line. Track click time and count in the window, the way
      `internal/window/pointer.go` already tracks the down edge.
- [x] Dragging a selection past the top or bottom edge scrolls while held.
      This is the one auto-scroll case worth doing; reuse the wheel clamp in
      `internal/window/view.go:21`.
- [x] A selection that survives a redraw must not turn into a stale range. The
      merge in Phase 1 clamps the offsets to the new value length.
- [x] Test: drag across a wrapped field and assert the selection covers the
      visible span, then redraw and assert the range is clamped and intact.

Exit: one pointer press, move, release selects a span on screen.

## Phase 5: the context menu

- [x] Right-click opens a shell-drawn menu at the cursor: cut, copy, paste,
      select all, undo, redo. The window draws it the way it draws the badge
      and the scrollbar thumbs, with `vector` and the existing face
      (`internal/window/badge.go:33`).
- [x] Items come from the page state: cut and copy need a non-empty selection,
      paste needs a focused editable field, undo and redo are enabled when the
      handler exists. Add `host.ContextMenu() []host.MenuItem` beside
      `Ticker` (`internal/host/screen.go:14`).
- [x] Actions call `Cut`, `Copy`, `Paste`, `SelectAll`, `Undo`, and `Redo`,
      which already exist on the screen interface.
- [x] A left click that is not on a row or Escape closes the menu; a right
      click elsewhere moves the menu to the new point. The menu never enters
      the page picture or a `Redraw`; it is chrome, the same rule the fallback
      badge follows.
- [x] Test: open the menu over a focused field with text selected, assert the
      enabled rows, click paste, assert the value.

Exit: right-click over a field pastes into the middle of a value through the
menu.

## Phase 6: cursor shapes

- [x] `internal/window/cursor.go` maps the hovered box to a shape: a text
      input or textarea to `CursorShapeText`, a link, button, checkbox, radio,
      or select to `CursorShapePointer`, a scrollbar thumb to a resize cursor,
      everything else to the default.
- [x] The window does not know a box is a form control. Add
      `host.CursorShape() host.Shape` to the screen, or expose the hovered
      control kind through the existing `Hover` path. Pick one in the phase
      and write it down.
- [x] Call `ebiten.SetCursorShape` only when the shape changes, not per frame.
- [x] wasm and mobile ignore the call; the cursor value still reads back for
      tests.
- [x] Test: move the cursor over a field, a button, and the page background,
      and assert the shape sequence.

Exit: the I-beam appears over a field and the hand over a button.

## Phase 7: scrolling, touch, and fullscreen

- [x] `Page.ScrollTo(x, y)` and `Page.ScrollBy(dx, dy)` store a request. The
      window consumes it through `host.ScrollRequester`, clamps to
      `contentSize` (`internal/window/scrollbar.go:21`), redraws the thumbs,
      and clears it. The offset stays owned by the shell.
- [x] Touch drag on the content scrolls instead of tapping. Keep the tap
      suppression rule in `pointer_touch.go:7` so one gesture is not both.
- [x] Pinch sets a scale on the shell. The replay or bitmap draws through it
      the way `drawReplayScaled` does (`internal/window/replay_scale.go:12`),
      and `contentPoint` (`internal/window/pointer_touch.go:18`) divides by it
      so clicks stay on the right glyph.
- [x] F11 toggles `ebiten.SetFullscreen` on desktop. wasm and mobile keep
      their own fullscreen and the call is a no-op.
- [x] Guard the overlay features: the devtools overlay and the partial repaint
      work in the other 0.0.2 plans must not fight a zoom or a scroll request.
      Add the zoom factor to those tests.
- [x] Test: `ScrollTo` past the bottom clamps, touch drag moves the offset,
      and a pinch of 1.5 puts a click at the same box.

Exit: a program can scroll a page from Go, a phone drag scrolls it, and F11
fills the screen.

## Phase 8: IME, recorded, the API exists

- [x] Record the Ebiten route in this file and in `../PHASES.md`: Ebiten
      v2.10.4 ships the experimental `exp/textinput` package, where
      `Composer.OnComposition` carries the preedit text,
      `SessionOptions.CaretBounds` the caret rectangle, and `OnCommit` and
      `OnEnd` the lifecycle. The window wires the Composer; the note sits
      where typing is wired (`internal/page/page_type.go`).
- [x] Sketch the page side so it can land next cycle: a composing run drawn
      under the caret, commit through the existing `Type` path, cancel on
      blur.
- [x] No fake composition, no platform code in this release.
- [x] Record the IME gap in `documentation/features.md` next to the other
      absences, with the Ebiten version checked.

Exit: the ask and the design are written down and no partial IME path ships.

## Phase 9: docs, example, and the ledger

- [x] New `examples/input`: two fields, a checkbox, a link, a long page, and a
      scrollbar, so Tab, caret keys, drag, menu, cursor, scroll, and F11 are
      all visible in one window. `-web` on port 8126.
- [x] Update `documentation/keys.md`: Tab and Shift+Tab are consumed when a
      field exists; the caret keys; what still reaches the handler.
- [x] Update `documentation/pointer.md`: drag, double, triple, and the context
      menu; `documentation/forms.md` and `documentation/editing.md`: the caret
      and the range.
- [x] Update `documentation/features.md`: move cursor shape, touch scroll and
      pinch, programmatic scrolling, and fullscreen out of Still absent. Keep
      IME in it.
- [x] New `documentation/interaction.md` linked from
      `documentation/README.md`, with the call shapes.
- [x] Add the new files to the map in `../../AGENTS.md`, and record what
      shipped in `../PHASES.md`.
- [x] `examples/readme.md` row.

## Risks and limits

- Click-to-offset needs glyph metrics the engine shapes internally. Phase 2
  has a recorded fallback. Do not ship a caret that sits at a wrong offset on
  a page whose run cannot be measured.
- The caret and anchor must merge like values do or a redraw loses the cursor
  position mid-typing. Phase 1 writes the test before the merge.
- Tab is a behavior change. A page that wants Tab today can keep it by not
  implementing `host.Focuser`, and the guide says which pages get the new
  path.
- The context menu, cursor, and zoom are shell state. None of them may enter
  `Page.PNG`, the display list, or the generation counter, the same guard the
  devtools plan carries.
- Pinch zoom touches pointer mapping, scrollbar math, and the devtools
  overlay. It is the riskiest phase for silent regressions, so it lands last
  among the interactive phases and only with the devtools byte test green.
- IME is not wired today. Ebiten ships `exp/textinput`, so the work is
  window and page code, not an upstream ask. The phase records the design so
  the gap is not forgotten.
- Up and Down in a textarea, and selection beyond the visible scroll of a
  field, are out of scope. The value model is a single line.

## Page side: what shipped and the measured limits

The caret and the anchor live on `formState` as rune offsets. The range is
`[min, max)`, the caret is the moving end, and an edit collapses the range to
the insertion point. `SelectAll` sets `[0, len)`. The merge after `Redraw`
clamps both offsets to the focused value's new length and drops them when the
focus is gone, so a redraw never moves a surviving caret. A caret move fires
neither `BeforeEdit` nor `Change`.

`SelectAt` measures the clicked glyph by walking the page's display-list text
runs in `internal/textrun`. A wrapped run keeps its offsets: the measure lines
the runs up with the value and lets only whitespace drop at a line break.
Three shapes fall back to the end of the value instead of a guessed offset: a
page with no display list (the bitmap fallback), a run whose drawn text does
not line up with the shown value (a placeholder, or a text transform that
changes the rune count), and a run whose font cannot be parsed. The caret
paints as a split in the field span at the offset, and a non-empty range
paints `data-ownframe-selection` around the selected runes.

The page's `KeyDown` moves the caret itself when a text field is focused:
`arrowleft`, `arrowright`, `home`, and `end` move it, `ctrl+` or `alt+` jumps
by word, `shift+` extends from the anchor, and `escape` clears the focus. The
window passes key names as it does today; the modifier prefixes are the
convention for the window side to name a chord when it wants the extended
behavior.

## IME: the route and the design

Ebiten v2.10.4 ships the experimental `exp/textinput` package, checked
against the v2.10.4 module. `Composer.OnComposition` carries the composing
text, `SessionOptions.CaretBounds` the caret rectangle, and `OnCommit` and
`OnEnd` the lifecycle, on Windows, macOS, Linux, iOS, Android, and browsers.

The page side takes it in three steps. A composing run is drawn under the
caret without touching the stored value, a commit goes through the existing
`Type` path so `BeforeEdit` and `Change` fire once, and a blur or an escape
drops the composing run. `internal/page/page_type.go` carries the same note
at the typing entry point. The window wiring and the page drawing are the
first input item of the next cycle. No IME code ships in 0.0.2.

