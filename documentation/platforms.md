# Platforms

One page opens in a desktop window, a browser canvas, a phone view, or an
HTTP picture page. `Run` covers the window and the canvas, `BindMobile` the
phone, and `Serve` the picture page. Each draws the same template and calls
the same handlers.

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
([crash.md](crash.md)).

## Browser canvas

```
sh scripts/browser.sh
```

The script builds `./examples/login` with `GOOS=js GOARCH=wasm`, copies
`wasm_exec.js` and `browser/index.html`, and serves the folder at
http://127.0.0.1:8092/. `Run` is the same call; on wasm the browser canvas is
the window. Resizing the browser lays the page out at the new size.

Ticks run each frame ([frames.md](frames.md)), key events arrive from the
canvas ([keys.md](keys.md)), and registered images paint. The clipboard uses
the memory copy ([clipboard.md](clipboard.md)). There is no file dialog, so a
file input focuses and keeps the typed name ([forms.md](forms.md)).

## Phone

`BindMobile` registers the page with Ebitengine's mobile view. Call it from
the package that `ebitenmobile bind` compiles, and do not call `Run` from that
package. [examples/login/mobile](../examples/login/mobile) is the pattern:

- `mobile.go` holds an `init` that calls `start` and an exported `Dummy` so
  the bind tool compiles the package.
- `start_mobile.go`, behind `//go:build android || ios`, calls
  `gpui.BindMobile`.
- `start_other.go`, behind the opposite tag, returns nil so desktop builds
  still compile.

[examples/platform/mobile](../examples/platform/mobile) repeats the pattern
for the platform example.

With the Android SDK or Xcode installed:

```
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
ebitenmobile bind -target android -javapkg com.chinmaysawant.gogpui -o go-gpui.aar ./examples/login/mobile
ebitenmobile bind -target ios -o go-gpui.xcframework ./examples/login/mobile
```

The generated view fills the screen, so rotating the phone or changing the
split lays the page out again. A tap arrives as a click
([pointer.md](pointer.md)). Typing needs a hardware keyboard; nothing in this
tree shows a soft keyboard. The clipboard uses the memory copy, and there is
no file dialog.

## Picture page

`Serve` serves the latest PNG and drives the page over HTTP. Start it with
`-web` and change the address with `-addr`:

```
go run ./examples/login -web
```

The default ports are login 8091, web 8110, and platform 8115.
[web.md](web.md) has the routes, the click map, and what the mode drops.

## Capability matrix

| | Desktop `Run` | Browser canvas | Phone `BindMobile` | `-web` `Serve` |
|---|---|---|---|---|
| Tick | yes | yes | yes | no |
| Key events | yes | yes | yes | no |
| Form typing | yes | yes | hardware keyboard | `/type` and `/backspace` |
| Clipboard | OS where supported | memory copy | memory copy | none |
| File dialog | desktop dialog | typed name | typed name | typed name |
| Audio context | 48 kHz | 48 kHz | 48 kHz | none |
| Crash report | yes | yes | yes | no |
| PNG over HTTP | no | no | no | `/frame.png` |

The clipboard is the X11, Windows, or macOS one where that works, and the
memory copy on Wayland, macOS without cgo, wasm, and mobile
([clipboard.md](clipboard.md)). `Run` and `BindMobile` create the 48 kHz
Ebiten audio context; `Serve` does not ([window.md](window.md),
[features.md](features.md)). The window modes replay the display list or blit
the fallback bitmap; the picture page serves a PNG painted from the template
source ([screen.md](screen.md)).
