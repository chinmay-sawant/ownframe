package replay

import "sync/atomic"

// replayCounts count the work a paint actually did: single ops drawn and
// chunk images blitted. A benchmark turns the gate on; shipped frames pay
// one atomic load per draw.
var replayCounts struct {
	on    atomic.Bool
	ops   atomic.Int64
	blits atomic.Int64
}

func countOp() {
	if replayCounts.on.Load() {
		replayCounts.ops.Add(1)
	}
}

func countBlit() {
	if replayCounts.on.Load() {
		replayCounts.blits.Add(1)
	}
}
