package masking

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	xdraw "golang.org/x/image/draw"
)

func maskLayerPixels(bounds image.Rectangle, layer layout.MaskLayer, desc *layout.MaskDescriptor) ([]uint8, error) {
	if bounds.Dx() < 0 || bounds.Dy() < 0 {
		return nil, fmt.Errorf("mask: invalid mask box")
	}
	if bounds.Empty() {
		return []uint8{}, nil
	}
	if bounds.Dx() > maxMaskPixels/bounds.Dy() {
		return nil, fmt.Errorf("mask: box exceeds %d pixels", maxMaskPixels)
	}
	alpha := make([]uint8, bounds.Dx()*bounds.Dy())
	if strings.EqualFold(strings.TrimSpace(layer.Source), "none") {
		return alpha, nil
	}
	originBox := maskBoxRect(bounds, layer.Origin, desc)
	clipBox := maskBoxRect(bounds, layer.Clip, desc)
	if originBox.Empty() || clipBox.Empty() {
		return alpha, nil
	}
	source, _, err := image.Decode(bytes.NewReader(layer.ImageBytes))
	if err != nil {
		return nil, fmt.Errorf("mask: decode image: %w", err)
	}
	b := source.Bounds()
	if b.Empty() {
		return nil, fmt.Errorf("mask: empty image")
	}
	w, h, err := maskDimensions(layer.Size, b.Dx(), b.Dy(), originBox.Dx(), originBox.Dy())
	if err != nil {
		return nil, err
	}
	px, py, err := maskOffsets(layer.Position, originBox.Dx(), originBox.Dy(), w, h)
	if err != nil {
		return nil, err
	}
	ox := float64(originBox.Min.X - bounds.Min.X)
	oy := float64(originBox.Min.Y - bounds.Min.Y)
	localClip := clipBox.Sub(bounds.Min)
	scaled := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(scaled, scaled.Bounds(), source, b, draw.Src, nil)
	return paintMaskPixels(alpha, bounds.Dx(), bounds.Dy(), localClip, scaled, ox+px, oy+py, layer), nil
}
