# Compare with Wails

Wails renders frontend assets in the system web view and binds Go methods to JavaScript. OwnFrame is a Go process with its own HTML and CSS engine: no WebView and no Chromium.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

The Wails program below is complete: `main.go` next to the frontend HTML. `NewApp` comes from the generated `app.go`; `wails init` scaffolds both plus `wails.json`.

### Wails v2

```go
package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "Hello world",
		Width:  480,
		Height: 320,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
```

```html
<div>Hello world</div>
```

This goes in the frontend's HTML. See the [Wails architecture](https://wails.io/docs/howdoesitwork/) and [complete project setup](https://wails.io/docs/gettingstarted/firstproject/).

### OwnFrame

```go
ownframe.Config{HTML: `<div>Hello world</div>`}
```

HTML describes the screen. Blinkless parses HTML and CSS in process, and Ebitengine supplies the window and drawing.

## Where the code lives

| You write | Wails | OwnFrame |
|---|---|---|
| Screen | HTML frontend assets | HTML templates |
| Styles | CSS | CSS |
| App logic | Go bound to JavaScript | Go data and handlers |

Wails shows its interface through the system web view. OwnFrame renders with Blinkless, its own HTML and CSS engine written in Go, with no browser DOM or JavaScript runtime.
