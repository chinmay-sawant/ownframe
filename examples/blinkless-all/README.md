# blinkless-all

One window for the 61 CSS properties that blinkless had not yet implemented.

The top of the window states the catalog counts: how many properties the
catalog lists, how many are implemented, and how many of these 61 are
implemented, partial or unsupported now. The numbers come from
`blinkless/testdata/css/catalog/properties.json`, so they move when a row is
flipped.

A family is a group of properties that one example covers. Two example boxes
can stand for eight properties, so the number on a tab is the number of
properties, not the number of boxes. Each row names the properties it also
covers.

| Tab | Properties |
|---|---|
| Navigation and UI | 20 |
| Draft borders and shapes | 18 |
| Motion path | 6 |
| Fragmentation and generated content | 7 |
| Reading order and misc | 5 |
| SVG leftovers | 5 |

Each tab button loads a page that holds only that tab's section. The other five
sections are not in the document, so nothing of them is laid out or painted.

Run it:

    go run ./examples/blinkless-all
    go run ./examples/blinkless-all -tab motion
    go run ./examples/blinkless-all -tab shapes -png shapes.png

`-png` lays the tab out and writes one picture without opening a window. Flags:
`-width`, `-height`, `-tab`, `-png`.

Test it:

    go test ./examples/blinkless-all

The engine stores each property and exports it as `Box.Declared` (or acts on it
in layout, for border clips, motion path, bookmarks and block-ellipsis). Where
the browser would act on a property, this program decides what happens in plain
functions (`nav_focus.go`, `read_order.go`, `motion_host.go`, `frag_host.go`)
that the tests call without a window. On the Navigation and UI tab the examples are live: arrow keys and on-screen buttons move focus, carets blink, fields take typing, sliders drag, the mock window drags and the image frames cycle (`nav_session.go`, `nav_draw.go`). `-tab nav -demo -demo-out DIR` plays a scripted session without a window and writes one PNG per step. A row stays `partial` in the catalog
until its behavior is checked by hand in the window.

Files: `app.go`, `catalog.go`, `page.go`, `tabs.go` and `main.go` are the shell.
Each tab has a `tab_<id>.go` and its own host and test files.

`-web ADDR` also has the live sessions (`web_live.go`): hover, click, key and
typed text from the web routes drive the same sessions the window uses. A click
carries the box under the point, not the point, so a slider click lands
mid-track, and dragging needs the window. `scripts/gui-drive.sh` in the
repository root of ownframe drives the real window; see
`documentation/testing.md`.
