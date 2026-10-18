## v0.0.2

Second release of **ownframe**: one Go process that shows an HTML template in a window and calls Go functions for clicks, keys, and typing. The project and its engine were renamed in this cycle, `go-gpui` to **ownframe** and `gowkhtmltopdf` to **blinkless**, and the engine is now a tagged security release, `blinkless v0.0.2`. On top of v0.0.1 the window keeps the parsed and styled document between frames, repaints only the changed region, and gains a DevTools dock, opt-in performance sampling, hot reload, drag and drop, a full input layer with IME and touch gestures, an accessibility tree, and eight Android apps. There is still no JavaScript engine and no second process. `blinkless` parses the HTML, applies the CSS, and lays the page out; a page whose operations the replay accepts keeps vector operations and no bitmap, and any other page falls back to the engine image.

- Module: `github.com/chinmay-sawant/ownframe` (the v0.0.1 tag declared `github.com/chinmay-sawant/go-gpui`)
- Go: 1.26.4
- Ebiten: v2.10.4, the floor for text replay
- Engine: `blinkless v0.0.2` (`23cc975`, 2026-10-15), the security release, no local replace; the previous pin was `v0.0.0-20261013163225-c085b3423260`
- Direct requires: `blinkless`, Ebiten, `golang.org/x/image`, `golang.org/x/text`
- Docs: `documentation/` (45 guides, up from 20), plus the site in `frontend/` built to `docs/`
- Examples: `examples/` is its own module joined by `go.work`; 52 directories, up from 25, with 49 top-level demos, three benchmark apps, shared `music`, and the `android-host` Java library
- Review: `plans/v0.0.2/report-html-2.0.0/v0.0.2-report.html`; security audit in `plans/v0.0.2/security/`
- Migration: `documentation/rebrand.md`
- Previous release: v0.0.1

### Highlights

| Area | What you get in v0.0.2 |
|------|------------------------|
| Rename | `go-gpui` to `ownframe`, engine `gowkhtmltopdf` to `blinkless`; `data-ownframe-*` attributes, `OWNFRAME_*` settings. |
| Pipeline split | `internal/render` caches the parsed tree and the styled document; a size or state change calls `css.Relayout` instead of reparsing. |
| Repaint | `TakeDirtyRects` and a 256 px tile grid paint the changed region; shapes, text runs, and element chunks bake into bounded raster caches. |
| Pacing | The loop parks after five clean frames and wakes on input or a `Page.WantFrameAt` deadline; `NoIdlePace` opts out. |
| DevTools | F12 dock with Elements, Frame, and Ops tabs; Frame carries performance, pipeline, pacing, and memory rows. |
| Performance | Sampling is opt-in through `WithPerf`, `Config.Perf`, `WindowOptions.Perf`, and `ServeOptions{Perf: true}`; off means no cost. |
| Hot reload | `Config.File` and `ThemeFile` watch and swap the page in 250 ms, keeping state and history. |
| Drag and drop | `host.Dropper` and `Handlers.Drop` receive the files of one frame. |
| Input | Caret and selection, click-to-glyph, Tab traversal, context menu, cursor shapes, scroll requests, touch gestures, pinch zoom, F11. |
| IME | Android and iOS soft-keyboard sessions; a desktop preedit composition path behind `Page.SetIMEComposition`. |
| Accessibility | `Page.AccessibilityTree` returns a detached semantic snapshot; `AccessibilityPerform` runs focus, activate, and set-value. |
| Android | Eight Gradle projects, a shared Java host, an AAR through `ebitenmobile`, seven committed signed arm64 APKs. |
| Security | Same-origin POST endpoints, bind validation, fetch caps, crash-dir and source-size validation, plus the engine's own security work. |
| Examples | 28 new demos, including Telegram, Teams, Spotify at eight screens, Desktop Cat, Eyes, and seven SQLite apps. |
| Docs and site | 25 new guides, a rebuilt GitHub Pages site, and `showcase.md`. |

### Migrating from v0.0.1

The v0.0.1 tag declared the module `github.com/chinmay-sawant/go-gpui` and package `gpui`. This release declares `github.com/chinmay-sawant/ownframe` and package `ownframe`. Update imports and selectors (`gpui.New` becomes `ownframe.New`) and keep one copy of the library in a build; `documentation/rebrand.md` has the full checklist. Template attributes renamed from `data-gpui-*` to `data-ownframe-*`, and the debug switches from `GPUI_FILEPICK_DEBUG` and `GPUI_BROWSER_PORT` to `OWNFRAME_*`. Error strings now start with `ownframe:`.

