package scene

import "github.com/chinmay-sawant/ownframe/examples/tetris/game"

// previewShapes are the four cells of each kind inside a 4x4 preview box,
// top-left aligned. The core game owns rotation and wall kicks; this is
// only the next-piece picture.
var previewShapes = [8][4]game.Point{
	{},
	{{X: 0, Y: 1}, {X: 1, Y: 1}, {X: previewMid, Y: 1}, {X: previewMax, Y: 1}}, // I
	{{X: 1, Y: 0}, {X: previewMid, Y: 0}, {X: 1, Y: 1}, {X: previewMid, Y: 1}}, // O
	{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: previewMid, Y: 0}, {X: 1, Y: 1}},          // T
	{{X: 1, Y: 0}, {X: previewMid, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}},          // S
	{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: previewMid, Y: 1}},          // Z
	{{X: 0, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: previewMid, Y: 1}},          // J
	{{X: previewMid, Y: 0}, {X: 0, Y: 1}, {X: 1, Y: 1}, {X: previewMid, Y: 1}}, // L
}

// nextSet marks the preview cells for a piece in a 4x4 box.
func nextSet(k game.Piece) [4][4]bool {
	var out [4][4]bool
	if !k.Valid() {
		return out
	}

	for _, p := range previewShapes[k] {
		out[p.Y][p.X] = true
	}

	return out
}
