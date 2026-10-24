# Compare with Ebitengine

Ebitengine gives you a drawing loop for custom interfaces and games: a `Game` implements `Update`, `Draw`, and `Layout`. OwnFrame builds on that same loop but takes HTML and CSS for the screen, with app logic in Go.
These examples display the same plain text, `Hello world`, with no added styles.

## The same screen

The Ebitengine program below is complete and follows the official hello world tour, with plain text.

### Ebitengine

```go
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, "Hello world")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello world")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
```

This call goes inside `Draw`, where `screen` is the drawing target. It uses Ebitengine's built-in debug font. See the [complete Ebitengine hello world](https://ebitengine.org/en/tour/hello_world.html).

### OwnFrame

```go
ownframe.Config{HTML: `<div>Hello world</div>`}
```

HTML describes the screen. Blinkless parses HTML and CSS in process, and Ebitengine supplies the window and drawing.

## Where the code lives

| You write | Ebitengine | OwnFrame |
|---|---|---|
| Screen | `Draw` calls on the game loop | HTML templates |
| Styles | Drawing code and fonts | CSS |
| App logic | Go `Update` and event handling | Go data and handlers |
