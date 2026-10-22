package dino

import "math/rand/v2"

// phase is the game state.
type phase int

const (
	ready phase = iota
	running
	over
)

// obstacle kinds.
const (
	cactusSmall = iota
	cactusBig
	bird
)

// Scene geometry, in CSS pixels.
const (
	sceneW  = 900
	sceneH  = 300
	groundY = 250
	dinoX   = 70
	slotMax = 4
)

// obstacle is one cactus or bird moving left.
type obstacle struct {
	kind   int
	x      float64 // left edge
	bottom float64 // bottom above the ground line
	w, h   float64
	flap   float64
}

// game is the simulation the tick advances.
type game struct {
	phase     phase
	high      int
	score     int
	speed     float64 // CSS pixels per second
	distance  float64
	nextSpawn float64 // pixels of travel until the next obstacle
	feet      float64 // dinosaur feet above the ground line
	vy        float64 // upward speed
	onGround  bool
	ducking   bool
	jumpHeld  bool
	run       float64 // seconds, drives the legs
	obstacles []obstacle
}

func newGame() game {
	return game{
		phase:     ready,
		speed:     startSpeed,
		nextSpawn: startGap,
		onGround:  true,
	}
}

// step advances the simulation by dt seconds.
func (g *game) step(dt float64, rng *rand.Rand) {
	if g.phase != running {
		g.run += dt

		return
	}

	g.run += dt
	g.distance += g.speed * dt
	g.score = int(g.distance / pixelsPerPoint)
	g.speed = min(maxSpeed, startSpeed+g.distance*speedGain)

	g.nextSpawn -= g.speed * dt
	if g.nextSpawn <= 0 {
		g.spawn(rng)
		g.nextSpawn = g.gap(rng)
	}

	g.move(dt)
	g.fall(dt)

	if g.hits() {
		g.phase = over
		g.high = max(g.high, g.score)
	}
}
