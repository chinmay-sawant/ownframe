# go-gpui

Go library that opens an HTML template in a window. The sign-in program lives in `examples/login` and is one caller of the library.

## Where code lives

Package `gpui` at the module root is the public API. `New`, `SetData`, and `Handle` build a page. `Config.Theme` and `SetTheme` layer an extra stylesheet after the template's own styles. `Page.SetTick` registers a function the window calls once per frame before drawing; `Serve` does not tick. `Run` opens the desktop or wasm window. `Serve` shows the picture page. `BindMobile` registers the phone view. The root files are aliases and those entry points. The page implementation is `internal/page`.

`internal/window` is the Ebiten loop for desktop, phone, and `GOOS=js GOARCH=wasm`. `browser/index.html` loads that wasm build. `scripts/browser.sh` builds and serves it.

`internal/web` is the `go run ./examples/login -web` page. It encodes PNG for `GET /frame.png`.

`internal/host` defines `Screen`. The window and web packages call that interface. They do not import package `gpui`.

`internal/clipboard` reads and writes the OS clipboard, and keeps an in-memory copy. Tests call `UseMemory`.

`internal/filepick` opens the desktop file dialog. Linux runs `zenity`, `qarma`, `matedialog`, or `kdialog`, and under WSL it runs the Windows dialog through `powershell.exe` or `pwsh.exe` first, found on `PATH` or under `/mnt/<drive>/Windows`, and converts the path with `wslpath`. Windows calls `comdlg32!GetOpenFileNameW`; macOS runs `osascript`. `Run` installs it on the page. wasm, mobile, and `Serve` have no dialog, and a file input there takes a typed name. Tests install a fake with `page.InstallPicker`. `GPUI_FILEPICK_DEBUG=1` prints fallback reasons to stderr.

`internal/render` paints the screen. `Redraw` calls `render.DisplayList` first. A page `render.Replayable` accepts keeps a `layout.Display`; any other page calls `render.Paint`, which is `html.Parse`, `css.Apply`, and `layout.Lay`. Neither path writes a PDF. `render.State.Theme` hands the page theme to `css.Options.Extra`, which applies it after the template's own styles.

`internal/replay` draws the display list on the Ebiten canvas: fills with circular corners, axis-aligned border lines, grid runs, and shaped text. `internal/window/draw.go` picks the display list or the fallback image; a fallback frame shows a `bitmap fallback` badge in its top-right corner.

`internal/frame` finds fill and text operations in a `layout.Display` by hit box and colour, so a `Page.SetTick` callback can animate without a Redraw. `documentation/frames.md` has the call shapes and the cost.

`internal/ipc` is in-process `Send`, `Listen`, `Handle`, and `Request`. `ipc.go` re-exports them.

`Page.Load`, `Back`, `Forward`, `Route`, and `HTML` live in `internal/page`. `nav.go` exports `ErrNoHistory`.

`internal/crash` writes a local report. `Run` and `BindMobile` recover a panic and return that path. `crash.go` exports `Report` and `SetCrashDir`.

`internal/fetch` sends one http or https request and stores no cookies. `fetch.go` exports `Fetch`, `XHR`, `FetchResponse`, and `ErrScheme`.

`github.com/chinmay-sawant/gowkhtmltopdf` parses the HTML, applies the CSS, and lays the page out. A replayable page keeps `layout.Display` operations and no picture; any other page paints `Page.Image`. The window replays or blits accordingly. Text replay needs Ebiten v2.10.4 or newer, because gowkhtmltopdf requires `go-text/typesetting` v0.3.4 and older Ebiten builds its font face without the lookup cache v0.3.4 added. `go.mod` replaces the module with `../gowkhtmltopdf` until upstream carries `Display.Boxes`; drop the replace and bump the pin then.

`examples/music` is shared example support: it searches the Openverse API for royalty-free MP3s, caches downloads, decodes MP3 and WAV, and plays through Ebiten audio. The audio player and Spotify player examples import it; package `gpui` does not. `Run` and `BindMobile` create the Ebiten audio context, because Ebiten v2.10 does not; `Serve` has none. `examples/audio-player` and `examples/spotify-player` use `Page.SetTick` to move the seek bar and equalizer from the audio position.

Read `documentation/features.md` before changing paint, theming, frames, IPC, navigation, crash reports, fetch, or the clipboard. The call shapes and the limits are in the other files under `documentation/`.

A page taller or wider than the window scrolls on the mouse wheel. A picture that matches the window stays at one CSS pixel per window pixel. The sign-in example accepts the email `secret` and the password `secret`.

## Go file size

Every Go file is at most 2000 characters. Count with `wc -m`. That number includes newlines. When a file would exceed 2000 characters, split it inside the same package and keep the behavior. One file in the package carries the package comment. The other files start with the package clause.

This limit applies to Go files only. HTML, CSS, Markdown, and scripts have no character cap.

## Writing

Load the `unslop` skill before writing prose: replies, docs, commit messages, and PR or issue text. Apply its fixes before sending.

## Before you finish

Run `gofmt` on every Go file you edit. From this directory, run `make test`. It calls `go test -p 1 ./...` so the packages do not all build and run at once. Raise the limit when a faster run is worth the load: `make test TEST_P=4`.

Leave `~/.Xauthority` untouched. The window toolkit logs a missing authority file and still opens the window.

Keep clipboard tests off the desktop clipboard. Call `clipboard.UseMemory` in those tests.