Two signatures changed: `Page.Route` and `SetCrashDir` return `error` now. Everything else in the public API is additive. The root package grew from 27 to 88 exported identifiers and `Page` from 44 to 159 methods, with no exported name removed.

The engine require also renamed, from `github.com/chinmay-sawant/gowkhtmltopdf` to `github.com/chinmay-sawant/blinkless`. There is no replace directive; the pin is the tagged `v0.0.2`.

### Pipeline split and the engine

`Page.Redraw` still executes the template, but the executed bytes key a cache. `internal/render.Cache` holds the `*html.Document` and the styled `*css.Document` from one parse and one cascade. An unchanged source calls `css.Relayout` and lays the cached document out again; a changed source builds a new cache. `SetSize`, hover, press, release, focus, and theme changes all take the relayout path, so a drag or a hover reparses nothing. `TestResizeReusesParseAndCascade` pins one parse and one cascade across ten redraws.

The engine kept the `layout.DisplayList` boundary through the rename, and bitmap painting moved in-repo as `internal/bitmap`. `Page.Stats` reports parses, cascades, layouts, repaints, and relayouts; the DevTools Frame tab and `GET /debug/state` show the same counters.

### Repaint, tiles, and caches

`Page.Invalidate(id)` marks a box, its children, and its previous rectangle; `Redraw` diffs the old and new display lists; `TakeDirtyRects` (or the `TakeDirty` bounding box) hands the window the regions. A diff over 8 changed ops or a dirty area over a third of the frame falls back to a full frame. The window keeps a 256 px tile grid sized to the content, with full, buffer, rect, and blit plans, and clips repaint and blit to the strip left of the open DevTools dock.

Repeated work bakes into bounded raster caches: shapes at 64 entries, text runs at 128, element chunks at 64, each capped at 4 million pixels; chunks re-bake when their content signature changes. Content past the Ebiten image limit replays directly or through a 256 px-overscan viewport image instead of panicking. Pinned or fixed layers (`SetViewportPinZ`) take a direct replay, which is what Telegram uses for its bars. Bitmap-fallback pages, ticking pages without `UseFrameDirty`, and `Page.PNG` still repaint in full.

`TestRepaintMatchesFullRedraw` byte-compares an incremental paint with a full `Redraw` for every repaint case.

Numbers from the branch commits: `Invalidate` plus `TakeDirty` went from 4.6 µs to 7 ns per op and box id lookup from 15 µs to 3.5 µs (82ca687). A tile-repaint probe went from 120 to 760 frames per 10 s and from 23 to 54 MB allocated per frame down to 4 KB (1316914). The shape cache cut heap use from 345.1 MB to 20.1 MB, with the p99 repaint at 19.2 ms (was 22.9 ms) (063a59b).

### The window loop

`Update` runs in 19 stages, from devtools sync through IME, keys, resize, pointer, touch, hold, drop, tick, reload, scroll, devtools refresh, and wheel. A desktop window that stays clean for five frames parks its loop for up to 250 ms, wakes on input or at a `Page.WantFrameAt` deadline, and keeps a 125 ms tick otherwise; `WindowOptions.NoIdlePace` turns that off, and `OWNFRAME_PACE_DEBUG=1` prints presents per second. The second pacing pass moved idle presents from 8/s to 1 to 2/s, took a card repaint from 8,589 ns to 4,432 ns, and an animated frame from 170,854 ns to 176 ns (29079c6). Idle CPU on the paced build fell from 27.4% to 7.0%, with a repaint at 35 µs (de786e6). A drag relayouts at most every 100 ms and always commits the final size.

### DevTools

F12 or Ctrl+Shift+I toggles a right dock, 340 px wide by default, 220 px minimum, resizable on its left edge. Elements shows the hovered or pinned box as pretty JSON. Frame groups counters under WINDOW, RENDERING, PERFORMANCE, PIPELINE, PIPELINE DETAIL, PACING, MEMORY, and RELOAD; the performance and memory rows read `-` until the app opts in. Ops lists every operation in paint order with a colour per kind; clicking a row copies it, and the picker outlines the operation at its painted bounds, with Alt+click forwarding the click to the page. The dock is touch-scrollable and never enters `Page.PNG`, the display list, or the box list. `Config.DevTools` and `Page.SetDevTools` open it without a key. `go run ./examples/devtools` serves web mode on 8122.

### Performance, opt-in first

