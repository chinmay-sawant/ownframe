package chadcn

import (
	"math"

	"github.com/chinmay-sawant/ownframe/ui"
)

// aspectRatioDefault is used when the options leave the ratio unset.
const aspectRatioDefault = 16.0 / 9.0

// aspectRatioPercent converts a width-to-height ratio into a percentage
// padding height.
const aspectRatioPercent = 100.0

// AspectRatioOptions describes the content box ratio.
type AspectRatioOptions struct {
	// Ratio is width divided by height. Zero and invalid values fall back to
	// 16:9.
	Ratio float64
}

// AspectRatio keeps its children inside a box with a fixed width-to-height
// ratio. The wrapper works without the CSS aspect-ratio property by using a
// percentage padding box.
func AspectRatio(options AspectRatioOptions, children ...*ui.Element) *ui.Element {
	ratio := options.Ratio
	if ratio <= 0 || math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		ratio = aspectRatioDefault
	}
	return ui.Div().Class("of-aspect-ratio").
		PaddingTop(ui.Percent(aspectRatioPercent / ratio)).
		Children(ui.Div().Class("of-aspect-ratio-content").Children(children...))
}
