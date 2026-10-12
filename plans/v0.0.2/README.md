# v0.0.2 plan

Recorded 2026-10-03 against `d6026e0`.

Two workstreams land in v0.0.2. Both come from one root cause: `Page.Redraw`
keeps nothing it already computed, so every pointer move, every window drag, and
every click reparses the HTML, recollects the stylesheets, relays the page out,
and repaints it end to end.

| Plan | Problem |
|---|---|
| [dynamic-resize.md](dynamic-resize.md) | Drag the window and the inner layout stops responding. The placement is laid out once at the clamped frame size, so text does not rewrap, `@media` and `vw` do not follow, and `:hover` is resolved against boxes from the previous layout. |
| [incremental-repaint.md](incremental-repaint.md) | Click the counter in `examples/platform` and the whole 640x480 page is reparsed, relaid out, and replayed to change one number. |

Phase 1 of both plans is the same work, the pipeline split, so it is written
once in `dynamic-resize.md` Phase 1 and referenced from the other plan. Do that
phase first. The rest of the two plans are independent and can run in parallel.

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
  `chore/changes-for-go-gpui` did for v0.0.1. Push order and the pinned commit
  are listed in `../../PHASES.md` under Pending, integration.
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
- `make test` passes, `go vet ./...` is clean, and
  `GOOS=js GOARCH=wasm go build ./...` still builds.
- `documentation/window.md`, `documentation/frames.md`,
  `documentation/features.md`, and `documentation/theming.md` describe the new
  behaviour, and `examples/resize` shows it on screen.
