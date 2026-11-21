package bitmap

import (
	"image"
	"image/color"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

func paintShape(img *image.NRGBA, op *layout.DisplayOp) {
	c := opColor(op)
	if c.A == 0 {
		return
	}
	bounds := shapeBounds(op).Intersect(img.Bounds())
	if bounds.Empty() {
		return
	}
	if op.Kind == layout.DisplayOpLine {
		lx, ly, w, h, width := op.PaintLineGeometry()
		lineW := max(width, 1/pxPerPt)
		if math.Hypot(w, h) == 0 {
			return
		}
		paintCoverage(img, op, bounds, func(x, y float64) bool {
			return onLine(x, y, lx, ly, w, h, lineW)
		})
		return
	}
	rx, ry := masking.RadiiXY(op)
	if op.Kind == layout.DisplayOpFillRect {
		if !op.XformSet {
			paintAxisFillRect(img, bounds, op.X, op.Y, op.W, op.H, rx, ry, c)
			return
		}
		paintCoverage(img, op, bounds, func(x, y float64) bool {
			return inRound(x-op.X, y-op.Y, op.W, op.H, rx, ry)
		})
		return
	}
	width := max(op.Width, 1/pxPerPt)
	if op.StrokeMask != 0 {
		paintCoverage(img, op, bounds, func(x, y float64) bool {
			return onStrokeMask(op, x, y, width)
		})
		return
	}
	half := width / halfDivisor
	outerX, outerY, innerX, innerY := strokeRadii(rx, ry, half)
	if !op.XformSet {
		paintAxisStrokeRect(img, bounds, op.X, op.Y, op.W, op.H, width, half, outerX, outerY, innerX, innerY, c)
		return
	}
	paintCoverage(img, op, bounds, func(x, y float64) bool {
		return inRound(x-op.X+half, y-op.Y+half, op.W+width, op.H+width, outerX, outerY) &&
			!inRound(x-op.X-half, y-op.Y-half, op.W-width, op.H-width, innerX, innerY)
	})
}

func paintAxisFillRect(img *image.NRGBA, bounds image.Rectangle, ox, oy, w, h float64,
	rx, ry [4]float64, c color.NRGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	minX := max(bounds.Min.X, int(math.Floor(ox*pxPerPt)))
	maxX := min(bounds.Max.X, int(math.Ceil((ox+w)*pxPerPt)))
	minY := max(bounds.Min.Y, int(math.Floor(oy*pxPerPt)))
	maxY := min(bounds.Max.Y, int(math.Ceil((oy+h)*pxPerPt)))
	if minX >= maxX || minY >= maxY {
		return
	}
	const (
		subLo = 0.25
		subHi = 0.75
	)
	for y := minY; y < maxY; y++ {
		py0 := (float64(y)+subLo)/pxPerPt - oy
		py1 := (float64(y)+subHi)/pxPerPt - oy
		if py1 < 0 || py0 >= h {
			continue
		}
		rowInsideY := py0 >= 0 && py1 < h
		safeX0, safeX1 := maxX, minX
		if rowInsideY {
			var leftR, rightR float64
			if ry[0] > 0 && py0 < ry[0] {
				leftR = max(leftR, rx[0])
			}
			if ry[3] > 0 && py1 > h-ry[3] {
				leftR = max(leftR, rx[3])
			}
			if ry[1] > 0 && py0 < ry[1] {
				rightR = max(rightR, rx[1])
			}
			if ry[2] > 0 && py1 > h-ry[2] {
				rightR = max(rightR, rx[2])
			}
			sx0 := int(math.Ceil((ox+leftR)*pxPerPt - subLo))
			sx1 := int(math.Floor((ox+w-rightR)*pxPerPt - subHi))
			if sx0 < minX {
				sx0 = minX
			}
			if sx1 > maxX {
				sx1 = maxX
			}
			if sx0 < sx1 {
				safeX0, safeX1 = sx0, sx1
				blendSpan(img, y, safeX0, safeX1, c)
			}
		}
		leftEnd := min(safeX0, maxX)
		for x := minX; x < leftEnd; x++ {
			paintAxisFillPixel(img, x, y, ox, oy, w, h, rx, ry, py0, py1, c)
		}
		if safeX0 < safeX1 {
			for x := safeX1; x < maxX; x++ {
				paintAxisFillPixel(img, x, y, ox, oy, w, h, rx, ry, py0, py1, c)
			}
		}
	}
}

const (
	subpixelLo = 0.25
	subpixelHi = 0.75
)

func paintAxisFillPixel(img *image.NRGBA, x, y int, ox, oy, w, h float64,
	rx, ry [4]float64, py0, py1 float64, c color.NRGBA) {
	px0 := (float64(x)+subpixelLo)/pxPerPt - ox
	px1 := (float64(x)+subpixelHi)/pxPerPt - ox
	count := 0
	if inRound(px0, py0, w, h, rx, ry) {
		count++
	}
	if inRound(px1, py0, w, h, rx, ry) {
		count++
	}
	if inRound(px0, py1, w, h, rx, ry) {
		count++
	}
	if inRound(px1, py1, w, h, rx, ry) {
		count++
	}
	if count == 0 {
		return
	}
	src := c
	if count < shapeSamples {
		src.A = uint8(float64(src.A) * float64(count) / shapeSamples)
	}
	overPixel(img, x, y, src)
}

func paintAxisStrokeRect(img *image.NRGBA, bounds image.Rectangle, ox, oy, w, h, width, half float64,
	outerX, outerY, innerX, innerY [4]float64, c color.NRGBA) {
	outerOX, outerOY, outerW, outerH := ox-half, oy-half, w+width, h+width
	innerOX, innerOY, innerW, innerH := ox+half, oy+half, w-width, h-width
	minX := max(bounds.Min.X, int(math.Floor(outerOX*pxPerPt)))
	maxX := min(bounds.Max.X, int(math.Ceil((outerOX+outerW)*pxPerPt)))
	minY := max(bounds.Min.Y, int(math.Floor(outerOY*pxPerPt)))
	maxY := min(bounds.Max.Y, int(math.Ceil((outerOY+outerH)*pxPerPt)))
	if minX >= maxX || minY >= maxY {
		return
	}
	const (
		subLo = 0.25
		subHi = 0.75
	)
	for y := minY; y < maxY; y++ {
		py0 := (float64(y) + subLo) / pxPerPt
		py1 := (float64(y) + subHi) / pxPerPt
		oPy0, oPy1 := py0-outerOY, py1-outerOY
		if oPy1 < 0 || oPy0 >= outerH {
			continue
		}
		iPy0, iPy1 := py0-innerOY, py1-innerOY
		holeX0, holeX1 := maxX, minX
		if innerW > 0 && innerH > 0 && iPy0 >= 0 && iPy1 < innerH {
			var leftR, rightR float64
			if innerY[0] > 0 && iPy0 < innerY[0] {
				leftR = max(leftR, innerX[0])
			}
			if innerY[3] > 0 && iPy1 > innerH-innerY[3] {
				leftR = max(leftR, innerX[3])
			}
			if innerY[1] > 0 && iPy0 < innerY[1] {
				rightR = max(rightR, innerX[1])
			}
			if innerY[2] > 0 && iPy1 > innerH-innerY[2] {
				rightR = max(rightR, innerX[2])
			}
			hx0 := int(math.Ceil((innerOX+leftR)*pxPerPt - subLo))
			hx1 := int(math.Floor((innerOX+innerW-rightR)*pxPerPt - subHi))
			if hx0 < minX {
				hx0 = minX
			}
			if hx1 > maxX {
				hx1 = maxX
			}
			if hx0 < hx1 {
				holeX0, holeX1 = hx0, hx1
			}
		}
		leftEnd := min(holeX0, maxX)
		for x := minX; x < leftEnd; x++ {
			paintAxisStrokePixel(img, x, y, outerOX, outerW, outerH, innerOX, innerW, innerH,
				outerX, outerY, innerX, innerY, oPy0, oPy1, iPy0, iPy1, c)
		}
		if holeX0 < holeX1 {
			for x := holeX1; x < maxX; x++ {
				paintAxisStrokePixel(img, x, y, outerOX, outerW, outerH, innerOX, innerW, innerH,
					outerX, outerY, innerX, innerY, oPy0, oPy1, iPy0, iPy1, c)
			}
		}
	}
}

func paintAxisStrokePixel(img *image.NRGBA, x, y int,
	outerOX, outerW, outerH, innerOX, innerW, innerH float64,
	outerX, outerY, innerX, innerY [4]float64,
	oPy0, oPy1, iPy0, iPy1 float64, c color.NRGBA) {
	px0 := (float64(x) + subpixelLo) / pxPerPt
	px1 := (float64(x) + subpixelHi) / pxPerPt
	oPx0, oPx1 := px0-outerOX, px1-outerOX
	iPx0, iPx1 := px0-innerOX, px1-innerOX
	count := 0
	if inRound(oPx0, oPy0, outerW, outerH, outerX, outerY) && !inRound(iPx0, iPy0, innerW, innerH, innerX, innerY) {
		count++
	}
	if inRound(oPx1, oPy0, outerW, outerH, outerX, outerY) && !inRound(iPx1, iPy0, innerW, innerH, innerX, innerY) {
		count++
	}
	if inRound(oPx0, oPy1, outerW, outerH, outerX, outerY) && !inRound(iPx0, iPy1, innerW, innerH, innerX, innerY) {
		count++
	}
	if inRound(oPx1, oPy1, outerW, outerH, outerX, outerY) && !inRound(iPx1, iPy1, innerW, innerH, innerX, innerY) {
		count++
	}
	if count == 0 {
		return
	}
	src := c
	if count < shapeSamples {
		src.A = uint8(float64(src.A) * float64(count) / shapeSamples)
	}
	overPixel(img, x, y, src)
}

func overPixel(img *image.NRGBA, x, y int, src color.NRGBA) {
	off := img.PixOffset(x, y)
	if src.A == channelMax {
		img.Pix[off] = src.R
		img.Pix[off+1] = src.G
		img.Pix[off+2] = src.B
		img.Pix[off+3] = src.A
		return
	}
	dst := color.NRGBA{R: img.Pix[off], G: img.Pix[off+1], B: img.Pix[off+2], A: img.Pix[off+3]}
	out := over(dst, src)
	img.Pix[off] = out.R
	img.Pix[off+1] = out.G
	img.Pix[off+2] = out.B
	img.Pix[off+3] = out.A
}

func blendSpan(img *image.NRGBA, y, x0, x1 int, src color.NRGBA) {
	off := img.PixOffset(x0, y)
	if src.A == channelMax {
		for x := x0; x < x1; x++ {
			img.Pix[off] = src.R
			img.Pix[off+1] = src.G
			img.Pix[off+2] = src.B
			img.Pix[off+3] = src.A
			off += 4
		}
		return
	}
	for x := x0; x < x1; x++ {
		dst := color.NRGBA{R: img.Pix[off], G: img.Pix[off+1], B: img.Pix[off+2], A: img.Pix[off+3]}
		out := over(dst, src)
		img.Pix[off] = out.R
		img.Pix[off+1] = out.G
		img.Pix[off+2] = out.B
		img.Pix[off+3] = out.A
		off += 4
	}
}

// paintCoverage paints one operation's ink at four subsamples per pixel,
// scaling the colour's alpha by how many subsamples land inside. hit reports
// whether one point in the operation's local space is ink; the caller bounds
// the loop because a rounded rectangle and a shaped path bound themselves
// differently.
func paintCoverage(img *image.NRGBA, op *layout.DisplayOp, bounds image.Rectangle,
	hit func(x, y float64) bool) {
	bounds = bounds.Intersect(img.Bounds())
	c := opColor(op)
	if c.A == 0 {
		return
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			count := 0
			for _, oy := range [2]float64{.25, .75} {
				for _, ox := range [2]float64{.25, .75} {
					px, py, ok := untransform(op, (float64(x)+ox)/pxPerPt, (float64(y)+oy)/pxPerPt)
					if ok && hit(px, py) {
						count++
					}
				}
			}
			if count == 0 {
				continue
			}
			src := c
			if count < shapeSamples {
				src.A = uint8(float64(src.A) * float64(count) / shapeSamples)
			}
			overPixel(img, x, y, src)
		}
	}
}

func shapeBounds(op *layout.DisplayOp) image.Rectangle {
	x, y, w, h := op.X, op.Y, op.W, op.H
	if op.Kind == layout.DisplayOpLine {
		x, y, w, h, _ = op.PaintLineGeometry()
	}
	pad := max(op.Width, 1)
	x0, y0, x1, y1 := min(x, x+w)-pad, min(y, y+h)-pad, max(x, x+w)+pad, max(y, y+h)+pad
	loX, loY, hiX, hiY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range [][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} {
		xx, yy := transform(op, p[0], p[1])
		loX, loY, hiX, hiY = min(loX, xx), min(loY, yy), max(hiX, xx), max(hiY, yy)
	}
	return image.Rect(int(math.Floor(loX*pxPerPt)), int(math.Floor(loY*pxPerPt)),
		int(math.Ceil(hiX*pxPerPt)), int(math.Ceil(hiY*pxPerPt)))
}
