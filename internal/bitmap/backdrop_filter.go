package bitmap

import (
	"image"
	"math"

	"github.com/chinmay-sawant/ownframe/internal/backdrop"
)

func applyBackdropFilter(dst *image.NRGBA, area image.Rectangle, radius float64, filter backdrop.Filter) {
	shape := area
	area = area.Intersect(dst.Bounds())
	if area.Empty() {
		return
	}
	blurRadius := int(math.Ceil(filter.BlurPx))
	rows := image.Rect(area.Min.X, max(dst.Bounds().Min.Y, area.Min.Y-blurRadius),
		area.Max.X, min(dst.Bounds().Max.Y, area.Max.Y+blurRadius))
	width := area.Dx()
	height := area.Dy()
	horizontal := make([][4]uint16, width*rows.Dy())
	filtered := make([][4]byte, width*height)

	for y := rows.Min.Y; y < rows.Max.Y; y++ {
		lo := max(dst.Bounds().Min.X, area.Min.X-blurRadius)
		hi := min(dst.Bounds().Max.X, area.Min.X+blurRadius+1)
		var sum [4]int
		for x := lo; x < hi; x++ {
			addPixel(&sum, dst.Pix[dst.PixOffset(x, y):])
		}
		for x := area.Min.X; x < area.Max.X; x++ {
			count := hi - lo
			index := (y-rows.Min.Y)*width + x - area.Min.X
			for channel := range 4 {
				horizontal[index][channel] = uint16((sum[channel] + count/2) / count)
			}
			nextLo := max(dst.Bounds().Min.X, x+1-blurRadius)
			nextHi := min(dst.Bounds().Max.X, x+blurRadius+2)
			if nextLo > lo {
				removePixel(&sum, dst.Pix[dst.PixOffset(lo, y):])
			}
			if nextHi > hi {
				addPixel(&sum, dst.Pix[dst.PixOffset(hi, y):])
			}
			lo, hi = nextLo, nextHi
		}
	}

	for x := area.Min.X; x < area.Max.X; x++ {
		lo := max(rows.Min.Y, area.Min.Y-blurRadius)
		hi := min(rows.Max.Y, area.Min.Y+blurRadius+1)
		var sum [4]int
		for y := lo; y < hi; y++ {
			addSample(&sum, horizontal[(y-rows.Min.Y)*width+x-area.Min.X])
		}
		for y := area.Min.Y; y < area.Max.Y; y++ {
			if insideRoundedRect(x, y, shape, radius) {
				count := hi - lo
				alpha := (sum[3] + count/2) / count
				var red, green, blue float64
				if alpha > 0 {
					red = float64(sum[0]) * 255 / float64(count*alpha)
					green = float64(sum[1]) * 255 / float64(count*alpha)
					blue = float64(sum[2]) * 255 / float64(count*alpha)
				}
				luma := 0.213*red + 0.715*green + 0.072*blue
				index := (y-area.Min.Y)*width + x - area.Min.X
				filtered[index] = [4]byte{
					clampByte(luma + (red-luma)*filter.Saturation),
					clampByte(luma + (green-luma)*filter.Saturation),
					clampByte(luma + (blue-luma)*filter.Saturation),
					byte(alpha),
				}
			}

			nextLo := max(rows.Min.Y, y+1-blurRadius)
			nextHi := min(rows.Max.Y, y+blurRadius+2)
			if nextLo > lo {
				removeSample(&sum, horizontal[(lo-rows.Min.Y)*width+x-area.Min.X])
			}
			if nextHi > hi {
				addSample(&sum, horizontal[(hi-rows.Min.Y)*width+x-area.Min.X])
			}
			lo, hi = nextLo, nextHi
		}
	}

	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			if !insideRoundedRect(x, y, shape, radius) {
				continue
			}
			dx, dy := backdropLensOffset(float64(x)+0.5, float64(y)+0.5, shape, radius, filter.RefractionPx)
			pixel := sampleBackdropPixels(filtered, width, height,
				float64(x-area.Min.X)+dx, float64(y-area.Min.Y)+dy)
			copy(dst.Pix[dst.PixOffset(x, y):], pixel[:])
		}
	}
}

