package dirty

import (
	"github.com/chinmay-sawant/blinkless/layout"
)

// SameOp reports whether two operations paint the same pixels.
// The element id is not paint, so it is ignored. Numeric paint fields
// compare one by one with an early-out, so a position or color mismatch
// returns before the remaining fields are read.
func SameOp(a, b *layout.DisplayOp) bool {
	sameOpCalls.Add(1)
	if a.Kind != b.Kind || a.Text != b.Text || a.Font != b.Font ||
		a.Grid != b.Grid || a.Bold != b.Bold || a.FakeOblique != b.FakeOblique ||
		a.IsJPEG != b.IsJPEG || a.IsBackground != b.IsBackground ||
		a.Fixed != b.Fixed || a.Pinned != b.Pinned ||
		a.Positioned != b.Positioned || a.ZIndex != b.ZIndex ||
		a.ZIndexSet != b.ZIndexSet || a.StickyID != b.StickyID ||
		a.StrokeMask != b.StrokeMask || a.LineInset != b.LineInset ||
		a.RotateDeg != b.RotateDeg || a.XformSet != b.XformSet {
		return false
	}

	if a.X != b.X || a.Y != b.Y || a.W != b.W || a.H != b.H ||
		a.R != b.R || a.G != b.G || a.B != b.B || a.Alpha != b.Alpha ||
		a.Width != b.Width || a.Size != b.Size ||
		a.LetterSpacing != b.LetterSpacing || a.InkDescent != b.InkDescent ||
		a.Radius != b.Radius || a.RadiusY != b.RadiusY ||
		a.RadiusTopLeft != b.RadiusTopLeft ||
		a.RadiusTopRight != b.RadiusTopRight ||
		a.RadiusBottomRight != b.RadiusBottomRight ||
		a.RadiusBottomLeft != b.RadiusBottomLeft ||
		a.RadiusTopLeftY != b.RadiusTopLeftY ||
		a.RadiusTopRightY != b.RadiusTopRightY ||
		a.RadiusBottomRightY != b.RadiusBottomRightY ||
		a.RadiusBottomLeftY != b.RadiusBottomLeftY {
		return false
	}

	return samePayload(a, b)
}

// samePayload compares the rare payload: link, image, blend, outline,
// opacity, shaping, and transform.
func samePayload(a, b *layout.DisplayOp) bool {
	equal, sharedImg := a.SameExtra(b)
	if sharedImg {
		sameImageIdentity.Add(1)
	}
	return equal
}
