# Examples

## Maintained teaching examples

Start with `login`, then use `bind`, `forms`, and `list` for individual library capabilities. `desktop-cat` teaches desktop window behavior; `telegram` is the Android integration reference. These six examples are the maintained teaching set in [maintenance.json](maintenance.json).

The rest of this catalog contains experiments and shared example support. They remain in `make test` for compatibility, but their app scope is not a promise of library support. Add a focused teaching example only when the maintained set cannot demonstrate a new public capability.


Copy [login](login) or [ops-console](ops-console), not the brand clones.

Most examples open one HTML template in a window. `go run ./examples/<name>`
opens a desktop window. `go run ./examples/<name> -web` serves the same picture
in a browser on the printed address. Android-only examples document their
platform build in the example README. `examples/music` is shared code the
player imports, not an example of its own.

The feature list these examples follow is
[../documentation/features-examples.md](../documentation/features-examples.md).

When a change adds a distinct user-facing capability, add a small runnable
example or extend a focused feature example so people can try the behavior
without navigating a large app demo.

## Dev loop

The reload example reads its HTML from disk, so the edit is the build step:

```sh
go run ./examples/reload
```

Edit `examples/reload/index.html` and the open window redraws within a quarter
second. The counter keeps its value. `-reload=false`
turns the watch off, and `-web` serves the same page in a browser. The
behavior and the limits are in
[../documentation/hot-reload.md](../documentation/hot-reload.md).

## Original examples

| Folder | Shows |
|---|---|
| [login](login) | Sign-in screen with real inputs, clipboard, and an app-level undo stack. |
| [list](list) | Fixed-height row window helper: start, count, and spacers. Import it. There is no `go run`. |
| [ops-console](ops-console) | Five-screen yard ops console: login, ticket list, detail, settings, and a shift report. The board windows rows through examples/list. One shell template, one `View`, `data-action` clicks. Not a brand clone. |
| [forms](forms) | Every form control: text, password, textarea, checkbox, radio, select, file. |
| [bind](bind) | `data-bind` two-way binding of controls to a struct. |
| [input-lab](input-lab) | All nine form and input demos in one window: plain controls, two-way binding, a locked bound field, edit undo, OS clipboard, scroll list, sign-in, and CSS states. |
| [accessibility](accessibility) | Labeled form, desktop IME composition, semantic tree snapshot and action, password redaction, and scrolling thumb cursor. |

## Feature examples

