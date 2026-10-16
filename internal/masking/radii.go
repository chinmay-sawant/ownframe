package masking

import "github.com/chinmay-sawant/blinkless/layout"

// RadiiXY resolves a display fill or stroke's corner radii on both axes in
// canvas points, in CSS order: top-left, top-right, bottom-right, bottom-left.
// A missing Y radius copies the X radius and a zero X radius forces a zero Y
// radius, matching the engine's resolution.
func RadiiXY(op *layout.DisplayOp) (rx, ry [4]float64) {
	rx = [4]float64{op.Radius, op.Radius, op.Radius, op.Radius}
	if op.RadiusTopLeft > 0 || op.RadiusTopRight > 0 ||
		op.RadiusBottomRight > 0 || op.RadiusBottomLeft > 0 {
		rx = [4]float64{
			op.RadiusTopLeft, op.RadiusTopRight,
			op.RadiusBottomRight, op.RadiusBottomLeft,
		}
	}

	ry = [4]float64{op.RadiusY, op.RadiusY, op.RadiusY, op.RadiusY}
	if op.RadiusTopLeftY > 0 || op.RadiusTopRightY > 0 ||
		op.RadiusBottomRightY > 0 || op.RadiusBottomLeftY > 0 {
		ry = [4]float64{
			op.RadiusTopLeftY, op.RadiusTopRightY,
			op.RadiusBottomRightY, op.RadiusBottomLeftY,
		}
	}

	for i := range rx {
		if ry[i] <= 0 {
			ry[i] = rx[i]
		}

		if rx[i] <= 0 {
			rx[i], ry[i] = 0, 0
		}
	}

	return rx, ry
}

// FillRadii resolves a fill's circular corner radii. ok is false when a
// corner is elliptical, because the fill replay draws circular arcs only.
func FillRadii(op *layout.DisplayOp) (radii [4]float64, ok bool) {
	rx, ry := RadiiXY(op)
	for i := range rx {
		if rx[i] != ry[i] {
			return [4]float64{}, false
		}
	}
	return rx, true
}
