# Platforms

Target native hosts (currently in the `v0.0.x` preview cycle; no `v1.0` tag yet)
are desktop `Run` (Linux, Windows, macOS) and Android `BindMobile`. That is a
native window for an operator you control: one compiled Go process and an Ebiten
window or Activity, no Chromium, no WebView. On Linux, Orca and Accerciser can
read the page tree over AT-SPI; AT-SPI focus, activation, and `EditableText` for
enabled non-password text controls reach the page through the window loop.
VoiceOver and TalkBack cannot inspect it because the macOS VoiceOver and
Android TalkBack adapters are not shipped. Windows has a preview UIA provider
that is not yet validated with NVDA or Narrator. Native does not mean
OS widgets or a complete screen-reader adapter.

`scripts/browser.sh` is a wasm canvas preview. `-web` / `Serve` is a loopback
browser preview and debugger (live HTML document with a `?dom=0` PNG shell), not
a release host. iOS has no Xcode project and is not a planned `v1.0` target.

## Current (`v0.0.x`) vs planned (`v1.0`)

| Area | Current (`v0.0.x`) | Planned (`v1.0`) |
|---|---|---|
| **Desktop windowing** | Linux, Windows, macOS via Ebiten `Run` | Same, with real-host Windows & macOS validation records |
| **Android** | `BindMobile` + shared `android-host` Activity (arm64 APK) | Same, with reproducible CI/release APK builds |
| **IME** | Linux & Android validated; Windows wired through Ebiten IMM32 and covered by headless tests, but no live Japanese IME session yet; macOS open (`plans/README.md`) | Validated preedit/commit/cancel on Windows & macOS |
| **Clipboard** | X11, Windows, macOS (cgo); Wayland uses built-in data-control where offered; GNOME and Weston/WSLg still require `wl-clipboard` on `PATH` | Wayland without `wl-clipboard` on compositors lacking data-control (needs a focused-surface `wl_data_device` path from the window layer) |
| **Accessibility** | Go snapshot + live Linux AT-SPI bridge with `EditableText` | Explicit `v1.0` policy or native VoiceOver / TalkBack adapters and recorded NVDA validation of the Windows UIA provider |

Each host draws the same template and calls the same handlers. What the host
can do is not the same. Use the matrix below.

## Desktop window

```
go run ./examples/login
```

`Run` opens a window with a title bar. Drag an edge to resize it; the page is
laid out again at the new size, inside the min and max range
([window.md](window.md)). Clicks, the mouse wheel, and the keyboard work as
the other guides describe ([pointer.md](pointer.md),
[scrolling.md](scrolling.md), [keys.md](keys.md)).

`Run` installs the desktop file dialog ([forms.md](forms.md)), creates the
audio context ([window.md](window.md)), draws the page when it has not been
drawn yet, and enters the Ebiten loop. It returns when the window closes or a
handler returns an error, and it recovers a panic into a crash report
([crash.md](crash.md)). Idle park is desktop only
(`internal/window/idle_supported.go`).

## Android native

`BindMobile` registers the page with Ebitengine's mobile view. Call it from
the package that `ebitenmobile bind` compiles, and do not call `Run` from that
package. [examples/login/mobile](../examples/login/mobile) is the pattern:

- `mobile.go` holds an `init` that calls `start` and an exported `Dummy` so
  the bind tool compiles the package.
- `start_mobile.go`, behind `//go:build android || ios`, calls
  `ownframe.BindMobile`.
- `start_other.go`, behind the opposite tag, returns nil so desktop builds
  still compile.

[examples/android-host](../examples/android-host) is the shared Activity
helper: insets, Back, lifecycle, and `TouchCancellation`. Release APKs are
listed in [examples/android-host/APPS.md](../examples/android-host/APPS.md)
(login, platform, dino, lg-remote, telegram, eyes). Build with
`scripts/android-release.sh`. Unset `GOWORK` for the bind: gomobile copies
packages into a temporary module. Setup steps are in
[android-app-creation.md](android-app-creation.md).

The generated view fills the screen, so rotating the phone or changing the
split lays the page out again. A tap arrives as a click
([pointer.md](pointer.md)). Focusing a text field opens the phone's soft
keyboard: the window drives an `exp/textinput` session and the page paints
the preedit ([interaction.md](interaction.md)). The clipboard is process
memory ([clipboard.md](clipboard.md)). There is no file dialog
([forms.md](forms.md)). Idle park is off
(`internal/window/idle_supported_other.go`). There is no TalkBack node. A
file-backed page reads its file once at startup
([hot-reload.md](hot-reload.md)).

Telegram and eyes are the 1.0 exhibits ([examples/readme.md](../examples/readme.md)).

## Wasm canvas preview

```
sh scripts/browser.sh
```

The script builds `./examples/login` with `GOOS=js GOARCH=wasm`, copies
`wasm_exec.js` and `browser/index.html`, and serves the folder at
http://127.0.0.1:8092/. A first argument picks another example:
`sh scripts/browser.sh drop` serves the drag-and-drop example, where a file
dropped on the canvas prints the entry name. `Run` is the same call; on wasm
the browser canvas is the window. Resizing the browser lays the page out at
the new size.

