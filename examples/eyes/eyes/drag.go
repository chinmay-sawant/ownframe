package eyes

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onPress looks toward the pressed point relative to center.
func (a *App) onPress(_ context.Context, b ownframe.Box, x, y float64) (bool, error) {
	w, h := a.page.Size()
	if w > 0 && h > 0 {
		a.st.tLookX = clamp((x-float64(w)/2)*0.12, -48, 48)
		a.st.tLookY = clamp((y-float64(h)/2)*0.12, -30, 30)
	}
	return false, nil
}

// onDragStart claims the drag so moves steer the look.
func (a *App) onDragStart(_ context.Context, _ ownframe.Box) (bool, error) {
	a.st.drag = true
	return true, nil
}

// onDragMove steers the look target while dragging.
func (a *App) onDragMove(_ context.Context, dx, dy float64) error {
	a.st.tLookX = clamp(a.st.tLookX+dx*0.4, -48, 48)
	a.st.tLookY = clamp(a.st.tLookY+dy*0.4, -30, 30)
	return nil
}

// onDragEnd releases the look back to auto saccades.
func (a *App) onDragEnd(_ context.Context) error {
	a.st.drag = false
	a.st.moodT = a.now()
	a.st.sacT = a.now()
	return nil
}
