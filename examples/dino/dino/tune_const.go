package dino

// Collision boxes stay a little inside the sprites, so near misses count
// as misses.
const (
	dinoBoxInset = 4
	duckBoxW     = 46
	duckBoxH     = 26
	standBoxW    = 38
	standBoxH    = 48
	obInset      = 3
	obShrink     = 6
)

// Run tuning in CSS pixels, points, and seconds.
const (
	startSpeed     = 330
	startGap       = 640
	maxSpeed       = 880
	pixelsPerPoint = 18
	speedGain      = 0.008
	diveVelocity   = -520
	fpsRound       = 0.5
	maxFrameDT     = 0.1
	heldGravity    = 1500
	fallGravity    = 2500
)

// Sky starts in CSS pixels.
const (
	cloudAX     = 620
	cloudAY     = 54
	cloudBX     = 300
	cloudBY     = 86
	pebbleStart = 40
	pebbleStep  = 140
)

// Pose sizes in CSS pixels and the leg-cycle timing.
const (
	dinoStandH   = 52
	legLong      = 8
	legShort     = 4
	stridePeriod = 0.09
	strideCycle  = 2
	duckHeight   = 30
)

// Sky drift and obstacle spawn tuning.
const (
	cloudDrift    = 0.14
	cloudRespawn  = 40
	cloudGap      = 60
	pebbleRespawn = 20
	pebbleGap     = 13
	gapBase       = 0.62
	gapVar        = 0.55
	bigCactusAt   = 0.6
	spawnAhead    = 20
	evenChance    = 0.5
	flapPeriod    = 0.16
	flapCycle     = 2
	opsPerCloud   = 3
	maxFPS        = 999
)
