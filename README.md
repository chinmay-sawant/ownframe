# ownframe

Formerly known as **go-gpui**.

<p align="left">
  <img src="assets/brand/gopher.webp" alt="ownframe Gopher mascot" width="140" />
</p>

Write screens in HTML and CSS, and application logic in Go. ownframe renders with [blinkless](https://github.com/chinmay-sawant/blinkless) and opens a window with Ebiten. It ships no Chromium, WebKit, WebView, or JavaScript runtime.

1.0 ships native Go apps on desktop (Linux, Windows, macOS) and Android. Native means one compiled Go process and an Ebiten window or Activity. It does not mean OS widgets or a screen reader.

`scripts/browser.sh` builds a wasm canvas preview. `-web` is a loopback PNG debugger. The capability matrix is in [documentation/platforms.md](documentation/platforms.md).

## Demo

ownframe preview. HTML and CSS layout with a layout engine written in Go, no Chromium or WebKit.

[![ownframe preview](assets/screenshots/demos/ownframe-preview.webp)](https://x.com/chinmay_sawant_/status/2106788230154871126)

[Watch the full preview on X](https://x.com/chinmay_sawant_/status/2106788230154871126)

Desktop cat overlay. Transparent, click-through cat that reports what is happening inside opencode.

[![Desktop cat overlay](assets/screenshots/demos/desktop-cat.webp)](https://x.com/chinmay_sawant_/status/2106829998409789929)

## Get started

```sh
go get github.com/chinmay-sawant/ownframe
```

The root package is `ownframe`. The repository and module path already use the new name; the first tagged release at that path is `v0.0.2`. The current library cycle is `v0.0.3`. See the [rename and migration notes](documentation/rebrand.md).

```go
package main

import (
	"context"
	"log"

	"github.com/chinmay-sawant/ownframe"
)

func main() {
	page, err := ownframe.New(ownframe.Config{
		Title:  "Hello",
		HTML:   `<h1>{{.Title}}</h1>`,
		Width:  480,
		Height: 640,
	})
	if err != nil {
		log.Fatal(err)
	}
	page.SetData(struct{ Title string }{"Hello"})
	if err := ownframe.Run(context.Background(), page); err != nil {
		log.Fatal(err)
	}
}
```

## Run the demos

From this checkout:

```sh
go run ./examples/login              # desktop native
sh scripts/android-release.sh login  # Android native APK
sh scripts/android-eyes.sh           # Android native (eyes)
go run ./examples/login -web         # PNG debugger preview
sh scripts/browser.sh                # wasm canvas preview
go run ./examples/desktop-cat        # desktop native overlay
```

The login demo accepts `secret` for both fields. [All examples](examples/readme.md), [desktop cat setup](examples/desktop-cat/readme.md), and [Android setup](documentation/android-app-creation.md) have the details. See [showcase.md](showcase.md) for the Telegram, Teams, Spotify Player, Pet Care, and other examples.

Nested module markers keep the website, generated site, screenshots, and development folders out of the library module archive. The examples remain in their own module.

## Documentation

Start with the [documentation index](documentation/README.md) or [feature list and limits](documentation/features.md).

- [Windows and sizing](documentation/window.md)
- [Desktop, Android, wasm preview, and Serve](documentation/platforms.md)
- [Templates and rendering](documentation/screen.md)
- [Forms and data binding](documentation/binding.md)
- [Theming](documentation/theming.md)
- [Packaging](documentation/packaging.md)
- [Comparisons with other UI libraries](documentation/compare-syntax.md)

The screen is a laid-out page, with no live DOM, JavaScript, video, canvas, or WebGL. 1.0 ships a Go accessibility snapshot; it does not connect AT-SPI, UI Automation, NSAccessibility, or TalkBack. The [feature list](documentation/features.md) records the supported behavior and remaining limits.

## Development

`make build` compiles both Go modules without linking executables. `make test` runs their tests one package at a time. `make lint` runs the Go linters, the file-size gate, and the website lint.

The React documentation and demo website lives in [frontend/](frontend/README.md). Run `npm ci` and `npm run dev` there with Node.js 22.12 or newer, then open the `/ownframe/` address Vite prints.

## AI assistance

I build this repository with the help of AI coding tools. I review and test the output before it lands. Generated content can be wrong, so I treat it as a draft, not a source of truth.

## Trademarks

The examples include demos that resemble third-party products. They are unofficial and built for illustration only. This project is not affiliated with, sponsored by, or endorsed by Microsoft, Spotify, Telegram, the Tetris Company, Google, Wispr, LG Electronics, Netflix, Amazon, or Disney. Product names, marks, and artwork belong to their owners.
