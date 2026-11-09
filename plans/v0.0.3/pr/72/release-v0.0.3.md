## v0.0.3

Third release of **ownframe**, prepared from tag `v0.0.2` (`4bbd832`, 2026-10-18) through `1202b5f2` (2026-11-10, merge of pull request #75). That range is 426 commits (390 non-merge, 36 merges). This file is the release body. Creating it does not create the git tag or a GitHub Release.

`ownframe.Page` is no longer an alias of the implementation type. It is a struct with an unexported inner page and an explicit method set. Callers that used methods promoted through the alias move those calls onto `Form`, `Edit`, `Scroll`, `Window`, `Access`, `Probe`, or `Testing`. CSS animations and transitions run from the page itself. Linux desktop `Run` serves the page tree on the accessibility bus. `Page.PatchElement` edits one element of a retained tree, and `Page.PatchElements` batches a set of edits through one shared style pass.

- Module: `github.com/chinmay-sawant/ownframe`
- Go: 1.26.10 (v0.0.2 was 1.26.4)
- Ebiten: `v2.11.0-alpha.0.20261103125056-58976b602532` (a pseudo-version, not a stable tag)
- Engine: `github.com/chinmay-sawant/blinkless v0.1.1`, no replace
- Examples module: still requires `ownframe v0.0.2` and still has `replace => ../`. The published `examples/v0.0.2` tag has no replace. A tagged examples module cannot keep one. Drop the replace only after this library tag exists.
- Previous release: v0.0.2
- Range: 4,002 files, +339,545 / −21,295

`go get github.com/chinmay-sawant/ownframe@v0.0.2` does not contain this `Page` wrapper.

### Highlights

| Area | What landed after v0.0.2 |
|------|--------------------------|
| **Public Page** | Defined wrapper. Doors are `Form`, `Edit`, `Scroll`, `Window`, `Access`, `Probe`, and `Testing`. `NewGame` drives the Ebiten loop in tests. |
| **Motion** | `@keyframes` and `transition` sample on the window tick. `SetAnimations` and `SetTransitions` turn them off. |
| **Patch** | `Page.PatchElement` returns an error. A refused edit is `ErrPatchRefused` and leaves the last display. `Page.PatchElements` batches edits. `LoadOptions.AdoptTree` takes an adopted source without a parse. |
| **Replay** | `backdrop-filter` paints `blur`, `saturate`, and host `refract`. Fixed layers stay at translation 0, 0 when a chunk blits. |
| **Pacing** | `WindowOptions.MaxFPS` caps desktop presents. `0` is uncapped. Accepted caps are 30 through 1000. |
| **Accessibility** | Linux `Run` discovers the accessibility bus and serves the page tree. AT-initiated actions return false. |
| **Components** | `ui`, `ui/theme`, and `components/chadcn`. The builder is experimental. Seven HTML chart sets joined in this cycle. |
| **Perf pipeline** | Batched patches, one patch display per batch, pooled vector paths, zero-alloc pointer moves, stable patch IDs, scoped hover and transition capture, `Display.Release` lifecycle. Internal only, no new required calls. |
| **Examples** | 17 new top-level directories, 69 total. None of the v0.0.2 directories were removed. New here are the Discord clone, the Blackhole GPU demo, and the invoice bench harness. |
| **Checks** | `make test` defaults to `-p 4`. Lint reads the plain golangci-lint version line. The pin is v2.15.0. |

### Migrating from v0.0.2

The root type changed from `type Page = page.Page` to a struct. Methods that the alias used to promote are not all methods of `Page`. `SetData`, `Handle`, `Redraw`, navigation, sizing, `Click`, `Boxes`, `WantFrameAt`, `WantFrameIn`, `ScrollTo`, and `ScrollBy` stay on `Page`.

| v0.0.2 call | v0.0.3 call |
|-------------|-------------|
| `FormValue`, `SetFormValue`, `FormChecked`, `SetFormChecked`, `FormSelected`, `Focus`, `FocusID`, `FocusNext`, `FocusPrev` | `Form()` |
| `Copy`, `Cut`, `Paste`, `SelectAll`, `Undo`, `Redo`, `Type`, `Backspace`, `DeleteWord`, `Submit`, `KeyDown`, `KeyUp`, `Hover`, `Press`, `Release`, `Drop`, `CancelTouches`, `LastFocusID`, `FocusedField` | `Edit()` |
| `AccessibilityTree`, `AccessibilityPerform` | `Access()` |
| `Stats`, `Display`, `PNG`, `Image`, `Element`, `Elements`, `ElementRegistry`, `TakeDirty`, `TakeDirtyRects`, `Generation`, `PatchDisplay`, `PatchOp`, `LastPatch` | `Probe()` |
| `PollReload`, `SetPerf`, `Perf`, `SetDevTools`, `DevTools`, `SetHotReload`, `Watching`, touch-scroll and touch-zoom setters | `Window()` |
| `SetWindowing`, `Windowing`, `SetAllowScroll`, `AllowScroll`, `ScrollOffset`, `SetScrollOffset`, `SetViewportPinZ`, `ViewportPinZ` | `Scroll()` |
| `Tick`, drag and long-press test hooks, `IMEReplace`, `SelectAt`, `PNGWithContext`, `TakeScroll`, `PrepareMobile`, `SetPicker` | `Testing()` |

`PatchElement` stays on `Page` and returns `error`. v0.0.2 returned `bool`. `ErrPatchRefused` means the edit was dropped and the cached tree was discarded. The last display stays. A page that does not retain state refuses every patch. Set `Config.RetainState` first.

`PatchElements(ctx, patches)` is the batched form with one shared style re-resolve and one patch display per batch. It is additive. `LoadOptions.AdoptTree` is the incremental builder path that adopts a patched source without a fresh parse and draw. It is additive.

These names were reachable through the alias and have no same-named export on the wrapper: `State`, `SetState`, `BindState`, `SyncState`, `ContextMenu`, `PreviousFocusID`, `BlurredFocusID`, `LastFocused`, `Drag`, `SelectWordAt`, `SelectLineAt`, `IMEAnchor`, `SetIMEComposition`, `CursorShape`, `SetDrawTime`, `FallbackDisplay`, `ScrollbarMode`, `PatchReport`, and the changed-property synonyms. `Probe.Element` and `Edit.LastFocusID` are the names that stayed. The store type is not re-exported. `internal/page` still has `State() *state.Store` for the implementation. External code cannot call it.

Later commits added calls that were not in the wrapper's first merge. `Load` is `Load(ctx, html, options ...LoadOptions) error`. A two-argument call still compiles. `LoadOptions.ReplaceCurrent` replaces the current history entry. `Page.ContextMenuFor` and `Page.MenuAction` are the menu API, with `type MenuItem`. That does not restore `ContextMenu() []host.MenuItem`. `Window` gained `SetAnimations`, `SetTransitions`, `Close`, `Minimize`, `ToggleMaximize`, `Maximized`, and the zoom allow-flags. `Probe.Motion` returns `MotionStats`. `Form` gained `RequestFocus` and `SetFocusScope`.

Root aliases that stay include `Box`, `Config`, `Handlers`, `Display`, `DisplayOp`, `Drop`, `Element`, `FetchResponse`, `ServeOptions`, and `WindowOptions`. The cycle also exports `type Secret` and `type Stats`. `pagetest` is a public package for hit tests. Clipboard exports `Read`, `Write`, `Clear`, `UseMemory`, and `ErrUnavailable`.

`Page.Testing` is for tests. Example programs do not call it. One `Page` has one owning loop. Concurrent calls on the same page are unsupported.

Applications use `Run` or `BindMobile`. Tests that need the game loop use `ownframe.NewGame`. The host screen type is not returned.

### Hosts

Desktop `Run` on Linux, Windows, and macOS, and Android `BindMobile`, are the 1.0 native hosts. Wasm from `scripts/browser.sh` is a canvas preview. `Serve` and `-web` are a loopback PNG debugger. iOS has no Xcode project. Play packaging is outside this release. `scripts/package.sh` builds unsigned desktop packages.

### Motion

Pull request #36 added the samplers and the 43-card catalog. Pull request #42 binds `@keyframes` and transitions on each publish and samples them before `SetTick`. Both default on. `Page.Window().SetAnimations(false)` and `SetTransitions(false)` are the page-wide switches. `Probe().Motion()` is the counter door.

`Run`, `BindMobile`, and wasm draw the sampled frame. `Serve` ticks, then `Redraw` replaces an active sample, so the browser preview stays on the declared paint.

The driver writes 2D transform, opacity, the nearest owner's colour, a 3D restamp, one layout relayout per frame, and one inline SVG restyle per frame. Transitions cover colour, opacity, 2D transform, `width`, `height`, `top`, and `left`. Discrete `visibility`, `content-visibility`, `z-index`, and `position` need `transition-behavior: allow-discrete`. A `display` cross-fade is not in this release. `filter`, `clip-path`, mask position, and the `translate` / `rotate` / `scale` longhands stay at the base paint. View transitions stay inside the package. The page does not capture an old and a new display for them.

An operation must not be driven by CSS and by `Page.SetTick` at the same time. Turn the matching switch off before a manual player writes the same elements.

The engine pin in the motion pull requests is older than HEAD. The pin in `go.mod` is the one named at the top of this note.

### Paint, patch, and the window

`Page.PatchElement` edits the cached tree and re-lays out that subtree. `SetText` drops children and skips emoji images. A size or container-query change refuses. A class that moves between two nodes is not one patch. The css-preview catalog at HEAD clicks through `SetData` and `Redraw` again. The method is there for callers who retain state.

`Page.PatchElements` takes a batch of element edits, re-resolves shared styles once, and builds one patch display for the batch. Stable IDs across the chadcn widgets, the inventory rows, and the Loglens surfaces keep the batch on the retained path. A refused batch reports `ErrPatchRefused` and keeps the last display.

`examples/trade` is the Meridian terminal: seven screens, synthetic data, CSS motion, and no `SetTick`. Hover can take the subtree relayout because `RetainState` is on. It does not call `PatchElement`.

`backdrop-filter` accepts one `blur()` in px, capped at 64, one `saturate()` number or percent, capped at 3, and host `refract()` in px, capped at 16. A second function, or any other function, rejects the value. A backdrop frame replays the whole display.

Pull request #64 keeps fixed chrome from jumping with the page scroll on a chunked blit. The pinned draw uses translation 0, 0. The per-op path already did.

`WindowOptions.MaxFPS` on `RunWithOptions` caps desktop presents and leaves VSync off. The merged default is `0`, uncapped. Values below 30, above 1000, or negative fall back to uncapped. Idle park is unchanged. The Frame tab shows MaxFPS, Present FPS, and Presents.

Windows RSS uses `GetProcessMemoryInfo` on the current process, sampled every 500 ms. An unknown read prints `-` on the MEMORY row. Web `GET /debug/state` omits `rss_bytes` when the reader fails. Linux stays on `/proc/self/statm`.

The late-cycle perf pass is internal: `Display.Release` lifecycle wiring, pooled vector paths with heap-free draw options, zero-alloc pointer moves, filtered transition capture, bound pooled memory, one reused hit spot with a lazy kid index, hashed adopted sources, and lengths kept without per-element maps. Nothing there changes a signature. Loglens scoped its interactive updates to the affected subtree, releases superseded displays, records hover-patch failures by stage and target, and keeps a stable closed tooltip label.

### Linux accessibility

v0.0.2 returned a detached accessibility snapshot and did not register it with an operating system. Linux desktop `Run` now discovers `org.a11y.Bus`, claims an ownframe name, embeds with the registry, and serves Accessible, Component, Text, Value, Action, Application, and Cache. It emits `state-changed` and `children-changed` with the `siiva{sv}` signature. The bus closes when `Run` returns.

Password values are empty. The node stays editable and does not advertise SetText. `DoAction` and `GrabFocus` return false.

The recorded check is Ubuntu 22.04 under Xvfb. libatspi read the tree, roles, names, bounds, and focus events. Accerciser 3.38 showed the tree only after a local Accerciser workaround, which is not an ownframe change. Orca 42 per-control speech was not verified. Speech Dispatcher captured the startup utterance `Screen reader on.` Wayland and a second monitor are outside that matrix. Issue #48 stays open.

Windows UI Automation, macOS VoiceOver, and Android TalkBack are not shipped. `Serve`, wasm, and Android do not use this bridge. `plans/README.md` still defers `NATIVE-A11Y`.

### Serve, fetch, and local data

Pull request #39 is the caller-visible hardening. `plans/v0.0.3/security/SECURITY_AUDIT_3.md` still has GEN-1 through GEN-50 checked. Branch and tag rules, the dependency-review action, Gradle verification metadata, push protection, private reporting, ClusterFuzzLite, and signing when the key is missing stay open in that audit's follow-ups.

- `ServeOptions{Perf, Debug}` still compile and are deprecated. Use `InsecureAllowPprof` and `InsecureAllowDebugState`. A non-loopback bind errors unless `InsecureAllowRemoteBind` is set.
- Mutating POST routes require `Origin` and fail closed without it. Debug and pprof routes are compiled only with `-tags ownframe_debug`.
- DOM depth stops at 256 and node count at 20,000. Field maxlength clamps at 4,096. Paste and programmatic writes clamp at 16,384 runes. `data:image/*` caps at 1 MiB. `file:` and `javascript:` are rejected when a URL is resolved. Open targets allow http, https, and mailto.
- Fetch uses a pinned transport, TLS 1.2 minimum, and rejects private addresses unless the caller uses `AllowPrivate`. Redirects cannot downgrade or hop to a private address.
- `secret.Secret` redacts format verbs and slog. `safego.Go` recovers into the crash report. `Run` errors no longer embed the crash path. Read `LastReportPath`.
- Crash directories reject dot segments and NUL, resolve symlinks, and stay mode `0700` / `0600`.
- Desktop cat notify defaults to a Unix socket. TCP is `-notify-addr`. The address must not be a wildcard.
- Download-manager job IDs are 16 bytes of hex. A CSPRNG failure panics.

### Components and examples

Import `github.com/chinmay-sawant/ownframe/ui`, `ui/theme`, and `components/chadcn`. `ui/builder.go` still calls the GPUI-style builder experimental. There is no reconciliation and no GPUI runtime parity. `theme.Light` and `theme.Dark` are not the chadcn sheets. Append `chadcn.Stylesheet` after one of them. ChadCN CSS sets `box-shadow` to `none`.

Pull request #70 added `Search`, `AreaChart`, `FilterBar`, `HexGrid`, `StatCard`, `StatusSegments`, and `SidebarSection`. No new `Page` method came with them. The pane scrollbar is `chadcn.ScrollState`. The thumb minimum is 24 px, and a range that fits draws no bar. This cycle also added seven HTML chart sets (apexcharts, chartjs, echarts, nivo, recharts, tremor, shadcnchart) with ui-builder galleries.

Runnable builder programs: `examples/builder/jmap-mail`, `ui-builder`, `loglens`, `mailpulse`, `inventory`, `discord`, plus `examples/compare/ownframe`. LogLens and jmap-mail both listen on `127.0.0.1:8132` unless `-addr` is set. ui-builder listens on `127.0.0.1:8131`, inventory on `127.0.0.1:8135`, mailpulse on `127.0.0.1:8148`, discord on `127.0.0.1:8152`. `examples/compare/mygo` and `examples/mygo-inventory` are separate modules and need Go 1.27.3 or newer. `examples/maintenance.json` marks `builder`, `compare`, and `mygo-inventory` experimental.

New top-level example directories since v0.0.2: `animations`, `blackhole`, `builder`, `code-view`, `compare`, `css-preview`, `depth-alert`, `frontend`, `grid`, `hello`, `list`, `motion-conformance`, `mygo-inventory`, `ops-console`, `performance`, `testing`, `trade`.

`examples/code-view` is the Code Review viewer. It opens the current directory, or the 17-file sample when the root is empty. A user can browse, edit, save, search, preview Markdown, diff, stage, discard, and revert a change block. The two editor groups scroll apart. Devtools stay off unless `-dev` is set or a session restores the dock. Issues #44 (TalkBack) and #45 (a Chromium comparison) stay open. An embedded terminal and a merge-conflict viewer were closed as not planned.

`examples/depth-alert` is an ARCore center-depth prototype with Walk, Stop, Stop now, and Depth unknown. Walk needs twelve clear frames. It is not a safe-path claim. Pull request #57 replaced the signed arm64 APKs for lg-remote, login, and platform only.

`examples/blackhole` is a GPU raymarched black hole with an HTML HUD. `examples/builder/discord` is a Discord clone demo with windowed transcript, fullscreen settings, and live typing. `examples/testing/ownframe` with `examples/testing/mygo` is the invoice benchmark harness. All three are demos, not supported hosts.

`examples/hello` serves `-web` on `127.0.0.1:8091`. `examples/grid` is desktop only. `examples/animations` serves `-web` on `127.0.0.1:8142`. `examples/performance` is the ten-scenario render-cost suite. `documentation/why-ownframe.md` is the About page on the site.

### Measurements recorded with the work

These numbers are the ones written in the pull request records. They were not remeasured for this note.

- Example tests, pull request #37: Spotify 355 s to 7.60 s, corebackend 32 s to 1.84 s, LG remote 26.3 s to 2.34 s, Wispr 10.8 s to 8.20 s, Teams 26.7 s to about 10 s in the full suite.
- `make test` on a 24-core machine, pull request #47: 149.91 s at `-p 1`, 50.46 s at `-p 4`.
- Hot-path pass, pull request #55: row windowing, soft wrap, Go highlight, warm dashboard and card, click, dirty drain, and idle tick cut to nanoseconds and microseconds at zero allocs. Quoted from that record, not remeasured here.
- `examples/performance`, one `warm=5` paste in the pull request #61 record: warm 1.33 µs to 4.49 µs on nine scenarios, 67.99 µs on the 1,000-row case, cold up to 593.33 ms. The docs do not carry a filled v0.0.3 table for that suite.

`documentation/performance-baseline.md` says its 2026-10-23 block is an earlier local sample. Do not read it as this release's result.

### Limits

- No `v0.0.3` tag, and no GitHub Release, is created with this note.
- Windows composing-input validation stays open (`NATIVE-IME`). macOS input validation stays deferred until a host exists.
- The 2026-10-23 Pixel 7 session is one reviewed snapshot. It does not prove every rotation or every input method.
- Pull request #50 (narrow code-view tabs) closed unmerged. Pull request #68 (a pluggable renderer) closed as a draft, unmerged. `internal/renderbackend` is not in this tree.
- Issue #51 (a JSX compiler), issue #67 (a native GPU backend), and issue #69 (native box-shadow parity) stay open.
- Fuzzy file ranking, symbol search, go-to-line, and a phone layout are not in Code Review.
- `filter` and a `display` cross-fade are not driven by the motion package.

### Install and run

```
go run ./examples/login
go run ./examples/hello -web
go run ./examples/code-view
```

The library:

```
go get github.com/chinmay-sawant/ownframe@v0.0.3
```

That `go get` works only after the tag exists. Until then, use this checkout.
