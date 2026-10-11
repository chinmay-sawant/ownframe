# Electron gaps

Recorded 2026-10-03 against `5a428f4`.

The dated inventory below records the original v0.0.2 decision. Current
implementation work and status live in
[`accessibility-desktop-ime.md`](accessibility-desktop-ime.md).

`documentation/features.md:69` keeps the still-absent list, and the v0.0.1
scan is `../v0.0.1/compare.md`. This file answers that list for the v0.0.2
cycle: DevTools, IME, accessibility, video, canvas, WebGL, printing, drag and
drop, auto-update, and packaging. DevTools has its own plan in
[devtools.md](devtools.md). Four of the ten land in v0.0.2, and the rest are
recorded here with what each one waits on, so nobody reads the absence list
and assumes v0.0.2 covers it. The v0.0.1 framing still holds: there is no
Chromium, no V8, and no second process.

## Where each one stands

| Item | v0.0.2 | Today | What it waits on |
|---|---|---|---|
| [DevTools](devtools.md) | in, own plan | Nothing on screen exposes the boxes, ops, or timing. | Library work only. |
| Drag and drop | in | Ebiten delivers dropped files per frame; ownframe never asks. | A handler, a file shape, one example. |
| Printing | in, as PDF export and a save path | The engine dependency writes PDF from HTML; the window never calls it. | `Page.PDF` / `WritePDF`, `SavePDF`, docs. |
| Packaging | in, as scripts and docs | `scripts/browser.sh` and the documented `ebitenmobile bind` are the whole story. | A release script, per-OS archive layout, caveats. |
| Canvas | no, v0.0.3 | The engine ignores `<canvas>`; there is no drawing surface for one. | Engine element support, a Go canvas call. |
| IME | follow-up in progress | Mobile IME is wired. Desktop builds connect Ebiten's `exp/textinput` Composer and paint preedit separately; native input-method verification remains. | Finish platform checks and record supported desktop backends. |
| Accessibility | page tree foundation implemented | The page exposes a semantic tree and control actions. There is no native screen-reader bridge, CSS visibility filtering, or platform verification. | Native adapters and computed-visibility support. |
| Video | no, dies on the no-modules rule | No `<video>`; `examples/music` decodes MP3 and WAV only. | A codec module and A/V sync. |
| WebGL | no | Ebiten owns the graphics context; no public GL API. | An engine and toolkit project. |
| Auto-update | no, not a library job | None. | A product-owned updater, or the OS store. |

## In v0.0.2

### Drag and drop

Ebiten v2.10.4 exposes `ebiten.DroppedFiles()`, an `fs.FS` scoped to the
`Update` frame it is read in. The doc comment on it says it works on desktops
and browsers, and since 2.10 it also implements `io/fs.ReadFileFS`, with an
`AbsPath` on each entry and file on desktops. ownframe never calls it, so a
file dropped on the window is ignored.

Shape: the window reads `ebiten.DroppedFiles()` in `Update`
(`internal/window/game.go:61`) after the pointer pass. An empty FS means no
drop. A non-empty FS becomes a slice of `host.Drop` values, and a screen that
implements an optional `host.Dropper` interface receives them:

```go
type Drop struct {
    Name  string
    Size  int64
    IsDir bool
    Path  string // absolute on desktop, empty in a browser
    Read  func() ([]byte, error) // valid only during the call
}

type Dropper interface {
    Drop(ctx context.Context, files []Drop) error
}
```

The frame scope is the honest limit: a handler that keeps the path can use it
later on desktop, and a handler that keeps the bytes has them. A handler that
returns without reading gets nothing, and the docs say so.

- [x] `internal/host`: `Drop` and `Dropper`, beside `Ticker`
      (`internal/host/screen.go:14`).
- [x] `internal/window`: a `drop.go` pass that builds the values and calls
      the screen, plus the `fs.WalkDir` case for a dropped directory. Pass
      the directory as one entry with `IsDir`, not a flattened tree.
- [x] `internal/page`: `Handlers.Drop` and `Page.Drop`, and a redraw after
      the handler, the way `Click` redraws (`internal/page/page_click.go:40`).
      An error returns before the redraw.
