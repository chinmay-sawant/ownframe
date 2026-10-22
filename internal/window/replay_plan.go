package window

import "image"

// repaintMode is what one frame does with the persistent buffer.
type repaintMode int

const (
	// repaintBuffer repaints the whole buffer.
	repaintBuffer repaintMode = iota
	// repaintRect repaints one dirty rect in the buffer.
	repaintRect
	// repaintBlit leaves the buffer alone. Its pixels are current.
	repaintBlit
)

// repaintPlan is the buffer work for one frame.
type repaintPlan struct {
	mode repaintMode
	rect image.Rectangle
}

// maxRepaintRects caps the dirty rects of one frame. Each rect costs a scan
// of the operation list, so a longer list rebuilds the whole buffer.
const maxRepaintRects = 8
