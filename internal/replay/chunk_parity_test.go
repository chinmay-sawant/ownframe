package replay

import (
	"bytes"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// checkChunkParity draws the card page direct and chunked and requires
// byte-identical pixels, then repeats with a partial rect over one card. The
// opaque cards fold into runs; the translucent card stays direct, because a
// run with a translucent op cannot promise byte parity.
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
