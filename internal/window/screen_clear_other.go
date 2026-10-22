//go:build js || android || ios

package window

// setupScreenClear keeps Ebiten's per-frame clear where screen retention is
// not guaranteed: browser canvases and mobile views may lose the surface
// between frames. frameStale requires a complete tile blit on each cleared
// frame, while damage still limits which tiles need repainting.
func setupScreenClear() {}
