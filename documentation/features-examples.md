# Implemented features

Verified scan of code, docs, and tests on 2026-10-05; `make test` passes on `master` for the v0.0.2 release.
This is a point-in-time report; [features.md](features.md) remains the canonical
feature index with links to the detailed docs.

## Core rendering

- HTML template -> CSS -> layout pipeline via `blinkless`, with hit-test boxes (`internal/render/paint.go`)
- Optional theme stylesheet layered after the template's own styles, swappable at runtime (`Config.Theme`, `Page.SetTheme`; `internal/page/theme.go`, `internal/render/state.go`)
- Display-list replay (fills, strokes, lines, grids, images, shaped text); non-replayable pages fall back to a bitmap with a "bitmap fallback" badge (`internal/render/display.go`, `internal/window/draw.go`)
- Rounded/elliptical/masked strokes, rounded fills, pixel-snapped borders, letter-spaced text, cached fonts/images (`internal/replay/`)
- `Page.PNG` on-demand rasterization with caching (`internal/page/page_image.go`)
- Frame animation: `Page.SetTick` runs a callback each window frame, and the retained `Display` operations can change paint fields without a Redraw (`internal/page/page_tick.go`, `internal/frame/`)

## Forms and interaction

- Click focus/toggle/radio/select activation; typing with auto-repeat Backspace, DeleteWord (`internal/page/form_*.go`, `internal/window/keys.go`)
- `data-bind` two-way binding to struct fields via reflection, `BeforeEdit`/`Change` handlers (`internal/page/bind.go`)
- Copy/Cut/Paste/SelectAll/Undo/Redo ([editing.md](editing.md)); form values survive redraw (`internal/page/page_clip.go`, `internal/page/page_edit.go`, `internal/page/form_merge.go`)
- Host CSS states: `:focus`, `:focus-visible`, `:hover`, `:active`, `:checked` (`internal/page`)
- Mouse-wheel and touch scrolling with configurable page scrollbar thumbs ([scrolling.md](scrolling.md)); click, hover, and tap input ([pointer.md](pointer.md)) (`internal/window/fit.go`, `internal/window/pointer.go`)

## Navigation and platform

- Browser-style history `Load`/`Back`/`Forward`/`HTML` with `ErrNoHistory`; `data-action` route map (`internal/page/page_nav.go`)
- Desktop window via Ebiten, wasm browser build, and mobile binding; sizing is in [window.md](window.md) and the run modes are in [platforms.md](platforms.md) (`internal/window/run.go`, `scripts/browser.sh`, `BindMobile`)
- DevTools overlay: a right dock with Elements (element JSON), Frame (window and pipeline counters), and Ops (operation list with outlines) tabs; F12 or Ctrl+Shift+I toggles it, `Config.DevTools` starts it on, and `Page.SetDevTools` changes it at runtime ([devtools.md](devtools.md), `internal/window/devtools*.go`, `internal/host/inspect.go`)
- Web mode HTTP server: `GET /`, `GET /frame.png`, `GET /debug/state` (with `ServeOptions{Debug: true}` or `{Perf: true}`), `POST /click`, `POST /type`, `POST /backspace` ([web.md](web.md), `internal/web/server.go`)

## Services

- In-process IPC `Send`/`Listen`/`Handle`/`Request` with `ErrNoHandler` (`internal/ipc/`)
- `Fetch` GET + `XHR` arbitrary method; http/https only, no cookies (`fetch.go`, `internal/fetch/`)
- OS clipboard (X11/Windows/macOS) with in-memory fallback and `UseMemory` test mode (`internal/clipboard/`)
- Panic recovery writing local timestamped crash reports (`internal/crash/`)

## Examples and tests

