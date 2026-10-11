# v0.0.2 plan

Recorded 2026-10-03. The pipeline plans are against `d6026e0`; devtools, hot
reload, and the gap list are against `5a428f4`, which moved docs and examples
only.

Status, 2026-10-05: all six workstreams shipped on `feature/v0.0.2` and every
checklist is closed. The devtools overlay became a right dock with Elements,
Frame, and Ops tabs ([devtools.md](devtools.md)), and the guides named in the
definition of done are linked from `documentation/README.md`. The engine branch
is pushed and `go.mod` pins it with no local `replace`; the release state is in
`../../PHASES.md`.

The follow-up checklist for desktop IME, accessibility, and the remaining
platform gaps is [accessibility-desktop-ime.md](accessibility-desktop-ime.md).

Six workstreams land in v0.0.2. Two come from one root cause: `Page.Redraw`
keeps nothing it already computed, so every pointer move, every window drag, and
every click reparses the HTML, recollects the stylesheets, relays the page out,
and repaints it end to end. The other four make the running page easier to see,
quicker to edit, and richer to use, and they close out the still-absent list
from `documentation/features.md`.

| Plan | Problem |
|---|---|
| [dynamic-resize.md](dynamic-resize.md) | Drag the window and the inner layout stops responding. The placement is laid out once at the clamped frame size, so text does not rewrap, `@media` and `vw` do not follow, and `:hover` is resolved against boxes from the previous layout. |
| [incremental-repaint.md](incremental-repaint.md) | Click the counter in `examples/platform` and the whole 640x480 page is reparsed, relaid out, and replayed to change one number. |
| [devtools.md](devtools.md) | The boxes, the display operations, and the cost of a frame are invisible in the running window. An overlay shows them, and `Serve` gets a JSON route with the same data. |
| [hot-reload.md](hot-reload.md) | Every example embeds its HTML, so an edit needs a rebuild and a restart. A file-backed page watches its file by default and redraws on change. |
| [electron-gaps.md](electron-gaps.md) | The still-absent list, sorted for this cycle. Drag and drop, printing, and packaging land; IME, accessibility, video, canvas, WebGL, and auto-update are recorded with what each one waits on. |
| [accessibility-desktop-ime.md](accessibility-desktop-ime.md) | Follow-up phases for accessibility, desktop IME, shell integrations, and planning-only browser APIs. |
| [input-interaction.md](input-interaction.md) | Typing can only append, Tab moves nothing, there is no drag selection, context menu, cursor shape, touch scroll, pinch, fullscreen, or page scroll call. The plan adds the caret model first, then the interactions that need it. |

Phase 1 of dynamic-resize and incremental-repaint is the same work, the
pipeline split, so it is written once in `dynamic-resize.md` Phase 1 and
referenced from the other plan. Do that phase first. Devtools Phase 1 turns the
counters incremental-repaint Phase 1 adds into a snapshot a running window can
read; land one shape for both. Hot reload invalidates the same caches the
pipeline split introduces and can start in parallel with it. The gap plans are
independent, except that drag and drop, the input phases, the reload poll, and
the devtools toggle all touch the same `Update` and pointer paths, and both the
context menu and the devtools overlay are shell chrome drawn after the page.
Pinch zoom lands last because it touches pointer mapping, scrollbar math, and
the devtools overlay together.

## Ground rules

Carried from `AGENTS.md`, restated because every phase here touches Go code.

- Every Go file stays at or under 2000 characters, counted with `wc -m`. Split
  inside the same package when a file would go over.
- `gofmt` every file you edit, then run `make test` from the repo root.
- No new modules. The build stays on Ebiten and `gowkhtmltopdf`.
- The `replace` for `../gowkhtmltopdf` stays until upstream carries
  `Display.Boxes`. Phases that need a new engine call land in the sibling
  checkout first and keep the replace.
- The engine work for v0.0.2 goes on its own branch in `gowkhtmltopdf`, the way
  `chore/changes-for-go-gpui` did for v0.0.1. The push order and the pinned
  commit are recorded in `../../PHASES.md`.
- Never touch `~/.Xauthority`. Keep clipboard tests on `clipboard.UseMemory`.

## Definition of done

- Dragging a window edge relayouts the page at the real window size. Text
  rewraps, percentages resolve against the new width, and `@media` blocks flip
  at the size the window actually is.
- A hover, press, or focus change paints the state at the position the element
  has after the relayout, not at the position it had before it.
- A click that changes one counter repaints that counter's box. The rest of the
  frame is byte identical to the frame before the click.
- A relayout during a drag costs one layout pass per committed size, not one
  full parse and repaint per mouse event.
- F12 or Ctrl+Shift+I shows an overlay that outlines the box under the cursor,
  pins one, outlines the display operations, and prints the frame stats. With
  the overlay off, `Page.PNG` bytes and `Page.Generation` are the same as
  before the feature.
- An edit to `Config.File` on disk appears in the open window without a
  restart. A parse error keeps the last good picture, and `-reload=false` turns
  the watch off.
- A file dropped on the window reaches `Handlers.Drop` on desktop and in the
  wasm canvas; `-web` has no window loop and receives none. `Page.PDF` writes a
  file the system reader opens. `scripts/package.sh` leaves a runnable archive
  on Linux.
- Click places the caret at the glyph, typing inserts there, arrows and
  Home/End move it, Shift extends a selection, Tab and Shift+Tab move focus in
  document order, and a drag selects a span. The right click menu, the hover
  cursor, touch scroll, programmatic scroll, and the F11 toggle work on
  desktop. IME stays absent with the design recorded.
- `make test` passes, `go vet ./...` is clean, and
  `GOOS=js GOARCH=wasm go build ./...` still builds.
- `documentation/window.md`, `documentation/frames.md`,
  `documentation/features.md`, and `documentation/theming.md` describe the new
  behaviour, and `examples/resize` shows it on screen. The new guides,
  `documentation/devtools.md`, `documentation/hot-reload.md`,
  `documentation/drag-drop.md`, `documentation/printing.md`,
  `documentation/packaging.md`, and `documentation/interaction.md`, are linked
  from `documentation/README.md`.
