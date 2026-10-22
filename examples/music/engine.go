package music

import (
	"sync"
	"time"
)

// Result is one resolved request waiting for the frame thread.
type Result struct {
	Gen  uint64
	Clip Clip
	Err  error
}

// Engine resolves one track off the frame thread and drives one Voice.
// Want, Update, and readers may race; mu guards state. Voice.Position is
// thread-safe; the pointer swap is not.
type Engine struct {
	// Open turns encoded bytes into a Voice. Tests replace it with a fake.
	Open func(data []byte) (Voice, error)

	// Timeout caps one background resolve, so a hung request cannot leave
	// the UI loading forever. Zero means defaultTimeout.
	Timeout time.Duration

	res      Resolver
	latestMu sync.Mutex
	latest   Result

	mu       sync.Mutex
	gen      uint64
	applied  uint64
	loading  bool
	clip     Clip
	voice    Voice
	wantPlay bool
	volume   float64
	err      error
}

// NewEngine returns an engine over res, playing through NewSound.
func NewEngine(res Resolver) *Engine {
	return &Engine{
		Open:    func(data []byte) (Voice, error) { return NewSound(data) },
		Timeout: defaultTimeout,
		res:     res,
		volume:  1,
	}
}

// defaultTimeout is the resolve cap when Timeout is zero.
const defaultTimeout = 20 * time.Second

// Update applies a resolved track. Call it from the page tick; it never
// blocks. Decode runs unlocked; a newer Want discards this result.
func (e *Engine) Update() {
	e.latestMu.Lock()
	res := e.latest
	e.latestMu.Unlock()

	e.mu.Lock()
	if !e.loading || res.Gen <= e.applied {
		e.mu.Unlock()

		return
	}

	e.applied = res.Gen
	e.loading = false
	if res.Err != nil {
		e.err = res.Err
		e.mu.Unlock()

		return
	}
	e.mu.Unlock()

	voice, err := e.Open(res.Clip.Data)
	if err != nil {
		e.mu.Lock()
		if e.gen == res.Gen {
			e.err = err
		}
		e.mu.Unlock()

		return
	}

	e.mu.Lock()
	if e.gen != res.Gen {
		e.mu.Unlock()
		_ = voice.Close()

		return
	}
	e.mu.Unlock()

	e.replace(voice, res.Clip)
}
