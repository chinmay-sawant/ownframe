package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// ProgressOptions configures a progress bar beyond its label and value.
type ProgressOptions struct {
	// ID is the stable host ID. The value fill derives its own ID from it,
	// so a value rebuild patches instead of reparsing.
	ID string
}

// Progress displays a percentage, clamped to 0..100, with progressbar
// semantics. The label is exposed to assistive technology, not painted.
func Progress(label string, value int) *ui.Element {
	return ProgressWithOptions(label, value, ProgressOptions{})
}

// ProgressWithOptions builds a Progress with a stable host ID.
func ProgressWithOptions(label string, value int, options ProgressOptions) *ui.Element {
	const maximum = 100
	value = min(max(value, 0), maximum)
	fill := ui.Div().Class("of-progress-fill").Width(ui.Percent(float64(value)))
	if fillID := partID(options.ID, "-fill"); fillID != "" {
		fill.ID(fillID)
	}
	bar := ui.Div().Class("of-progress").
		Attr("role", "progressbar").
		Attr("aria-label", label).
		Attr("aria-valuemin", "0").
		Attr("aria-valuemax", strconv.Itoa(maximum)).
		Attr("aria-valuenow", strconv.Itoa(value)).
		Children(fill)
	if options.ID != "" {
		bar.ID(options.ID)
	}
	return bar
}
