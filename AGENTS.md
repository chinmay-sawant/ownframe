# ownframe

Go library that opens an HTML template in a window. The sign-in program lives in `examples/login` and is one caller of the library.

## Where code lives

Package `ownframe` at the module root is the public API. Keep it there: do not move the library into `api/`. `New`, `SetData`, and `Handle` build a page. `ownframe.Page` holds an unexported `*page.Page` and delegates an explicit method set. `NewWithOptions` applies `Option` values first; `WithPerf(true)` opts into pipeline timing, dirty counters, and alloc tracking, off by default so shipped apps pay nothing. `Config.Theme` and `SetTheme` layer an extra stylesheet after the template's own styles. `Page.SetTick` registers a function the window calls once per frame before drawing. `Page.Testing().Tick` runs that callback and blinks the focused field's caret. The window calls the inner screen. `Serve` does not tick. `Run` opens the desktop or wasm window. `NewGame` returns that same Ebiten game for tests that drive the loop without `Run`. `Serve` shows the picture page. `BindMobile` registers the phone view. 1.0 native hosts are desktop `Run` (Linux, Windows, macOS) and Android `BindMobile`. Wasm and `-web` are previews, not 1.0 hosts. The page implementation is `internal/page`. `examples/` is its own Go module, `github.com/chinmay-sawant/ownframe/examples`, so the library module zip carries no examples. A committed `go.work` joins both modules. `examples/go.mod` requires `github.com/chinmay-sawant/ownframe` at the last published tag (`v0.0.2`) until `v0.0.3` is tagged, and `replace`s it with `../` for this checkout. The current library cycle is `v0.0.3`. `make test` and `make build` run both modules. A release tags the library, bumps the examples require to that tag, and drops the relative replace before tagging `examples/vX.Y.Z`; a tagged examples module cannot keep `replace => ../`.

`examples/telegram` is the phone-first demo: a Telegram-like chat list with search, contacts, settings with a dark theme, and one conversation with a composer. `examples/telegram/mobile` is the `ebitenmobile` package and `examples/telegram/android` is a Gradle project whose `:telegram` module wraps the generated `telegram.aar`. `MainActivity` intercepts the Android back key and calls the exported `mobile.Back`; it polls `mobile.AttachRequest` to open the camera or the photo picker and returns the bytes through `mobile.SetPhoto`; it reports the status bar, cutout, and IME insets through `mobile.SetInsets`, and `App.Tick` applies the queued work on the game loop. The chat top bar and composer are absolutely positioned and counter-moved by the scroll offset through `Page.SetScrollWindow`, so they stay pinned while the messages scroll and painted after the messages so a scrolled thread cannot cover them. `scripts/android.sh` binds, builds the debug APK, and with `install` runs `adb install -r`; `examples/telegram/android/README.md` has the SDK, NDK, and device steps. Do not export `GOWORK` for the bind: gomobile copies the packages into a temporary module and builds there, and the workspace file would disown that module.

`internal/window` is the Ebiten loop for desktop, phone, and `GOOS=js GOARCH=wasm`. `Update` runs devtools sync, the phone text input session, keys, resize, the passthrough update, pointer, drop, tick, a reload poll, sync, the page scroll request, a devtools refresh, and wheel. A drag relayouts at most every 100 ms and commits the final size; a partial repaint replays a dirty rect into a 256 px tile grid sized to the content; a desktop window that stays clean for five frames and sees no input parks in the event wait and wakes on input or at the deadline the page declared with `Page.WantFrameAt`, falling back to a 125 ms keep-alive tick, and `WindowOptions.NoIdlePace` opts out; while the devtools dock is open the tile repaint and blit clip to the strip the panel leaves visible; Tab and Shift+Tab move focus through `host.Focuser`; a right click draws a shell context menu; touch drags scroll and pinch zooms; a press held within 8 px for about 450 ms fires `host.LongPresser`; a claimed hold keeps routing movement to the selection instead of scrolling; F11 toggles fullscreen; F12 or Ctrl+Shift+I toggles a devtools dock on the right, 340 px wide and resizable on its left edge, with Frame (window and pipeline counters, the tab the dock opens on), Elements (JSON element properties), and Ops (the display operation list with outlines) tabs; `Config.DevTools` and `Page.SetDevTools` open it without a key. `WindowOptions.Perf` turns on frame-ring and runtime sampling for the PERFORMANCE and MEMORY rows; without it those rows read `-`. `browser/index.html` loads that wasm build. `scripts/browser.sh` builds and serves it.

