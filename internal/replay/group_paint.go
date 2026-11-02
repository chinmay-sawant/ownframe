package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// opPicker decides whether an op paints in this call and where. ok false
// skips the op; a group paints when any member is picked.
type opPicker func(op *layout.DisplayOp) (dx, dy float64, ok bool)

// paintDisplay replays display in paint order, compositing each non-normal
// group once. A page with no groups takes the direct path and allocates
// nothing.
func (renderer *Renderer) paintDisplay(dst *ebiten.Image, d *layout.Display, pick opPicker) {
	if dst == nil || d == nil {
		return
	}

	if !hasGroups(d) {
		for _, i := range d.Order {
			if i < 0 || i >= len(d.Ops) {
				continue
			}

			if dx, dy, ok := pick(&d.Ops[i]); ok {
				renderer.drawOp(dst, &d.Ops[i], dx, dy)
			}
		}

		return
	}

	groupPick := func(op *layout.DisplayOp) (float64, float64, bool) {
		dx, dy, ok := pick(op)
		if ok && op.Fixed {
			// Fixed ops keep their viewport position. Group layers apply a
			// second, local offset when they draw into the offscreen image.
			dx, dy = 0, 0
		}
		return dx, dy, ok
	}
	renderer.paintNode(dst, d, renderer.planFor(d), groupPick, false)
}

// paintNode paints one node: a non-normal group buffers its members into a
// layer and composites once; every other node paints item by item.
func (renderer *Renderer) paintNode(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker, local bool) {
	if n.group != nil && (!normalMode(n.group.Mode) || n.group.Mask != nil || n.group.BackdropFilter != "") {
		renderer.paintGroup(dst, d, n, pick)

		return
	}

	renderer.paintItems(dst, d, n, pick, local)
}

// paintItems paints a node's ops in order, recursing into child groups.
func (renderer *Renderer) paintItems(dst *ebiten.Image, d *layout.Display, n *groupNode, pick opPicker, local bool) {
	for i := range n.items {
		it := &n.items[i]
		if it.child != nil {
			renderer.paintNode(dst, d, it.child, pick, local)

			continue
		}

		if dx, dy, ok := pick(it.op); ok {
			if local {
				renderer.drawOpInLayer(dst, it.op, dx, dy)
			} else {
				renderer.drawOp(dst, it.op, dx, dy)
			}
		}
	}
}
