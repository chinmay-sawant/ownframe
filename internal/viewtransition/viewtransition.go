package viewtransition

import (
	"errors"

	"github.com/chinmay-sawant/ownframe"
)

// DocumentRoot selects the full display snapshot.
const DocumentRoot = -1

// ErrDisplay means a snapshot is missing or unusable.
var ErrDisplay = errors.New("viewtransition: invalid display")

// ErrScope means a root index is outside its display's boxes.
var ErrScope = errors.New("viewtransition: invalid scope")

// ErrDuplicateName means a scope contains the same capture name twice.
var ErrDuplicateName = errors.New("viewtransition: duplicate name")

// Scope selects the old and new subtree roots by box index.
type Scope struct{ OldRoot, NewRoot int }

// Rect holds a box in CSS pixels.
type Rect struct{ X, Y, W, H float64 }

// ClassStyle changes opacity or scale for captures carrying a class.
type ClassStyle struct{ Opacity, Scale *float64 }

// Transition pairs named elements from two retained displays.
type Transition struct {
	old, next *ownframe.Display
	groups    []group
	base      []int
	styles    map[string]ClassStyle
}
type group struct {
	info           Group
	oldBox, newBox int
	oldOps, newOps []int
}

// Group is a named capture and its parent in the generated group tree.
type Group struct {
	Name, Parent, Mode     string
	Classes                []string
	From, To               Rect
	OldPresent, NewPresent bool
}

func (t *Transition) Groups() []Group {
	if t == nil {
		return nil
	}
	out := make([]Group, len(t.groups))
	for i, g := range t.groups {
		out[i] = g.info
		out[i].Classes = append([]string(nil), g.info.Classes...)
	}
	return out
}