`internal/web` is the `go run ./examples/login -web` page. It encodes PNG for `GET /frame.png`, and calls the reloader before it answers when a page watches files. `ServeWithOptions` registers `GET /debug/state` only with `ServeOptions{Debug: true}` or `ServeOptions{Perf: true}`, exposes `GET /debug/pprof/*` only with `ServeOptions{Perf: true}`, and `POST /click` was moved off GET; plain `Serve` exposes no debug or profiler endpoint.

`examples/spotify-player` is a multi-screen demo. The sidebar switches `View.Nav` between home, search, library, liked, browse, radio, queue, and profile. The six library screens each get a `screen_<name>.go` handler for their `components/` fragment; `build.go` wires the fragments and `screens.go` routes each screen's action prefix to its `click<Name>` handler.

`examples/desktop-cat` is the transparent, click-through overlay: `RunWithOptions` with `WindowOptions{Transparent, Borderless, Floating, FixedSize, BottomRight, Margin}`, `cat.Draggable` hit-tests the current PNG's alpha so a four-pixel drag moves the window, and `cat.Interactive` adds the bubble, so every other pointer event passes through. The `cat` package embeds 30 expression PNGs, switches one every eight seconds, and draws an HTML/CSS bubble. The notify listener defaults to `auto`, a Unix socket. `POST /notify` takes an expression and message (`GET /expressions`, `GET /state`). TCP `127.0.0.1:6969` is opt-in with `-notify-addr=127.0.0.1:6969`. `.agents/skills/catnotify` sends to that listener. `-web` serves a still preview on 8128; the Windows-only `nowplaying` reader samples Chrome's system media session every two seconds (`-media`, `-media-snapshot`).

`examples/wispr-flow-dashboard` is a 13-page app shell. `app/app.go` parses one template with a collapsible sidebar and the active page fragment; `View` carries every page's data, and a nav click sets `ActivePage`. Each page is its own package exposing `HTML` and `CSS`, listed in `app/pages.go` in sidebar order; `app/actions.go` handles the sidebar, tabs, streak chevrons, share, and get-app controls.

`internal/atspi` maps `AccessibilityTree` to AT-SPI roles over a bus seam. Linux desktop `Run` wires it from `internal/window/a11y_atspi.go`. The live D-Bus client is not implemented, so `Live` stays false. Windows UIA, macOS VoiceOver, and Android TalkBack are not shipped. Tests use `FakeBus` and must not log accessible text.

`internal/host` defines `Screen` and the optional interfaces a screen can add: `Ticker`, `Inspector`, `Reloader`, `Dropper`, `Focuser`, `ContextMenu`, `CursorShape`, `ScrollRequester`, `IME`, and `LongPresser`. `Dropper` receives `Drop.Path`, a sensitive local desktop path; copy bytes from `Drop.Read` during the call. The window and web packages call those. They do not import package `ownframe`.

`internal/clipboard` reads and writes the OS clipboard, and keeps an in-memory copy. It is desktop and window only: `Serve` and `/debug/state` expose no clipboard route. Tests call `UseMemory`.

