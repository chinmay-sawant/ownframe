package replay

import "github.com/chinmay-sawant/blinkless/layout"

// textProjection preserves the engine's already projected baseline while
// applying its remaining linear projection to the glyphs. Unlike a retained
// 2D patch, this metadata has XformSet false and must not move the baseline twice.
func textProjection(op *layout.DisplayOp) (layout.Matrix2D, bool) {
	if op.XformSet || (op.Kind != layout.DisplayOpText && op.Kind != layout.DisplayOpBullet) {
		return layout.Matrix2D{}, false
	}
	m := op.Transform()
	if m == (layout.Matrix2D{}) || m.IsIdentity() {
		return layout.Matrix2D{}, false
	}
	m.E = op.X - m.A*op.X - m.C*op.Y
	m.F = op.Y - m.B*op.X - m.D*op.Y
	return m, true
}
