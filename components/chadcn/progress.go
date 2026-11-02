package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// Progress displays a percentage, clamped to 0..100, with progressbar
// semantics. The label is exposed to assistive technology, not painted.
func Progress(label string, value int) *ui.Element {
	const maximum = 100
	value = min(max(value, 0), maximum)
	return ui.Div().Class("of-progress").
		Attr("role", "progressbar").
		Attr("aria-label", label).
		Attr("aria-valuemin", "0").
		Attr("aria-valuemax", strconv.Itoa(maximum)).
		Attr("aria-valuenow", strconv.Itoa(value)).
		Children(ui.Div().Class("of-progress-fill").Width(ui.Percent(float64(value))))
}
