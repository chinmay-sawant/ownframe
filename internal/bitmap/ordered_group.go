package bitmap

import (
	"fmt"
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/backdrop"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

const (
	groupBoundsPadPx  = 2
	groupHalfWidthDiv = 2
)

func paintGroup(dst *image.NRGBA, group *layout.DisplayGroup,
	items map[*layout.DisplayGroup][]paintItem,
	areas map[*layout.DisplayGroup]image.Rectangle,
	radii map[*layout.DisplayGroup]float64,
	faces *faceCache,
) error {
	for _, item := range items[group] {
		if item.op != nil {
			if err := paintOp(dst, item.op, faces); err != nil {
				return err
			}
			continue
		}
		if item.group.BackdropFilter != "" {
			filter, supported := backdrop.Parse(item.group.BackdropFilter)
			area, hasArea := areas[item.group]
			if supported && !hasArea {
				return fmt.Errorf("bitmap: backdrop-filter group has no border box")
			}
			if supported {
				applyBackdropFilter(dst, area, radii[item.group], filter)
			}
		}
		// Tight group bounds: a masked group zeros every pixel outside its
		// mask area, and a non-masked blend/isolation group only paints where
		// its ops land. Allocating only bounds ∩ dst.Bounds() avoids O(N*W*H)
		// full-canvas buffers and composites when many elements carry groups.
		bufRect := dst.Bounds()
		if item.group.Mask != nil {
			area, ok := areas[item.group]
			if !ok {
				return fmt.Errorf("bitmap: masked group has no border box")
			}
			bufRect = area.Intersect(dst.Bounds())
			if bufRect.Empty() {
				continue
			}
		} else if tight, ok := groupOpBounds(item.group, items, areas); ok {
			bufRect = tight.Intersect(dst.Bounds())
			if bufRect.Empty() {
				continue
			}
		}
		buffer := image.NewNRGBA(bufRect)
		if err := paintGroup(buffer, item.group, items, areas, radii, faces); err != nil {
			return err
		}
		if item.group.Mask != nil {
			area := areas[item.group]
			if err := masking.Apply(buffer, area, item.group.Mask); err != nil {
				return err
			}
		}
		composite(dst, buffer, item.group.Mode)
	}
	return nil
}

// groupOpBounds returns a conservative pixel bounding box for all ops inside
// group and its descendants, or ok false when any descendant cannot be bounded
// safely (such as a path or rotated text).
func groupOpBounds(
	group *layout.DisplayGroup,
	items map[*layout.DisplayGroup][]paintItem,
	areas map[*layout.DisplayGroup]image.Rectangle,
) (image.Rectangle, bool) {
	var union image.Rectangle
	for _, item := range items[group] {
		var (
			box image.Rectangle
			ok  bool
		)
		switch {
		case item.group != nil && item.group.Mask != nil:
			box, ok = areas[item.group]
		case item.group != nil:
			box, ok = groupOpBounds(item.group, items, areas)
		case item.op != nil:
			box, ok = bitmapOpBounds(item.op)
		default:
			ok = true
		}
		if !ok {
			return image.Rectangle{}, false
		}
		union = union.Union(box)
	}
	return union, true
}

func bitmapOpBounds(op *layout.DisplayOp) (image.Rectangle, bool) {
	switch op.Kind {
	case layout.DisplayOpNoop, layout.DisplayOpLinkURI:
		return image.Rectangle{}, true
	case layout.DisplayOpFillRect, layout.DisplayOpImage:
		return bitmapMappedRect(op, op.X, op.Y, op.X+op.W, op.Y+op.H, 0)
	case layout.DisplayOpStrokeRect:
		grow := math.Max(op.Width, 1)/groupHalfWidthDiv + 1
		return bitmapMappedRect(op, op.X, op.Y, op.X+op.W, op.Y+op.H, grow)
	case layout.DisplayOpLine:
		x, y, w, h, width := op.PaintLineGeometry()
		grow := math.Max(width, 1)/groupHalfWidthDiv + 1
		return bitmapMappedRect(op, min(x, x+w), min(y, y+h), max(x, x+w), max(y, y+h), grow)
	case layout.DisplayOpGridRun:
		if op.Grid == nil {
			return image.Rectangle{}, true
		}
		var union image.Rectangle
		for i := range op.Grid.Segs {
			line := masking.GridLine(op, i)
			b, ok := bitmapOpBounds(&line)
			if !ok {
				return image.Rectangle{}, false
			}
			union = union.Union(b)
		}
		return union, true
	case layout.DisplayOpText, layout.DisplayOpBullet:
		if op.XformSet || op.RotateDeg != 0 || op.FakeOblique {
			return image.Rectangle{}, false
		}
		pad := op.Size * pxPerPt * textPadScale
		if strokeW, _, ok := op.TextStroke(); ok && strokeW > 0 {
			pad += strokeW * pxPerPt / groupHalfWidthDiv
		}
		return image.Rect(
			int(math.Floor(op.X*pxPerPt-pad))-groupBoundsPadPx,
			int(math.Floor(op.Y*pxPerPt-pad))-groupBoundsPadPx,
			int(math.Ceil((op.X+op.W)*pxPerPt+pad))+groupBoundsPadPx,
			int(math.Ceil(op.Y*pxPerPt+pad))+groupBoundsPadPx,
		), true
	default:
		return image.Rectangle{}, false
	}
}

func bitmapMappedRect(op *layout.DisplayOp, x0, y0, x1, y1, growPt float64) (image.Rectangle, bool) {
	if op.RotateDeg != 0 {
		return image.Rectangle{}, false
	}
	l, t, r, b := x0-growPt, y0-growPt, x1+growPt, y1+growPt
	if op.XformSet && !op.Transform().IsIdentity() {
		m := op.Transform()
		loX, loY := math.Inf(1), math.Inf(1)
		hiX, hiY := math.Inf(-1), math.Inf(-1)
		for _, p := range [4][2]float64{{l, t}, {r, t}, {r, b}, {l, b}} {
			x, y := m.Apply(p[0], p[1])
			loX, loY = math.Min(loX, x), math.Min(loY, y)
			hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
		}
		l, t, r, b = loX, loY, hiX, hiY
	}
	return image.Rect(
		int(math.Floor(l*pxPerPt))-groupBoundsPadPx,
		int(math.Floor(t*pxPerPt))-groupBoundsPadPx,
		int(math.Ceil(r*pxPerPt))+groupBoundsPadPx,
		int(math.Ceil(b*pxPerPt))+groupBoundsPadPx,
	), true
}
