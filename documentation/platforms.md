# Platforms

1.0 native hosts are desktop `Run` (Linux, Windows, macOS) and Android
`BindMobile`. That is a native window for an operator you control: one
compiled Go process and an Ebiten window or Activity, no Chromium, no
WebView. Orca, NVDA, VoiceOver, and TalkBack cannot inspect it. Native
does not mean OS widgets or a screen-reader product for other people.

`scripts/browser.sh` is a wasm canvas preview. `-web` / `Serve` is a loopback
PNG debugger. iOS has no Xcode project and is not a 1.0 target.

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

## Serve PNG debugger

`Serve` serves the latest PNG and drives the page over HTTP. Start it with
`-web` and change the address with `-addr`:

```
go run ./examples/login -web
```

Every example has a default port; [examples/readme.md](../examples/readme.md)
lists them all. login defaults to 8091, web to 8110, and platform to 8115.
[web.md](web.md) has the routes, the click map, and what the mode drops. A
file-backed page is polled before `GET /` and `GET /frame.png`, and the shell
page refreshes the image while a watch is active
([hot-reload.md](hot-reload.md)). There is no tick, no audio, and no
relayout when the browser tab resizes. A handler panic becomes HTTP 500.
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

| | Desktop `Run` | Android `BindMobile` | Wasm preview | `Serve` PNG debugger |
|---|---|---|---|---|
| Input | pointer, keys, wheel ([pointer.md](pointer.md), [keys.md](keys.md)) | tap, Back, long press; see [android-app-creation.md](android-app-creation.md) | canvas pointer and keys | `POST /click`, `/type`, `/backspace` ([web.md](web.md)) |
| Clipboard | X11, Windows, macOS+cgo; Wayland uses `wl-copy`/`wl-paste` when on PATH; missing helpers are a no-op/`ErrUnavailable` ([clipboard.md](clipboard.md)) | process memory (`internal/clipboard/clipboard_other.go`) | memory copy | none |
| Files | desktop dialog (`internal/filepick`) | typed name; no file dialog. Telegram camera/gallery is app Java | typed name | typed name |
| IME | Composer path with preedit; native IM unverified ([interaction.md](interaction.md)) | `exp/textinput` | canvas text | `/type` only |
| A11y | Go snapshot only. Experimental Linux AT-SPI mapping, `Live` is false; no UIA, no NSAccessibility. Orca and NVDA cannot inspect ([accessibility.md](accessibility.md), [atspi.md](atspi.md)) | Go snapshot; no TalkBack | Go snapshot | snapshot on the page object; no AT |
| Packaging | unsigned `scripts/package.sh` ([packaging.md](packaging.md)) | `scripts/android-release.sh` sideload APK; Play is not 1.0 ([play-store.md](play-store.md), [APPS.md](../examples/android-host/APPS.md)) | wasm zip preview | not a ship target |
| Idle | desktop park after clean frames (`idle_supported.go`) | off (`idle_supported_other.go`) | off | none |
| Tick | yes ([frames.md](frames.md)) | yes | yes | no |
| Hot reload | file watch, 250 ms poll ([hot-reload.md](hot-reload.md)) | none; file read once | none; file read once | poll per `GET /` and `/frame.png` |
| Crash | report file 0600 ([crash.md](crash.md)) | panic recover and file | recover, no file | handler 500, no file |
| Resize | window drag relayouts ([window.md](window.md)) | rotation and split relayout | browser canvas resize | no; picture stays at `Page.Size()` |

Audio: `Run` and `BindMobile` create the 48 kHz Ebiten audio context; `Serve`
does not ([window.md](window.md), [features.md](features.md)). Drop: desktop
and wasm canvas only ([drag-drop.md](drag-drop.md)). The window modes replay
the display list or blit the fallback bitmap; `Serve` paints a PNG from the
template source ([screen.md](screen.md)).
