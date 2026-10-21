package window

import "image/color"

// Shared devtools UI numbers used by two or more files. File-local
// numbers stay in their own file, except devScrollMin and
// devScrollMargin, kept here to keep the scrollbar file under the cap.
const (
	// devTabPad is the horizontal pad around a tab name.
	devTabPad = 20.0
	// devClosePad is the hit pad around the header close key.
	devClosePad = 12.0
	// devHeaderGap is the gap between header buttons.
	devHeaderGap = 8.0
	// devLabelInset is the label text inset and outline gap.
	devLabelInset = 4.0
	// devCenterDiv centers by halving a height difference.
	devCenterDiv = 2.0
	// devScrollMin floors the scrollbar thumb.
	devScrollMin = 24.0
	// devScrollMargin insets the scrollbar thumb from the edge.
	devScrollMargin = 2.0
)

const (
	devShiftR = 24
	devShiftG = 16
	devShiftB = 8
)

// devRGBA builds an opaque-ish ink from packed 0xRRGGBBAA.
func devRGBA(hex uint32) color.RGBA {
	return color.RGBA{
		R: uint8(hex >> devShiftR), //nolint:gosec // reason: shift keeps channel in byte range.
		G: uint8(hex >> devShiftG), //nolint:gosec // reason: shift keeps channel in byte range.
		B: uint8(hex >> devShiftB), //nolint:gosec // reason: shift keeps channel in byte range.
		A: uint8(hex),              //nolint:gosec // reason: low byte holds alpha.
	}
}
