# Compare with webview_go

webview_go passes HTML to the system web view from Go. OwnFrame defines screens in HTML and CSS too, but renders with its own engine: no WebView and no Chromium.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

The webview_go program below is complete, following the official basic example with `SetHtml` instead of navigation. Here, `w` is a webview instance.

### webview_go

```go
package main

import "github.com/webview/webview_go"

func main() {
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Hello world")
	w.SetSize(480, 320, webview.HintNone)
	w.SetHtml("<div>Hello world</div>")
	w.Run()
}
```

See the [SetHtml API](https://github.com/webview/webview_go/blob/master/webview.go) and [complete basic app](https://github.com/webview/webview_go/blob/master/examples/basic/main.go).

### OwnFrame

```go
ownframe.Config{HTML: `<div>Hello world</div>`}
```

HTML describes the screen. Blinkless parses HTML and CSS in process, and Ebitengine supplies the window and drawing.

## Where the code lives

| You write | webview_go | OwnFrame |
|---|---|---|
| Screen | HTML passed to the system web view | HTML templates |
| Styles | CSS | CSS |
| App logic | Go driving the web view | Go data and handlers |