`Config.Perf`, `WithPerf(true)`, `WindowOptions.Perf`, and `ServeOptions{Perf: true}` turn sampling on; `Page.SetPerf` flips it at runtime. Off, a shipped app pays nothing and the untracked `Stats` fields stay zero. On, `host.Stats` adds `LastTemplate`, `LayoutTime`, `DisplayListTime`, `PaintTime`, `DirtyOps`, `DirtyRegions`, `ChangedOps`, and `AllocFrame`, a 120-frame ring reports average, p95, p99, and long frames over 25 ms, and a 500 ms sampler records heap, alloc per frame, goroutines, and RSS. `ServeOptions{Debug: true}` exposes `GET /debug/state`; `Perf: true` also exposes `GET /debug/pprof/*`. Fixtures: `examples/perf-benchmarks` (normal, large, and flappy apps on 8131, 8132, and 8133), `examples/perf-complex` (a 480-row dashboard on 8135), `examples/perf-stress` (8134), and the scripts `perf-baseline.sh`, `perf-idle.sh`, `perf-longrun.sh`, and `perf-dump.sh`. The guides are `documentation/performance.md` and `performance-baseline.md`.

### Hot reload

`Config.File` and `Config.ThemeFile` read the template and theme from disk and watch them. The window polls every 250 ms before it syncs. A broken edit keeps the last good picture and retries; a deleted file keeps the picture and reports once. A reload replaces the current history entry in place, keeps history length, typed values, focus, hover, active state, images, and tick, and fires no `Change` callback. `Serve` polls before it answers; `-reload=false` and `Config.DisableHotReload` turn the watch off. `go run ./examples/reload` serves web mode on 8123.

### Drag and drop

`host.Dropper` receives the dropped entries of one frame from `ebiten.DroppedFiles()`; a screen without it ignores files. `Handlers.Drop` handles them on the page, and `Drop.Read` is valid only during the call. A directory stays one entry. `-web` has no window loop and receives no drops; the wasm canvas runs the same pass. `go run ./examples/drop` serves web mode on 8124.

### Input, editing, and touch

- A form field carries a caret and an anchor as rune offsets; a selection normalizes to `[min, max)` and keys replace it. A click focuses the field and places the caret at the clicked glyph through `internal/textrun`. Drag, double-click, and triple-click select by word and line, an edge drag auto-scrolls, and the caret blinks on a 530 ms cycle in the field's own text colour.
- Tab and Shift+Tab move focus through `host.Focuser` in document order, skipping disabled controls and `tabindex < 0`. Escape closes the context menu or clears focus, and the menu and cursor shapes come from live page state through `host.ContextMenu` and `host.CursorShape`.
- `ScrollTo` and `ScrollBy` queue a request the window clamps; touch drags scroll with an 8 px slop, pinch clamps between 0.25 and 4, and F11 toggles fullscreen. Long press fires after 450 ms inside 8 px; a swipe needs 24 px and reports through `Handlers.Swipe`.
- IME: focusing a field on Android or iOS opens an `exp/textinput` session, the window paints the preedit as field text, and a commit goes through `Type` or `IMEReplace`. Desktop gets a preedit composition path through `Page.SetIMEComposition` and `IMEAnchor`, marked experimental.
- `Page.SetViewportPinZ` marks fixed layers so pinned chrome (the Telegram header and composer) stays still while the rest scrolls.
- `DispatchDeepLink` routes allowlisted schemes (http, https, and mailto) and rejects `javascript:`, `data:`, `vbscript:`, `file:`, `blob:`, and `about:`.

### Accessibility

`Page.AccessibilityTree` returns a detached semantic snapshot in document order: roles, names, descriptions, form state, actions, and bounds when the layout can provide them, with password fields redacted. `Page.AccessibilityPerform` runs focus, activate, and set-value against a document token, on the page's owning loop. The guides are `documentation/accessibility.md` and `accessibility-native.md`; the latter states the boundary: ownframe exposes the semantics but does not register them with an operating system, so screen readers cannot inspect the window through this API yet. `go run ./examples/accessibility`.

### Android

Eight Gradle projects: digital-wellbeing, dino, eyes, lg-remote, login, pet-care, platform, and telegram. They share `examples/android-host`, a Java library (`dev.ownframe.android`) with `AndroidHost.attach`, `onBackPressed`, `setSystemBarStyle`, and `close`, callbacks for insets, configuration, viewport, IME, and Back, and `TouchCancellation` to work around Ebiten 2.10.4 ignoring `ACTION_CANCEL`. The projects target minSdk 23, compileSdk 34, AGP 8.5.1 or newer, and Java 17, and set `enableOnBackInvokedCallback`.

