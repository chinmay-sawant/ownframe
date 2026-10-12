//go:build js || android || ios

package window

// setupScreenClear keeps Ebiten's per-frame clear where screen retention is
// not guaranteed: browser canvases and mobile views may lose the surface
// between frames, so each frame starts from a cleared screen and the full
// branches repaint everything.
func setupScreenClear() {}
