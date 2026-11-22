package bitmap

import (
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/backdrop"
	"github.com/chinmay-sawant/ownframe/internal/masking"
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
		// mask area, so allocating only area ∩ dst.Bounds() avoids O(N*W*H)
		// full-canvas buffers and composites when many elements carry masks.
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
