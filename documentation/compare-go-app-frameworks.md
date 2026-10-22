# Compare Go UI frameworks

Each example shows the same plain text, "Hello world". These are UI excerpts with imports, window setup, and application loops omitted for every project. They compare how you describe a screen, not complete app size. Fonts and default spacing differ between frameworks.

## OwnFrame

```go
ownframe.Config{HTML: `<div>Hello world</div>`}
```

HTML describes the screen. Blinkless parses HTML and CSS in process. OwnFrame
connects data and actions to Go, and Ebitengine supplies the window and
drawing. 1.0 native hosts are desktop and Android: one compiled Go process, no
Chromium, no WebView. That is not native OS widgets. iOS is not shipped. See
the [complete window example](window.md) and [rendering guide](screen.md).

## Wails v2

```html
<div>Hello world</div>
```

This goes in the frontend's HTML. Wails displays frontend assets in the system web view and can bind Go methods to JavaScript. See the [Wails architecture](https://wails.io/docs/howdoesitwork/) and [complete project setup](https://wails.io/docs/gettingstarted/firstproject/).

## Fyne

```go
widget.NewLabel("Hello world")
```

A Go widget describes the text. The application mounts it as window content. See [Fyne's complete first app](https://docs.fyne.io/started/hello/).

## Gio

```go
material.Label(theme, theme.TextSize, "Hello world").Layout(gtx)
```

Go emits the label's layout and drawing operations during a frame. Here, `theme` is a Material theme and `gtx` is the frame's layout context. See the [Label API](https://pkg.go.dev/gioui.org/widget/material#Label) and [complete Gio app](https://gioui.org/doc/learn/get-started).

## webview_go

```go
w.SetHtml(`<div>Hello world</div>`)
```

The Go binding passes HTML to the system web view. Here, `w` is a webview instance. See the [SetHtml API](https://github.com/webview/webview_go/blob/master/webview.go) and [complete basic app](https://github.com/webview/webview_go/blob/master/examples/basic/main.go).

## Ebitengine

```go
ebitenutil.DebugPrint(screen, "Hello world")
```

This call goes inside `Draw`, where `screen` is the drawing target. It uses Ebitengine's built-in debug font. Ebitengine provides the drawing loop rather than an HTML or widget toolkit. See the [complete Ebitengine hello world](https://ebitengine.org/en/tour/hello_world.html).

## Choose by how you build the UI

Wails and webview_go use the system browser engine for HTML screens. Fyne uses Go widgets, and Gio uses Go layout and drawing operations. Ebitengine gives you a drawing loop for custom interfaces and games.

OwnFrame gives you HTML and CSS screens with Go behavior through our own rendering engine. It has no browser DOM or JavaScript runtime, and it does not promise compatibility with arbitrary websites. The same page definition feeds desktop `Run` and Android `BindMobile`. wasm and `-web` are previews. iOS is not shipped. Check the [platform setup](platforms.md) and [supported features and limits](features.md), including the Go-snapshot accessibility limit, before choosing it for an app.
