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
	faces faceCache,
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
		// One full-canvas buffer per group level, capped at 7 levels above,
		// on the cold bitmap fallback only. No pool: shared scratch would
		// need lifetime and zeroing rules across recursive depths.
		buffer := image.NewNRGBA(dst.Bounds())
		if err := paintGroup(buffer, item.group, items, areas, radii, faces); err != nil {
			return err
		}
		if item.group.Mask != nil {
			area, ok := areas[item.group]
			if !ok {
				return fmt.Errorf("bitmap: masked group has no border box")
			}
			if err := masking.Apply(buffer, area, item.group.Mask); err != nil {
				return err
			}
		}
		composite(dst, buffer, item.group.Mode)
	}
	return nil
}
