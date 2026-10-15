package window

import "github.com/hajimehoshi/ebiten/v2"

// panelFinger is one finger inside the devtools dock: its landing point,
// latest y, and whether it dragged past the tap slop. A finger that lifts
// without dragging taps; a dragged finger scrolls and never fires.
type panelFinger struct {
	id    ebiten.TouchID
	x, y  int
	lastY int
	drag  bool
}

// panelTrack is the dock's finger state across frames.
type panelTrack struct {
	fingers []panelFinger
}

// panelFind returns the tracked finger with an id, or -1.
func panelFind(fs []panelFinger, id ebiten.TouchID) int {
	for i, f := range fs {
		if f.id == id {
			return i
		}
	}
	return -1
}