`internal/filepick` opens the desktop file dialog. Linux runs `zenity`, `qarma`, `matedialog`, or `kdialog`, and under WSL it runs the Windows dialog through `powershell.exe` or `pwsh.exe` first, found on `PATH` or under `/mnt/<drive>/Windows`, and converts the path with `wslpath`. Windows calls `comdlg32!GetOpenFileNameW`; macOS runs `osascript`. `Run` installs it on the page. wasm, mobile, and `Serve` have no dialog, and a file input there takes a typed name. Tests install a fake with `page.InstallPicker`. `OWNFRAME_FILEPICK_DEBUG=1` prints fallback reasons to stderr.

`internal/render` owns the cached document. A cache holds the parsed HTML tree and the styled CSS document from one parse and one cascade, keyed on the executed source; a size or state change calls the engine's `css.Relayout` and lays the cached document out again, so a drag or a hover never reparses or recollects. `Redraw` keeps a `layout.Display` for a page `render.Replayable` accepts and a picture for any other page. Neither path writes a file. `render.State.Theme` hands the page theme to `css.Options.Extra`, which applies it after the template's own styles.

`internal/replay` draws the display list on the Ebiten canvas: fills with circular corners, axis-aligned border lines, stroked diagonal segments, grid runs, and shaped text. Images draw with linear filtering. A repeated shape bakes once into a bounded raster cache and blits, and a text run does the same while its face, content, color, and sub-pixel origin match. Consecutive operations inside one element box bake into a chunk that blits on later damage, re-baking when a content signature changes. A finite 2D transform replays through its matrix, and a separable blend or isolation group buffers into a layer and composites once through a Kage shader; non-separable modes, transformed grids, and rotated glyphs keep the bitmap. `internal/emoji` paints thirty-one color emoji as inline images backed by bundled Twemoji PNGs, resolved through the page image lookup. `DrawRect` replays only the operations that intersect a dirty rect, which the window keeps in a tile grid sized to the content rather than the window. `internal/window/draw.go` picks the display list or the fallback image; a fallback frame shows a `bitmap fallback` badge in its top-right corner.

`internal/frame` finds fill and text operations in a `layout.Display` by hit box and colour, so a `Page.SetTick` callback can animate without a Redraw. `internal/frame/anim*.go` drives compositor-style animations: translate, fade, scale, and rotate patch the retained op through `X`/`Y`, alpha, and `SetXform`, report the old and new boxes as damage, and go inactive when finished. `documentation/frames.md` has the call shapes and the cost.

`internal/ipc` is in-process `Send`, `Listen`, `Handle`, and `Request`. `ipc.go` re-exports them.

`Page.Load`, `Back`, `Forward`, `Route`, and `HTML` live in `internal/page`. `Config.File` and `ThemeFile` read from disk, and a page watches them by default; `Page.PollReload` swaps the template without touching history, and `-reload=false` turns the watch off. `Redraw` counts parses, cascades, layouts, and repaints for `host.Inspector`. `Config.Perf` and `Page.SetPerf` gate the extra timing, dirty counters, and alloc tracking behind an opt-in; untracked redraws leave those `Stats` fields at zero. A click dirties the changed box, and `page.Probe().TakeDirty` hands the window the region to repaint. The window calls the inner screen. The caret, the selection range, focus order, context menu state, cursor shape, and scroll requests are page state too. The caret span takes the field's text color through `currentColor`, so it shows on a dark theme, and a whole-value selection paints a selection span, with no field background, so only the selected text is highlighted. `internal/textrun` turns a point inside a shaped text run into a rune offset, so a click can place the caret at the glyph. `nav.go` exports `ErrNoHistory`.

`internal/crash` writes a local report. `Run` and `BindMobile` recover a panic and return that path. The default folder is `ownframe/crashes` under `os.UserConfigDir`, or `ownframe-crashes` under `os.TempDir`, created with mode `0700`; each report is written with mode `0600`. `SetCrashDir` rejects empty, control-character, dot-segment, or traversal paths before replacing a previously configured directory. `crash.go` exports `Report` and `SetCrashDir`.

`internal/fetch` sends one http or https request and stores no cookies. `fetch.go` exports `Fetch`, `XHR`, `FetchResponse`, and `ErrScheme`.

