package dirty

import (
	"bytes"

	"github.com/chinmay-sawant/blinkless/layout"
)

// paintFields is how many numeric paint fields numOf compares at once.
const paintFields = 22

// SameOp reports whether two operations paint the same pixels.
// The element id is not paint, so it is ignored.
func SameOp(a, b *layout.DisplayOp) bool {
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

	return numOf(a) == numOf(b) && samePayload(a, b)
}

// numOf collects the numeric paint fields so one array comparison covers
// them.
func numOf(o *layout.DisplayOp) [paintFields]float64 {
	return [paintFields]float64{
		o.X, o.Y, o.W, o.H, o.R, o.G, o.B, o.Alpha, o.Width,
		o.Size, o.LetterSpacing, o.InkDescent, o.Radius, o.RadiusY,
		o.RadiusTopLeft, o.RadiusTopRight, o.RadiusBottomRight,
		o.RadiusBottomLeft, o.RadiusTopLeftY, o.RadiusTopRightY,
		o.RadiusBottomRightY, o.RadiusBottomLeftY,
	}
}

// samePayload compares the rare payload: link, image, blend, outline,
// opacity, shaping, and transform.
func samePayload(a, b *layout.DisplayOp) bool {
	if a.Outline() != b.Outline() || a.Opacity() != b.Opacity() ||
		a.LinkURI() != b.LinkURI() || a.ImageAlt() != b.ImageAlt() ||
		a.BlendModeName() != b.BlendModeName() ||
		a.Transform() != b.Transform() ||
		a.FontFeatures() != b.FontFeatures() ||
		a.TextLanguage() != b.TextLanguage() ||
		a.TextAutospace() != b.TextAutospace() ||
		a.TextTransformValue() != b.TextTransformValue() ||
		a.NoFakeBoldValue() != b.NoFakeBoldValue() ||
		a.GroupBoundary() != b.GroupBoundary() {
		return false
	}

	ab, aw, ah := a.ImageBytes()
	bb, bw, bh := b.ImageBytes()

	return aw == bw && ah == bh && bytes.Equal(ab, bb)
}
