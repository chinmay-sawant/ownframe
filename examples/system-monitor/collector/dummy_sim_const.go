package collector

// Dummy host simulation tuning. The same seed and the same call sequence
// always produce the same counters, so these constants are the documented
// fixture behavior, not guesses.
const (
	// CPU load model: base plus jitter, with a rotating spike.
	loadBase    = 0.06
	loadJitter  = 0.12
	spikeStride = 4
	spikeParity = 2
	spikeBase   = 0.82
	spikeJitter = 0.1
	driftCenter = 0.5
	driftAmp    = 0.3
	loadFloor   = 0.05
	loadCeil    = 16
	spikeEvery  = 37
	spikeWidth  = 4
	// Memory walk: jitter range and center, then the clamp bounds.
	memStepRange = 2 << 27
	memStepHalf  = 1 << 27
	memLow       = 3 << 30
	memHigh      = 30 << 30
	swapHigh     = 6 << 30
	swapChance   = 0.3
	// Disk model: IO spans, the burst every diskBurstEvery ticks, and the
	// free-space floor.
	diskWriteSpan  = 1 << 22
	diskBurstEvery = 23
	diskBurst      = 384 << 20
	diskReadSpan   = 1 << 24
	diskWobble     = 128 << 20
	diskWobbleBase = 64 << 20
	diskFreeFloor  = 10 << 30
	// Network model: IO spans, the burst every netBurstEvery ticks, the USB
	// tick, and the error rate.
	netSpan       = 1 << 20
	netBurstEvery = 19
	netBurst      = 96 << 20
	netTxSpan     = 1 << 18
	usbAtTick     = 30
	netErrRate    = 0.004
)

// Initial host state: core times, memory, disks, and boot offset.
const (
	coreBusyStep   = 100
	coreBusyOffset = 3
	coreTotalBase  = 100_000
	memTotalInit   = 32 << 30
	memUsedInit    = 12 << 30
	swapInit       = 512 << 20
	diskSdaTotal   = 512 << 30
	diskSdaFree    = 220 << 30
	diskNvmeTotal  = 1 << 40
	diskNvmeFree   = 700 << 30
	bootBackHours  = 72
)

// Sample rendering: the mem-available quarter divisor, the swap total, and
// the temperature baseline and span.
const (
	memQuarterDiv = 4
	swapTotal     = 8 << 30
	tempBase      = 38
	tempSpan      = 25
)