`github.com/chinmay-sawant/blinkless` `v0.0.3-0.20261021155643-c1054ff46202` parses the HTML, applies the CSS, and lays the page out. That require is the only engine pin. There is no replace. That commit carries both one-pass grid paths and the element patch work (`layout.RelayoutElement` with `html.SetText` and `html.SetAttribute`). `layout.DisplayList` is the placement. A replayable page keeps those operations and no picture. Any other page replays the same list into `page.Probe().Image`. The window replays or blits accordingly. blinkless does not paint a page bitmap. Text replay needs Ebiten v2.10.4 or newer, because blinkless requires `go-text/typesetting` v0.3.4 and older Ebiten builds its font face without the lookup cache v0.3.4 added. `css.Relayout` re-places an already styled document at a new viewport and state.

`examples/music` is shared example support: it decodes MP3 and WAV and plays them through Ebiten audio. The Spotify player example imports it and supplies clips from the bundled CC0 soundtrack in `spotify-player/player/assets/audio/`; `DemoTune` renders a short generated WAV for tests. Package `ownframe` does not import it. `Run` and `BindMobile` create the Ebiten audio context, because Ebiten v2.10 does not; `Serve` has none. `examples/spotify-player` uses `Page.SetTick` to move the seek bar and equalizer from the audio position.

Read `documentation/features.md` before changing paint, theming, frames, IPC, navigation, crash reports, fetch, the clipboard, drag and drop, or the devtools overlay. The call shapes and the limits are in the other files under `documentation/`.

A page taller or wider than the window scrolls on the mouse wheel. A picture that matches the window stays at one CSS pixel per window pixel. The sign-in example accepts the email `secret` and the password `secret`.

## Public API after the Page wrap

`ownframe.Page` is a defined type with an unexported inner page. Do not embed `*page.Page` again. Do not add an exported `Inner()`. The host loop stays unexported: do not return `host.Screen`. Tests that need a window game use `ownframe.NewGame`.

- No new `type X =` alias at the module root except documented re-exports named in `plans/v0.0.3/prod-ready-task/phase-01-public-api.md`.
- No new `func (p *Page)` on `internal/page` that exists only so `ownframe.Page` grows.
- Public `Page` keeps the app set: `SetData`, `Handle`, `AddHandlers`, `SetTheme`, `SetTick`, `SetImage`, `Redraw`, `RedrawIfChanged`, `Load`, `Back`, `Forward`, `Route`, `HTML`, `Size`, `MinSize`, `MaxSize`, `Clamp`, `SetSize`, `PatchElement`, `SetScrollWindow`, `SetViewportPinZ`, `ViewportPinZ`, `FormValue`, `SetFormValue`, `FormChecked`, `SetFormChecked`, `FormSelected`, `Focus`, `FocusID`, `FocusNext`, `FocusPrev`, `Probe`, plus the app calls for input, scroll, frames, and the accessibility snapshot. Paint, dirty regions, and counters are `page.Probe()`. `Page.Testing()` is for test files only. Example programs do not call it.
- No new example import of `github.com/chinmay-sawant/ownframe/internal/`. Existing pairs stay in `scripts/example-internal-allowlist.txt` until a public export replaces them. `make internal-import-check` records the set.

v0 may break. Rules for names, errors, context, and examples are in `plans/v0.0.3/prod-ready-task/field-go-practices.md`.

## Working directory and paths

- Always work from an explicit absolute path. For the main repository use
  `/home/chinmay/ChinmayPersonalProjects/ownframe`. For a named worktree use its
  absolute path, for example `/home/chinmay/ChinmayPersonalProjects/ownframe-wt-p1`.
- Never run `find`, `grep`, or any filesystem scan from `/` or from a home
  directory. Restrict searches to the repository or worktree root you are working
  in, by giving the exact absolute path as the first argument, for example
  `find /home/chinmay/ChinmayPersonalProjects/ownframe -name '*.go'`.
