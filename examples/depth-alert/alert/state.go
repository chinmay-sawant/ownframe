package alert

// State is the haptic state derived from a fresh depth observation.
type State int32

const (
	Unknown State = iota
	Clear
	Stop
	StopNow
)

const (
	MinimumCoveragePercent = 15
	StopDistanceMM         = 3000
	ClearDistanceMM        = 3500
	StopNowDistanceMM      = 1200
	ClearFramesRequired    = 12
)

// Observation contains a valid distance summary for the center view.
type Observation struct {
	NearestMM int32
	Coverage  int32
	Tracking  bool
}

// Analyzer requires repeated clear observations before it reports Clear.
type Analyzer struct {
	state       State
	clearFrames int
}

// State returns the most recent decision without changing the history.
func (a *Analyzer) State() State { return a.state }

func (s State) String() string {
	switch s {
	case Clear:
		return "WALK"
	case Stop:
		return "STOP"
	case StopNow:
		return "STOP NOW"
	default:
		return "DEPTH UNKNOWN"
	}
}

// Observe classifies valid observations and requires repeated clear frames.
func (a *Analyzer) Observe(o Observation) State {
	if !o.Tracking || o.Coverage < MinimumCoveragePercent || o.NearestMM <= 0 {
		a.clearFrames = 0
		if a.state != Stop && a.state != StopNow {
			a.state = Unknown
		}
		return a.state
	}
	if o.NearestMM <= StopNowDistanceMM {
		a.state = StopNow
		a.clearFrames = 0
		return a.state
	}
	if o.NearestMM <= StopDistanceMM {
		a.state = Stop
		a.clearFrames = 0
		return a.state
	}
	if o.NearestMM < ClearDistanceMM {
		if a.state == Clear {
			a.state = Unknown
		}
		a.clearFrames = 0
		return a.state
	}
	a.clearFrames++
	if a.clearFrames >= ClearFramesRequired {
		a.state = Clear
	}
	return a.state
}

// MarkUnavailable reports that no fresh observation arrived before the deadline.
func (a *Analyzer) MarkUnavailable() {
	a.state = Unknown
	a.clearFrames = 0
}

// Reset discards prior observations before scanning starts or resumes.
func (a *Analyzer) Reset() { *a = Analyzer{} }
