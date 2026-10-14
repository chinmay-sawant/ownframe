# Hello world across Go UI libraries

Each excerpt describes the same interface: plain `Hello world` text. Imports, window setup, and error handling are omitted for every library. These are UI excerpts, not complete programs. They use default styles and contain no state or event handlers.

## OwnFrame

HTML in a Go page configuration:

```go
ownframe.Config{
	HTML: `<div>Hello world</div>`,
}
```

OwnFrame passes the HTML to Blinkless for layout.

[Page configuration](screen.md) · [Complete window setup](window.md)

## go-gui

A Go label widget with the default text style:

```go
gui.Label("Hello world", gui.TextStyle{})
```

[Label API and default style](https://github.com/go-gui-org/go-gui#its-just-go) · [Complete window example](https://github.com/go-gui-org/go-gui/blob/main/examples/get_started/main.go)

## gogpu/ui

A Go text widget:

```go
primitives.Text("Hello world")
```

[Text constructor](https://github.com/gogpu/ui/blob/main/primitives/text.go) · [Complete window setup](https://github.com/gogpu/ui#quick-start)

## MyGo web frontend

HTML in the frontend page:

```html
<div>Hello world</div>
```

MyGo loads the page in the operating system's webview.

[Web frontend guide](https://github.com/egoist/mygo/blob/main/docs/frontend.md) · [Complete web app setup](https://github.com/egoist/mygo/blob/main/docs/getting-started.md#a-web-frontend)

## MyGo native UI

A Go text element inside a view. `c` is the view's `*ui.Context`:

```go
ui.Text(c, "Hello world")
```

[Text API](https://github.com/egoist/mygo/blob/main/docs/ui/text.md) · [Complete native app setup](https://github.com/egoist/mygo/blob/main/docs/getting-started.md#native-ui)
