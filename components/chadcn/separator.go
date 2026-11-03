package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// SeparatorOrientation selects the separator axis. The zero value is
// horizontal.
type SeparatorOrientation uint8

const (
	// SeparatorHorizontal draws a full-width rule.
	SeparatorHorizontal SeparatorOrientation = iota
	// SeparatorVertical draws a full-height rule between inline items.
	SeparatorVertical
)

// SeparatorOptions describes the separator axis.
type SeparatorOptions struct {
	Orientation SeparatorOrientation
}

// Separator draws a hairline divider with separator semantics.
func Separator(options SeparatorOptions) *ui.Element {
	orientation := "of-separator--horizontal"
	axis := "horizontal"
	if options.Orientation == SeparatorVertical {
		orientation = "of-separator--vertical"
		axis = "vertical"
	}
	return ui.Div().Class(classes("of-separator", orientation)).
		Attr("role", "separator").Attr("aria-orientation", axis)
}
