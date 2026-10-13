package replay

import "github.com/chinmay-sawant/blinkless/layout"

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
			continue
		}

		node := attachNode(root, nodes, op.Group())
		node.items = append(node.items, paintItem{op: op})
	}

	return root
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
