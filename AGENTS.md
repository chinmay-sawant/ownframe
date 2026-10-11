# ownframe

Go library that opens an HTML template in a window. The sign-in program lives in `examples/login` and is one caller of the library.

## Where code lives

Package `ownframe` at the module root is the public API. `New`, `SetData`, and `Handle` build a page. `NewWithOptions` applies `Option` values first; `WithPerf(true)` opts into pipeline timing, dirty counters, and alloc tracking, off by default so shipped apps pay nothing. `Config.Theme` and `SetTheme` layer an extra stylesheet after the template's own styles. `Page.SetTick` registers a function the window calls once per frame before drawing; `Tick` also blinks the focused field's caret; `Serve` does not tick. `Run` opens the desktop or wasm window. `Serve` shows the picture page. `BindMobile` registers the phone view. The root files are aliases and those entry points. The page implementation is `internal/page`. `examples/` is its own Go module, `github.com/chinmay-sawant/ownframe/examples`, so the library module zip carries no examples. A committed `go.work` joins both modules. `examples/go.mod` replaces `github.com/chinmay-sawant/ownframe` with the parent checkout until a release tag exists. `make test` and `make build` run both modules. A release bumps the examples require to the new tag, drops the pin once the tag is pushed, and tags `examples/v0.0.2`.

`examples/telegram` is the phone-first demo: a Telegram-like chat list with search, contacts, settings with a dark theme, and one conversation with a composer. `examples/telegram/mobile` is the `ebitenmobile` package and `examples/telegram/android` is a Gradle project whose `:telegram` module wraps the generated `telegram.aar`. `MainActivity` intercepts the Android back key and calls the exported `mobile.Back`; it polls `mobile.AttachRequest` to open the camera or the photo picker and returns the bytes through `mobile.SetPhoto`; it reports the status bar, cutout, and IME insets through `mobile.SetInsets`, and `App.Tick` applies the queued work on the game loop. The chat top bar and composer are absolutely positioned and counter-moved by the scroll offset through `Page.SetScrollWindow`, so they stay pinned while the messages scroll and painted after the messages so a scrolled thread cannot cover them. `scripts/android.sh` binds, builds the debug APK, and with `install` runs `adb install -r`; `examples/telegram/android/README.md` has the SDK, NDK, and device steps. Do not export `GOWORK` for the bind: gomobile copies the packages into a temporary module and builds there, and the workspace file would disown that module.

`internal/window` is the Ebiten loop for desktop, phone, and `GOOS=js GOARCH=wasm`. `Update` runs devtools sync, the phone text input session, keys, resize, the passthrough update, pointer, drop, tick, a reload poll, sync, the page scroll request, a devtools refresh, and wheel. A drag relayouts at most every 100 ms and commits the final size; a partial repaint replays one dirty rect into a persistent buffer sized to the content; Tab and Shift+Tab move focus through `host.Focuser`; a right click draws a shell context menu; touch drags scroll and pinch zooms; a press held within 8 px for about 450 ms fires `host.LongPresser`; a claimed hold keeps routing movement to the selection instead of scrolling; F11 toggles fullscreen; F12 or Ctrl+Shift+I toggles a devtools dock on the right, 340 px wide and resizable on its left edge, with Elements (JSON element properties), Frame (window and pipeline counters), and Ops (the display operation list with outlines) tabs; `Config.DevTools` and `SetDevTools` open it without a key. `WindowOptions.Perf` turns on frame-ring and runtime sampling for the PERFORMANCE and MEMORY rows; without it those rows read `-`. `browser/index.html` loads that wasm build. `scripts/browser.sh` builds and serves it.

`internal/web` is the `go run ./examples/login -web` page. It encodes PNG for `GET /frame.png`, serves `GET /pdf` when the screen can render PDF bytes, serves `GET /debug/state` with boxes, stats, and operation counts, and calls the reloader before it answers when a page watches files. `ServeWithOptions` with `ServeOptions{Perf: true}` adds `GET /debug/pprof/*`; plain `Serve` exposes no profiler.

`examples/spotify-player` is a multi-screen demo. The sidebar switches `View.Nav` between home, search, library, liked, browse, radio, queue, and profile. The six library screens each get a `screen_<name>.go` handler for their `components/` fragment; `build.go` wires the fragments and `screens.go` routes each screen's action prefix to its `click<Name>` handler.

