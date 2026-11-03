package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

// ButtonGroupOptions labels the group and selects its direction.
type ButtonGroupOptions struct {
	Label    string
	Vertical bool
}

// ButtonGroup joins its button children into one segmented control. Adjacent
// borders overlap and only the outer corners keep their radius, so pass the
// buttons in order. The group is structural: the host still owns each action.
func ButtonGroup(options ButtonGroupOptions, children ...*ui.Element) *ui.Element {
	groupClass := "of-button-group"
	if options.Vertical {
		groupClass += " of-button-group--vertical"
	}
	group := ui.Div().Class(groupClass).Attr("role", "group")
	if options.Label != "" {
		group.Attr("aria-label", options.Label)
	}
	return group.Children(children...)
}
