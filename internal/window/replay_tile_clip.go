package window

import (
	"image"
	"math"
)

// dockClipX returns the first page x the devtools dock hides this frame:
// page pixels at or right of it sit under the panel. The panel is placed
// from the live dock width, the same geometry devLayout paints later this
// frame. With the dock off the content width comes back, so nothing clips.
func (s *shell) dockClipX(contentW int) int {
	if !s.dev.on || s.dev.mobile {
		return contentW
	}

	panel := devDockRect(s.screenW, s.screenH, s.devDockWidth())
	x := int(math.Ceil(panel.X)) + s.scrollX
	if x < 0 {
		x = 0
	}

	if x > contentW {
		x = contentW
	}

	return x
}

// clipDockX trims a page rect to the page x the dock leaves visible. The
// dock is full height, so only x clips.
func clipDockX(r image.Rectangle, clipX int) image.Rectangle {
	if r.Max.X > clipX {
		r.Max.X = clipX
	}

	if r.Min.X >= r.Max.X {
		return image.Rectangle{}
	}

	return r
}

// dockStrip is the page strip the previous frame's dock hid and this frame
// shows. Tiles there were not repainted while hidden, so the frame must
// catch them up. Empty unless the visible clip grew.
func (st partialState) dockStrip(clipX, contentH int) image.Rectangle {
	if clipX <= st.clipX || contentH <= 0 {
		return image.Rectangle{}
	}

	return image.Rect(st.clipX, 0, clipX, contentH)
}
