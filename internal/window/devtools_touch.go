package window

import "github.com/hajimehoshi/ebiten/v2"

// devTouch routes finger input while the devtools dock is open. Lifts
// without a drag fire the panel action; drags scroll the dock; every other
// finger joins the touch pipeline for the content beside the dock.
func (s *shell) devTouch(frameW, frameH int) error {
	var all []touchPos
	for _, id := range ebiten.TouchIDs() {
		x, y := ebiten.TouchPosition(id)
		all = append(all, touchPos{id: id, x: x, y: y})
	}
	content, taps, dy := devPanelStep(all, s.dev.panel, &s.dev.panelTrack)
	if dy != 0 {
		s.dev.scroll += float64(dy)
		s.devClampScroll()
	}
	for _, p := range taps {
		s.devPanelPress(p.x, p.y)
	}
	return s.applyTouches(content, false, frameW, frameH)
}

// devPanelStep sorts panel fingers from content fingers. It returns the
// content fingers, taps that lifted still, and upward scroll pixels.
// A zero-size panel owns nothing.
func devPanelStep(all []touchPos, panel devRect, tr *panelTrack) (content, taps []touchPos, dy int) {
	live := panel.W > 0 && panel.H > 0
	next := make([]panelFinger, 0, len(tr.fingers)+1)
	for _, p := range all {
		if !live || !devInRect(panel, float64(p.x), float64(p.y)) {
			content = append(content, p)
			continue
		}
		i := panelFind(tr.fingers, p.id)
		if i < 0 {
			next = append(next, panelFinger{id: p.id, x: p.x, y: p.y, lastY: p.y})
			continue
		}
		f := tr.fingers[i]
		if absInt(p.y-f.y) > touchSlop {
			f.drag = true
		}
		if f.drag {
			dy += f.lastY - p.y
		}
		f.lastY = p.y
		next = append(next, f)
	}
	var ids []ebiten.TouchID
	for _, p := range all {
		ids = append(ids, p.id)
	}
	for _, f := range tr.fingers {
		if !hasTouch(ids, f.id) && !f.drag {
			taps = append(taps, touchPos{id: f.id, x: f.x, y: f.y})
		}
	}
	tr.fingers = next
	return content, taps, dy
}
