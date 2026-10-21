package masking

import (
	"fmt"
	"image"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
)

func Image(bounds image.Rectangle, desc *layout.MaskDescriptor) (*image.NRGBA, error) {
	if desc == nil {
		return nil, fmt.Errorf("mask: nil descriptor")
	}
	if len(desc.Layers) == 0 {
		return nil, fmt.Errorf("mask: no layers")
	}
	if bounds.Dx() < 0 || bounds.Dy() < 0 || bounds.Dx() > maxMaskPixels/max(1, bounds.Dy()) {
		return nil, fmt.Errorf("mask: invalid or oversized box")
	}
	img := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	if bounds.Empty() {
		return img, nil
	}
	alpha, err := maskLayerPixels(bounds, desc.Layers[len(desc.Layers)-1])
	if err != nil {
		return nil, err
	}
	for i := len(desc.Layers) - topLayerOffset; i >= 0; i-- {
		layer, err := maskLayerPixels(bounds, desc.Layers[i])
		if err != nil {
			return nil, err
		}
		mode := strings.TrimSpace(desc.Layers[i].Composite)
		for p := range alpha {
			alpha[p] = combineMask(layer[p], alpha[p], mode)
		}
	}
	for i, value := range alpha {
		img.Pix[i*4], img.Pix[i*4+1], img.Pix[i*4+2], img.Pix[i*4+3] = channelMax, channelMax, channelMax, value
	}
	return img, nil
}

func Apply(group *image.NRGBA, bounds image.Rectangle, desc *layout.MaskDescriptor) error {
	mask, err := Image(bounds, desc)
	if err != nil {
		return err
	}
	for y := 0; y < group.Rect.Dy(); y++ {
		for x := 0; x < group.Rect.Dx(); x++ {
			gx, gy := x+group.Rect.Min.X, y+group.Rect.Min.Y
			pixel := group.NRGBAAt(gx, gy)
			if !image.Pt(gx, gy).In(bounds) {
				pixel.A = 0
			} else {
				maskAlpha := mask.NRGBAAt(gx-bounds.Min.X, gy-bounds.Min.Y).A
				pixel.A = channel(float64(pixel.A) / channelMax * float64(maskAlpha) / channelMax)
			}
			group.SetNRGBA(gx, gy, pixel)
		}
	}
	return nil
}
