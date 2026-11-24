package bitmap

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
)

const (
	degreesQuarterTurn      = 90
	degreesThreeQuarterTurn = 270
	degreesFullTurn         = 360
	nrgbaAlphaOffset        = 3
)

// drawString draws s with letter spacing, text-autospace gap, and tab-stop
// expansion. The gap is the extra advance layout put at each ideograph and
// Latin boundary, in pixels. runW is the run's laid-out width in pixels, used
// to advance across literal tab characters without drawing missing-glyph boxes.
// x/image's face does no OpenType shaping, so this painter draws a run without
// its font features (small caps, tabular numerals, ligature control): the
// glyphs keep their default forms and advances. That is the degraded picture a
// fallback page shows instead of failing; the display-list replay applies the
// features.
func drawString(d *font.Drawer, s string, spacing, gap, runW float64) {
	if strings.ContainsRune(s, '\t') {
		drawTabbedString(d, s, spacing, gap, runW)
		return
	}
	drawSegment(d, s, spacing, gap)
}

func drawTabbedString(d *font.Drawer, s string, spacing, gap, runW float64) {
	parts := strings.Split(s, "\t")
	tabCount := len(parts) - 1
	if tabCount <= 0 {
		drawSegment(d, s, spacing, gap)
		return
	}

	var totalSegW fixed.Int26_6
	for _, seg := range parts {
		totalSegW += measureSegment(d, seg, spacing, gap)
	}

	runWFixed := fixed.Int26_6(math.Round(runW * fixedPointScale))
	var tabW fixed.Int26_6
	if runWFixed > totalSegW {
		tabW = (runWFixed - totalSegW) / fixed.Int26_6(tabCount)
	}

	for i, seg := range parts {
		if seg != "" {
			drawSegment(d, seg, spacing, gap)
		}
		if i < tabCount {
			d.Dot.X += tabW
		}
	}
}

func measureSegment(d *font.Drawer, s string, spacing, gap float64) fixed.Int26_6 {
	if s == "" {
		return 0
	}
	if spacing == 0 && gap == 0 {
		return d.MeasureString(s)
	}
	var total fixed.Int26_6
	var prev rune
	for _, r := range s {
		total += fixedSpacing(masking.AutospaceGapBetween(prev, r, gap))
		total += d.MeasureString(string(r))
		total += fixedSpacing(spacing)
		prev = r
	}
	return total
}

func drawSegment(d *font.Drawer, s string, spacing, gap float64) {
	if spacing == 0 && gap == 0 {
		d.DrawString(s)
		return
	}
	var prev rune
	for _, r := range s {
		d.Dot.X += fixedSpacing(masking.AutospaceGapBetween(prev, r, gap))
		d.DrawString(string(r))
		d.Dot.X += fixedSpacing(spacing)
		prev = r
	}
}

// tightAlphaBounds returns the smallest sub-rectangle of img that contains a
// non-zero alpha pixel, trimming the generous ascent/overhang padding before
// rotating or resampling glyph pixels.
func tightAlphaBounds(img *image.NRGBA) image.Rectangle {
	r := img.Rect
	minX, minY := r.Max.X, r.Max.Y
	maxX, maxY := r.Min.X, r.Min.Y
	for y := r.Min.Y; y < r.Max.Y; y++ {
		off := img.PixOffset(r.Min.X, y) + nrgbaAlphaOffset
		rowHit := false
		for x := r.Min.X; x < r.Max.X; x++ {
			if img.Pix[off] != 0 {
				if x < minX {
					minX = x
				}
				if x+1 > maxX {
					maxX = x + 1
				}
				rowHit = true
			}
			off += nrgbaBytesPerPixel
		}
		if rowHit {
			if y < minY {
				minY = y
			}
			if y+1 > maxY {
				maxY = y + 1
			}
		}
	}
	if minX >= maxX || minY >= maxY {
		return image.Rectangle{}
	}

	return image.Rect(minX, minY, maxX, maxY)
}

