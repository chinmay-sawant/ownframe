package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"math"
)

// xformBounds bounds a transformed op in CSS pixels: the corners of its local
// box map through the matrix, so a rotation or skew bounds as the mapped
// parallelogram instead of its layout box. ok is false for a kind the
// painter cannot bound, such as a transformed grid.
func xformBounds(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	return matrixBounds(op, ppt, op.Transform())
}

func matrixBounds(op *layout.DisplayOp, ppt float64, m layout.Matrix2D) (image.Rectangle, bool) {
	l, t, r, b, ok := xformPointsBox(op)
	if !ok {
		return image.Rectangle{}, false
	}

	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for _, p := range [4][2]float64{{l, t}, {r, t}, {r, b}, {l, b}} {
		x, y := m.Apply(p[0], p[1])
		loX, loY = math.Min(loX, x), math.Min(loY, y)
		hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
	}

	pad := float64(boundsPad)

	return image.Rect(
		int(math.Floor(loX/ppt-pad)), int(math.Floor(loY/ppt-pad)),
		int(math.Ceil(hiX/ppt+pad)), int(math.Ceil(hiY/ppt+pad)),
	), true
}

// xformPointsBox is the op's local painted box in points, before its baked
// transform: the same geometry the untransformed bounds cover.
func xformPointsBox(op *layout.DisplayOp) (l, t, r, b float64, ok bool) {
	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		top := op.Y - textAscent(op)
		bottom := op.Y + op.InkDescent
		if bottom < top {
			bottom = top
		}

		return op.X, top, op.X + op.W, bottom, true
	case layout.DisplayOpLine:
		x, y, w, h, width := op.PaintLineGeometry()
		grow := math.Max(width, 1) / halfDiv

		return min(x, x+w) - grow, min(y, y+h) - grow,
			max(x, x+w) + grow, max(y, y+h) + grow, true
	case layout.DisplayOpStrokeRect:
		grow := math.Max(op.Width, 1) / halfDiv

		return op.X - grow, op.Y - grow,
			op.X + op.W + grow, op.Y + op.H + grow, true
	case layout.DisplayOpFillRect:
		return op.X, op.Y, op.X + op.W, op.Y + op.H, true
	}

	return 0, 0, 0, 0, false
}

// drawXform draws one op through its baked transform: a shape comes from the
// raster cache and blits through the matrix, and a run of glyphs draws with
// the matrix as its GeoM. A transformed grid carries no paint, because the
// gate keeps it on the bitmap path.
func (renderer *Renderer) drawXform(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		renderer.drawXformText(dst, op, dx, dy)
	case layout.DisplayOpGridRun:
	default:
		renderer.drawXformShape(dst, op, dx, dy)
	}
}

// drawXformShape draws a fill, stroke, or line through the op matrix. The
// cached raster is reused across transforms; a shape too large to cache
// rasterizes into a scratch image instead.
func (renderer *Renderer) drawXformShape(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	img, at, ok := renderer.xformRaster(op)
	if !ok {
		renderer.drawXformScratch(dst, op, dx, dy)

		return
	}

	dst, ok = scissorOverflow(dst, op, dx, dy)
	if !ok {
		return
	}

	geom := xformGeom(op, dx, dy)
	shiftLocal(&geom, float64(at.X)-op.X*pxPerPt, float64(at.Y)-op.Y*pxPerPt)
	opts := &ebiten.DrawImageOptions{
		GeoM:   geom,
		Filter: ebiten.FilterLinear,
	}
	opts.ColorScale.ScaleWithColor(rgba(op))
	dst.DrawImage(img, opts)
}

// xformPad is the pixel slack around a transformed shape's local raster.
const xformPad = 2

// xformGeom maps an op's local device pixels through its baked transform into
// dst pixels. Local coordinates count from the op's own origin in CSS pixels;
// the matrix itself is unitless, so only its translation scales from points.
func xformGeom(op *layout.DisplayOp, dx, dy float64) ebiten.GeoM {
	m := op.Transform()

	var g ebiten.GeoM
	g.SetElement(0, 0, m.A)
	g.SetElement(0, 1, m.C)
	g.SetElement(0, geomTransCol, pxPerPt*(m.A*op.X+m.C*op.Y+m.E)+dx)
	g.SetElement(1, 0, m.B)
	g.SetElement(1, 1, m.D)
	g.SetElement(1, geomTransCol, pxPerPt*(m.B*op.X+m.D*op.Y+m.F)+dy)

	return g
}

// shiftLocal moves a geom's origin by (x, y) in its local space, before the
// matrix. Ebiten's GeoM.Translate moves the origin in dst space instead.
func shiftLocal(g *ebiten.GeoM, x, y float64) {
	g.SetElement(0, geomTransCol, g.Element(0, geomTransCol)+g.Element(0, 0)*x+g.Element(0, 1)*y)
	g.SetElement(1, geomTransCol, g.Element(1, geomTransCol)+g.Element(1, 0)*x+g.Element(1, 1)*y)
}