`examples/desktop-cat` is the transparent, click-through overlay: `RunWithOptions` with `WindowOptions{Transparent, Borderless, Floating, FixedSize, BottomRight, Margin}`, `cat.Draggable` hit-tests the current PNG's alpha so a four-pixel drag moves the window, and `cat.Interactive` adds the bubble, so every other pointer event passes through. The `cat` package embeds 30 expression PNGs, switches one every eight seconds, and draws an HTML/CSS bubble. `POST /notify` on 127.0.0.1:6969 takes an expression and message (`GET /expressions`, `GET /state`), and `.agents/skills/catnotify` sends to it. `-web` serves a still preview on 8128; the Windows-only `nowplaying` reader samples Chrome's system media session every two seconds (`-media`, `-media-snapshot`).

`examples/wispr-flow-dashboard` is a 13-page app shell. `app/app.go` parses one template with a collapsible sidebar and the active page fragment; `View` carries every page's data, and a nav click sets `ActivePage`. Each page is its own package exposing `HTML` and `CSS`, listed in `app/pages.go` in sidebar order; `app/actions.go` handles the sidebar, tabs, streak chevrons, share, and get-app controls.

`internal/print` hands a PDF to the OS print path. Linux runs `lp` or `xdg-open`, macOS `osascript` or `open`, Windows PowerShell `Start-Process -Verb Print`; wasm, Android, and iOS report `ErrNoPrinter`. `OWNFRAME_PRINT_DEBUG=1` logs fallback reasons. `Page.PDF`, `WritePDF`, `SavePDF`, and `Print` live in `internal/page`.

`internal/host` defines `Screen` and the optional interfaces a screen can add: `Ticker`, `Inspector`, `Reloader`, `Dropper`, `Focuser`, `ContextMenu`, `CursorShape`, `ScrollRequester`, `IME`, and `LongPresser`. The window and web packages call those. They do not import package `ownframe`.

`internal/clipboard` reads and writes the OS clipboard, and keeps an in-memory copy. Tests call `UseMemory`.

`internal/filepick` opens the desktop file dialog. Linux runs `zenity`, `qarma`, `matedialog`, or `kdialog`, and under WSL it runs the Windows dialog through `powershell.exe` or `pwsh.exe` first, found on `PATH` or under `/mnt/<drive>/Windows`, and converts the path with `wslpath`. Windows calls `comdlg32!GetOpenFileNameW`; macOS runs `osascript`. `Run` installs it on the page. wasm, mobile, and `Serve` have no dialog, and a file input there takes a typed name. Tests install a fake with `page.InstallPicker`. `OWNFRAME_FILEPICK_DEBUG=1` prints fallback reasons to stderr.

`internal/render` owns the cached document. A cache holds the parsed HTML tree and the styled CSS document from one parse and one cascade, keyed on the executed source; a size or state change calls the engine's `css.Relayout` and lays the cached document out again, so a drag or a hover never reparses or recollects. `Redraw` keeps a `layout.Display` for a page `render.Replayable` accepts and a picture for any other page. Neither path writes a PDF; `Page.PDF` re-renders from source. `render.State.Theme` hands the page theme to `css.Options.Extra`, which applies it after the template's own styles.

`internal/replay` draws the display list on the Ebiten canvas: fills with circular corners, axis-aligned border lines, stroked diagonal segments, grid runs, and shaped text. Images draw with linear filtering. `internal/emoji` paints thirty-one color emoji as inline images backed by bundled Twemoji PNGs, resolved through the page image lookup. `DrawRect` replays only the operations that intersect a dirty rect, which the window keeps in a persistent buffer sized to the content rather than the window. `internal/window/draw.go` picks the display list or the fallback image; a fallback frame shows a `bitmap fallback` badge in its top-right corner.

`internal/frame` finds fill and text operations in a `layout.Display` by hit box and colour, so a `Page.SetTick` callback can animate without a Redraw. `documentation/frames.md` has the call shapes and the cost.

`internal/ipc` is in-process `Send`, `Listen`, `Handle`, and `Request`. `ipc.go` re-exports them.

