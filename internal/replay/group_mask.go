package replay

import (
	"image"

	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
)

func groupMaskTexture(n *groupNode) *ebiten.Image {
	if n == nil || n.group == nil || n.group.Mask == nil || !n.hasMask {
		return nil
	}
	n.maskMu.Lock()
	defer n.maskMu.Unlock()
	if n.maskTex != nil {
		return n.maskTex
	}
	box := image.Rect(0, 0, n.maskBox.Dx(), n.maskBox.Dy())
	pixels, err := masking.Image(box, n.group.Mask)
	if err != nil {
		return nil
	}
	n.maskTex = ebiten.NewImageFromImage(pixels)
	return n.maskTex
}

func disposePlan(n *groupNode) {
	if n == nil {
		return
	}
	n.maskMu.Lock()
	if n.maskTex != nil {
		n.maskTex.Dispose()
		n.maskTex = nil
	}
	n.maskMu.Unlock()
	for i := range n.items {
		disposePlan(n.items[i].child)
	}
}
