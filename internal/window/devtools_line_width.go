package window

import (
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// width is the measured width of the line, swatches included. A span whose
// builder measured it reports w; the rest measure here.
func (l devLine) width() float64 {
	w := 0.0

	for _, span := range l.spans {
		if span.swatch > 0 {
			w += span.swatch + devSwatchGap

			continue
		}

		if span.w > 0 {
			w += span.w

			continue
		}

		sw, _ := text.Measure(span.text, badgeFace, 0)
		w += sw
	}

	return w
}
