package cssmotion

import "github.com/chinmay-sawant/blinkless/layout"

// Capture returns paint values for operations inside a layout box.
func Capture(d *layout.Display, box layout.Box) []Paint {
	var out []Paint
	for _, op := range Operations(d, box) {
		out = append(out, Read(op))
	}
	return out
}

// Operations returns pointers to display operations whose centers lie in box.
func Operations(d *layout.Display, box layout.Box) []*layout.DisplayOp {
	return Resolve(d, OperationIndices(d, box))
}

// Resolve finds the current display operations at previously captured indices.
func Resolve(d *layout.Display, indices []int) []*layout.DisplayOp {
	var out []*layout.DisplayOp
	if d == nil {
		return out
	}
	for _, index := range indices {
		if index >= 0 && index < len(d.Ops) {
			out = append(out, &d.Ops[index])
		}
	}
	return out
}

// Contains reports whether an operation's center lies inside a layout box.
func Contains(d *layout.Display, box layout.Box, op layout.DisplayOp) bool {
	x, y := op.X+op.W/halfDivisor, op.Y+op.H/halfDivisor
	if op.Kind == layout.DisplayOpText {
		y = op.Y
	}
	p := d.PointsPerPixel
	return x >= box.X*p && x <= (box.X+box.W)*p &&
		y >= box.Y*p && y <= (box.Y+box.H)*p
}
