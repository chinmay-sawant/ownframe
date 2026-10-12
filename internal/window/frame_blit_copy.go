package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// blitRect copies one buffer rect to dst at rect minus the scroll, replacing
// the pixels there. A rect with no screen intersection is skipped.
func (s *shell) blitRect(dst, buf *ebiten.Image, rect image.Rectangle) {
	src := rect.Intersect(buf.Bounds())
	if src.Empty() {
		return
	}

	offset := image.Pt(s.scrollX, s.scrollY)
	at := src.Sub(offset).Intersect(dst.Bounds())
	if at.Empty() {
		return
	}

	sub, ok := buf.SubImage(at.Add(offset)).(*ebiten.Image)
	if !ok {
		return
	}

	var op ebiten.DrawImageOptions
	op.Blend = ebiten.BlendCopy
	op.GeoM.Translate(float64(at.Min.X), float64(at.Min.Y))
	dst.DrawImage(sub, &op)
}
