package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// flattenNode appends one node's paint order to p. A direct op joins the
// current run when it lands in the same element box and the caps still hold;
// anything else flushes the run first.
func flattenNode(d *layout.Display, n *groupNode, p *chunkPlan, next *int) {
	if n.group != nil && (!normalMode(n.group.Mode) || n.group.Mask != nil) {
		p.items = append(p.items, chunkPaint{group: n})

		return
	}

	run := make([]*layout.DisplayOp, 0, chunkMaxOps)
	box, union := -1, image.Rectangle{}

	flush := func() {
		if len(run) > 1 {
			// The chunk keeps this backing array; a fresh run starts
			// below, so no copy is needed to protect the emitted ops.
			ops := run
			p.items = append(p.items, chunkPaint{run: &chunkRun{id: *next, ops: ops}})
			*next++
			run = make([]*layout.DisplayOp, 0, chunkMaxOps)
		} else if len(run) == 1 {
			p.items = append(p.items, chunkPaint{op: run[0]})
		}

		run, box, union = run[:0], -1, image.Rectangle{}
	}

	for i := range n.items {
		it := &n.items[i]
		if it.child != nil {
			flush()
			flattenNode(d, it.child, p, next)

			continue
		}

		idx, b, ok := chunkBoxFor(d, it.op, p)
		if !ok {
			flush()
			p.items = append(p.items, chunkPaint{op: it.op})

			continue
		}

		if len(run) > 0 && (idx != box || len(run) >= chunkMaxOps || !chunkFits(union.Union(b))) {
			flush()
		}

		if len(run) == 0 {
			box = idx
		}

		union = union.Union(b)
		run = append(run, it.op)
	}

	flush()
}