// xformLocalBox is an op's painted box in its own device pixels, before the
// baked transform: the geometry the untransformed painter would draw.
func xformLocalBox(op *layout.DisplayOp) (l, t, r, b float64, ok bool) {
	k := pxPerPt

	switch op.Kind {
	case layout.DisplayOpFillRect:
		return op.X * k, op.Y * k, (op.X + op.W) * k, (op.Y + op.H) * k, true
	case layout.DisplayOpStrokeRect:
		grow := math.Max(op.Width, 1) / halfDiv * k

		return op.X*k - grow, op.Y*k - grow,
			(op.X+op.W)*k + grow, (op.Y+op.H)*k + grow, true
	case layout.DisplayOpLine:
		x, y, w, h, width := op.PaintLineGeometry()
		grow := math.Max(width, 1) / halfDiv * k
		x0, x1 := min(x, x+w), max(x, x+w)
		y0, y1 := min(y, y+h), max(y, y+h)

		return x0*k - grow, y0*k - grow, x1*k + grow, y1*k + grow, true
	}

	return 0, 0, 0, 0, false
}

// xformRaster returns a cached raster of the op's local shape and the local
// integer origin to blit it at, or ok false when the shape is too large to
// cache. The raster is keyed on the shape and color alone, so one shape at
// any transform reuses it, and a stable image keeps the linear-filter blit
// deterministic.
func (renderer *Renderer) xformRaster(op *layout.DisplayOp) (img *ebiten.Image, at image.Point, ok bool) {
	x, y := float32(op.X*pxPerPt), float32(op.Y*pxPerPt)
	w, h := float32(op.W*pxPerPt), float32(op.H*pxPerPt)

	switch op.Kind {
	case layout.DisplayOpFillRect:
		return renderer.fillRaster(op, x, y, w, h)
	case layout.DisplayOpStrokeRect:
		return renderer.strokeRaster(op, x, y, w, h)
	}

	return nil, image.Point{}, false
}

func (renderer *Renderer) fillRaster(op *layout.DisplayOp, x, y, w, h float32) (*ebiten.Image, image.Point, bool) {
	radii, ok := masking.FillRadii(op)
	if !ok {
		return nil, image.Point{}, false
	}

	left, top, size, ok := rasterBox(x, y, w, h, 0)
	if !ok {
		return nil, image.Point{}, false
	}

	r := radius4(radii)
	key := rasterKey{
		variant: rasterFill,
		w:       w, h: h,
		r:  r,
		fx: x - float32(left),
		fy: y - float32(top),
	}

	img := renderer.rasters.raster(key, size, func(img *ebiten.Image) {
		path := acquireVectorPath()
		populateRoundedFillPath(path, key.fx, key.fy, w, h, r)

		var options vector.DrawPathOptions
		options.AntiAlias = true

		vector.FillPath(img, path, nil, &options)
		releaseVectorPath(path)
	})

	return img, image.Pt(left, top), true
}

// strokeRaster caches a transformed border. The key carries the stroke width,
// the side mask, and the elliptical radii, as the direct painter's bake does.
func (renderer *Renderer) strokeRaster(op *layout.DisplayOp, x, y, w, h float32) (*ebiten.Image, image.Point, bool) {
	if op.StrokeMask&^0x0F != 0 {
		return nil, image.Point{}, false
	}

	width := float32(op.Width * pxPerPt)
	if width < 1 {
		width = 1
	}

	left, top, size, ok := rasterBox(x, y, w, h, width/2+1)
	if !ok {
		return nil, image.Point{}, false
	}

	rx, ry := masking.RadiiXY(op)
	key := rasterKey{
		variant: rasterStroke,
		mask:    op.StrokeMask,
		w:       w, h: h, width: width,
		r:  radius4(rx),
		ry: radius4(ry),
		fx: x - float32(left),
		fy: y - float32(top),
	}

	img := renderer.rasters.raster(key, size, func(img *ebiten.Image) {
		path := acquireVectorPath()
		populateStrokePath(path, key)

		var stroke vector.StrokeOptions
		stroke.Width = width
		stroke.LineJoin = vector.LineJoinRound

		var options vector.DrawPathOptions
		options.AntiAlias = true

		vector.StrokePath(img, path, &stroke, &options)
		releaseVectorPath(path)
	})

	return img, image.Pt(left, top), true
}

// maxScratchPixels caps one scratch raster, mirroring maxXformPixels.
const maxScratchPixels = 4 << 20

