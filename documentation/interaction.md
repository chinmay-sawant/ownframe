# Interaction

The window owns the focus keys, the pointer gestures, the context menu, the cursor shape, touch gestures, fullscreen, and the scroll offset. The page owns the caret, the selection range, the focus order, and the scroll requests. A page that implements none of the optional surfaces keeps the v0.0.1 behavior.

## Focus

Tab and Shift+Tab move focus in document order when the screen implements `host.Focuser` and a field takes the focus. The window calls `FocusNext` or `FocusPrev`, then reads `FocusID`. An empty id means no field moved, so the key reaches the page handler as before. When a field moves, the window consumes the press and its release, so `Handlers.KeyDown` does not see `"tab"`.

The page sets the same `:focus-visible` host state a click sets, so the focus ring draws at each stop. Escape closes the context menu first. With the menu closed it calls `Focus(ctx, "")` and clears the ring. Escape itself still reaches the handler. `internal/window/focus_keys.go` holds the window side.

Buttons with an id join the focus order. Disabled buttons are skipped and ignore clicks. Focusing an offscreen button requests a scroll to reveal it. `FocusID` reports button focus. `page.FocusedField` reports fields only. The application's `KeyDown` handler can activate a focused button on Enter or Space by passing its box center to `Page.Click`. Enabled buttons keep their `data-action` routes.

## Caret and selection

A mouse press in a field calls `page.Testing().SelectAt(ctx, x, y)`, which focuses the field under the point and puts the caret at the glyph. The window calls the inner screen. It runs before the click handler, so a handler that focuses or selects a field keeps it. A move with the button down calls `Drag(ctx, x, y)`, so one press, move, release selects a span. Dragging past the top or bottom edge scrolls one wheel step per frame, clamped to the content.

Two rapid presses at one spot call `SelectWordAt`; a third calls `SelectLineAt`. The gap is 400 ms and the spot may move 4 px. `internal/window/drag_select.go` and `internal/window/click_watch.go` hold the window side. The page holds the caret and the range ([editing.md](editing.md)).

A press held within 8 px for 450 ms calls `Page.Testing().LongPress`. A text field under the point selects its word, the same selection a double click makes, and claims the press. Any other box reaches `Handlers.LongPress`, which receives the innermost box under the point the way `page.Click` finds it. A claim keeps the gesture: each move until release calls `Drag`, and the lift is not a tap, so a phone can hold a word and then extend the selection. A press that moves past the slop before the delay never fires. `internal/window/long_press.go` holds the delay and the slop, and `internal/window/hold.go` checks the timer once per frame.

## Context menu

A right click asks the page for rows with `ContextMenu() []host.MenuItem` and draws them at the cursor. Row ids are `cut`, `copy`, `paste`, `select-all`, `undo`, and `redo`. An action calls the matching method on the inner screen. Tests use `page` for `Copy`, `Cut`, `Paste`, `SelectAll`, `Undo`, and `Redo`. Paste reads the clipboard the way the Ctrl+V chord does. A disabled row draws grey and does nothing.

A left click on a row runs the action and closes the menu. A left click elsewhere or Escape closes it, and a new right click reopens it at the new point. The menu is chrome. `internal/window/menu_draw.go` paints it after the page, so it never enters `Page.PNG`, the display list, or a `Redraw`. The page is only asked for its rows when the menu opens.

## Cursor shape

When the screen implements `host.CursorShape`, the window reads `CursorShape()` after each hover and sets the Ebiten cursor only when the shape changed. Text fields and textareas map to the I-beam, links and controls to the hand, and everything else to the default arrow. wasm and mobile keep their own cursor and ignore the call. The last shape stays on the shell for tests.

## Touch

`Page.PressedID()` returns the control held by the current pointer gesture, or an empty string after `page.Release`. A frame callback can use it to stop application-specific hold repeats when the pointer releases.

A finger that lifts without moving sends press, click, and release as a tap. A finger held within 8 px for 450 ms is a long press: a text field under the point selects its word and any other box reaches `Handlers.LongPress`, after which the finger drags the selection instead of the page and the lift is not a tap. A second finger cancels the hold. A finger that moves past 8 px drags the page instead: the content follows the finger, clamped to the content ends. Two fingers pinch a zoom that starts at the span of the first two fingers and stays between 0.25 and 4. `Ctrl+=`, `Ctrl+-`, `Ctrl+0`, and Ctrl+wheel change the same zoom; the keyboard steps are in [keys.md](keys.md). The replay and bitmap paths draw through the zoom, and `contentPoint` divides it out, so a click under a pinch lands on the same box. A tap that the system also reports as a mouse click does not tap twice.

## Captured pointer gestures

`Handlers.DragStart(ctx, box)` can return `true` to claim a mouse or single-finger gesture that begins inside that box. `DragMove(ctx, dx, dy)` then receives movement in CSS pixels as the pointer moves, including movement outside the starting box. `DragEnd(ctx)` ends the capture on release or when a second finger appears. These callbacks do not automatically redraw after movement; redraw explicitly when they change the page.

A captured drag suppresses page scrolling and long press. A stationary press still clicks on release. An unclaimed gesture keeps the normal selection and scrolling behavior. `examples/lg-remote` uses capture only inside its Wi-Fi pointer pad. `Handlers.Swipe` remains the whole-page callback for a completed touch swipe.

## Programmatic scroll

`Page.ScrollTo(x, y)` and `Page.ScrollBy(dx, dy)` store one request. The window consumes it through `host.ScrollRequester`, clamps it to the content size, and clears it. The offset stays owned by the shell, so the wheel, a touch drag, and a page request all move the same value ([scrolling.md](scrolling.md)).

## Fullscreen

F11 toggles `ebiten.SetFullscreen` on desktop and the window consumes the key. wasm and mobile keep their own fullscreen, the call is a no-op there, and F11 reaches the page. `internal/window/fullscreen.go` holds the window side.

## Limits

- Up and Down in a textarea, and selection beyond the visible scroll of a field, are out of scope. The value model is a single line.
- IME uses Ebiten's experimental `exp/textinput` on Android, iOS, and desktop builds. A phone shows its soft keyboard. Desktop preedit is drawn as a temporary underlined run; it leaves the stored value alone until commit. Commits use `page.Type` or `page.Testing().IMEReplace`, and the window skips its own typing on ticks the IME consumed. The window calls the inner screen. The Android and iOS adapter still displays preedit through the field edit path. Native desktop composition needs verification with each supported OS input method. wasm uses browser key input.
- The page exposes a semantic accessibility tree and actions for supported controls. The window has no native screen-reader bridge, and the tree does not filter content hidden only by CSS. See [accessibility.md](accessibility.md).
- The context menu is not native. It is a shell rectangle drawn with the badge face.
