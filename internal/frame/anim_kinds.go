package frame

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// pairStride packs one op's x and y in the from/to slices.
const pairStride = 2

// TranslateTo animates ops by (dx, dy) points from where they are now.
// Nil entries are skipped defensively.
func TranslateTo(d *page.Display, ops []*page.DisplayOp, dx, dy float64,
	dur time.Duration, ease Easing) *Anim {
	a := newAnim(d, ops, animTranslate, dur, ease)
	for i, op := range ops {
		if op == nil {
			continue
		}
		j := i * pairStride
		a.from[j], a.from[j+1] = op.X, op.Y
		a.to[j], a.to[j+1] = op.X+dx, op.Y+dy
	}

	return a
}

// FadeTo animates ops to alpha to, from current alpha. A zero target
// stops at minAlpha, since the engine reads zero as opaque. Nil entries
// are skipped defensively.
func FadeTo(d *page.Display, ops []*page.DisplayOp, to float64,
	dur time.Duration, ease Easing) *Anim {
	a := newAnim(d, ops, animFade, dur, ease)
	for i, op := range ops {
		if op == nil {
			continue
		}
		a.from[i], a.to[i] = effectiveAlpha(op), clampAlpha(to)
	}

	return a
}

// ScaleTo scales ops to factor times their size about each op's painted-box
// center. A baked transform on the op stays: the scale composes outside it.
func ScaleTo(d *page.Display, ops []*page.DisplayOp, factor float64,
	dur time.Duration, ease Easing) *Anim {
	return xformAnim(d, ops, animScale, factor, dur, ease)
}

// RotateTo rotates ops to degrees about each op's painted-box center. A
// baked transform on the op stays: the rotation composes outside it.
func RotateTo(d *page.Display, ops []*page.DisplayOp, degrees float64,
	dur time.Duration, ease Easing) *Anim {
	return xformAnim(d, ops, animRotate, degrees, dur, ease)
}
