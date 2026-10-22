package collector

import "time"

// Dummy process simulation tuning: spawn mix, CPU and memory ranges, state
// chances, and table churn. The same seed always replays the same table.
const (
	// Spawn: retired-identity reuse chance, CPU cap, memory base and span,
	// and the thread and nice ranges.
	reuseChance   = 0.35
	cpuCapNS      = 3_600_000_000_000
	procMemBase   = 8 << 20
	procMemSpan   = 400 << 20
	maxNewThreads = 24
	niceSpan      = 10
	// Evolve: CPU delta range, settle and spike chances, memory jitter and
	// clamp, IO chances and spans, and the running chance.
	evolveSpanNS  = 300_000
	settleChance  = 0.15
	settleBase    = time.Millisecond
	settleExtra   = 8 * time.Millisecond
	spikeEvolve   = 0.9
	spikeEvolveDT = 20 * time.Millisecond
	procJitter    = 2 << 20
	procJitterMid = 1 << 20
	procMemFloor  = 1 << 20
	procMemCeil   = 2 << 30
	readChance    = 0.1
	readSpan      = 4 << 20
	writeSpan     = 1 << 20
	runChance     = 0.4
	// Churn: exits per advance, the table floor, and the retire cap.
	exitSpan  = 3
	procFloor = 16
	retireCap = 32
	// Detail: virtual memory is resident times this factor.
	virtMult = 4
)
