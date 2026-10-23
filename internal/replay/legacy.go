package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// Package functions retain the shared default for callers without an owner.
// Window hosts use their own Renderer instead.
func DrawRect(dst *ebiten.Image, display *layout.Display, rect image.Rectangle, dx, dy float64) {
	defaultRenderer.DrawRect(dst, display, rect, dx, dy)
}

func DrawUnfixedZ(dst *ebiten.Image, d *layout.Display, z int) {
	defaultRenderer.DrawUnfixedZ(dst, d, z)
}

func DrawRectUnfixedZ(dst *ebiten.Image, d *layout.Display, rect image.Rectangle, z int) {
	defaultRenderer.DrawRectUnfixedZ(dst, d, rect, z)
}

func DrawVisibleUnfixedZ(dst *ebiten.Image, d *layout.Display, dx, dy float64, z int) {
	defaultRenderer.DrawVisibleUnfixedZ(dst, d, dx, dy, z)
}

func DrawPinnedZ(dst *ebiten.Image, d *layout.Display, z int, dx, dy float64) {
	defaultRenderer.DrawPinnedZ(dst, d, z, dx, dy)
}

func Draw(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	defaultRenderer.Draw(dst, display, dx, dy)
}

func DrawUnfixed(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	defaultRenderer.DrawUnfixed(dst, display, dx, dy)
}

func DrawFixed(dst *ebiten.Image, display *layout.Display) {
	defaultRenderer.DrawFixed(dst, display)
}

func DrawVisible(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	defaultRenderer.DrawVisible(dst, display, dx, dy)
}

func DrawVisiblePinned(dst *ebiten.Image, display *layout.Display, dx, dy float64, pinZ int, pinDx, pinDy float64) {
	defaultRenderer.DrawVisiblePinned(dst, display, dx, dy, pinZ, pinDx, pinDy)
}