`Page.Load`, `Back`, `Forward`, `Route`, and `HTML` live in `internal/page`. `Config.File` and `ThemeFile` read from disk, and a page watches them by default; `PollReload` swaps the template without touching history, and `-reload=false` turns the watch off. `Redraw` counts parses, cascades, layouts, and repaints for `host.Inspector`. `Config.Perf` and `SetPerf` gate the extra timing, dirty counters, and alloc tracking behind an opt-in; untracked redraws leave those `Stats` fields at zero. A click dirties the changed box, and `TakeDirty` hands the window the region to repaint. The caret, the selection range, focus order, context menu state, cursor shape, and scroll requests are page state too. The caret span takes the field's text color through `currentColor`, so it shows on a dark theme, and a whole-value selection paints a selection span, with no field background, so only the selected text is highlighted. `internal/textrun` turns a point inside a shaped text run into a rune offset, so a click can place the caret at the glyph. `nav.go` exports `ErrNoHistory`.

`internal/crash` writes a local report. `Run` and `BindMobile` recover a panic and return that path. `crash.go` exports `Report` and `SetCrashDir`.

`internal/fetch` sends one http or https request and stores no cookies. `fetch.go` exports `Fetch`, `XHR`, `FetchResponse`, and `ErrScheme`.

`github.com/chinmay-sawant/blinkless` `v0.0.0-20261010174808-34d93862f9ed` parses the HTML, applies the CSS, and lays the page out. That require is the only engine pin. There is no replace. `layout.DisplayList` is the placement. A replayable page keeps those operations and no picture. Any other page replays the same list into `Page.Image`. The window replays or blits accordingly. blinkless does not write a PDF and does not paint a page bitmap. `Page.PDF`, `WritePDF`, `SavePDF`, and `Print` return `ErrNoPDF`. Text replay needs Ebiten v2.10.4 or newer, because blinkless requires `go-text/typesetting` v0.3.4 and older Ebiten builds its font face without the lookup cache v0.3.4 added. `css.Relayout` re-places an already styled document at a new viewport and state.

`examples/music` is shared example support: it searches the Openverse API for royalty-free MP3s, caches downloads, decodes MP3 and WAV, and plays through Ebiten audio. The Spotify player example imports it; package `ownframe` does not. `Run` and `BindMobile` create the Ebiten audio context, because Ebiten v2.10 does not; `Serve` has none. `examples/spotify-player` uses `Page.SetTick` to move the seek bar and equalizer from the audio position.

Read `documentation/features.md` before changing paint, theming, frames, IPC, navigation, crash reports, fetch, the clipboard, printing, drag and drop, or the devtools overlay. The call shapes and the limits are in the other files under `documentation/`.

A page taller or wider than the window scrolls on the mouse wheel. A picture that matches the window stays at one CSS pixel per window pixel. The sign-in example accepts the email `secret` and the password `secret`.

## Go file size

Every Go file is at most 2000 characters. Count with `wc -m`. That number includes newlines. When a file would exceed 2000 characters, split it inside the same package and keep the behavior. One file in the package carries the package comment. The other files start with the package clause.

This limit applies to Go files only. HTML, CSS, Markdown, and scripts have no character cap.

## Writing

Load the `unslop` skill before writing prose: replies, docs, commit messages, and PR or issue text. Apply its fixes before sending.

## Builds

Do not run `go build ./...`. Every run links an executable for each example, takes about a minute, and leaves the binaries in this directory; the default `-p 24` loads the machine. For the same compile check, run `make build`. It calls `go vet -p 1 ./... ./examples/...`, which compiles every package including the examples, links nothing, and writes nothing. Raise the limit with `make build BUILD_P=4`. To check one example, name it: `go vet -p 1 ./examples/login`.

## Before you finish

Run `gofmt` on every Go file you edit. From this directory, run `make test`. It calls `go test -p 1 ./... ./examples/...` so the packages do not all build and run at once. Raise the limit when a faster run is worth the load: `make test TEST_P=4`.

Leave `~/.Xauthority` untouched. The window toolkit logs a missing authority file and still opens the window.

Keep clipboard tests off the desktop clipboard. Call `clipboard.UseMemory` in those tests.
