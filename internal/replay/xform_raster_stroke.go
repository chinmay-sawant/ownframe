package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// strokeRaster caches a transformed border. The key carries the stroke width,
// the side mask, and the elliptical radii, as the direct painter's bake does.
func strokeRaster(op *layout.DisplayOp, x, y, w, h float32) (*ebiten.Image, image.Point, bool) {
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

	rx, ry := render.RadiiXY(op)
	key := rasterKey{
		variant: rasterStroke,
		mask:    op.StrokeMask,
		col:     rgba(op),
		w:       w, h: h, width: width,
		r:  radius4(rx),
		ry: radius4(ry),
		fx: x - float32(left),
		fy: y - float32(top),
	}

	img := rasters.raster(key, size, func(img *ebiten.Image) {
		path := strokePathAt(key)
		stroke := &vector.StrokeOptions{Width: width, LineJoin: vector.LineJoinRound}
		options := &vector.DrawPathOptions{AntiAlias: true}
		options.ColorScale.ScaleWithColor(key.col)
		vector.StrokePath(img, &path, stroke, options)
	})

	return img, image.Pt(left, top), true
}