| Folder | Feature | `-web` port |
|---|---|---|
| [perf-complex](perf-complex) | Complex 480-row performance baseline, optional live row windowing, and headless profiling with `-dump`. | 8135 |
| [layout](layout) | HTML -> CSS -> layout pipeline with hit-test boxes. | 8100 |
| [replay](replay) | Display-list replay and the bitmap fallback. | 8101 |
| [grid](grid) | CSS grid layout with clickable tiles. | 8140 |
| [css-masking](css-masking) | Retained mask-group compositing with alpha and luminance masks, SVG `mask-type`, and two-layer `exclude` composition. | 8139 |
| [png](png) | `Page.Probe().PNG` on-demand rasterization and caching. | 8103 |
| [controls](controls) | Click to focus, toggle, check, and cycle controls; typing edits; a file control opens the OS dialog. | 8104 |
| [bind-hooks](bind-hooks) | Two-way binding plus `BeforeEdit` veto and `Change` callback. | 8105 |
| [editing](editing) | Select all, undo, redo, and edits that survive a redraw. | 8106 |
| [states](states) | Host `:focus`, `:focus-visible`, `:hover`, `:active`, `:checked` styles. | 8107 |
| [scrolling](scrolling) | Oversized pages scroll on the wheel and touch drags, with no visible page scrollbar. | 8108 |
| [history](history) | `Load`, `Back`, `Forward`, `HTML`, and `data-action` routes. | 8109 |
| [web](web) | Plain `Serve` routes: `GET /`, `GET /frame.png`, `POST /click`, `POST /type`, `POST /backspace`. `GET /debug/state` stays off unless `ServeOptions` sets `InsecureAllowDebugState` or `InsecureAllowPprof` and the build uses `-tags ownframe_debug`. | 8110 |
| [ipc](ipc) | In-process `Send`, `Listen`, `Handle`, `Request`, with registered state and cancel. | 8111 |
| [fetch](fetch) | `Fetch` GET and `XHR` with method, headers, and body; a fetched PNG or JPEG becomes the page background. | 8112 |
| [clipboard](clipboard) | Copy, cut, paste, select all, undo, and redo through the page APIs. | 8113 |
| [crash](crash) | `Report` files and panic recovery by `Run`. | 8114 |
| [platform](platform) | Counter page on desktop native, Android native, and wasm preview. The counter click calls `Invalidate("count")`, so only that box repaints. | 8115 |
| [theme](theme) | `Config.Theme` and `SetTheme` restyle a running page; the toggle swaps custom properties. | 8116 |
| [wispr-flow-dashboard](wispr-flow-dashboard) | Wispr Flow app clone: collapsible sidebar with the Flow Pro logo, dictation history, notetaker, insights with usage/voice/leaderboard tabs, dictionary, snippets, style, transforms, scratchpad, invite, free month, settings, and help. UI only, sample data. | 8117 |
| [teams](teams) | Microsoft Teams-like app clone: dark theme by default with a light theme toggle, six rail menus (Activity, Chat, Teams and channels, Calendar, Calls, Files), a More apps flyout, a profile flyout with status, chat compose, live chat and search filtering, expandable channel threads with emoji reactions, a meeting card with date and time picks, a dial pad, and file details. The state persists in SQLite, the initial data comes from `examples/teams/store/seed/*.sql`, and `-db` selects the database file (`:memory:` keeps it in memory). Marvel-cast sample data. | 8118 |
| [spotify-player](spotify-player) | Dark Spotify-like player: eight screens (home, search, library, liked, browse, radio, queue, profile), live iTunes data, local free-music playback, and an animated now-bar equalizer. | 8119 |
| [telegram](telegram) | Telegram-like chat demo sized for a phone: chat list with search, contacts, settings with a dark theme, and one open conversation with a composer. The Android project under `telegram/android` builds an installable APK. | 8129 |
| [dino](dino) | Chrome-style dinosaur game: keyboard jump and duck, cacti and birds, running score, and a live frames-per-second readout. | 8120 |
| [lg-remote](lg-remote) | LG webOS remote for a UP7750-class TV. Dark by default, with a light theme toggle. Desktop uses Wi-Fi. Android adds Bluetooth HID. | 8131 |
| [flappy-bird](flappy-bird) | Flappy Bird: HTML/CSS scene, gravity and flap physics, scrolling pipe pairs, score and best, keyboard or click. | 8121 |
| [devtools](devtools) | Inspector dock: JSON element properties, frame counters, and operation outlines; F12 or Ctrl+Shift+I. | 8122 |
| [reload](reload) | File-backed page that redraws when `index.html` changes; the counter survives the reload. | 8123 |
| [drop](drop) | Dropped files print their absolute path on desktop and the entry name in a browser; a file control takes a path from the picker. | 8124 |
| [input](input) | Focus traversal, caret keys, drag selection, context menu, cursor shapes, touch scroll and pinch, page scrolling, and F11. | 8126 |
| [input-lab](input-lab) | The nine form and input demos (forms, controls, bind, bind-hooks, editing, clipboard, input, login, states) in one window. | 8130 |
| [desktop-cat](desktop-cat) | Transparent orange backpack cat with 30 expression PNGs, native click-through, a notification listener on the default Unix socket (`-notify-addr=auto`), TCP `127.0.0.1:6969` as opt-in, and a WASM canvas preview; web mode is a still picture. | 8128 |
| [resize](resize) | Window resize: two columns switch at a media query, a 100vw bar, rewrapping text, and a hover control. | 8127 |
| [snipping-tool](snipping-tool) | Desktop screen capture: toolbar capture, form-field crop, save and copy, OS tools with demo fallback. | 8136 |
| [departure-board](departure-board) | Split-flap departure board: eight rows of per-character tiles over a seeded timetable, four liveries with a night variant, tick-driven flaps that rebind per `Page.Probe().Generation()`, and a refresh bar found by the livery accent. | 8137 |
| [theme-studio](theme-studio) | Theme editor with a live preview: Apply sets the page theme and writes the watched theme file, a token strip shows colour swatches, four presets. | 8138 |
| [ops-console](ops-console) | Five-screen yard ops teaching sample: login, list, detail, settings, report. | 8141 |
| [download-manager](download-manager) | Download queue with bounded workers: progress and speed, pause/resume with validated HTTP range resume, retry, cancel, and durable SQLite history. | — |
| [live-log-viewer](live-log-viewer) | Multi-file log viewer: source sidebar, fixed-height rows, filters, follow/pause with unread count, entry detail, bounded export, and SQLite history. | — |
| [spreadsheet](spreadsheet) | Sparse editable grid with a formula bar, SUM/AVERAGE formulas, range selection, undo/redo, CSV import/export, and SQLite workbooks. | — |
| [system-monitor](system-monitor) | CPU, memory, disk, and network panels with downsampled graphs, a paged process table with frozen snapshots, detail view, and dummy or live collectors. | — |
| [tetris](tetris) | Playable falling-block game: SRS rotation, 7-bag randomizer, fixed-step tick, held-key repeat, next preview, and SQLite score history. | — |

