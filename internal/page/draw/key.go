package draw

import (
	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Viewport is the frame size in CSS pixels.
type Viewport struct {
	Width  int
	Height int
}

// Key is every input that changes the layout of a frame. Images are a
// counter because a resolver function cannot be compared.
type Key struct {
	Source               string
	Width, Height        int
	Focus, Hover, Active string
	Theme                *css.Sheet
	Images               uint64
}

// KeyFor drops pointer state the sheets cannot match. Leaving it out of the
// key lets a hover or a press land without forcing a layout.
func KeyFor(source string, view Viewport, state render.State, pseudos render.StatePseudos, images uint64) Key {
	if !pseudos.Hover {
		state.Hover = ""
	}
	if !pseudos.Active {
		state.Active = ""
	}
	if !pseudos.Focus {
		state.Focus = ""
	}

	return Key{
		Source: source,
		Width:  view.Width,
		Height: view.Height,
		Focus:  state.Focus,
		Hover:  state.Hover,
		Active: state.Active,
		Theme:  state.Theme,
		Images: images,
	}
}

// Unchanged reports that this frame was already committed.
func Unchanged(hasLast bool, last, next Key) bool {
	return hasLast && last == next
}
