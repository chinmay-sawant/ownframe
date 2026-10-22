package frame

type genSrc interface {
	Generation() uint64
}

// Bind records the page's display generation, so Step goes inactive once
// a Redraw replaces the list instead of patching freed operations. It
// returns the animation for chaining with its constructor. An unbound
// animation keeps the old behavior: Step patches until it ends.
func (a *Anim) Bind(p genSrc) *Anim {
	a.src = p
	if p != nil {
		a.gen = p.Generation()
	}
	return a
}

// Generation returns the display generation recorded by Bind, or zero
// when the animation is unbound.
func (a *Anim) Generation() uint64 { return a.gen }

// stale reports whether a Redraw has replaced the bound display list.
func (a *Anim) stale() bool {
	return a.src != nil && a.src.Generation() != a.gen
}