// xformScratch returns the local raster's integer origin and size, or ok
// false when the local box cannot make an image.
func xformScratch(op *layout.DisplayOp) (left, top int, size image.Point, ok bool) {
	l, t, r, b, ok := xformLocalBox(op)
	if !ok {
		return 0, 0, image.Point{}, false
	}

	left = int(math.Floor(l)) - xformPad
	top = int(math.Floor(t)) - xformPad
	right := int(math.Ceil(r)) + xformPad
	bottom := int(math.Ceil(b)) + xformPad
	size = image.Pt(right-left, bottom-top)

	if size.X <= 0 || size.Y <= 0 || oversizedImage(size) {
		return 0, 0, image.Point{}, false
	}

	return left, top, size, true
}

// oversizedImage reports a raster no Ebiten image can hold. A zero limit
// means the game is not running, where no GPU image exists to overflow.
func oversizedImage(size image.Point) bool {
	if size.X <= 0 || size.Y <= 0 {
		return true
	}
	limit := ebiten.MaxImageSize()
	if limit > 0 && (size.X > limit || size.Y > limit) {
		return true
	}
	return int64(size.X)*int64(size.Y) > maxScratchPixels
}

// drawXformScratch rasterizes a shape too large for the raster cache into a
// local scratch image and blits it through the op matrix.
func (renderer *Renderer) drawXformScratch(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	left, top, size, ok := xformScratch(op)
	if !ok {
		return
	}

	scratch := ebiten.NewImage(size.X, size.Y)
	defer scratch.Deallocate()

	ox, oy := float64(-left), float64(-top)

	switch op.Kind {
	case layout.DisplayOpFillRect:
		renderer.fillRect(scratch, op, ox, oy)
	case layout.DisplayOpStrokeRect:
		renderer.drawStrokeRect(scratch, op, ox, oy)
	case layout.DisplayOpLine:
		drawLine(scratch, op, ox, oy)
	}

	// Blit the interior: a bilinear tap at a whole image's boundary reads a
	// texel outside its region, whose content is undefined, and the shape
	// keeps a two-pixel pad, so one pixel of inset costs no ink.
	src := scratch
	if size.X > 2 && size.Y > 2 {
		if inner, ok := scratch.SubImage(image.Rect(1, 1, size.X-1, size.Y-1)).(*ebiten.Image); ok {
			src = inner
			left++
			top++
		}
	}

	geom := xformGeom(op, dx, dy)
	shiftLocal(&geom, float64(left)-op.X*pxPerPt, float64(top)-op.Y*pxPerPt)
	clipped, paint := scissorOverflow(dst, op, dx, dy)
	if !paint {
		return
	}
	clipped.DrawImage(src, &ebiten.DrawImageOptions{
		GeoM:   geom,
		Filter: ebiten.FilterLinear,
	})
}

// drawXformText draws one shaped run through its baked transform. The op's
// baseline origin maps through the matrix, and the run's own geometry stays
// in text space, so a rotation tilts the glyphs instead of stair-stepping
// them. Letter spacing advances in text space, as in the direct painter.
func (renderer *Renderer) drawXformText(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	face := renderer.textFace(op)
	if face == nil || op.Text == "" {
		return
	}

	dst, paint := scissorOverflow(dst, op, dx, dy)
	if !paint {
		return
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	geom := xformGeom(op, dx, dy)
	if !op.XformSet {
		geom.SetElement(0, geomTransCol, op.X*pxPerPt+dx)
		geom.SetElement(1, geomTransCol, op.Y*pxPerPt+dy)
	}
	shiftLocal(&geom, 0, -face.Metrics().HAscent)

	if op.LetterSpacing == 0 {
		renderer.drawXformRun(dst, content, face, geom, op)

		return
	}

	pen := 0.0

	for i := 0; i < len(content); {
		glyph, size := glyphAt(content, i)
		i += size
		at := geom
		shiftLocal(&at, pen, 0)
		renderer.drawXformRun(dst, glyph, face, at, op)

		width, _ := text.Measure(glyph, face, 0)
		pen += width + op.LetterSpacing*pxPerPt
	}
}

// drawXformRun draws one string at the text origin of geom. Fake bold strikes
// the string twice one local pixel apart, as the direct painter does.
func (renderer *Renderer) drawXformRun(dst *ebiten.Image, content string, face *text.GoTextFace, geom ebiten.GeoM, op *layout.DisplayOp) {
	options := &text.DrawOptions{}
	options.GeoM = geom
	options.ColorScale.ScaleWithColor(rgba(op))

	text.Draw(dst, content, face, options)

	if layout.DisplayFakeBold(op) {
		bold := geom
		shiftLocal(&bold, 1, 0)
		options.GeoM = bold
		text.Draw(dst, content, face, options)
	}
}
