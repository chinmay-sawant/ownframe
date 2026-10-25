package replay

import "github.com/chinmay-sawant/blinkless/layout"

// textProjection preserves the engine's already projected baseline while
// applying its remaining linear projection to the glyphs. Unlike a retained
// 2D patch, this metadata has XformSet false and must not move the baseline
// twice.
//
// The read goes through XformValue, a pointer-receiver accessor that returns the
// raw matrix without copying the op or binding the shared empty extra.
// Transform has a value receiver, so it copies the whole 256-byte op on every
// text run, which dominated the draw-rect scan. Reading the promoted Xform
// field directly is not an option: the payload is an embedded pointer, so a
// hand-built op leaves it nil, and binding the shared extra to guard the read
// would make Transform report a zero matrix where it documents an identity.
// internal/page/dirty compares Transform to diff a painted display against the
// next one, so that bind would dirty every text run on each redraw.
func textProjection(op *layout.DisplayOp) (layout.Matrix2D, bool) {
	if op.XformSet || (op.Kind != layout.DisplayOpText && op.Kind != layout.DisplayOpBullet) {
		return layout.Matrix2D{}, false
	}
	m := op.XformValue()
	if m == (layout.Matrix2D{}) || m.IsIdentity() {
		return layout.Matrix2D{}, false
	}
	m.E = op.X - m.A*op.X - m.C*op.Y
	m.F = op.Y - m.B*op.X - m.D*op.Y
	return m, true
}
