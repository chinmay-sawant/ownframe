# blinkless-scroll

The host half of CSS scroll-driven timelines and triggers.

The blinkless engine reads these styles and hands them over as plain values on
each layout box. It never scrolls, never reads the pointer, and never starts a
clock. This program does those jobs. It opens the fixture
`blinkless/testdata/css/v012/scroll/page.html` in a window and, on every frame,
turns the scroll offset and pointer position into progress numbers, then writes
them back into the page as style patches.

What you see when you scroll:

- the red bar at the top grows with page scroll progress
- the three view cards fade and slide in as they enter the window
- the pointer boxes fill up as the mouse moves across them
- the trigger boxes turn green while their trigger is on
- the event boxes turn green after a click and off on the end event
- the interest cards turn amber after the hover delay and clear after the end delay
- the inert panel gets no pointer hits or focus, so clicking it does nothing
- clicking a box that declares `interactivity` in its own style outlines it in
  blue until the next click, so a click on the auto panel is visible. Boxes
  that do not declare it, such as a button, are never outlined

Run it:

    go run ./examples/blinkless-scroll
    go run ./examples/blinkless-scroll -png out.png -scroll 900

`-png` lays the page out, applies the timelines at the scroll offset you give,
and writes one picture of the whole page without opening a window. Flags:
`-width`, `-height`, `-png`, `-scroll`. A page path can be passed as the last
argument.

Test it:

    go test ./examples/blinkless-scroll

Files: `host.go` is the engine, `progress.go`, `scope.go`, `pointer.go`,
`trigger.go`, `events.go`, `interest.go` and `inert.go` hold one rule each,
`styles.go` and `apply.go` write results into the page, `adapter.go` reads
the engine's boxes, `window.go` wires input, and `main.go` is the command.

Trigger sources: a dashed name means the timeline that name resolves to from
the trigger's box (own, ancestor, or hoisted by timeline-scope). `view()` means
the box's own view timeline on the y axis with no insets. `scroll()` and `auto`
mean the page scroll timeline on y. `none` never fires. With several sources,
the first one that resolves wins.

Demo flags. Neither moves your mouse or touches another window.

- `-demo [-demo-out DIR]` plays a scripted session with no window, writes one
  PNG per step into DIR, and builds an mp4 when `ffmpeg` is on PATH.
- `-demo-window [-demo-out DIR]` opens the real window and plays the same
  script inside it, then closes the window. The window runs the script and logs
  each step. Frames and the mp4 are rendered after it closes, so saving
  pictures does not slow the window down. The scripted cursor is not drawn in
  the live window, only in the saved frames.

The two scroll bars at the top are `position: fixed`, so they stay on screen
while the page scrolls. The pointer boxes only follow the pointer while it is
over the box.
