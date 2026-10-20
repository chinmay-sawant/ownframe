package viewtransition

import "github.com/chinmay-sawant/ownframe"

// New pairs names found in two display snapshots inside one scope.
func New(old, next *ownframe.Display, scope Scope, styles map[string]ClassStyle) (*Transition, error) {
	if old == nil || next == nil {
		return nil, ErrDisplay
	}
	if !validRoot(old, scope.OldRoot) || !validRoot(next, scope.NewRoot) {
		return nil, ErrScope
	}
	a, ao, err := discover(old, scope.OldRoot)
	if err != nil {
		return nil, err
	}
	b, bo, err := discover(next, scope.NewRoot)
	if err != nil {
		return nil, err
	}
	t := &Transition{old: old, next: next, styles: styles}
	add := func(name string) {
		x, xok := a[name]
		y, yok := b[name]
		chosen := y
		if !yok {
			chosen = x
		}
		g := group{oldBox: -1, newBox: -1}
		g.info = Group{Name: name, Mode: mode(chosen.box.ViewTransitionGroup), Classes: classNames(chosen.box.ViewTransitionClass), OldPresent: xok, NewPresent: yok}
		if xok {
			g.oldBox = x.index
			g.info.From = rect(x.box)
		}
		if yok {
			g.newBox = y.index
			g.info.To = rect(y.box)
		}
		if !xok {
			g.info.From = g.info.To
		}
		if !yok {
			g.info.To = g.info.From
		}
		t.groups = append(t.groups, g)
	}
	for _, name := range bo {
		add(name)
	}
	for _, name := range ao {
		if _, ok := b[name]; !ok {
			add(name)
		}
	}
	setParents(t.groups, a, b)
	for i := range t.groups {
		g := &t.groups[i]
		if g.oldBox >= 0 {
			g.oldOps = captureOps(old, g.oldBox, a)
		}
		if g.newBox >= 0 {
			g.newOps = captureOps(next, g.newBox, b)
		}
	}
	t.base = baseOps(next, t.groups)
	return t, nil
}
