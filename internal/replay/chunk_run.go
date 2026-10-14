package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// pickRun returns the shared translation of the run's picked ops. picked is
// false when none is; shared is false when two picked ops disagree, which no
// painter the plan feeds does.
func pickRun(c *chunkRun, pick opPicker) (tx, ty float64, picked, shared bool) {
	shared = true

	for _, op := range c.ops {
		dx, dy, ok := pick(op)
		if !ok {
			continue
		}

		if picked && (dx != tx || dy != ty) {
			shared = false
		}

		tx, ty, picked = dx, dy, true
	}

	return tx, ty, picked, shared
}

// runOpaque reports a run whose every op is fully opaque. A tick can lower an
// op's paint opacity after the plan was built, so the run rechecks it.
func runOpaque(c *chunkRun) bool {
	for _, op := range c.ops {
		if op.Opacity() < 1 {
			return false
		}
	}

	return true
}

// drawRunDirect paints the picked ops one by one.
func drawRunDirect(dst *ebiten.Image, c *chunkRun, pick opPicker) {
	for _, op := range c.ops {
		if dx, dy, ok := pick(op); ok {
			countOp()
			drawOp(dst, op, dx, dy)
		}
	}
}
