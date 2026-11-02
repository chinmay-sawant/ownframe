package bitmap

import (
	"image"
	"math"

	"github.com/chinmay-sawant/ownframe/internal/backdrop"
)

// CPU backdrop-filter tuning.
const (
	// blurWindowPad extends the sliding blur window past the exclusive end
	// by one step, so the next column starts with the window covered.
	blurWindowPad = 2
	// pixelCenter is the half-texel offset from a pixel corner to its center.
	pixelCenter = 0.5
	// refractionSpan scales the lens falloff distance from the strength.
	refractionSpan = 2
	// lumaRedWeight, lumaGreenWeight and lumaBlueWeight are the luma
	// coefficients used for the saturate pass.
	lumaRedWeight   = 0.213
	lumaGreenWeight = 0.715
	lumaBlueWeight  = 0.072
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
				average := (sum[channel] + count/2) / count
				horizontal[index][channel] = uint16(min(max(average, 0), math.MaxUint16))
			}
			nextLo := max(dst.Bounds().Min.X, x+1-blurRadius)
			nextHi := min(dst.Bounds().Max.X, x+blurRadius+blurWindowPad)
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
					red = float64(sum[0]) * channelMax / float64(count*alpha)
					green = float64(sum[1]) * channelMax / float64(count*alpha)
					blue = float64(sum[2]) * channelMax / float64(count*alpha)
				}
				luma := lumaRedWeight*red + lumaGreenWeight*green + lumaBlueWeight*blue
				index := (y-area.Min.Y)*width + x - area.Min.X
				filtered[index] = [4]byte{
					clampByte(luma + (red-luma)*filter.Saturation),
					clampByte(luma + (green-luma)*filter.Saturation),
					clampByte(luma + (blue-luma)*filter.Saturation),
					byte(max(0, min(channelMax, alpha))),
				}
			}

			nextLo := max(rows.Min.Y, y+1-blurRadius)
			nextHi := min(rows.Max.Y, y+blurRadius+blurWindowPad)
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
			dx, dy := backdropLensOffset(float64(x)+pixelCenter, float64(y)+pixelCenter, shape, radius, filter.RefractionPx)
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
	halfWidth, halfHeight := float64(shape.Dx())*pixelCenter, float64(shape.Dy())*pixelCenter
	radius = min(radius, halfWidth, halfHeight)
	centerX := x - (float64(shape.Min.X) + halfWidth)
	centerY := y - (float64(shape.Min.Y) + halfHeight)
	qx := math.Abs(centerX) - (halfWidth - radius)
	qy := math.Abs(centerY) - (halfHeight - radius)
	outX, outY := max(qx, 0), max(qy, 0)
	outLength := math.Hypot(outX, outY)
	distance := outLength + min(max(qx, qy), 0) - radius
	edge := min(max(1+distance/max(strength*refractionSpan, 1), 0), 1)
	var normalX, normalY float64
	switch {
	case outLength > 0:
		normalX = math.Copysign(outX/outLength, centerX)
		normalY = math.Copysign(outY/outLength, centerY)
	case qx > qy:
		normalX = math.Copysign(1, centerX)
	default:
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
		sum[channel] += (int(pixel[channel])*alpha + channelMax/2) / channelMax
	}
	sum[3] += alpha
}

func removePixel(sum *[4]int, pixel []byte) {
	alpha := int(pixel[3])
	for channel := range 3 {
		sum[channel] -= (int(pixel[channel])*alpha + channelMax/2) / channelMax
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
	radius = min(radius, float64(area.Dx())/halfDivisor, float64(area.Dy())/halfDivisor)
	if radius <= 0 {
		return true
	}
	px, py := float64(x)+pixelCenter, float64(y)+pixelCenter
	var cx, cy float64
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
	return byte(max(0, min(channelMax, int(math.Round(value)))))
}