- A subagent must `cd` into the absolute repository or worktree path before
  running any command. Do not rely on inherited current directories, and do
  not traverse outside the assigned project tree.
- Do not scan or inspect files outside the ownframe repository unless the user
  explicitly asks for them.

## Package and file cohesion

Group code by responsibility. Keep a type and its related methods together when that makes the behavior easier to follow. Split packages at ownership boundaries, with explicit inputs and outputs. File length is a review signal, not a character limit.

Keep related test scenarios and their fixtures together. Do not split files to satisfy a size count, create synonym helpers, or move literals into unrelated constant files just for lint. Review changes for the number of files a reader must open to understand one operation. Use the cohesion questions in CONTRIBUTING.md during review.

## Lint

`make lint` is the lint gate. It runs the pinned `golangci-lint` (v2.15.0, see `GOLANGCI_LINT_VERSION` in the Makefile) over both modules, then the documentation and plan checks, then `make internal-import-check`, then the website lint from `frontend/`. The library at the root and the examples module under `examples/` are separate Go modules, so each gets its own pass; the examples tree is linted too, because it ships to users.

The website gate is ESLint 10 through `npm run lint` in `frontend/`. Its flat config lives in `frontend/eslint.config.js` and needs no separate make target; `make lint` runs it in the same pass. `eslint-plugin-react` still declares a peer range that stops at ESLint 9, so `frontend/package.json` pins the range with an `overrides` entry and the config sets `settings.react.version` to skip the plugin's version detection, which calls an API ESLint 10 removed.

The policy lives in `.golangci.yml`. It is a curated enable list, not `enable-all`: the runner defaults (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) plus `gosec`, `gocritic`, `misspell`, `unconvert`, `mnd`, `nolintlint`, `depguard`, `bodyclose`, `noctx`, `contextcheck`, `errorlint`, `forcetypeassert`, `gochecknoglobals`, and `forbidigo`. `//nolint` is a last resort and needs a written reason.

GitHub CI (`.github/workflows/ci.yml`) runs `make test` and `make lint` in the `gates` job, and `make test-race` (library module only) in the `test-race` job, plus the sec, cross, secrets, lint-actions, and reproducible jobs. The release workflow (`.github/workflows/release.yml`) runs `make test` and `make lint` before it creates the GitHub Release. Both set up Node.js 22 and install the website dependencies, so the website lint runs in CI.

## Writing

Load the `unslop` skill before writing prose: replies, docs, commit messages, and PR or issue text. Apply its fixes before sending.

## Builds

Do not run `go build ./...`. Every run links an executable for each example, takes about a minute, and leaves the binaries in this directory; the default `-p 24` loads the machine. For the same compile check, run `make build`. It calls `go vet -p 1 ./... ./examples/...`, which compiles every package including the examples, links nothing, and writes nothing. Raise the limit with `make build BUILD_P=4`. To check one example, name it: `go vet -p 1 ./examples/login`.

## Before you finish

For documentation or maintenance tooling, run `make docs-check plans-check examples-check`. For paint or window changes, also run `make graphics-check`; see `documentation/testing.md` for prerequisites and limits.

Run `gofmt` on every Go file you edit. From this directory, run `make test`. It calls `go test -p 1 ./... ./examples/...` so the packages do not all build and run at once. Raise the limit when a faster run is worth the load: `make test TEST_P=4`.

Leave `~/.Xauthority` untouched. The window toolkit logs a missing authority file and still opens the window.

Keep clipboard tests off the desktop clipboard. Call `clipboard.UseMemory` in those tests.

## Blocked work

If the user hands over a task and you cannot do it right as it stands, update the user and stop. Name the part that is blocked, the input or decision you need, and what you can do once you have it.

Fill no gap with a guess, a stub, or a silent default, and never report the task as done while the blocked part is still missing. Work that does not depend on the missing input can go ahead, as long as you mark what is on hold. The task stays open until the user answers.
