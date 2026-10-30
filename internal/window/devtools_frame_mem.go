package window

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// bytesPerKB and bytesPerMB size the MEMORY rows.
const (
	bytesPerKB = 1024
	bytesPerMB = 1024 * bytesPerKB
)

// devFrameRow builds one right-aligned row. The label and separator widths
// come from the panel memo; only the value, which moves as counters change,
// is measured per call. The value span right-aligns at draw time, so no
// padding run is built or drawn.
func devFrameRow(w float64, l, v string, ink color.RGBA) devRow {
	vw, _ := text.Measure(v, badgeFace, 0)

	return devRow{
		line: devLine{spans: []devSpan{
			{text: l, ink: devDim, w: devTextWidth(l)},
			{text: " ", ink: devPunctInk, w: devTextWidth(" ")},
			{text: v, ink: ink, w: vw, right: true},
		}},
		w: w,
	}
}

// addMemoryRows prints the MEMORY section. A shell without Perf reports
// !ok, so end users see dashes and developers see live runtime numbers.
func (s *shell) addMemoryRows(rs *[]devRow, w float64, h func(string), n func(string, any)) {
	h("Memory")
	if s.perfHooks.devRuntimePerf == nil {
		for _, l := range []string{"Alloc/frame", "Go heap", "RSS", "Goroutines"} {
			n(l, "-")
		}
		return
	}
	alloc, heap, rss, gr, ok := s.perfHooks.devRuntimePerf()
	if !ok {
		for _, l := range []string{"Alloc/frame", "Go heap", "RSS", "Goroutines"} {
			n(l, "-")
		}
		return
	}
	n("Alloc/frame", devBytes(alloc))
	n("Go heap", devBytes(heap))
	n("RSS", devRSS(rss))
	n("Goroutines", gr)
}

// devRSS prints a resident byte count for the panel. Zero means the
// platform reader reported nothing, since genuine RSS is never zero,
// so the row reads "-" instead of "0B".
func devRSS(b uint64) string {
	if b == 0 {
		return "-"
	}
	return devBytes(b)
}

// devBytes prints a byte count for the panel.
func devBytes(b uint64) string {
	if b < bytesPerKB {
		return fmt.Sprintf("%dB", b)
	}
	if b < bytesPerMB {
		return fmt.Sprintf("%.1fKB", float64(b)/bytesPerKB)
	}
	return fmt.Sprintf("%.1fMB", float64(b)/bytesPerMB)
}
