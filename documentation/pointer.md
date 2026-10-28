# Pointer

The window sends hover, press, release, and click events to a page. `internal/window/pointer.go` calls the inner screen each frame. A caller uses `page` for the same `Hover`, `Press`, `Release`, and `Click` calls.

## Input calls

`Page.Edit().Hover(ctx, x, y)` sets the hovered id to the innermost box with an id at the point. A point that hits no box with an id clears it. It redraws only when the id changed. A hover change repaints the union of the box it left and the box it entered, not the rest of the frame.

`Page.Edit().Press(ctx, x, y)` sets the pressed id the same way and redraws only when it changed. `Page.Edit().Release(ctx)` clears the pressed id and redraws only when it was set.

`Page.Click(ctx, x, y)` hit-tests the last box that contains the point, which is the innermost box in document order. A `position: fixed` box is painted over the content, so it wins the hit test against any content box it overlaps, and a control inside it still wins against the fixed box around it. A box without an id falls back to the innermost id-bearing element under the point, so a click on a child icon reaches its control. Then:

- A form control activates before the handler runs and never follows a route ([forms.md](forms.md)).
- Any other hit blurs the focused control.
- After a non-control hit, a non-empty `data-action` with a registered route loads that HTML and skips the handler ([navigation.md](navigation.md)).
- A nil `Handlers.Click` ignores the event.
- An error from the handler or the load returns before the redraw. Otherwise the page redraws after the handler.
- `Handlers.ClickPaint` runs after the handler, and after a control's own change. When it reports that it painted the frame, the page skips the redraw and keeps the damage the handler marked for the window. A handler that patched the retained display with `Page.PatchElement` uses it to keep the click off the layout path.

The inner screen's `LongPress(ctx, x, y)` handles a held press. A text control under the point selects the word there, the same selection `SelectWordAt` makes. Otherwise a `Handlers.LongPress` handler runs on the innermost box under the point, found the way `page.Click` finds it. It reports true when it claimed the press, from a text control or a handler, and the page redraws after the handler. A nil handler with no text control under the point reports false.

## Window mapping

The window sends one hover when the point moves, and again after a redraw that changed the built generation ([interaction.md](interaction.md)). Press and click both fire on the mouse-down edge; release fires on the mouse-up edge.

A mouse press in a field calls the inner screen's `SelectAt` before it runs the click handler, so a handler that focuses or selects a field is not blurred by the caret placement. Each move with the button down calls `Drag`, so one press, move, release selects a span. Two rapid presses call `SelectWordAt` and a third calls `SelectLineAt`. A press held within 8 px for 450 ms calls the inner screen's `LongPress` once, on a mouse or a touch. The inner screen's `LongPress` selects the word under a text field or runs `Handlers.LongPress`, and a true answer claims the press. On a claimed press the moves until release call `Drag` instead of scrolling, and the lift is not a tap. A press that moves past the slop before the delay never fires. `internal/window/long_press.go` holds the watch and `internal/window/hold.go` the frame check. Dragging past the top or bottom edge scrolls while the button stays down. The web `/click` route calls `SelectAt` before `Click` on the inner screen too ([web.md](web.md)). The window calls the inner screen. [interaction.md](interaction.md) has the details.

A right click asks the page for its context menu rows and draws them at the cursor. A left click on a row runs its action; a left click elsewhere or Escape closes the menu.

The cursor shape follows the hovered element when the page implements `host.CursorShape`: an I-beam over a field, a hand over a link or button.

Visible scrollbar thumbs belong to the window. They show a vertical or horizontal resize cursor on hover, and dragging one scrolls the page without pressing the underlying content. `Config.Scrollbar` controls whether thumbs appear; hidden thumbs are not hit-testable.

While the devtools overlay is on ([devtools.md](devtools.md)), hover pauses and a click pins the box under the cursor instead of reaching `page.Click`. Clicks on the dock switch tabs, fold JSON nodes, toggle the outlines, or pick an operation, and never reach the page. The pointer drags the dock's left edge to resize it. Alt+click forwards the press and the click to the page. Closing the overlay sends one hover for the current cursor.

The coordinates arrive in page space, with the scroll offset added or the stretch scale applied ([scrolling.md](scrolling.md)).

A touch that lifts without moving sends press, click, and release as a tap. A moved touch drags the page, and two fingers pinch a zoom; a second finger cancels a pending hold. A touch held within 8 px for 450 ms fires the inner screen's `LongPress` first, and after a claim the finger drags the selection instead of the page. A mouse click in the same frame suppresses the tap, so one gesture is not delivered twice.

## CSS states

The hovered and pressed ids become the host states `:hover` and `:active`. `:focus` comes from clicks and from Tab, and the focus ring draws at each keyboard stop ([interaction.md](interaction.md)). [forms.md](forms.md) lists the state attributes and the focus rules.
