package bitmap

import (
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

type paintItem struct {
	op    *layout.DisplayOp
	group *layout.DisplayGroup
}

func paintOrdered(dst *image.NRGBA, d *layout.Display, faces faceCache) error {
	items := map[*layout.DisplayGroup][]paintItem{}
	seen := map[*layout.DisplayGroup]bool{}
	areas := map[*layout.DisplayGroup]image.Rectangle{}
	radii := map[*layout.DisplayGroup]float64{}
	for _, index := range d.Order {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		op := &d.Ops[index]
		if op.GroupBoundary() != 0 {
			if op.IsGroupBegin() && op.Group() != nil {
				group := op.Group()
				if group.Mask != nil || group.BackdropFilter != "" {
					areas[group] = masking.GroupBounds(op, d)
				}
				if group.BackdropFilter != "" {
					if err := attachGroups(group, items, seen); err != nil {
						return err
					}
				}
			}
			continue
		}
		group := op.Group()
		if err := attachGroups(group, items, seen); err != nil {
			return err
		}
		items[group] = append(items[group], paintItem{op: op})
		if group != nil && group.BackdropFilter != "" {
			if _, found := radii[group]; !found &&
				(op.Kind == layout.DisplayOpFillRect || op.Kind == layout.DisplayOpStrokeRect) {
				radii[group] = max(
					op.Radius, op.RadiusTopLeft, op.RadiusTopRight,
					op.RadiusBottomRight, op.RadiusBottomLeft, op.RadiusTopLeftY,
					op.RadiusTopRightY, op.RadiusBottomRightY, op.RadiusBottomLeftY,
				) * d.PixelPerPoint
			}
		}
	}
	return paintGroup(dst, nil, items, areas, radii, faces)
}

func attachGroups(group *layout.DisplayGroup,
	items map[*layout.DisplayGroup][]paintItem,
	seen map[*layout.DisplayGroup]bool,
) error {
	depth := 0
	for current := group; current != nil; current = current.Parent {
		depth++
		if depth > maxBlendDepth {
			return fmt.Errorf("bitmap: blend nesting exceeds 7")
		}
		if !seen[current] {
			items[current.Parent] = append(items[current.Parent], paintItem{group: current})
			seen[current] = true
		}
	}
	return nil
}
