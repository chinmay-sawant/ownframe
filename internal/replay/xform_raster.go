package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// xformRaster returns a cached raster of the op's local shape and the local
// integer origin to blit it at, or ok false when the shape is too large to
// cache. The raster is keyed on the shape and color alone, so one shape at
// any transform reuses it, and a stable image keeps the linear-filter blit
// deterministic.
func xformRaster(op *layout.DisplayOp) (img *ebiten.Image, at image.Point, ok bool) {
	x, y := float32(op.X*pxPerPt), float32(op.Y*pxPerPt)
	w, h := float32(op.W*pxPerPt), float32(op.H*pxPerPt)

	switch op.Kind {
	case layout.DisplayOpFillRect:
		return fillRaster(op, x, y, w, h)
	case layout.DisplayOpStrokeRect:
		return strokeRaster(op, x, y, w, h)
	}

	return nil, image.Point{}, false
}

func fillRaster(op *layout.DisplayOp, x, y, w, h float32) (*ebiten.Image, image.Point, bool) {
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

	return img, image.Pt(left, top), true
}