// paintQuarterTurnText composites an exact 90/180/270-degree rotated glyph
// buffer directly into dst, skipping transparent pixels instead of scanning
// through xdraw.BiLinear.Transform's generic RGBA64Image loop.
func paintQuarterTurnText(dst, src *image.NRGBA, srcBounds image.Rectangle, matrix f64.Aff3) {
	dstBounds := dst.Bounds()
	m0, m1, m2 := matrix[0], matrix[1], matrix[2]
	m3, m4, m5 := matrix[3], matrix[4], matrix[5]

	for sy := srcBounds.Min.Y; sy < srcBounds.Max.Y; sy++ {
		off := src.PixOffset(srcBounds.Min.X, sy)
		fy := float64(sy) + blendMidpoint
		for sx := srcBounds.Min.X; sx < srcBounds.Max.X; sx++ {
			sa := src.Pix[off+3]
			if sa != 0 {
				fx := float64(sx) + blendMidpoint
				dx := int(math.Floor(m0*fx + m1*fy + m2))
				dy := int(math.Floor(m3*fx + m4*fy + m5))
				if dx >= dstBounds.Min.X && dx < dstBounds.Max.X &&
					dy >= dstBounds.Min.Y && dy < dstBounds.Max.Y {
					overPixel(dst, dx, dy, color.NRGBA{
						R: src.Pix[off],
						G: src.Pix[off+1],
						B: src.Pix[off+2],
						A: sa,
					})
				}
			}
			off += nrgbaBytesPerPixel
		}
	}
}

func paintTextOp(dst *image.NRGBA, op *layout.DisplayOp, faces *faceCache) error {
	if !op.XformSet && op.RotateDeg == 0 && !op.FakeOblique {
		return paintText(dst, op, faces)
	}
	if op.XformSet && op.RotateDeg == 0 && !op.FakeOblique {
		m := op.Transform()
		if m.A == 1 && m.B == 0 && m.C == 0 && m.D == 1 {
			shifted := *op
			shifted.X += m.E
			shifted.Y += m.F
			shifted.XformSet = false
			return paintText(dst, &shifted, faces)
		}
	}
	// Bound the untransformed glyph canvas, including ascent and italic overhang.
	pad := op.Size * pxPerPt * textPadScale
	b := image.Rect(int(op.X*pxPerPt-pad), int(op.Y*pxPerPt-pad),
		int((op.X+op.W)*pxPerPt+pad), int(op.Y*pxPerPt+pad))
	depth := 1
	for g := op.Group(); g != nil; g = g.Parent {
		depth++
	}
	budget := maxPixels - depth*dst.Bounds().Dx()*dst.Bounds().Dy()
	if b.Empty() || b.Dx() > budget/max(1, b.Dy()) {
		return fmt.Errorf("bitmap: transformed text exceeds buffer budget")
	}
	tmp := image.NewNRGBA(b)
	if err := paintText(tmp, op, faces); err != nil {
		return err
	}
	tight := tightAlphaBounds(tmp)
	if tight.Empty() {
		return nil
	}
	m := op.Transform()
	if !op.XformSet {
		m.A, m.B, m.C, m.D, m.E, m.F = 1, 0, 0, 1, 0, 0
	}
	rotMod := math.Mod(float64(op.RotateDeg), degreesFullTurn)
	if rotMod < 0 {
		rotMod += degreesFullTurn
	}
	isQuarterTurn := math.Mod(rotMod, degreesQuarterTurn) == 0
	var c, s float64
	if isQuarterTurn {
		switch rotMod {
		case 0:
			c, s = 1, 0
		case degreesQuarterTurn:
			c, s = 0, 1
		case degreesInHalfCircle:
			c, s = -1, 0
		case degreesThreeQuarterTurn:
			c, s = 0, -1
		}
	} else {
		angle := float64(op.RotateDeg) * math.Pi / degreesInHalfCircle
		c, s = math.Cos(angle), math.Sin(angle)
	}
	shear := 0.0
	if op.FakeOblique {
		shear = fakeObliqueShear
	}
	a, bv, cv, d := c, s, c*shear-s, s*shear+c
	x, y := op.X*pxPerPt, op.Y*pxPerPt
	tx, ty := x-a*x-cv*y, y-bv*x-d*y
	matrix := f64.Aff3{m.A*a + m.C*bv, m.A*cv + m.C*d, m.A*tx + m.C*ty + m.E*pxPerPt,
		m.B*a + m.D*bv, m.B*cv + m.D*d, m.B*tx + m.D*ty + m.F*pxPerPt}

	if !op.XformSet && !op.FakeOblique && isQuarterTurn {
		paintQuarterTurnText(dst, tmp, tight, matrix)
		return nil
	}

	xdraw.BiLinear.Transform(dst, matrix, tmp, tight, draw.Over, nil)
	return nil
}
