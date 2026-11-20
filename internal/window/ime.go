package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"image"
	"math"
	"time"
)

// imeScreenRect maps the already scroll-adjusted control box to window pixels.
// Density conversion belongs to EbitenView, so it is not applied here.
func (s *shell) imeScreenRect(box layout.Box) image.Rectangle {
	sx, sy, ox, oy := s.zoom(), s.zoom(), 0.0, 0.0
	switch {
	case s.viewLocked():
		sx, ox, oy = s.fit()
		sy = sx
	case s.stretched():
		w, h := s.frameSize()
		viewW, viewH := s.viewSize()
		if w > 0 && h > 0 {
			sx *= float64(viewW) / float64(w)
			sy *= float64(viewH) / float64(h)
		}
	default:
		// Page.IMEContext has already subtracted the CSS scroll offset.
		ox = float64(s.scrollX) * (sx - 1)
		oy = float64(s.scrollY) * (sy - 1)
	}
	oy += float64(s.titlebarInset())
	return image.Rect(int(math.Floor(box.X*sx+ox)), int(math.Floor(box.Y*sy+oy)),
		int(math.Ceil((box.X+box.W)*sx+ox)), int(math.Ceil((box.Y+box.H)*sy+oy)))
}

const imeRotationWait = 250 * time.Millisecond

type imeRotationState struct {
	landscape bool
	known     bool
	rotatedAt time.Time
	pendingAt time.Time
	rotated   bool
}

func (r *imeRotationState) layout(w, h int, now time.Time) {
	landscape := w > h
	if r.known && r.landscape != landscape {
		r.rotatedAt = now
		r.rotated = !r.pendingAt.IsZero()
	}
	r.landscape, r.known = landscape, true
}

func (r *imeRotationState) endByUser(now time.Time) bool {
	if !r.rotatedAt.IsZero() && now.Sub(r.rotatedAt) < imeRotationWait {
		return true
	}
	r.pendingAt, r.rotated = now, false
	return false
}

func (r *imeRotationState) resolve(now time.Time) (wait, preserve bool) {
	if r.rotated {
		r.pendingAt, r.rotated = time.Time{}, false
		return false, true
	}
	if now.Sub(r.pendingAt) < imeRotationWait {
		return true, false
	}
	r.pendingAt = time.Time{}
	return false, false
}

// Late callbacks from a cancelled session must not edit a newly focused field.
// Preedit already shown in the old field remains part of its draft.
func (s *shell) imeTargetCurrent(field string) bool {
	focus, ok := s.app.(host.Focuser)
	return ok && field != "" && focus.FocusID() == field
}