- `examples/login` (sign-in, secret/secret), `examples/forms` (all control types), `examples/bind` (two-way binding), `examples/input-lab` (all nine form and input demos in one window: forms, controls, bind, bind-hooks, editing, clipboard, input, login, states), `examples/theme` (light/dark switch), `examples/wispr-flow-dashboard` (Wispr Flow app clone: a collapsible sidebar, one package per screen, dictation history, notetaker, insights with usage, voice, and leaderboard tabs, dictionary, snippets, style, transforms, scratchpad, invite, free month, settings, and help; the grids stack on a narrow window), `examples/spotify-player` (dark web-player clone with eight screens for home, search, library, liked songs, browse, radio, queue, and profile, live iTunes data, local free-music playback, and an animated now-bar equalizer), `examples/teams` (Microsoft Teams-like app clone with Activity, Chat, Teams and channels, Calendar, Calls, and Files menus, a dark theme and a light theme toggle, a More apps flyout, a profile flyout with presence, chat compose, live chat and search filtering, expandable channel threads with emoji reactions, a meeting card with date and time picks, a dial pad, and file details; the state persists in SQLite, seeded from `examples/teams/store/seed/*.sql`, and `-db` selects the database file (`:memory:` keeps it in memory); Marvel-cast sample data), `examples/dino` (keyboard-only dinosaur game with a per-frame tick and an FPS readout), `examples/flappy-bird` (Flappy Bird with an HTML/CSS scene, gravity and pipe physics, and keyboard or click control)
- `examples/devtools` (right-dock inspector with Elements JSON, Frame counters, and Ops outlines; F12 or Ctrl+Shift+I), `examples/reload` (a file-backed page that redraws when `index.html` changes; the counter keeps its value), `examples/drop` (a dropped file prints its path; a file control takes a path from the picker), `examples/input` (focus traversal, caret keys, drag selection, context menu, cursor shapes, touch scroll and pinch, page scrolling, and F11), `examples/desktop-cat` (transparent desktop companion with 30 expression PNGs, native click-through and dragging, a notification server on 127.0.0.1:6969, and a wasm canvas preview), `examples/resize` (media-query columns, a `100vw` bar, rewrapping text, and a hover control)
- `examples/music` is shared example support: MP3/WAV decode, playback through Ebiten audio, and a generated demo tune for tests. The Spotify player supplies clips from its bundled CC0 soundtrack. Package `ownframe` does not import it.
- `examples/accessibility` demonstrates desktop IME composition, the semantic tree and its actions, password redaction, and scrollbar thumb feedback.
- `examples/depth-alert` is an experimental Android app that samples ARCore center-view depth and maps fresh, confident distance summaries to haptic alerts. A 3–3.5 m surface revokes a prior Clear state; Stop alerts survive brief depth gaps and Clear requires twelve consecutive frames. Its README documents the thresholds and safety limits; physical-device safety validation remains required.
- `examples/departure-board` is a split-flap board: per-character tiles over a seeded timetable, four liveries with a night variant, and tick-driven flaps that write only `op.Text` and rebind per `Page.Generation()`; the refresh bar is a `frame.Fill` lookup that needs the livery accent. `examples/theme-studio` is an editor beside a live preview: Apply calls `Page.SetTheme` and writes the watched `Config.ThemeFile`, a token strip renders `--name: value` pairs with colour swatches, and four presets ship with it.
- Go tests cover packages across `internal/` and `examples/`, including `internal/page`, `replay`, `render`, `window`, `ipc`, `fetch`, `crash`, and `clipboard`; `make test` runs both modules
- Other examples: `layout`, `replay`, `shapes`, `css-masking`, `png`, `controls`, `bind-hooks`, `editing`, `states`, `scrolling`, `history`, `web`, `ipc`, `fetch`, `clipboard`, `crash`, `platform`; index in [examples/readme.md](../examples/readme.md)

Update 2026-10-13: `examples/css-masking` demonstrates raster alpha and
luminance masks, SVG `mask-type`, and two-layer `mask-composite` through
retained display replay. Its test checks PNG pixels and retained display
selection (`examples/css-masking/page_test.go:11-24`). The replay package
compares GPU mask compositing with the software painter
(`internal/replay/gpu_group_mask_test.go:23-48`,
`internal/replay/gpu_group_pixels_test.go:13-35`).

## Known gaps

Documented in [features.md](features.md) and [../PHASES.md](../PHASES.md), not
implemented: JS engine/Chromium, cookies/sessions, native menus/tray,
auto-update, video/WebGL, a library-level audio API (the examples play
audio through `examples/music`), an accessibility tree, and some engine CSS limits
(transformed WOFF2, `:disabled`, `::placeholder`).
Desktop IME composition is implemented in the window and page packages, with
native input-method checks still open in the
[phase checklist](../plans/v0.0.2/accessibility-desktop-ime.md).

Update 2026-10-11: ownframe now exposes a page-level semantic tree and
supported control actions. The native screen-reader bridge remains open;
see [accessibility.md](accessibility.md).
