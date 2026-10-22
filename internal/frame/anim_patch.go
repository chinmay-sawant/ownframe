package frame

import (
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// SetAlpha sets ops to alpha, where zero means invisible here. It gives
// a FadeTo the hidden start the engine's zero-means-opaque alpha
// cannot. Nil entries are skipped defensively.
func SetAlpha(ops []*page.DisplayOp, alpha float64) {
	for _, op := range ops {
		if op == nil {
			continue
		}
		op.Alpha = clampAlpha(alpha)
	}
}

// patch applies the value at progress p to one op. A nil op is skipped.
func (a *Anim) patch(i int, op *page.DisplayOp, p float64) {
	if op == nil {
		return
	}
	switch a.kind {
	case animTranslate:
		j := i * pairStride
		op.X = lerp(a.from[j], a.to[j], p)
		op.Y = lerp(a.from[j+1], a.to[j+1], p)
	case animFade:
		op.Alpha = clampAlpha(lerp(a.from[i], a.to[i], p))
	case animScale:
		a.setScale(i, op, lerp(a.from[i], a.to[i], p))
	case animRotate:
		a.setRotate(i, op, lerp(a.from[i], a.to[i], p))
	}
}
