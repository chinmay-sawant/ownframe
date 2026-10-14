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
	for _, index := range d.Order {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		op := &d.Ops[index]
		if op.GroupBoundary() != 0 {
			if op.IsGroupBegin() && op.Group() != nil && op.Group().Mask != nil {
				areas[op.Group()] = masking.GroupBounds(op, d)
			}
			continue
		}
		group := op.Group()
		depth := 0
		for g := group; g != nil; g = g.Parent {
			depth++
			if depth > 7 {
				return fmt.Errorf("bitmap: blend nesting exceeds 7")
			}
			if !seen[g] {
				items[g.Parent] = append(items[g.Parent], paintItem{group: g})
				seen[g] = true
			}
		}
		items[group] = append(items[group], paintItem{op: op})
	}
	return paintGroup(dst, nil, items, areas, faces)
}

func paintGroup(dst *image.NRGBA, group *layout.DisplayGroup,
	items map[*layout.DisplayGroup][]paintItem,
	areas map[*layout.DisplayGroup]image.Rectangle, faces faceCache) error {
	for _, item := range items[group] {
		if item.op != nil {
			if err := paintOp(dst, item.op, faces); err != nil {
				return err
			}
			continue
		}
		buffer := image.NewNRGBA(dst.Bounds())
		if err := paintGroup(buffer, item.group, items, areas, faces); err != nil {
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
