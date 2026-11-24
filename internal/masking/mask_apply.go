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
	alpha, err := maskLayerPixels(bounds, desc.Layers[len(desc.Layers)-1], desc)
	if err != nil {
		return nil, err
	}
	for i := len(desc.Layers) - topLayerOffset; i >= 0; i-- {
		layer, err := maskLayerPixels(bounds, desc.Layers[i], desc)
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

const (
	maskBytesPerPixel = 4
	maskAlphaOffset   = 3
	maskRoundHalf     = 127
)

func Apply(group *image.NRGBA, bounds image.Rectangle, desc *layout.MaskDescriptor) error {
	mask, err := Image(bounds, desc)
	if err != nil {
		return err
	}
	gRect := group.Rect
	for gy := gRect.Min.Y; gy < gRect.Max.Y; gy++ {
		gOff := group.PixOffset(gRect.Min.X, gy)
		inY := gy >= bounds.Min.Y && gy < bounds.Max.Y
		var mRowOff int
		if inY {
			mRowOff = mask.PixOffset(0, gy-bounds.Min.Y)
		}
		for gx := gRect.Min.X; gx < gRect.Max.X; gx++ {
			a := group.Pix[gOff+maskAlphaOffset]
			if a != 0 {
				if !inY || gx < bounds.Min.X || gx >= bounds.Max.X {
					group.Pix[gOff+maskAlphaOffset] = 0
				} else {
					ma := mask.Pix[mRowOff+(gx-bounds.Min.X)*maskBytesPerPixel+maskAlphaOffset]
					scaled := (uint16(a)*uint16(ma) + maskRoundHalf) / channelMax
					group.Pix[gOff+maskAlphaOffset] = uint8(scaled & channelMax)
				}
			}
			gOff += maskBytesPerPixel
		}
	}
	return nil
}