The five app examples above (`download-manager`, `live-log-viewer`,
`spreadsheet`, `system-monitor`, and `tetris`) are desktop-only and have no
`-web` port. Each starts from dummy data, keeps its state in SQLite under
`os.UserConfigDir()/ownframe/<name>/` (with a data-directory flag to override
it), and documents recovery, platform limits, and measured numbers in its
README.

## Run

```sh
go run ./examples/layout          # desktop window
go run ./examples/layout -web     # browser on 127.0.0.1:8100
```

The platform example also builds for the browser canvas:

```sh
GOOS=js GOARCH=wasm go build ./examples/platform
```

Cursor keys, Space, and the other game keys arrive at the page through
`Handlers.KeyDown` and `Handlers.KeyUp`; the dino example is keyboard-only,
while flappy-bird flaps on a key or a click.
`Ctrl+C`/`Ctrl+X`/`Ctrl+V`, `Ctrl+Z`, and `Ctrl+Y` are handled by the
window, not by the examples.

## Android native apps

1.0 ships native Android apps through `BindMobile` and
`examples/android-host`. Native means a compiled Go process in an Ebiten
Activity, not OS widgets and not TalkBack. The APK index is
[android-host/APPS.md](android-host/APPS.md). Setup is in
[../documentation/android-app-creation.md](../documentation/android-app-creation.md).

`examples/telegram` is a native Android chat: list, search, IME, camera and
gallery through app Java, and pinned chrome. `scripts/android-release.sh
telegram` builds the release APK. `sh scripts/android.sh` binds a debug APK;
`sh scripts/android.sh install` also runs `adb install -r`. SDK steps are in
[telegram/android/README.md](telegram/android/README.md).

`examples/eyes` is a native Android landscape app: device motion, tick-driven
expressions, `scripts/android-eyes.sh` or `scripts/android-release.sh eyes`.
See [eyes/README.md](eyes/README.md).

## Audio

The Spotify player plays real music in the desktop window. It ships 30 CC0
tracks from the Open Lo-Fi collection and reads them from the binary, so
playback is offline and needs no API key and no network call. It plays them
through Ebiten audio while the frame tick moves the seek bar and equalizer.
`ownframe.Run` and `ownframe.BindMobile` create the Ebiten audio context the player
needs; `-web` has none. Every track's title, creator, license, and checksum are
in `spotify-player/player/assets/audio/catalog.json`, with the full attribution
table and the CC0 text in the README beside it. Tests inject a fake engine and
never touch the audio device or the network.

## Tests

```sh
make test
```

`make test` runs `go test -p 1 ./...` by default so the packages do not all
build and run at once. Set `TEST_P` for more parallelism: `make test TEST_P=4`.

## Perf benchmarks

| Folder | Shows | `-web` port |
|---|---|---|
| [perf-benchmarks/normal](perf-benchmarks/normal) | Benchmark A: small desktop form with a counter and a bound input, mostly idle. | 8131 |
| [perf-benchmarks/large](perf-benchmarks/large) | Benchmark B: thousand-row scrolling text list from one template range. | 8132 |
| [perf-benchmarks/flappy](perf-benchmarks/flappy) | Benchmark C: flappy-bird-like tick animation over retained display-list ops, space or click to flap. | 8133 |

The shared app code lives in `perf-benchmarks/benchutil`. The Go benchmarks
need no display:

```sh
go test ./examples/perf-benchmarks/ -run '^$' -bench . -benchmem
```

`BenchmarkNormalRedraw` times a warm small-form redraw,
`BenchmarkLargeRedraw` a warm thousand-row redraw alternating one pixel of
width, and `BenchmarkFlappyTick` one frame tick with no redraw.
