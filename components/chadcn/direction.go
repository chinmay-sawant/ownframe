package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// DirectionMode selects text direction. The zero value is left to right.
type DirectionMode uint8

const (
	// DirectionLTR lays content out left to right.
	DirectionLTR DirectionMode = iota
	// DirectionRTL lays content out right to left.
	DirectionRTL
)

// DirectionOptions describes the writing direction of a subtree.
type DirectionOptions struct {
	Mode DirectionMode
}

// Direction wraps children with a dir attribute so a subtree can override the
// page direction.
func Direction(options DirectionOptions, children ...*ui.Element) *ui.Element {
	dir := "ltr"
	if options.Mode == DirectionRTL {
		dir = "rtl"
	}
	return ui.Div().Class("of-direction").Attr("dir", dir).Children(children...)
}
