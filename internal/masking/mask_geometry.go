package masking

import (
	"fmt"
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

func GroupBounds(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
	ppp := d.PixelPerPoint
	if ppp <= 0 {
		ppp = 96.0 / 72.0
	}
	minX, minY := math.Floor(op.X*ppp), math.Floor(op.Y*ppp)
	maxX, maxY := math.Ceil((op.X+op.W)*ppp), math.Ceil((op.Y+op.H)*ppp)
	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

func maskDimensions(value string, iw, ih, bw, bh int) (int, int, error) {
	wv, hv, err := maskSize(value)
	if err != nil {
		return 0, 0, err
	}
	w, wp, err := cssLength(wv)
	if err != nil && wv != "auto" {
		return 0, 0, err
	}
	h, hp, err := cssLength(hv)
	if err != nil && hv != "auto" {
		return 0, 0, err
	}
	if wv == "auto" && hv == "auto" {
		w, h = float64(iw), float64(ih)
	} else if wv == "auto" {
		h = scaledLength(h, hp, bh)
		w = h * float64(iw) / float64(ih)
	} else if hv == "auto" {
		w = scaledLength(w, wp, bw)
		h = w * float64(ih) / float64(iw)
	} else {
		w, h = scaledLength(w, wp, bw), scaledLength(h, hp, bh)
	}
	if math.IsNaN(w) || math.IsInf(w, 0) || math.IsNaN(h) || math.IsInf(h, 0) ||
		w <= 0 || h <= 0 || w > maxMaskPixels || h > maxMaskPixels || w*h > maxMaskPixels {
		return 0, 0, fmt.Errorf("size is outside the bitmap limit")
	}
	return max(1, int(math.Round(w))), max(1, int(math.Round(h))), nil
}

func scaledLength(value float64, percent bool, base int) float64 {
	if percent {
		return value * float64(base)
	}
	return value
}

func maskOffsets(value string, bw, bh, iw, ih int) (float64, float64, error) {
	xv, yv, err := maskPosition(value)
	if err != nil {
		return 0, 0, err
	}
	x, xp, _ := cssPosition(xv)
	y, yp, _ := cssPosition(yv)
	return positionOffset(x, xp, bw, iw), positionOffset(y, yp, bh, ih), nil
}

func positionOffset(value float64, percent bool, box, imageSize int) float64 {
	if percent {
		return value * float64(box-imageSize)
	}
	return value
}
