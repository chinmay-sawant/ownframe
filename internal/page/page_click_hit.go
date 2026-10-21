package page

// hit returns the last box that contains x, y.
// Boxes are in document order, so that box is the inner element. It
// delegates to the shared boxAt scan so clicks agree with Press and Hover.
func hit(boxes []Box, x, y float64) (Box, bool) {
	spot := scanBoxes(boxes, x, y)

	return spot.box, spot.ok
}
