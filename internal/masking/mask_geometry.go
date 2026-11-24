package masking

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
)

func GroupBounds(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
	ppp := d.PixelPerPoint
	if ppp <= 0 {
		ppp = cssDPI / pointsPerInch
	}
	minX, minY := math.Floor(op.X*ppp), math.Floor(op.Y*ppp)
	maxX, maxY := math.Ceil((op.X+op.W)*ppp), math.Ceil((op.Y+op.H)*ppp)
	return image.Rect(int(minX), int(minY), int(maxX), int(maxY))
}

// maskBoxRect resolves a CSS mask-origin or mask-clip reference box
// (border-box, padding-box, or content-box) from the element's border-box
// bounds and the descriptor's point-space border/padding insets.
func maskBoxRect(bounds image.Rectangle, boxName string, desc *layout.MaskDescriptor) image.Rectangle {
	if desc == nil {
		return bounds
	}
	ppp := cssDPI / pointsPerInch
	var top, right, bottom, left float64
	switch strings.ToLower(strings.TrimSpace(boxName)) {
	case "padding-box":
		top = desc.BorderInsets[0] * ppp
		right = desc.BorderInsets[1] * ppp
		bottom = desc.BorderInsets[2] * ppp
		left = desc.BorderInsets[3] * ppp
	case "content-box":
		top = (desc.BorderInsets[0] + desc.PaddingInsets[0]) * ppp
		right = (desc.BorderInsets[1] + desc.PaddingInsets[1]) * ppp
		bottom = (desc.BorderInsets[2] + desc.PaddingInsets[2]) * ppp
		left = (desc.BorderInsets[3] + desc.PaddingInsets[3]) * ppp
	default:
		return bounds
	}
	minX := bounds.Min.X + int(math.Round(left))
	minY := bounds.Min.Y + int(math.Round(top))
	maxX := bounds.Max.X - int(math.Round(right))
	maxY := bounds.Max.Y - int(math.Round(bottom))
	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}
	return image.Rect(minX, minY, maxX, maxY).Intersect(bounds)
}

func maskDimensions(value string, iw, ih, bw, bh int) (int, int, error) {
	wv, hv, err := maskSize(value)
	if err != nil {
		return 0, 0, err
	}

	var w, h float64
	switch wv {
	case "cover":
		scale := math.Max(float64(bw)/float64(iw), float64(bh)/float64(ih))
		w, h = float64(iw)*scale, float64(ih)*scale
	case "contain":
		scale := math.Min(float64(bw)/float64(iw), float64(bh)/float64(ih))
		w, h = float64(iw)*scale, float64(ih)*scale
	default:
		var wp, hp bool
		w, wp, err = cssLength(wv)
		if err != nil && wv != "auto" {
			return 0, 0, err
		}
		h, hp, err = cssLength(hv)
		if err != nil && hv != "auto" {
			return 0, 0, err
		}
		switch {
		case wv == "auto" && hv == "auto":
			w, h = float64(iw), float64(ih)
		case wv == "auto":
			h = scaledLength(h, hp, bh)
			w = h * float64(iw) / float64(ih)
		case hv == "auto":
			w = scaledLength(w, wp, bw)
			h = w * float64(ih) / float64(iw)
		default:
			w, h = scaledLength(w, wp, bw), scaledLength(h, hp, bh)
		}
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
	xAxis, yAxis, err := maskPosition(value)
	if err != nil {
		return 0, 0, err
	}
	return axisOffset(xAxis, bw, iw), axisOffset(yAxis, bh, ih), nil
}

func axisOffset(axis maskAxisPos, box, imageSize int) float64 {
	if axis.hasOff {
		off := positionOffset(axis.value, axis.percent, box, imageSize)
		if axis.edge == "right" || axis.edge == "bottom" {
			return float64(box-imageSize) - off
		}
		return off
	}
	return positionOffset(axis.value, axis.percent, box, imageSize)
}

func positionOffset(value float64, percent bool, box, imageSize int) float64 {
	if percent {
		return value * float64(box-imageSize)
	}
	return value
}
