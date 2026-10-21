package frame

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// Step applies the value at now and reports the boxes it changed. The
// first step fixes the start time; the step that reaches the end applies
// the final value and reports false. A later step, or one after a Redraw
// replaced a bound list, changes nothing.
func (a *Anim) Step(now time.Time) (Damage, bool) {
	if a.stale() {
		a.live = false
		return Damage{}, false
	}
	if !a.live {
		return Damage{}, false
	}

	if a.start.IsZero() {
		a.start = now
	}

	raw := progress(now, a.start, a.dur)
	p := a.ease(raw)

	var dmg Damage

	for i, op := range a.ops {
		old := OpBox(a.d, op)
		a.patch(i, op, p)
		dmg.Old = dmg.Old.Union(old)
		dmg.New = dmg.New.Union(OpBox(a.d, op))
	}

	if raw >= 1 {
		a.live = false
	}

	return dmg, a.live
}

// Tick steps the animation and marks the changed box on m. It reports
// whether the animation is still running, so a tick can stop asking for
// frames once every animation is done.
func (a *Anim) Tick(m Marker, now time.Time) bool {
	dmg, live := a.Step(now)
	if m != nil && !dmg.Empty() {
		m.MarkRect(dmg.Box())
	}

	return live
}

// patch applies the value at progress p to one op.
func (a *Anim) patch(i int, op *page.DisplayOp, p float64) {
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

// progress clamps now against the animation's window to 0..1.
func progress(now, start time.Time, dur time.Duration) float64 {
	if dur <= 0 {
		return 1
	}

	p := float64(now.Sub(start)) / float64(dur)
	if p < 0 {
		return 0
	}

	if p > 1 {
		return 1
	}

	return p
}

// lerp is from + (to-from)*p.
func lerp(from, to, p float64) float64 { return from + (to-from)*p }
