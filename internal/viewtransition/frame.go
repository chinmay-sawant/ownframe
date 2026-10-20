package viewtransition

import (
	"github.com/chinmay-sawant/ownframe"
)

// Frame holds the new-state base and captured layers at one progress value.
type Frame struct {
	Base   *ownframe.Display
	Groups []Layer
}

// Layer holds the old and new paint for one generated view-transition group.
type Layer struct {
	Name, Parent, Mode     string
	Classes                []string
	Bounds                 Rect
	Old, New               []ownframe.DisplayOp
	OldOpacity, NewOpacity float64
	newOwners              []int
}

// At builds paint layers for one progress value in [0, 1].
func (t *Transition) At(progress float64) Frame {
	if t == nil || t.next == nil {
		return Frame{}
	}
	p := clamp(progress)
	f := Frame{Base: subset(t.next, t.base), Groups: make([]Layer, 0, len(t.groups))}
	for _, g := range t.groups {
		bounds := lerpRect(g.info.From, g.info.To, p)
		layer := Layer{Name: g.info.Name, Parent: g.info.Parent, Mode: g.info.Mode, Classes: append([]string(nil), g.info.Classes...), Bounds: bounds}
		layer.OldOpacity = styledAlpha(1-p, g.info.Classes, t.styles)
		layer.NewOpacity = styledAlpha(p, g.info.Classes, t.styles)
		layer.Old = paintOps(t.old, g.oldOps, g.info.From, bounds, 1-p, g.info.Classes, t.styles)
		layer.New = paintOps(t.next, g.newOps, g.info.To, bounds, p, g.info.Classes, t.styles)
		layer.newOwners = make([]int, len(g.newOps))
		for i, op := range g.newOps {
			layer.newOwners[i] = ownerAt(t.next, op)
		}
		f.Groups = append(f.Groups, layer)
	}
	return f
}
