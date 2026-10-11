package bitmap

import (
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

type paintItem struct {
	op    *layout.DisplayOp
	group *layout.DisplayGroup
}

func paintOrdered(dst *image.NRGBA, d *layout.Display, faces faceCache) error {
	items := map[*layout.DisplayGroup][]paintItem{}
	seen := map[*layout.DisplayGroup]bool{}
	for _, index := range d.Order {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		op := &d.Ops[index]
		if op.GroupBoundary() != 0 {
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
	return paintGroup(dst, nil, items, faces)
}

func paintGroup(dst *image.NRGBA, group *layout.DisplayGroup,
	items map[*layout.DisplayGroup][]paintItem, faces faceCache) error {
	for _, item := range items[group] {
		if item.op != nil {
			if err := paintOp(dst, item.op, faces); err != nil {
				return err
			}
			continue
		}
		buffer := image.NewNRGBA(dst.Bounds())
		if err := paintGroup(buffer, item.group, items, faces); err != nil {
			return err
		}
		composite(dst, buffer, item.group.Mode)
	}
	return nil
}
