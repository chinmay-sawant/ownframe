package animation

import "time"

// Player applies animation effects to retained display operations.
type Player struct {
	targets []Target
	effects []Effect
	states  []effectState
	start   time.Time
	started bool
	active  bool
	touched bool
}