func backdropLensOffset(x, y float64, shape image.Rectangle, radius, strength float64) (float64, float64) {
	if strength <= 0 {
		return 0, 0
	}
	halfWidth, halfHeight := float64(shape.Dx())*0.5, float64(shape.Dy())*0.5
	radius = min(radius, halfWidth, halfHeight)
	centerX := x - (float64(shape.Min.X) + halfWidth)
	centerY := y - (float64(shape.Min.Y) + halfHeight)
	qx := math.Abs(centerX) - (halfWidth - radius)
	qy := math.Abs(centerY) - (halfHeight - radius)
	outX, outY := max(qx, 0), max(qy, 0)
	outLength := math.Hypot(outX, outY)
	distance := outLength + min(max(qx, qy), 0) - radius
	edge := min(max(1+distance/max(strength*2, 1), 0), 1)
	var normalX, normalY float64
	if outLength > 0 {
		normalX = math.Copysign(outX/outLength, centerX)
		normalY = math.Copysign(outY/outLength, centerY)
	} else if qx > qy {
		normalX = math.Copysign(1, centerX)
	} else {
		normalY = math.Copysign(1, centerY)
	}
	amount := strength * 0.72 * edge * edge
	return -normalX * amount, -normalY * amount
}

func sampleBackdropPixels(pixels [][4]byte, width, height int, x, y float64) [4]byte {
	x = min(max(x, 0), float64(width-1))
	y = min(max(y, 0), float64(height-1))
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	x1, y1 := min(x0+1, width-1), min(y0+1, height-1)
	fx, fy := x-float64(x0), y-float64(y0)
	var result [4]byte
	for channel := range 4 {
		top := float64(pixels[y0*width+x0][channel])*(1-fx) + float64(pixels[y0*width+x1][channel])*fx
		bottom := float64(pixels[y1*width+x0][channel])*(1-fx) + float64(pixels[y1*width+x1][channel])*fx
		result[channel] = clampByte(top*(1-fy) + bottom*fy)
	}
	return result
}

func addPixel(sum *[4]int, pixel []byte) {
	alpha := int(pixel[3])
	for channel := range 3 {
		sum[channel] += (int(pixel[channel])*alpha + 127) / 255
	}
	sum[3] += alpha
}

func removePixel(sum *[4]int, pixel []byte) {
	alpha := int(pixel[3])
	for channel := range 3 {
		sum[channel] -= (int(pixel[channel])*alpha + 127) / 255
	}
	sum[3] -= alpha
}

func addSample(sum *[4]int, sample [4]uint16) {
	for channel := range 4 {
		sum[channel] += int(sample[channel])
	}
}

func removeSample(sum *[4]int, sample [4]uint16) {
	for channel := range 4 {
		sum[channel] -= int(sample[channel])
	}
}

func insideRoundedRect(x, y int, area image.Rectangle, radius float64) bool {
	radius = min(radius, float64(area.Dx())/2, float64(area.Dy())/2)
	if radius <= 0 {
		return true
	}
	px, py := float64(x)+0.5, float64(y)+0.5
	cx, cy := px, py
	switch {
	case px < float64(area.Min.X)+radius && py < float64(area.Min.Y)+radius:
		cx, cy = float64(area.Min.X)+radius, float64(area.Min.Y)+radius
	case px >= float64(area.Max.X)-radius && py < float64(area.Min.Y)+radius:
		cx, cy = float64(area.Max.X)-radius, float64(area.Min.Y)+radius
	case px >= float64(area.Max.X)-radius && py >= float64(area.Max.Y)-radius:
		cx, cy = float64(area.Max.X)-radius, float64(area.Max.Y)-radius
	case px < float64(area.Min.X)+radius && py >= float64(area.Max.Y)-radius:
		cx, cy = float64(area.Min.X)+radius, float64(area.Max.Y)-radius
	default:
		return true
	}
	dx, dy := px-cx, py-cy
	return dx*dx+dy*dy <= radius*radius
}

func clampByte(value float64) byte {
	return byte(max(0, min(255, int(math.Round(value)))))
}
