package host

import "github.com/chinmay-sawant/blinkless/layout"

// MaxSizeProvider optionally reports the largest window a screen wants.
// Window-only: the desktop window clamps to it; the browser sink has no
// window to clamp and ignores it.
type MaxSizeProvider interface {
	MaxSize() (int, int)
}

// FallbackDisplayProvider optionally exposes the retained operations
// behind a bitmap page. Window-only: the desktop window repaints viewport
// tiles from them; the browser sink serves PNG bytes and ignores them.
type FallbackDisplayProvider interface {
	FallbackDisplay() *layout.Display
}
