package collector

import (
	"math/rand"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/system-monitor/domain"
)

// DummyOptions configures the reproducible dummy source. The same seed and
// the same sequence of calls always produce the same counters.
type DummyOptions struct {
	Seed      int64
	Processes int
	// Stamp returns the current stamp. Nil uses wall and process time.
	Stamp func() domain.Stamp
}

// Dummy is the labeled fixture source. It reads no host state, so the example
// runs without permissions, and a fixture sample never mixes with live
// readings. Every call advances the simulation one tick; the mutex lets the
// summary and process loops call it concurrently.
type Dummy struct {
	opts DummyOptions
	mu   sync.Mutex
	rng  *rand.Rand

	tick     int64
	lastMono time.Duration
	boot     time.Time
	spike    bool

	cores    int
	cpu      []domain.CPUTimes
	load     [3]float64
	memTotal uint64
	memUsed  uint64
	swapUsed uint64
	disks    []dmDisk
	nets     []dmNet

	procs     []*dmProc
	retired   []domain.ProcessIdentity
	nextPID   int32
	nextStart uint64
}

// NewDummy returns a fixture source.
func NewDummy(opts DummyOptions) *Dummy {
	if opts.Processes <= 0 {
		opts.Processes = DefaultDummyProcesses
	}
	if opts.Stamp == nil {
		start := time.Now()
		opts.Stamp = func() domain.Stamp {
			return domain.Stamp{At: time.Now(), Mono: time.Since(start)}
		}
	}

	d := &Dummy{opts: opts, rng: rand.New(rand.NewSource(opts.Seed))} //nolint:gosec // reason: the fixture requires deterministic math/rand output for a reproducible simulation, not cryptographic randomness.
	d.init()

	return d
}

// NewStress returns the 10,000 process fixture used by stress tests.
func NewStress(seed int64) *Dummy {
	return NewDummy(DummyOptions{Seed: seed, Processes: StressProcesses})
}

// Name labels the source.
func (d *Dummy) Name() string { return "dummy" }
