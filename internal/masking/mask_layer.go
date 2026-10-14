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

func maskLayerPixels(bounds image.Rectangle, layer layout.MaskLayer) ([]uint8, error) {
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
	source, _, err := image.Decode(bytes.NewReader(layer.ImageBytes))
	if err != nil {
		return nil, fmt.Errorf("mask: decode image: %w", err)
	}
	b := source.Bounds()
	if b.Empty() {
		return nil, fmt.Errorf("mask: empty image")
	}
	w, h, err := maskDimensions(layer.Size, b.Dx(), b.Dy(), bounds.Dx(), bounds.Dy())
	if err != nil {
		return nil, err
	}
	px, py, err := maskOffsets(layer.Position, bounds.Dx(), bounds.Dy(), w, h)
	if err != nil {
		return nil, err
	}
	scaled := image.NewNRGBA(image.Rect(0, 0, w, h))
	xdraw.BiLinear.Scale(scaled, scaled.Bounds(), source, b, draw.Src, nil)
	return paintMaskPixels(alpha, bounds.Dx(), bounds.Dy(), scaled, px, py, layer), nil
}
