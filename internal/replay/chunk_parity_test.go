package replay

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// checkChunkParity requires byte-identical direct and chunked draws.
// Translucent ops stay direct.
func checkChunkParity() error {
	d, err := cardParityPage()
	if err != nil {
		return err
	}

	p := chunkPlanFor(d)
	if p == nil {
		return fmt.Errorf("the card page has no chunk plan")
	}

	if _, runs, _ := planStats(p); runs == 0 {
		return fmt.Errorf("the card page folded no run")
	}

	if runHolds(p, true) {
		return fmt.Errorf("a run holds a translucent op")
	}

	direct, chunked := ebiten.NewImage(640, 200), ebiten.NewImage(640, 200)
	defer direct.Dispose()
	defer chunked.Dispose()

	for _, i := range d.Order {
		drawOp(direct, &d.Ops[i], 0, 0)
	}

	before := chunkEntries()
	DrawVisible(chunked, d, 0, 0)
	baked := chunkEntries()
	if baked == before {
		return fmt.Errorf("the chunked draw baked no chunk")
	}

	DrawVisible(chunked, d, 0, 0)
	if chunkEntries() != baked {
		return fmt.Errorf("the second chunked draw re-baked a chunk")
	}

	a, b := make([]byte, 640*200*4), make([]byte, 640*200*4)
	direct.ReadPixels(a)
	chunked.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("chunked visible draw differs in %d channels: %s",
			countDiff(a, b), diffPixels(a, b, 640))
	}

	return chunkPartialParity(d, a)
}

// runHolds reports whether a run holds an op with opacity under one.
func runHolds(p *chunkPlan, want bool) bool {
	for i := range p.items {
		if p.items[i].run == nil {
			continue
		}

		for _, op := range p.items[i].run.ops {
			if (op.Opacity() < 1) == want {
				return true
			}
		}
	}

	return false
}

// diffPixels describes the first differing pixels for the failure message.
func diffPixels(a, b []byte, width int) string {
	out := ""

	for i := 0; i+4 <= len(a) && len(out) < 400; i += 4 {
		if bytes.Equal(a[i:i+4], b[i:i+4]) {
			continue
		}

		p := i / 4
		out += fmt.Sprintf("(%d,%d) %v vs %v ", p%width, p/width, a[i:i+4], b[i:i+4])
	}

	return out
}
