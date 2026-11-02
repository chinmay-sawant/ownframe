package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// buildPlan buckets the paint order by owning group. A group's item sits in
// its parent where the group first paints, so nesting follows the engine's
// group pointers rather than marker positions.
func buildPlan(d *layout.Display) *groupNode {
	root := &groupNode{}
	nodes := map[*layout.DisplayGroup]*groupNode{}

	for _, i := range d.Order {
		if i < 0 || i >= len(d.Ops) {
			continue
		}

		op := &d.Ops[i]
		if op.GroupBoundary() != 0 {
			if op.IsGroupBegin() && op.Group() != nil {
				group := op.Group()
				if group.Mask != nil || group.BackdropFilter != "" {
					node := attachNode(root, nodes, group)
					if group.Mask != nil {
						node.maskBox = masking.GroupBounds(op, d)
						node.hasMask = true
					}
					if group.BackdropFilter != "" {
						node.backdropBox = masking.GroupBounds(op, d)
						node.backdropMark = op
						node.hasBackdropBox = true
						node.backdrop, node.hasBackdrop = parseBackdropFilter(group.BackdropFilter)
					}
				}
			}
			continue
		}

		node := attachNode(root, nodes, op.Group())
		node.items = append(node.items, paintItem{op: op})
	}
	setBackdropRadii(root, d.PixelPerPoint)

	return root
}

func setBackdropRadii(node *groupNode, pixelsPerPoint float64) {
	if node.group != nil && node.hasBackdrop {
		for i := range node.items {
			item := &node.items[i]
			if item.op == nil || item.op.Kind != layout.DisplayOpFillRect {
				continue
			}

			op := item.op
			radius := max(op.Radius, op.RadiusTopLeft, op.RadiusTopRight,
				op.RadiusBottomRight, op.RadiusBottomLeft, op.RadiusTopLeftY,
				op.RadiusTopRightY, op.RadiusBottomRightY, op.RadiusBottomLeftY)
			node.backdropRadius = radius * pixelsPerPoint
			break
		}
	}

	for i := range node.items {
		if child := node.items[i].child; child != nil {
			setBackdropRadii(child, pixelsPerPoint)
		}
	}
}

// attachNode finds or creates the node for g and attaches it, with any
// missing ancestor, to its parent's items.
func attachNode(root *groupNode, nodes map[*layout.DisplayGroup]*groupNode, g *layout.DisplayGroup) *groupNode {
	if g == nil {
		return root
	}

	if node := nodes[g]; node != nil {
		return node
	}

	node := &groupNode{group: g}
	nodes[g] = node

	parent := attachNode(root, nodes, g.Parent)
	parent.items = append(parent.items, paintItem{child: node})

	return node
}
