package window

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// inputWatch reports whether input arrived since the previous call. Ebiten
// gives each tick its own input snapshot, so it must run in Update. The two
// key and rune buffers are distinct and swap, so the comparison never sees
// the append it is comparing against. Just-pressed and just-released keys
// catch a tap that starts and ends inside one tick, when the held set is
// already empty again.
type inputWatch struct {
	keys   []ebiten.Key
	next   []ebiten.Key
	chars  []rune
	runes  []rune
	tap    []ebiten.Key
	cursor image.Point
	touch  int
	ready  bool
}

func (w *inputWatch) changed() bool {
	x, y := ebiten.CursorPosition()
	keys := inpututil.AppendPressedKeys(w.next[:0])
	chars := ebiten.AppendInputChars(w.runes[:0])
	tap := len(inpututil.AppendJustPressedKeys(w.tap[:0])) > 0 ||
		len(inpututil.AppendJustReleasedKeys(w.tap[:0])) > 0
	touch := len(ebiten.TouchIDs())
	changed := !w.ready || x != w.cursor.X || y != w.cursor.Y ||
		!slices.Equal(keys, w.keys) || !slices.Equal(chars, w.chars) ||
		tap || touch != w.touch || justPressedMouse() || wheelMoved()

	w.next, w.keys = w.keys, keys
	w.runes, w.chars = w.chars, chars
	w.cursor, w.touch, w.ready = image.Pt(x, y), touch, true

	return changed
}

func justPressedMouse() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle)
}

func wheelMoved() bool {
	x, y := ebiten.Wheel()

	return x != 0 || y != 0
}