- [x] Root: re-export `Drop` and the handler field.
- [x] New `examples/drop`: drop a PNG or JPEG and the page shows it through
      `SetImage`, drop a `.txt` and the first lines appear as text. `-web` on
      port 8124. A test drives the fake FS with `testing/fstest.MapFS`.
- [x] `documentation/drag-drop.md`, a row in `documentation/README.md`, and a
      note in `documentation/features.md`.

Exit: a file dropped on the desktop window reaches the page handler with its
name and bytes, the wasm canvas does the same, and a dropped directory
arrives as one entry. `-web` has no window loop, so it receives no drops and
says so in [drag-drop.md](../../documentation/drag-drop.md).

### Printing

The engine dependency already writes PDF. `gowkhtmltopdf.Document.WritePDF`
and `Document.PDF` (`document.go:196`, `document.go:250`) take a page source,
page size, margins, headers, and a PDF profile, and the profile supports
PDF/A and PDF/UA (`document.go:116`). ownframe imports `css` and `layout` from
that module and stops before the PDF step. `../../AGENTS.md` records the
boundary: neither render path writes a PDF.

Shape: `Page.PDF(ctx, opts) ([]byte, error)` and
`Page.WritePDF(ctx, w, opts) error`, plus `SavePDF(ctx, path, opts) error`.
The render source is the last executed template output, `p.source`, with the
current theme injected as a `<style>` element in the head. The engine accepts
a user style sheet key as inert (`internal/settings/settings.go:322`), so
injection is the working route, not an engine option. The defaults are A4,
10 mm margins, no header or footer.

Printing lives beside saving: `Page.Print(ctx, opts)` writes a temp PDF and
hands it to the OS print path. On Linux that is `lp` or `xdg-open`, on macOS
`osascript` or `open`, on Windows PowerShell `Start-Process -Verb Print`.
`internal/filepick` is the precedent for a platform helper that shells out and
logs why it fell back (`OWNFRAME_FILEPICK_DEBUG=1`); printing gets
`OWNFRAME_PRINT_DEBUG=1`. When no helper is available, `Print` returns a typed
error that names `SavePDF`.

- [x] `internal/page/print.go` with the three calls. The opts struct carries
      page size, margin, and profile strings; a zero value means the engine
      defaults.
- [x] The theme injection reuses the source the next `Redraw` would execute,
      so the PDF matches the window's styles where the engine's print
      pagination allows.
- [x] `internal/print` for the OS helper, in the `filepick` shape: Linux,
      Windows, macOS, and a stub file for wasm and mobile. `Print` is a
      no-op with `ErrNoPrinter` there.
- [x] Tests: `PDF` starts with `%PDF-`, a two-paragraph page makes one page,
      a wrong profile is an error, and the helper picks the right command on
      each GOOS with a fake runner.
- [x] New `examples/print`: a report-like page with a Save PDF button and a
      Print button. `-web` on port 8125 saves through `GET /pdf`, which
      returns the bytes with `Content-Type: application/pdf`.
- [x] `documentation/printing.md`, a `documentation/web.md` route row, a
      `documentation/README.md` row, and a features.md section.
- [x] Record in `../../PHASES.md` that the live window still cannot print
      its own display list; the PDF is a re-render from source.

Exit: `SavePDF` writes a file the system PDF reader opens, `Print` reaches
the platform print path (a dialog on macOS and Windows, `lp` or an opened
viewer on Linux), and `GET /pdf` returns the same bytes.

### Packaging

`scripts/browser.sh` builds the wasm bundle and serves it, and the README
documents `ebitenmobile bind` for Android and iOS (`README.md:87`). Desktop
distribution is `go install` or `go run`, and there is no archive, bundle, or
checksum story.

Shape: `scripts/package.sh <example>` builds a release binary for the host OS
and arch with `-trimpath` and `-ldflags "-s -w"`, then lays out one archive:

- Linux: `tar.gz` with the binary, a `.desktop` file, and a README.
- Windows: `zip` with the `.exe`.
- macOS: `zip` with an `.app` directory; the script runs only on macOS
  because the toolkit needs the platform build tools.
- wasm: `zip` with `ownframe.wasm`, `wasm_exec.js`, and `browser/index.html`.

