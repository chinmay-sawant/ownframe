//go:build !android && !ios && !js

package window

import "github.com/hajimehoshi/ebiten/v2"

// setupScreenClear stops Ebiten from clearing the screen at the start of each
// frame. The partial replay keeps the last frame on the screen and repaints
// only the damage, so a per-frame clear would erase pixels the damage list
// does not cover. Ebiten may still discard the surface after an occlusion;
// the focus-regain guard in frameStale repaints the whole frame then.
func setupScreenClear() {
	ebiten.SetScreenClearedEveryFrame(false)
}
