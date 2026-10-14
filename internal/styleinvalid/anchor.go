package styleinvalid

// PromoteLayer records fixed/pinned layer promotion rules. Promoted reports
// whether a box with the given position value should be composited on its own
// layer.
func PromoteLayer(position string) bool {
	switch position {
	case "fixed", "sticky", "absolute":
		return true
	default:
		return false
	}
}

// Anchor resolves an anchor-based positioned element without requiring a full
// relayout: returns the offset delta relative to anchor box.
func Anchor(anchorX, anchorY, dx, dy float64) (float64, float64) {
	return anchorX + dx, anchorY + dy
}
