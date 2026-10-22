package flappy

// phase is the game state.
type phase int

const (
	ready phase = iota
	running
	over
)

// Scene geometry, in CSS pixels.
const (
	sceneW  = 480
	sceneH  = 720
	groundY = 624
	birdX   = 130
	birdR   = 16

	pipeW = 62
	capW  = 74
	capH  = 26
)

// Game tuning.
const (
	gravity      = 1700.0
	flapV        = -520.0
	maxFall      = 780.0
	speedStart   = 150.0
	speedMax     = 260.0
	spacingStart = 300.0
	spacingMin   = 230.0
	gapStart     = 190.0
	gapMin       = 150.0
	margin       = 70.0
)

// pipe is one pair of pipes moving left.
type pipe struct {
	x      float64 // left edge of the body
	gapY   float64 // centre of the gap
	gap    float64 // gap height
	scored bool
}

// game is the simulation the tick advances.
type game struct {
	phase    phase
	score    int
	best     int
	birdY    float64 // centre of the bird
	vy       float64
	flapAge  float64 // seconds since the last flap
	distance float64
	speed    float64 // CSS pixels per second
	nextPipe float64 // pixels of travel until the next pair
	pipes    []pipe
	run      float64 // seconds, drives the bob
}

// newGame returns the ready screen the window opens on.
func newGame() game {
	return game{
		phase:    ready,
		birdY:    sceneH / halfDiv,
		speed:    speedStart,
		nextPipe: spacingStart,
	}
}