Each app wires its own MainActivity duties: Telegram opens the camera or the photo picker and reports insets; Eyes reads the accelerometer and shake; LG Remote uses Bluetooth HID; Pet Care exports and restores JSON through SAF; Digital Wellbeing reads `UsageStatsManager`. The AAR is built with `ebitenmobile bind` and 16 KB page alignment, and `scripts/android-release.sh` signs arm64 APKs; seven signed APKs are committed under `examples/*/android/artifacts/`; `APK-SHA256SUMS` and `APPS.md` cover six of them. `documentation/android-build.md`, `android-app-creation.md`, and `android-runtime.md` cover the setup.

### Desktop integrations

`examples/desktop-cat` is the transparent, click-through cat: 30 expression PNGs, alpha hit-test dragging, a notification server on `127.0.0.1:6969` (`POST /notify`, `GET /expressions`, `GET /state`), the `.agents/skills/catnotify` sender, and a Windows-only reader that samples Chrome's media session. `examples/lg-remote` drives a webOS TV over Wi-Fi and adds Bluetooth HID on Android. `examples/snipping-tool`, `system-monitor`, `pet-care`, `digital-wellbeing`, and `theme-studio` cover capture, live graphs, a journal with SQLite, Android usage stats, and a live theme editor. File dialogs and the clipboard gained WSL path validation, `ErrUnavailable`, and the `OWNFRAME_FILEPICK_DEBUG` trace.

### Security

The security audit landed as PR #30; the architecture review in PR #31 then removed the print path entirely, so none of the print-command hardening is in this release. What shipped:

- Web mode: 5/15/15/60 s server timeouts, `nosniff` and `X-Frame-Options: DENY`, panic recovery to 500, bind-address validation that refuses wildcard binds, and `POST`-only same-origin `/click`, `/type`, and `/backspace` with a 256 KiB cap and NaN/Inf rejection. `GET /debug/state` needs `Debug` or `Perf`; `GET /debug/pprof/*` needs `Perf`.
- Fetch: 30 s timeout, http and https only, userinfo rejected, a 10-hop redirect limit, a 10 MiB body cap, a 1 MiB header cap, and no cookies. There is no SSRF allowlist; the caller sets host policy.
- Sources and crash: `MaxSourceSize` caps templates at 8 MiB with `ErrSourceTooLarge` on `New`, `Load`, `Route`, `SetTheme`, and reload. `SetCrashDir` validates its path and returns `ErrEmptyDir` or `ErrBadDir`; report directories are `0700` and files `0600`.
- Runtime: `ipc.Request` recovers a panicking handler as `ErrHandlerPanicked`, the clipboard no longer serves stale text after an OS failure, and Telegram's photo path bounds decoded images at 4096 px.
- Engine: `blinkless v0.0.2` is a security release. A stylesheet over 32 MiB is rejected, HTML nesting and list limits stop at 1024, images decode to at most 16 million pixels, SVG text is dropped before raster, `ParseWithLimit` defaults to 64 MiB when `html.Parse` is left uncapped, and local files open through `os.Root`. The audit lives in `plans/v0.0.2/security/`; the guide is `documentation/security.md`.

### Examples

52 example directories, up from 25, with 28 new ones: accessibility, android-host, css-masking, departure-board, desktop-cat, devtools, digital-wellbeing, download-manager, drop, eyes, flappy-bird, input, input-lab, lg-remote, live-log-viewer, perf-benchmarks, perf-complex, perf-stress, pet-care, reload, resize, snipping-tool, spreadsheet, system-monitor, teams, telegram, tetris, and theme-studio. The new demos' ports run from 8118 (teams) to 8140 (eyes); `examples/readme.md` maps most demos to their `-web` port.

Updated: spotify-player reached eight screens with a bundled CC0 soundtrack, a seek bar and equalizer driven by `Page.SetTick`, and a performance report with per-screen redraw cuts of 86% to 93%. wispr-flow-dashboard became a 13-page shell with one package per screen. dino moved its scene into template files and gained swipe controls and an Android project. clipboard, login, platform, fetch, replay, states, and music picked up smaller features; `examples/music` dropped its Openverse downloader and now uses bundled local tracks. `examples/audio-player` was removed after Spotify absorbed it.

Print and PDF were built on this cycle's dev line and removed before release: `Page.PDF`, `WritePDF`, `SavePDF`, `Print`, `SupportsPDF`, `PDFOptions`, `ErrNoPDF`, `ErrNoPrinter`, `internal/print`, `GET /pdf`, `examples/print`, and `documentation/printing.md` are gone. None of that was in v0.0.1.

### Documentation and the site

