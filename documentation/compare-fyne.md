# Compare with Fyne

Fyne describes screens with Go widgets. OwnFrame defines screens in HTML and CSS, with app logic in Go.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

The Fyne program below is complete and follows the official hello world, with plain text.

### Fyne

```go
package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Hello world")

	w.SetContent(widget.NewLabel("Hello world"))
	w.ShowAndRun()
}
```

See [Fyne's complete first app](https://docs.fyne.io/started/hello/).

### OwnFrame

```go
ownframe.Config{HTML: `<div>Hello world</div>`}
```

HTML describes the screen. Blinkless parses HTML and CSS in process, and Ebitengine supplies the window and drawing.

## Where the code lives

| You write | Fyne | OwnFrame |
|---|---|---|
| Screen | Go widgets | HTML templates |
| Styles | Go theme and widget properties | CSS |
| App logic | Go state and event handlers | Go data and handlers |