A `SHA256SUMS` file sits next to the archives. Custom icons need platform
resource tools, so v0.0.2 ships the default icon and records icons as
follow-up. Signing and notarization stay outside the script; the docs say how
each OS treats an unsigned archive.

- [x] `scripts/package.sh`, no new modules, no network calls.
- [x] `documentation/packaging.md`: the archive layouts, the signing and
      notarization caveats, and that auto-update is not included.
- [x] `documentation/README.md` row and a paragraph in `README.md` under the
      run modes.
- [x] Tests: `scripts/package.sh -n <example>` (dry run) lists the same files
      the docs describe. A shell test so CI can check layout without
      building.

Exit: running the script on Linux leaves a `tar.gz` with a binary that opens
a window and a `.desktop` entry that launches it.

## After v0.0.2

### Canvas

The engine has no canvas drawing surface. A bare `<canvas>` produces no box
or operation; inside a flex container the box is laid out and nothing paints
it. There is no JavaScript engine, so a canvas can only ever be Go-drawn. The
first slice is small and belongs in the engine and then here: treat
`<canvas id width height>` as a replaced element with a box, and let the page
fill it with `Page.SetCanvas(id string, img image.Image)` beside `SetImage`.
A `SetTick` that draws and calls `SetCanvas` then animates it. The HTML
Canvas 2D API does not follow from that, and the guide should say so.

### IME

The original plan assumed committed text was the only input available. The
current desktop window creates an Ebiten `exp/textinput.Composer` for the
focused field. The page paints preedit as an underlined temporary run and
leaves the stored value and edit handlers alone until commit. Commits use the
existing `Type` or `IMEReplace` path. The code has unit coverage; actual
composition still needs checks with native input methods on supported desktop
systems. See the live rows in
[`accessibility-desktop-ime.md`](accessibility-desktop-ime.md).

### Accessibility

The window is one Ebiten canvas. The page now exposes a semantic tree and
actions for supported controls, but ownframe has no native screen-reader
bridge. The first tree uses HTML parentage and joins bounds through unique
element IDs. It does not filter CSS-hidden nodes. Separate adapters for
AT-SPI, UI Automation, and NSAccessibility still need platform checks.
PDF/UA output does not expose the live window. The phase order and acceptance
checks are in [`accessibility-desktop-ime.md`](accessibility-desktop-ime.md).

### Video

`examples/music` decodes MP3 and WAV and plays through Ebiten audio
(`../../AGENTS.md`). Video needs a demuxer and a decoder: none in the Go
standard library, and a codec module breaks the no-new-modules ground rule in
[README.md](README.md). Frame pacing and A/V sync sit on top. If it ever
ships, it belongs in an example with the decoder at arm's length and the
library taking frames, so the core module stays clean.

### WebGL

Ebiten owns the graphics context on every backend and exposes no GL calls.
WebGL without a JavaScript engine is a Go GL binding plus engine work, which
is a different project. The same call applies to WebGPU. No.

### Auto-update

A library cannot replace its own host binary portably. The work is a signed
manifest, a download, an atomic swap, and a restart, and it fights the OS
stores and package managers that own the install. The packaging guide tells a
product how to do it in its own `cmd/updater`; this library does not grow the
feature.

## Risks and limits

- The four in-v0.0.2 items are deliberate scope, not a promise that the
  other six arrive soon. Video and WebGL need codec or graphics work. The
  first accessibility tree is in ownframe; native adapters and CSS visibility
  still need work. IME needs window and page work on top of Ebiten's
  `exp/textinput`, and is the closest of the six.
- Printing re-renders from source, so the paper page and the window page
  differ in pagination and in anything the print path does not implement.
  The guide has to show both pictures.
- Drag and drop is frame-scoped. A handler that stores a path works on
  desktop. A handler that stores a reader does not, and the API cannot hide
  that.
- The packaging script runs per OS because the desktop toolkit is per OS.
  A single Linux runner cannot produce the macOS archive.
- Engine work for canvas goes on the sibling `gowkhtmltopdf` branch, the way
  `chore/changes-for-go-gpui` did for v0.0.1. Nothing here adds a module.
