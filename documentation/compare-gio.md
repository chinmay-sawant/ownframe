# Compare with Gio

Gio emits the screen's layout and drawing operations from Go during each frame. OwnFrame defines screens in HTML and CSS, with app logic in Go.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

The Gio program below is complete and follows the official get-started guide, trimmed to unstyled text.

### Gio

```go
package main

import (
	"log"
	"os"

	"gioui.org/app"
	"gioui.org/op"
	"gioui.org/widget/material"
)

func main() {
	go func() {
		window := new(app.Window)
		err := run(window)
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func run(window *app.Window) error {
	theme := material.NewTheme()
	var ops op.Ops
	for {
		switch e := window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			title := material.H1(theme, "Hello world")
			title.Layout(gtx)
			e.Frame(gtx.Ops)
		}
	}
}
```

See the [Label API](https://pkg.go.dev/gioui.org/widget/material#Label) and [complete Gio app](https://gioui.org/doc/learn/get-started).

### OwnFrame

```go
ownframe.Config{HTML: `<div>Hello world</div>`}
```

HTML describes the screen. Blinkless parses HTML and CSS in process, and Ebitengine supplies the window and drawing.

## Where the code lives

| You write | Gio | OwnFrame |
|---|---|---|
| Screen | Go layout and drawing operations | HTML templates |
| Styles | Go theme values | CSS |
| App logic | Go state and event handlers | Go data and handlers |
