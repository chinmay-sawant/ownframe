package frame

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// TranslateTo animates ops by (dx, dy) points from where they are now.
func TranslateTo(d *ownframe.Display, ops []*ownframe.DisplayOp, dx, dy float64,
	dur time.Duration, ease Easing) *Anim {
	a := newAnim(d, ops, animTranslate, dur, ease)
	for i, op := range ops {
		j := i * 2
		a.from[j], a.from[j+1] = op.X, op.Y
		a.to[j], a.to[j+1] = op.X+dx, op.Y+dy
	}

	return a
}

// FadeTo animates ops to alpha to, from their current alpha. A target of
// zero stops at minAlpha, which paints nothing, because the engine reads a
// zero alpha as fully opaque. SetAlpha gives a hidden start.
func FadeTo(d *ownframe.Display, ops []*ownframe.DisplayOp, to float64,
	dur time.Duration, ease Easing) *Anim {
	a := newAnim(d, ops, animFade, dur, ease)
	for i, op := range ops {
		a.from[i], a.to[i] = effectiveAlpha(op), clampAlpha(to)
	}

	return a
}

// ScaleTo scales ops to factor times their size about each op's painted-box
// center. A baked transform on the op stays: the scale composes outside it.
func ScaleTo(d *ownframe.Display, ops []*ownframe.DisplayOp, factor float64,
	dur time.Duration, ease Easing) *Anim {
	return xformAnim(d, ops, animScale, factor, dur, ease)
}

// RotateTo rotates ops to degrees about each op's painted-box center. A
// baked transform on the op stays: the rotation composes outside it.
func RotateTo(d *ownframe.Display, ops []*ownframe.DisplayOp, degrees float64,
	dur time.Duration, ease Easing) *Anim {
	return xformAnim(d, ops, animRotate, degrees, dur, ease)
}

// SetAlpha sets ops to alpha, where zero means invisible here. It gives a
// FadeTo the hidden start the engine's zero-means-opaque alpha cannot.
func SetAlpha(ops []*ownframe.DisplayOp, alpha float64) {
	for _, op := range ops {
		op.Alpha = clampAlpha(alpha)
	}
}
