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