`documentation/` grew from 20 files to 45 guides. New: accessibility, accessibility-native, android-app-creation, android-build, android-runtime, compare-electron, compare-rust-gpui, compare-go-app-frameworks, compare-syntax, desktop-shell-integrations, devtools, drag-drop, emoji, hot-reload, interaction, migrate-experimental, packaging, performance, performance-baseline, rebrand, repaint, runtime-behavior, security, styleinvalid, and tricks-tips. The older guides were updated across the board.

`frontend/` is a React and Vite site built into `docs/` for GitHub Pages: a hero, an examples gallery, the engine diagram, get-started, a Rust GPUI comparison, and a mobile section, plus a docs view that renders every guide. `showcase.md` collects nine sections with screenshots and recordings (Telegram, Teams, Spotify, Pet Care, Digital Wellbeing, Desktop overlays, Wispr Flow, Dino Run, Eyes). The README gained the demo section, and the repository now carries `CONTRIBUTING.md` and an MIT `LICENSE`.

### Limits

- No JavaScript engine, Chromium, preload, `contextBridge`, or Node. One process.
- No cross-process IPC, native menus, tray, notifications, or more than one window. No window icon.
- No session, cookies, cache, or web storage.
- No auto-update or installer.
- No video, document canvas, or WebGL.
- No library audio API and no `<audio>` element. The examples play through `examples/music`.
- No PDF export or printing. `Page.PNG` and `GET /frame.png` are the raster outputs.
- The accessibility tree is a detached snapshot; ownframe does not register it with an operating system, so screen readers cannot use it yet. No spellcheck, no OS-global shortcuts.
- No sandbox, CSP, or context isolation.
- Desktop IME is an experimental preedit path; mobile IME runs through Ebiten's `exp/textinput`.
- wasm has no file dialog, no hot reload watch, and no clipboard route; F12 belongs to the browser.
- Pinned or fixed layers (`SetViewportPinZ`) skip tiling and replay in full each frame.
- Bitmap-fallback pages, ticking pages without `UseFrameDirty`, and `Page.PNG` repaint in full.
- Fetch has no SSRF allowlist; callers set host policy.
- The DevTools dock shows geometry and attributes, not computed styles.

### Install and run

```
go run ./examples/login
go run ./examples/login -web
sh scripts/browser.sh
sh scripts/android.sh
```

The library:

```go
page, err := ownframe.New(ownframe.Config{
    Title:  "Hello",
    HTML:   `<h1>{{.Title}}</h1>`,
    Width:  480,
    Height: 640,
})
if err != nil {
    return err
}
page.SetData(struct{ Title string }{"Hello"})
return ownframe.Run(context.Background(), page)
```

### Verification

- `make build` (`go vet -p 1 ./... ./examples/...`) passes with `blinkless v0.0.2` and no replace.
- `make test` (`go test -p 1 ./... ./examples/...`) passes on both modules; every package with tests reports ok. The tree holds 1,162 committed `_test.go` files, up from 195 at v0.0.1.
- Desktop and `GOOS=js GOARCH=wasm` builds compile.
- Incremental repaints are byte-equal to full redraws through `TestRepaintMatchesFullRedraw`.
- Eight Android projects bind through `ebitenmobile`, and seven signed arm64 APKs are committed with checksums.
- The security audit is closed in `plans/v0.0.2/security/`, and the six-part scope review is `plans/v0.0.2/report-html-2.0.0/v0.0.2-report.html` (the printing part predates the print-path removal).

### What's changed

- 362 commits between the v0.0.1 tag and this release (320 non-merge), 2026-10-03 to 2026-10-17, with 42 branch and PR merges (PR #3 to #32).
- The diff is 4,116 files, +188,695 and -6,675 lines.
- Documentation went from 20 files to 45 guides; examples from 25 directories to 52; committed test files from 195 to 1,162.
- The nine v0.0.2 workstreams (foundation, resize, repaint page and shell, hot reload, input window and page, devtools, print) landed on the feature branch and merged to master as PR #6. Later waves added performance observability, the Telegram and dino phone demos, seven SQLite examples, the Android host and runtime, accessibility, the damage list, security, the architecture review, and Eyes.
- Print and PDF were developed during the cycle and removed before release, along with `examples/audio-player` and the Openverse downloader.
- Engine history: the pin moved from `gowkhtmltopdf v0.2.7-0.20261003121325-2111b364213b` at v0.0.1 through the `blinkless` rename to the tagged `blinkless v0.0.2`.
- Full history: https://github.com/chinmay-sawant/ownframe/compare/v0.0.1...master
