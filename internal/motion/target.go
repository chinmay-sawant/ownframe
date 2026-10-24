package motion

import "github.com/chinmay-sawant/ownframe/internal/cssmotion"

// fieldMask names the display values one target may write. BindBox narrows it
// from the keyframe properties a player carries, so a transform-only animation
// stops rewriting paint and a colour animation stops writing the matrix.
type fieldMask uint16

// One bit per field a keyframe can change.
const (
	fieldX fieldMask = 1 << iota
	fieldY
	fieldMatrix
	fieldOpacity
	fieldRGB
)

// maskAll is every field a target can carry. A zero mask writes nothing.
const maskAll = fieldX | fieldY | fieldMatrix | fieldOpacity | fieldRGB

// fieldMaskOf narrows the driver's mask to the paint fields a player owns. A
// full snapshot maps to maskAll; W, H, and Alpha are not driver write fields,
// so they never appear here.
func fieldMaskOf(mask cssmotion.Mask) fieldMask {
	if mask == 0 || mask == cssmotion.MaskAll {
		return maskAll
	}
	var out fieldMask
	if mask&cssmotion.MaskX != 0 {
		out |= fieldX
	}
	if mask&cssmotion.MaskY != 0 {
		out |= fieldY
	}
	if mask&cssmotion.MaskMatrix != 0 {
		out |= fieldMatrix
	}
	if mask&cssmotion.MaskOpacity != 0 {
		out |= fieldOpacity
	}
	if mask&cssmotion.MaskRGB != 0 {
		out |= fieldRGB
	}
	return out
}

// geometry reports whether a mask can move or resize a target, so a sample
// only re-measures the painted bounds when the geometry may have changed.
func (m fieldMask) geometry() bool {
	return m&(fieldX|fieldY|fieldMatrix) != 0
}
