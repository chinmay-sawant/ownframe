package form

// RevealOrigin is the scroll origin that brings the box into the viewport.
// A box already inside is left where it is. The origin is never negative.
func RevealOrigin(boxX, boxY, boxW, boxH float64, x, y, width, height int) (int, int) {
	if boxY < float64(y) {
		y = int(boxY)
	} else if boxY+boxH > float64(y+height) {
		y = int(boxY+boxH) - height + 1
	}

	if boxX < float64(x) {
		x = int(boxX)
	} else if boxX+boxW > float64(x+width) {
		x = int(boxX+boxW) - width + 1
	}

	return max(0, x), max(0, y)
}
