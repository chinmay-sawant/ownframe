package frame

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// animKind selects what an Anim patches.
type animKind uint8

const (
	animTranslate animKind = iota
	animFade
	animScale
	animRotate
)

// minAlpha stands in for invisible: the engine reads an op alpha of zero as
// "no override", so a fade out stops just above it.
const minAlpha = 1e-4

// newAnim allocates the per-op value tracks an animation steps through.
func newAnim(d *page.Display, ops []*page.DisplayOp, k animKind,
	dur time.Duration, ease Easing) *Anim {
	n := len(ops)
	if k == animTranslate {
		n *= 2
	}

	a := &Anim{
		d: d, ops: ops, kind: k, dur: dur,
		ease: easeOr(ease), live: true,
		from: make([]float64, n), to: make([]float64, n),
	}
	if k == animScale || k == animRotate {
		a.base = make([]mat6, len(ops))
		for i, op := range ops {
			a.base[i] = baseOf(op)
		}
	}

	return a
}

// xformAnim builds a scale or rotate animation from factor 1.
func xformAnim(d *page.Display, ops []*page.DisplayOp, k animKind,
	to float64, dur time.Duration, ease Easing) *Anim {
	a := newAnim(d, ops, k, dur, ease)
	for i := range ops {
		a.from[i], a.to[i] = 1, to
	}

	return a
}

// effectiveAlpha is op's alpha, reading the engine's zero as opaque.
// A nil op reads as opaque.
func effectiveAlpha(op *page.DisplayOp) float64 {
	if op == nil {
		return 1
	}
	if op.Alpha > 0 && op.Alpha < 1 {
		return op.Alpha
	}

	return 1
}

// clampAlpha keeps an alpha inside the range the engine honors: zero means
// no override, so invisible stops at minAlpha.
func clampAlpha(alpha float64) float64 {
	if alpha < minAlpha {
		return minAlpha
	}

	if alpha > 1 {
		return 1
	}

	return alpha
}
