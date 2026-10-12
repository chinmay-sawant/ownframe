package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// cachedRoundedFill draws a rounded fill through the raster cache. ok is
// false when the shape is too large to cache; the caller draws it directly.
func cachedRoundedFill(dst *ebiten.Image, op *layout.DisplayOp, radii [4]float64, dx, dy float64) bool {
	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	left, top, size, ok := rasterBox(x, y, w, h, 0)
	if !ok {
		return false
	}

	r := radius4(radii)
	key := rasterKey{
		variant: rasterFill,
		col:     rgba(op),
		w:       w, h: h,
		r:  r,
		fx: x - float32(left),
		fy: y - float32(top),
	}
	img := rasters.raster(key, size, func(img *ebiten.Image) {
		path := roundedFillPath(key.fx, key.fy, w, h, r)
		options := &vector.DrawPathOptions{AntiAlias: true}
		options.ColorScale.ScaleWithColor(key.col)
		vector.FillPath(img, &path, nil, options)
	})
	blitRaster(dst, img, left, top)

	return true
}
