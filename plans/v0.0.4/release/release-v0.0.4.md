## v0.0.4

Fourth release of **ownframe**, prepared from tag `v0.0.3` (`085ba7ea`, 2026-11-10) through `012b39c9` (2026-11-23, merge of pull request #99). That range is 162 commits (134 non-merge, 28 merges) and 23 merged pull requests, #77 through #99, and it touches 2,071 files (+131,386 / −92,374). Most of the deletions are the removal of tracked WebAssembly and APK payloads and the merge of micro-files, not lost behavior. This file is the release body. Creating it does not create the git tag or a GitHub Release.

The headline is that every target native host now has a screen-reader bridge or a documented gap, and the render path is much cheaper per hover and per resize. Linux AT-SPI accepts actions, Windows ships a UI Automation provider, and Android ships a TalkBack bridge. macOS VoiceOver is the one adapter that is still missing. Viewport `position: sticky` and height media queries work. A retained hover on the 688-property catalog costs under 0.35 ms of delta. The v0.0.4 review passes closed a long list of races, injection paths, and cache-lifetime bugs.

- Module: `github.com/chinmay-sawant/ownframe`
- Go: 1.26.11 (v0.0.3 was 1.26.10)
- Ebiten: `v2.10.6`, a stable tag. v0.0.3 pinned the pseudo-version `v2.11.0-alpha.0.20261103125056-58976b602532`. Text replay needs Ebiten v2.10.4 or newer.
- Engine: `github.com/chinmay-sawant/blinkless v0.1.6`, no replace. v0.0.3 pinned `v0.1.1`.
- Examples module: `examples/go.mod` still requires `ownframe v0.0.2` and still has `replace => ../`. See [Release steps](#release-steps).
- Previous release: v0.0.3
- Range: `085ba7ea..012b39c9`

`go get github.com/chinmay-sawant/ownframe@v0.0.3` does not contain anything described here.

### Highlights

| Area | What landed after v0.0.3 |
|------|--------------------------|
| **Accessibility** | Linux AT-SPI accepts focus and activation from a live client and supports `EditableText`. A pure-Go Windows UI Automation provider (amd64 and arm64) is validated with NVDA and Narrator. An Android TalkBack bridge is paired with a Java `AccessibilityNodeProvider`. |
| **CSS** | Viewport `position: sticky` and `min-height` / `max-height` media queries (blinkless pull request 39, adopted in #79). The window paints and hit-tests the sticky view at the current scroll and zoom. |
| **Static HTML** | `Config.StaticHTML` and `LoadOptions.StaticHTML` make literal HTML an explicit source mode. `ui.NewPage` and `ui.Mount` opt in. |
| **Window** | One paint mode is chosen per frame and transitions release unused buffers. Desktop windows share customizable headers. Windows windows are fitted to their monitor. Windows IME composition is no longer committed early. |
| **Rendering and motion** | Nine-slice rounded fills, native quarter-turn text, OpenType feature and `text-autospace` replay, `-webkit-text-stroke` paint, a translucent-backdrop over-blend fix, retained-hover fixes, and a motion transform-base fix. |
| **Hover and resize cost** | `loglens` edge-region hover 2.08 ms to 0.80 ms and command-palette hover 1.77 ms to 0.66 ms (blinkless v0.1.4). The v0.1.5 harness reports 0 of 688 failed rows and two gradient combinations 15x to 44x faster. |
| **`-web`** | Serves the live application document with real inputs, server-sent snapshots, in-place tick patches, and browser playback of registered images and audio. |
| **Clipboard** | A pure-Go Wayland `data-control` client, tried before `wl-copy` and X11. |
| **Components** | `chadcn.Interactions` installs tab, slider, and dialog behavior once. `ui/virtual.Window` computes a windowed row range. |
| **Hardening** | Two review passes (#93, #98, #99) fixed COM lifetime, injection, XSS, password-copy, data-race, SSRF, crash-directory, and GPU-lifetime defects. `engineMu` is gone. |
| **Repo and CI** | The website and nine wasm demos build in GitHub Actions. Tracked wasm and APK payloads left the checkout. A native `windows-arm64` job runs. PR runs are lighter. `x/perf` and `govulncheck` moved out of the library module. |
| **Examples** | `blinkless-all` (one window per CSS property group, 688 properties), `blinkless-next`, `blinkless-scroll`, a refreshed LG remote, a Login app with TalkBack. Six small examples were removed. |

### Migrating from v0.0.3

Most v0.0.3 programs compile unchanged. The items below can break a build or change behavior.

| Change | Who is affected | What to do |
|--------|-----------------|------------|
| Chart packages left the library: `components/apexcharts`, `chartjs`, `echarts`, `nivo`, `recharts`, `shadcnchart`, `tremor` now live under `examples/builder/charts/<pkg>` | Anyone who imported them (#83, `ae94e106`) | Import from the examples module or copy the package. Nothing in the library imported them. |
| `Page.Audio`, `Page.AudioPlayback`, `Probe.Elements`, `Scroll.ViewportPinZ`, and `Testing.TakeChangedProperties` are removed | Callers of those five methods (#83, `3591cebf`) | None of them had a caller in `examples/`, `components/`, `ui/`, `cmd/`, or `documentation/`. `Page.SetAudio` and `Page.SetAudioPlayback` are the audio surface now. |
| Builder documents passed straight to `New` or `Load` need `StaticHTML: true` | Programs that hand a `ui` builder document to `ownframe.New` or `Load` without `ui.NewPage` or `ui.Mount` (#94, `39d651e6`) | Set `Config.StaticHTML` or `LoadOptions.StaticHTML`. `ui.NewPage` and `ui.Mount` already do. `StaticHTML` treats Go template actions in the source as literal text, and without it a builder document that contains `{{` is run as a template. |
| `Copy` and `Cut` on a password field return nothing | Code or tests that read a password back through the clipboard (#98, `17caf6c9`) | Read the value through your own state. This is intentional. |
| `Page.Window()` desktop controls act only on a page that `Run`, `RunWithOptions`, or `NewGame` attached | Code that calls `Minimize`, `ToggleMaximize`, `Maximized`, `Close`, or `CursorPosition` on a page it never ran (#83) | A detached page now no-ops. `Maximized` is false and `CursorPosition` is `0, 0`. |
| Blinkless custom property names are case-sensitive, and HTTP loads are GET-only | Stylesheets that spell `--Foo` and `--foo` alike (#98, engine change) | Match the case you declared. |
| Per-property `ui` value types such as `AccentColorValue` and `BorderBlockEndColorValue` are gone. `Element` setters accept the shared types `Value`, `LengthValue`, `NumericValue`, `LengthPercentageValue`, `LengthPercentageSpacingValue`, and `BorderStrokeValue`, or an interface that is still generated for that property (`WidthValue` remains) | Code that names one of the removed types in a signature or variable (#83, `b9dba989`) | Use the constructors (`ui.Px`, `ui.Rem`, `ui.Percent`, `ui.Hex`, and so on) and let the type be inferred, or name the shared type. About 530 generated type names left the `ui` package, and 140 shared names joined it. |
| Demo and `-png` output files are written with mode `0o600` and directories with `0o750` | Anyone who reads those files as a second user (#78) | `chmod` them, or change the writer in your own program. |
| The `-web` password value is redacted only when `InsecureAllowRemoteBind` is on. Loopback keeps the live value so typing round-trips | `-web` users on a non-loopback bind (#98) | None on loopback. A remote bind must be opted in as before. |
| `-web` theme overrides containing `</` are rejected | Theme sources that put `</style` in a value (#98, XSS fix) | Remove the sequence. |
| Crash directories are checked with `Lstat` before `MkdirAll` and always `Chmod`ed on Unix; existing directory permissions are kept | Crash-report consumers that expect a loosened directory (#93, #99) | Expect `0700` / `0600` as before. |
| Removed examples: `bind-hooks`, `clipboard`, `controls`, `editing`, `input`, `states` | Anyone who built them (#83, `b9dba989`) | The behavior stays covered by `forms`, `login`, `list`, and the tests. Added: `blinkless-all`, `blinkless-next`, `blinkless-scroll`. |
| Tracked Pages wasm demos and Android release APKs left the checkout | Anyone who linked `docs/demos/*.wasm` or an APK path from the tree (#85) | The site builds from `npm --prefix frontend run build`. The signed Android reference files stay on their existing GitHub release. The payloads remain in Git history. |

The `Page` wrapper, its doors (`Form`, `Edit`, `Scroll`, `Window`, `Access`, `Probe`, `Testing`), `NewGame`, `PatchElement`, and `PatchElements` are unchanged in shape. `Page.Testing` is still for tests only, and the five example harnesses that called it were rewritten (#85). `Page`, `Form`, and `Scroll` methods are now safe on a nil `*Page` and on a zero-value `&ownframe.Page{}` and return an error where one exists (#93, `TestNilAndZeroPageMethodsDoNotPanic`).

### Public API

Added to the module root:

- `Page.SetAudio(data, contentType)` and `Page.SetAudioPlayback(playing, volume)`, the audio route that `-web` plays in the browser tab (#77).
- `Config.StaticHTML` and `LoadOptions.StaticHTML` (#94).
- `Probe.PNGWithContext` and `Window.CursorPosition` on the doors (#83).
- Fetch sentinels you can match with `errors.Is`: `ErrBodyTooLarge`, `ErrHeadersTooLarge`, `ErrTooManyRedirects`, `ErrCredentialsInURL`, `ErrMethod`, `ErrBlockedHeader` (#98).
- `pagetest.ErrNilPage` (#93).

Added packages and symbols:

- `androidaccess`: `Tree()`, `Perform(document, node, action, value)`, and the action constants `ActionFocus`, `ActionActivate`, `ActionSetValue`. It exists for `ebitenmobile` bindings, which cannot export Go structs (#96).
- `components/chadcn`: `Interactions` with `NewInteractions`, `Mount`, `Update`, `Handle`, and `Close`; `Interactions` fields on `TabsOptions`, `SliderOptions`, `DialogOptions`, and `AlertDialogOptions`; `SliderOptions.OnValueChange` (#79).
- `ui/virtual`: `Window` and `Range` (`Start`, `End`, `TopSpacer`, `BottomSpacer`, `TotalHeight`) for a windowed row list (#83). A total height that would overflow is clamped (#99).
- `ui`: `ErrNilElement`, a generic `Value` family, and about 130 `Element` setters for properties the engine added in v0.1.2 and later, among them `BackdropFilter`, `WebkitTextStroke`, `ScrollTimeline`, `ViewTimeline`, `TimelineTrigger`, `OffsetPath`, `Marker`, `Interactivity`, `FieldSizing`, `TextFit`, `WindowDrag`, `Zoom`, and the SVG geometry properties. `SupportsProperty` reads the generated compatibility matrix.

Unchanged and still enforced by tests: no new root `type X =` alias outside the documented allowlist, no exported `Inner()`, `NewGame` for tests, and no new example imports of `ownframe/internal/` (`make internal-import-check`, 10 allowlisted pairs, down from 16). The `internal/` import gates `TestInternalNoTestingImport` and `TestInternalNoRootOwnframeImport` and the `.Testing()` call gate `TestNoUnauthorizedPageTestingCalls` are new.

### Hosts

Desktop `Run` on Linux, Windows, and macOS, and Android `BindMobile`, are still the target native hosts. This tree is `v0.0.x`. There is no `v1.0` tag. Wasm and `-web` stay previews and debuggers. `documentation/platforms.md` carries the current-versus-planned table.

| Host | Accessibility | Clipboard | Input and window |
|------|---------------|-----------|------------------|
| **Linux** | AT-SPI tree, actions, and `EditableText` (#79, #85) | Wayland `data-control` client, then `wl-copy` and X11 (#95) | WSLg Xlib error handler is serialized (#98, #99). Top-edge snap on WSL is kept. |
| **Windows** | UI Automation provider, amd64 and arm64 (#95) | Backend locks its OS thread (code in #98, proof pending a Windows runner) | Window fit to monitor (#95). IME composition no longer commits early (#95). Native file selection and `Run` checked on a Windows host (#85). |
| **macOS** | VoiceOver is not shipped. | Unchanged. | Unchanged. No host was available for validation. |
| **Android** | TalkBack bridge with the Login app wired (#96) | Unchanged. | `.debug` application-id suffix so debug and signed builds install side by side. |
| **Wasm and `-web`** | None | None | Live document, SSE snapshots, image and audio routes (#77). |

### Accessibility

**Linux AT-SPI actions (#79).** A live client can focus and activate controls. `internal/atspi/action_queue.go` queues the request and waits for the page's real result. Success means the page accepted the action, not that the request entered the queue. It handles cancel, shutdown, and a full queue. A reply has a two-second deadline, and the window loop runs up to four requests per `Update` and wakes the idle window so the action paints. Evidence is `go test -race -p 1 ./internal/atspi ./internal/page ./internal/window` (exit 0), `make graphics-check` (all 14 required checks, including the wake test), and `scripts/check-atspi-actions.py` for live D-Bus delivery. All of it is Linux. Later work added `EditableText` through the page's owning loop (#85). A private X11 session passed replace, insert, delete, copy, cut, and paste, and Orca 42.0 discovered `EditableText` and sent a focus announcement for an edited Unicode value to Speech Dispatcher on a non-audible backend. Orca speech through speakers was not recorded. Password values stay empty and the node does not advertise `SetText`. The newest snapshot wins when several arrive close together (#98).

**Windows UI Automation (#95, `32727657`, `15d53b96`, `982bf627`).** `internal/uia` is a pure-Go provider with no cgo. It maps the snapshot to UIA control types, names, state, bounds, and the Invoke, Value, Toggle, and SelectionItem patterns. Actions use the same queue as Linux. Two details decide whether any client sees an event. The provider must report `ProviderOptions_ServerSideProvider` (value 2), and the window must answer `WM_GETOBJECT` for every object ID UIA sends, not only the root. Before the fix every raise call returned `S_OK` and nothing arrived. Hit testing needs one assembly thunk per architecture (`hit_windows_amd64.s`, `hit_windows_arm64.s`) because a Go syscall callback cannot read the two double arguments of `ElementProviderFromPoint`. `ElementAt` picks the deepest element and the later sibling where two overlap. Validated with NVDA and Narrator. **JAWS is untested.** A `windows-arm64` job on the native `windows-11-arm` runner builds, vets, and tests `internal/uia` and `internal/window`, so the arm64 thunk runs on real hardware.

Review follow-ups (#98, #99) add COM `AddRef`/`Release` accounting and GC keep-alives for `*comObj`, a `tryAddRef` in `IUnknown::AddRef`, a guard so `syncA11y` does not `attachUIA` after cancel, and a detach on context done. **The Windows-host rows `SEC-UIA-01`, `SEC-UIA-02`, `SEC-UIA-03`, `CONC-CLIP-01`, `CORR-UIA-01`, and `PERF-UIA-01` are written and `GOOS=windows go vet` passes, but `go test -race ./internal/uia ./internal/clipboard ./internal/window` has not run on a Windows runner.** They stay `[~]` in `plans/v0.0.4/go-review-hardening/`.

**Android TalkBack (#96, `346d52a3`).** The Android window loop publishes the snapshot as JSON (password values omitted) and drains focus, activation, and text actions on the owning page loop. `dev.ownframe.android.AccessibilityBridge` in `examples/android-host` attaches an `AccessibilityNodeProvider` to the `EbitenView`, converts CSS pixels to screen pixels with the display density, emits `TYPE_WINDOW_CONTENT_CHANGED` and `TYPE_VIEW_FOCUSED`, and runs `ACTION_FOCUS`, `ACTION_CLICK`, and `ACTION_SET_TEXT` on a worker thread. Java polls every 250 ms only while `AccessibilityManager.isEnabled()` is true. Nodes with no geometry inherit their ancestor's bounds so Android does not prune the subtree. `ACTION_SET_TEXT` cannot write a password field. The only app wired is `examples/login`. A native arm64 device launch stays open. An x86_64 signed Login APK installed and rendered on an Android 14 emulator (#85). That does not prove TalkBack speech on a device.

**Not shipped.** macOS VoiceOver (`NSAccessibility`). Only the three bridges above exist.

### Window, replay, and render

- **Paint lifecycle (#94, `8f50dd30`).** One paint mode is selected per frame and the transitions among tile, viewport, backdrop, and transient replay buffers live in one place. The painted mode decides idle parking. Surface changes release buffers the next mode does not use. Pixel and lifecycle regression tests cover mode changes. `make test-race` was not run for #94 at merge time.
- **Retained-surface correctness (#88, #93).** A translucent `position: fixed` backdrop no longer composites repeatedly over its own retained pixels on partial or idle frames, which had darkened a Command Palette and checkerboarded 256-pixel tiles. `blitFromTiles` restores visible cached unfixed tiles with `BlendCopy` before pinned chrome paints, and `drawViewport` sets `BlendCopy` so transparent windows replace retained pixels. Also fixed: an offscreen-dirtied tile that later takes a sub-rect damage repaints its full visible tile, `layerDamage` is no longer dropped when page dirty rects are empty, stale promoted layer IDs across mode switches, viewport layer z-order inversion, and a tile-plan refresh on in-place motion. Frame pacing is phase-locked.
- **Blend and masks (#93).** `color-dodge`, `color-burn`, and group alpha in `group_blend.kage` are corrected. Group bounds are tight for non-masked blend groups. `rotate(360deg)` and a transformed `-webkit-text-stroke` replay correctly. Quarter-turn bitmap math is exact.
- **New paint and text (#78, #92).** `-webkit-text-stroke` outlines paint. OpenType font features and `text-autospace` replay. Rotated quarter-turn text runs natively. Dead 2D function blends are skipped. Rounded fills and borders use a nine-slice with analytical corner masks (`0814a184`).
- **Caches and buffers (#88, #93).** Pooled PNG encoder buffers. Replay LRU caches share `internal/replay/lru.go`. An unchunked image hash ring, raster keys quantized to 1/1024 px, a `layerMaxPixels` eviction on layer resize, and a `Radius <= 0` early-out in the backdrop path. The backdrop shader and scratch textures are pooled and deallocated on `Renderer.Close()`. `runWindow` disposes its GPU resources fully (#98).
- **Shared headers (#78).** Desktop applications can share a customizable header (`4be98873`). A Windows window is shrunk and moved only as much as needed to lie inside the work area of the monitor they settle on, once at startup and once after a monitor change. They are never enlarged, and the move waits while the window is maximized, minimized, full screen, or dragged. It uses `SWP_ASYNCWINDOWPOS`, since a synchronous move from `Update` crashed the window (#95).
- **IME (#95).** Windows reported the clause cursor at the end of a fresh preedit, so the glyph anchor moved with every key and the window committed the half-typed composition. A live composition is no longer restarted. Microsoft Japanese IME was validated on a Windows host. CJK glyphs in the IME field also need the engine change in blinkless pull request 47.
- **Sticky and height queries (#79).** `internal/page/page_sticky.go` and `internal/window/sticky.go` ask the engine for the sticky view at the current scroll and paint and hit-test from it. `SetScrollViewport` is a host-interface method and not on the public `Page`. `TestStickyViewportPixelsAndHits` covers activation, release at the containing block, scroll return, overlap, edits while scrolled, and zoom. `scripts/check-sticky-reference.mjs` and `scripts/compare-sticky-captures.py` compare eight frames with Chrome. The engine scope is viewport sticky. Nested scroll containers stay out of scope.
- **Resize (#79).** `internal/page/page_size.go` keeps template data clean when only the window size changes and no scroll callback is active. Layout still reruns. Same-host medians with perf off, 1,000-row list, 800 by 600 viewport: resize 25.7 ms to 23.4 ms (about 9% lower), bytes per op down about 11.7%, allocations down about 41.9%. Text update, unchanged redraw, and windowed scroll stayed within the 15% bound. These are page-work numbers, not native frame time. Source: `plans/v0.0.4/performance-evidence.md`.
- **Window screenshot and test harness (#78).** `-snapshot ADDR` serves a window screenshot on demand. `scripts/gui-drive.sh` drives the real window with `xdotool` and reads it back. `documentation/testing.md` has the new "Testing an app by driving it" section, and `AGENTS.md` carries the short form.

### Motion and hover

- **Hover latency (#88, #89, #90, #92).** Batched hover patching through `render.RelayoutElements` is the default (`paintOnlyBatchDisplay`, one pass over several roots). The dirty diff packs discrete op flags into one FNV-1a word and delegates rare payloads to `Op.SameExtra` and `Op.MixExtraStamp`. `KidIndex.Rebuild` reuses its backing arrays. `hoverGuardHash` compares `(sourceRev, sourceHash, sourceLen, themeSrc)` instead of re-hashing the document. Transition capture reuses runtime scratch maps, pairs aligned boxes in O(1), and walks declared properties through a bitmask.
- **`loglens` hover, blinkless v0.1.3 to v0.1.4 (#89).** `BenchmarkEdgeRegionHover` 2.08 ms to 0.80 ms (−62%), 1.74 MB to 203 KB (−88%), 3,455 to 1,047 allocations. `BenchmarkCommandPaletteHover` 1.77 ms to 0.66 ms, 1.89 MB to 221 KB, 3,350 to 892 allocations. Source: `plans/v0.0.4/pr/pr-subms-hover-relayout.md`.
- **Property harness v3 and blinkless v0.1.5 (#91, #92).** The 688-property harness reports 0 of 688 failed rows, 0 phase errors, 0 unchanged-redraw allocations, and 0 bitmap fallbacks. `bg_gradient + clip_shape` at 64 nodes 42.33 ms to 0.96 ms (44.1x) and its hover 90.71 ms to 2.04 ms. `bg_gradient + filter_fx` 12.03 ms to 0.77 ms (15.7x). `clip-path` resize delta 0.73 ms to 0.03 ms. The layout-animation tick at 64 and 256 boxes dropped 32% and 22% in time and 61% and 74% in allocations. Twelve 64-node UI recipes build in 0.59 to 1.65 ms and hover in 1.14 to 4.34 ms. These are the figures in the pull request records. They were not remeasured for this note, and they depend on the harness host.
- **Correctness (#90, #92, #93).** Live transitions survive scoped patches. `PatchElements` unhovers correctly. Sibling paint survives a retained redraw. The pointer redraw batch flushes before the tick. A hover patch and a layout-animation tick that rebind a paint player before the next sample keep the original paint base, so translation does not double-apply (`89e4e536`). `StepLayout` no longer spins on a delayed animation (#88).
- **Motion rules are unchanged.** One operation must not be driven by CSS motion and by `Page.SetTick` together. `MotionStats.ManualOverlap` and `ManualOverlapID` count a tick that writes an op an active CSS animation or transition owns (#83). All seven CSS motion packages live in `internal/motion`.

### `-web` live document (#77)

`GET /` returns the executed page with real text fields, passwords, textareas, selects, checkboxes, and radios. `?dom=0` and a page with no source keep the PNG shell. `POST /change` replaces a field value or moves the caret. Snapshots arrive over server-sent events with a separate `domGeneration`, so an in-place tick patches leaf text, inline geometry, and fill widths by element id without a refetch. The browser scrolls the document itself when it can. A click resolved by element id ignores the server frame scroll offset, which fixed the Loglens calendar. `GET /{name}` serves bytes stored with `SetImage`, `GET /audio` serves `SetAudio`, and `/state` reports `audioGen`, `audioPlaying`, and `audioVolume`. `Serve` creates the audio context `Run` uses.

Example fixes that came out of the sweep: Spotify navigation no longer restarts the track, Dino HI score, a Flappy Bird wasm canvas preview, Loglens calendar and scroll in `-web`, Blackhole frames go back through `SetData` (a `PatchElement` on the inline `<img>` was refused every push, so new frames never reached the screen), and Telegram and Teams thread view and word wrap. `-web` on `blinkless-all` now drives the same live sessions as the window (#78).

### Security and hardening

The second and third review passes are the largest hardening change since the audit in v0.0.3. Evidence rows are in `plans/v0.0.4/go-review-hardening/`.

- **Injection and XSS (#98).** The WSL file picker runs a static PowerShell script and passes the title through the environment. A smart-quote title no longer reaches the command line (`SEC-WSL-01`). `-web` theme overrides reject `</`, closing a `<style>` breakout (`SEC-WEB-01`). SVG and image routes use an explicit image MIME allowlist, `Content-Security-Policy: sandbox` on SVG, and `X-Content-Type-Options: nosniff`. `/dom`, `/state`, `/audio`, and the image routes enforce `sameOrigin` (#93). Raw-text tag detection now needs a delimiter after `</script` and `</style` (#99).
- **Secrets (#98).** `Copy` and `Cut` on a password field return nothing (`SEC-PWD-01`). The Linux, Android, and Windows accessibility bridges never expose or accept a password value.
- **Network (#98).** Fetch private-address checks use `netip` ranges, and the exported sentinels above let callers branch on the failure. `CheckDOM` also covers executed templates. `-web` writes its socket frames without holding the server lock, caps subscribers at 64, bounds shutdown to 2 s, and closes subscriber channels on cancel (#93). `POST /click` releases `s.mu` before it writes an error or redirect (#99). pprof routes carry per-route deadlines, and `debugWriteDeadline` seconds are capped (#98, #99).
- **Filesystem (#93, #99).** `crash.Write` keeps pre-existing directory permissions, and the `Lstat` check runs before `MkdirAll`. `filepick` preserves context errors.
- **Concurrency (#98).** Data races fixed in `-web` click, `bindWrite`, scroll and navigation (`p.dataMu` discipline), and `fetch.UseSystemProxy`. Xlib calls on WSLg run under one `x11Mu` with a scoped error handler (a C `pthread_mutex_t`), and unrelated `BadWindow` errors no longer discard a window ID that was already found (#99). `XResizeWindow` rejects non-positive sizes. `internal/render`'s `engineMu` is **removed** now that blinkless fixed the `EnsureAttrs` and `SheetCache` races. Removing it exposed a race in `TestMotionWrappersConcurrentWithParse`, which mutated one display from two goroutines. The test now gives each mutator its own page. Engine calls that panic are still recovered.
- **Correctness (#93, #98).** `NaN` guards in layout and replay input, an emoji parser fix, a display cache keyed on sequence rather than pointer, `ui/diff.go` patch budgets that no longer depend on order, a `View.Update` zero-patch generation check so ticks do not capture ops, and `recordPartialBounds` handling of an empty rect.
- **Performance and lifecycle (#98).** `isWSL()` is memoized, `textRasterCache` touch is O(1), motion easings are compiled once, `placeUIA()` is throttled, and the backdrop, blend, and `rendererPool` buffers are bounded and reset on return (#99).
- **Known gap.** `internal/bitmap` `TestSoftwarePainterReproducesAutospaceGap` was failing at #98 and also failed on the previous blinkless pin. It was recorded as pre-existing. Check that it passes on the final pin before tagging.

### Components and examples

- **`chadcn.Interactions` (#79).** An application installs tab keyboard handling, slider keys and pointer drag, and dialog focus and Escape once instead of rewriting them. `examples/builder/ui-builder` and `examples/builder/loglens` use it. The slider drag code in `ui-builder` shrinks by about 100 lines. `TestManagedInteractionsNative` drives the behavior on the real update loop. The component values stay owned by the host.
- **`ui` properties.** `ui/properties_generated.go` is generated from the compatibility matrix and gained setters for the properties blinkless added. `make ui-gen-check` guards drift.
- **`examples/blinkless-all` (#78, #82, #93).** One window for properties the engine implemented late, then generated tabs for all 688 CSS properties and 40 validated tabs. Each tab compares the description, the painted picture, and the readout. Fixes found by testing the real window: `shape-image-threshold` registered no mask, `copy-into` did not draw the heading it describes, and a scrolled action sheet painted over the pinned header. `properties.json` is read from the pinned `blinkless` module first, and generated tabs evaluate lazily, which removes about 1,254 eager startup layouts (#93).
- **`blinkless-scroll` and `blinkless-next`.** `blinkless-scroll` is the host side of scroll timelines. `blinkless-next` opens one blinkless HTML fixture in the window. The preview tool also opens the ten-style page with no arguments, and starts with no sound device.
- **`lg-remote` (#81).** Refreshed phone UI, a Comfort layout, settings, accents with blue as the default, and a TV keyboard.
- **`examples/login`** gained the TalkBack bridge (#96), and `.debug` APKs install beside signed releases.
- **`examples/code-view`, `examples/builder/loglens`.** Loglens modal layering was fixed: the top bar sits at z-index 45, above the topbar at 40 and below modal backdrops at 50 to 55 (#88).
- **Example tests** were stabilized for CI: an isolated Git fixture for Code Review startup, a spreadsheet test race fix, a gallery timeout sized for race instrumentation, and raster tests that finish under `-race` (#87, `c3163747`).
- **Layout.** 66 top-level example directories (69 in v0.0.3) after three additions and six removals. `examples/maintenance.json` marks the maintained teaching set `bind`, `desktop-cat`, `frontend`, `forms`, `list`, `login`, `telegram` and the rest experimental or support (`make examples-check`).

### Structure, CI, and tooling

- **Package cohesion (#83).** `internal/page` 317 to 52 files, `internal/window` 406 to 70, `internal/replay` 146 to 29, root non-test files 48 to 18, with no test dropped. Seven CSS motion packages merged into `internal/motion`, which no longer imports a preview helper. `internal/frame` and `internal/motion` share one bounds routine (`dirty.OpBounds`, `dirty.LocalPaintBox`), and stroke-width growth now applies on the identity-transform path too. `internal/viewtransition` no longer imports root `ownframe`. Preview packages moved to `cmd/internal/`. Exported receiver methods on `Page` went from 149 to 144. `AGENTS.md` shrank from about 31.5 KB to about 9 KB.
- **Footprint (#83, #85).** `scripts/check-repo-size.py` fails on any newly tracked file over 5 MiB against a frozen allowlist that can only shrink (10 entries at the end of this range). `make clean-local` removes local build and profile leftovers. The Pages wasm and Android release APK payloads left the tracked tree (#85). They remain in Git history, and the signed Android reference files stay on their existing release.
- **Pages (#84).** `.github/workflows/pages.yml` builds the site and nine wasm demos, checks the artifact, and deploys from `master` when Actions is the publishing source. The first run succeeded and a browser opened all nine demos before the payloads were removed. The site check runs 140 layout checks.
- **CI.** Pull requests skip the heavy jobs and push runs no longer duplicate (#80). A `windows-arm64` job on the `windows-11-arm` runner runs on master or with `[ci only=windows-arm64]` (#95). A `[ci only=examples-race]` job exists for the example race timeout (#87). `vuln` installs X11 headers before `govulncheck` loads packages. Blinkless module lookups and the snapshot fixture are bounded, and the `go.work` sibling path was dropped.
- **Module boundaries (#98).** `golang.org/x/perf` and the `govulncheck` tool moved out of the library `go.mod` into the `scripts/` module (`scripts/cmd/benchscore`). `make module-check` guards the split. The library module keeps `blinkless`, Ebiten, `x/image`, `x/text`, and `x/time`.
- **Lint and docs gates.** `make lint` adds doc-pin drift and relative-link checks (`scripts/doc_claims.py`), the generated-`ui` drift check, `internal/` import gates, and the repo-size check. The golangci-lint findings in `internal/window`, text stroke, and the examples were cleared (#78). `scripts/check-graphics.sh` and `scripts/test_graphics_gate.py` list the new pixel tests (20 Xvfb cases at #94).
- **Docs.** New: `documentation/updates.md` (which update API to call) and `documentation/property-profiling.md` (the v3 harness). Reworked: `platforms.md`, `accessibility-native.md`, `atspi.md`, `web.md`, `window.md`, `scrolling.md`, `css-motion.md`, `testing.md`, `features.md`, `compatibility-matrix.md`, and `README.md`. The three native-status documents were reconciled after #95 and #96 (#97), so none claims Windows UIA, Android TalkBack, or Windows IME is missing. `documentation/README.md` is four tiers with a three-document start path.

### Blinkless

This release pins `blinkless v0.1.6` (see the TODO above) in both `go.mod` files, equal in both, with no `replace`. The pin moved through these points in this range.

| Pin | Where | What the host gained |
|-----|-------|----------------------|
| `v0.1.1` | v0.0.3 | Starting point |
| `v0.1.2` pre-release `f68817e582e2` | #79 | Viewport sticky and height media queries (blinkless 39) |
| `v0.1.3` | #88 | Roughly 5x faster dashboard resize and layout, `patchBoxes` spare capacity so scoped hover stops falling back |
| `v0.1.4` | #89 | `SameBoxes` incremental box patching, cross-relayout flex override and style memo caching, pooled clone and batch buffers, pointer-receiver `Op.SameExtra`, `Op.MixExtraStamp`, `Op.HasExtra`, `Op.HasBoundExtra` |
| `v0.1.5` | #92 | Paint-only `background-color` restyle fast path, root margin splice escalation, single-slab op payload cloning, grouped hover splice stabilization, layout-animation base style caching, bounded image filter and clip-path mask PNG caches, zero-copy ASCII word breaking |
| pre-release `df3f742da29d` | #93 | `html.Document` leak fix and a 32 MiB raster cache cap (blinkless 45) |
| pre-release `b2fe5ba5105e`, `44951cb`, `8bbafad`, `658f99356cf1` | #95, #98, #99 | Fixes for the `EnsureAttrs` and `SheetCache` races (blinkless 48) and the review follow-ups (blinkless 49) |
| `v0.1.6` | release | Tag of the above. The engine's `plans/v0.1.6/` records its own changes. |

Blinkless pull request 47 (CJK glyphs in the Windows IME field) is a separate engine change. This release does not require it.

### Upgrade notes

1. Change the engine pin if you pinned it yourself. Keep one blinkless version across your module and any examples module.
2. Replace imports of `components/{apexcharts,chartjs,echarts,nivo,recharts,shadcnchart,tremor}` with the copy under `examples/builder/charts/` or your own package.
3. If you hand a `ui` document to `ownframe.New`, `Load`, or `Page.Load` without `ui.NewPage` or `ui.Mount`, set `StaticHTML`.
4. If your tests read a password through Copy or Cut, read it through state instead.
5. Android apps that want TalkBack call `androidaccess.Tree` and `androidaccess.Perform` from their `ebitenmobile` package and attach `dev.ownframe.android.AccessibilityBridge` to the `EbitenView`. `examples/login` is the working reference. Do not export `GOWORK` for `ebitenmobile bind`.
6. Windows builds: the UIA provider needs no cgo. Nothing in your `main` changes. NVDA and Narrator are the validated readers.
7. Tooling: `golang.org/x/perf` and `govulncheck` are no longer in the library `go.mod`. Run them from `scripts/`.

### Release steps

This note does not create a tag. In order:

1. Tag blinkless `v0.1.6`.
2. Bump both `go.mod` files to it and clear the TODO above. Run `make test`, `make lint`, `make graphics-check`, and `make test-race`.
3. Merge this note's pull request, then tag `v0.0.4` on the merge commit.
4. In `examples/go.mod`, require `github.com/chinmay-sawant/ownframe v0.0.4`, drop `replace github.com/chinmay-sawant/ownframe => ../`, and tag `examples/v0.0.4`. A tagged examples module cannot keep a replace.
5. Open pull request #100 (`fix/windows-arm64-ci`, Windows placement tests) is not in this range. Decide whether it lands before the tag.

### Validation recorded with the work

All of this is what the pull requests and `plans/v0.0.4/` record. It was not rerun for this note.

- #79 closed 22 accessibility rows and the CSS rows with retained evidence under `plans/v0.0.4/evidence/`, including `make graphics-check` (14 required Linux checks) and a Chrome comparison of the sticky fixture.
- #93: `make test`, `make lint`, `make graphics-check` passed.
- #94: `make test TEST_P=1`, `make graphics-check` (20 Xvfb cases), and the docs, plans, examples, tools, and repo-size checks passed. `make test-race` was not run.
- #95: `make lint` found 26 issues in the new Wayland code and `-race` found one data race in a test. Both were fixed, and eight consecutive `-race` runs of `internal/clipboard` were clean. Windows `Run`, custom close, clipboard read, and native file selection were exercised on a Windows host (#85).
- #98: `make lint`, `make graphics-check`, the maintenance checks, and `go test -race` on the affected packages passed. The full `make test` and `make test-race`, the Windows `native-package` job, and the Android `android-bind` job were left to CI, and the follow-up in #99 records the `android-bind` result that closed `HYG-MOD-02` and `HYG-API-02`.

### Limits

- No `v0.0.4` tag and no GitHub Release is created by this file.
- macOS has no accessibility adapter and no validated host. JAWS is untested. TalkBack speech on a physical device is untested, and a native arm64 Android launch is open.
- The six `[~]` Windows-host rows listed under Accessibility need a Windows runner before they can close.
- Wasm, `-web`, and `Serve` are preview and debugging hosts. `Serve` exposes no clipboard route, and `ServeWithOptions` keeps debug and pprof behind explicit `InsecureAllow*` flags.
- The performance figures are quoted from pull request records. `documentation/performance-baseline.md` carries older local samples and is not this release's result.
- Branch protection, Gradle verification metadata, ClusterFuzzLite, and signing when the key is missing remain open hardening items carried from v0.0.3.
