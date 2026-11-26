// Package ui is Quarry's ownframe builder UI: ui.Mount, *ui.Element trees, and
// components/chadcn styling, with all CSS kept in the embedded styles/ files.
package ui

import (
	_ "embed"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
)

//go:embed styles/base.css
var baseCSS string

//go:embed styles/dark.css
var darkCSS string

//go:embed styles/light.css
var lightCSS string

// ThemeCSS composes the page stylesheet: ChadCN tokens and components, the
// Quarry token sheet for the chosen theme, then the shared Quarry geometry.
func ThemeCSS(dark bool) string {
	if dark {
		return chadcn.Dark + "\n" + chadcn.Stylesheet + "\n" + darkCSS + "\n" + baseCSS
	}
	return chadcn.Light + "\n" + chadcn.Stylesheet + "\n" + lightCSS + "\n" + baseCSS
}