Ticks run each frame ([frames.md](frames.md)), key events arrive from the
canvas ([keys.md](keys.md)), and registered images paint. The clipboard uses
the memory copy ([clipboard.md](clipboard.md)). There is no file dialog, so a
file input focuses and keeps the typed name ([forms.md](forms.md)). A
file-backed page reads its file once at startup; wasm has no watch
([hot-reload.md](hot-reload.md)). A panic is recovered but writes no report
file ([crash.md](crash.md)). Idle park is off.

## Serve web preview and debugger

`Serve` serves the live HTML document by default (with real form controls and a
boot script that posts input back to Go), plus a `?dom=0` PNG fallback shell
(`GET /frame.png`), and drives the page over HTTP. It is a loopback preview and
debugger, not a release host. Start it with `-web` and change the address with
`-addr`:

```
go run ./examples/login -web
```

Every example has a default port; [examples/readme.md](../examples/readme.md)
lists them all. login defaults to 8091, web to 8110, and platform to 8115.
[web.md](web.md) has the routes, the click map, and what the mode drops. A
file-backed page is polled before `GET /` and `GET /frame.png`. When the page
registers a tick callback (`SetTick`), `Serve` runs a 60 Hz tick loop
(`internal/web/tick_loop.go`), pushing snapshot updates and in-place tick
patches over `/events` (with `/state` polling fallback) while a watch or tick is
active ([hot-reload.md](hot-reload.md)). `Serve` creates the 48 kHz Ebiten audio
context on the host machine and also serves `GET /audio` with `/state` playback
metadata when a page publishes a clip with `SetAudio`
(`internal/web/audio_http.go`). There is no clipboard route, no relayout when
the browser tab resizes, and keys, clicks, form changes, and wheel events arrive
from the browser page. A handler panic becomes HTTP 500.
`Serve` writes no crash-report file.

## iOS

iOS is not a 1.0 target. There is no Xcode project, no IPA pipeline, and no
App Store listing. `BindMobile` is the Android host today. The mobile
package already compiles under `//go:build android || ios`, so an iOS bind
would reuse that Go entry. Next work, if anyone starts it, is
`ebitenmobile bind -target ios`, an Xcode host that owns insets, Back, IME,
and lifecycle the way `examples/android-host` does on Android, then signing
and App Store review. None of that is in this repo.

## Capability matrix

| | Desktop `Run` | Android `BindMobile` | Wasm preview | `Serve` web preview/debugger |
|---|---|---|---|---|
| Input | pointer, keys, wheel ([pointer.md](pointer.md), [keys.md](keys.md)) | tap, Back, long press; see [android-app-creation.md](android-app-creation.md) | canvas pointer and keys | `POST /click`, `/change`, `/key`, `/hover`, `/wheel`, plus `/type` and `/backspace` in PNG shell ([web.md](web.md)) |
| Clipboard | X11, Windows, macOS+cgo; Wayland prefers built-in data-control, then `wl-copy`/`wl-paste`; if missing, X11 CLIPBOARD when DISPLAY is set; otherwise no-op/`ErrUnavailable` ([clipboard.md](clipboard.md)) | process memory (`internal/clipboard/clipboard_other.go`) | memory copy | none |
| Files | desktop dialog (`internal/filepick`) | typed name; no file dialog. Telegram camera/gallery is app Java | typed name | typed name |
| IME | Composer path with preedit; native IM unverified ([interaction.md](interaction.md)) | `exp/textinput` | canvas text | browser input via `/change` (`/type` in PNG shell) |
| A11y | Go snapshot plus an experimental live Linux AT-SPI bridge (read, focus, activation, and `EditableText` on enabled non-password text controls); Orca and Accerciser can inspect the window. No UIA, NSAccessibility, or TalkBack adapters ([accessibility.md](accessibility.md), [atspi.md](atspi.md)) | Go snapshot; no TalkBack | Go snapshot | snapshot on the page object; no AT |
| Packaging | unsigned `scripts/package.sh` ([packaging.md](packaging.md)) | `scripts/android-release.sh` sideload APK; Play is not 1.0 ([play-store.md](play-store.md), [APPS.md](../examples/android-host/APPS.md)) | wasm zip preview | not a ship target |
| Idle | desktop park after clean frames (`idle_supported.go`) | off (`idle_supported_other.go`) | off | none |
| Tick | yes ([frames.md](frames.md)) | yes | yes | yes (about 60 Hz via `internal/web/tick_loop.go`, [web.md](web.md)) |
| Hot reload | file watch, 250 ms poll ([hot-reload.md](hot-reload.md)) | none; file read once | none; file read once | poll per `GET /` and `/frame.png`, push over `/events` |
| Crash | report file 0600 ([crash.md](crash.md)) | panic recover and file | recover, no file | handler 500, no file |
| Resize | window drag relayouts ([window.md](window.md)) | rotation and split relayout | browser canvas resize | no; Go layout stays at `Page.Size()` |

Audio: `Run`, `BindMobile`, and `Serve` create the 48 kHz Ebiten audio context.
In `-web`, Ebiten audio plays on the host machine, and a page can also publish a
clip with `SetAudio` for browser playback via `GET /audio` and `/state`
(`internal/web/audio_http.go`, [web.md](web.md), [window.md](window.md),
[features.md](features.md)). Drop: desktop and wasm canvas only
([drag-drop.md](drag-drop.md)). The window modes replay the display list or blit
the fallback bitmap; `Serve` serves the live HTML document by default and paints
a PNG from the template source for `GET /frame.png` and `?dom=0`
([screen.md](screen.md)).
