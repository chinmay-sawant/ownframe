//go:build linux && !android && !ios && !js

package window

import (
	"github.com/chinmay-sawant/ownframe/internal/a11yqueue"
	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/hajimehoshi/ebiten/v2"
	"sync"
)

// a11yHost holds the Linux AT-SPI publisher. The render loop snapshots a
// changed tree and hands it to one worker, so a slow D-Bus call never
// stalls Update.
type a11yHost struct {
	actions   *a11yqueue.Actions
	pub       *atspi.Publisher
	gen       uint64
	published bool
	retry     bool

	mu        sync.Mutex
	queued    int
	running   bool
	closed    bool
	shutdown  bool
	jobs      chan a11yJob
	done      chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

type a11yJob struct {
	tree host.AccessibilityTree
	gen  uint64
}

func (s *shell) initA11y() {
	s.a11y.actions = a11yqueue.NewActions(s.ctx, func() {
		//nolint:staticcheck // reason: concurrent-safe wake for an AT request while the desktop is parked in WaitEvents.
		ebiten.ScheduleFrame()
	})
	s.a11y.pub = atspi.NewPublisher(atspi.NewLiveBus(s.a11y.actions.Request))
	s.a11y.jobs = make(chan a11yJob, 1)
	s.a11y.done = make(chan struct{})
	if s.ctxDone != nil {
		go s.watchA11y(s.ctxDone)
	}
}

// watchA11y closes the publisher when the shell context ends, which
// covers NewGame hosts that never call runWindow. It also exits when
// closeA11y runs on its own, so window tests do not leak it.
func (s *shell) watchA11y(ctxDone <-chan struct{}) {
	select {
	case <-ctxDone:
		s.closeA11y()
	case <-s.a11y.done:
	}
}

func (s *shell) closeA11y() {
	a := &s.a11y
	a.closeOnce.Do(func() {
		if a.actions != nil {
			a.actions.Close()
		}
		a.mu.Lock()
		a.closed = true
		a.mu.Unlock()
		if a.done != nil {
			close(a.done)
		}
		a.wg.Wait()
		if a.pub != nil {
			_ = a.pub.Close()
		}
		a.mu.Lock()
		a.shutdown = true
		a.mu.Unlock()
	})
}

func (s *shell) runA11yActions() {
	if s.a11y.actions == nil {
		return
	}
	if acc, ok := s.app.(host.Accessible); ok {
		s.a11y.actions.Drain(acc.AccessibilityPerform)
	}
}

func (s *shell) syncA11y() {
	if s.a11y.pub == nil {
		return
	}
	// Coordinates do not touch D-Bus. Keep them current on the render
	// loop so moving the window updates the root extents even when the
	// page tree itself did not change.
	x, y := 0, 0
	if s.a11yOrigin != nil {
		x, y = s.a11yOrigin()
	}
	s.a11y.pub.SetOrigin(x, y)
	s.a11y.pub.SetSize(s.screenW, s.screenH)
	if !s.a11y.pub.Ready() {
		return
	}
	s.publishA11y()
}

func (s *shell) publishA11y() {
	a := &s.a11y
	if a.pub == nil {
		return
	}
	acc, ok := s.app.(host.Accessible)
	if !ok {
		return
	}
	gen := s.app.Generation()
	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return
	}
	if a.published && !a.retry && a.gen == gen {
		a.mu.Unlock()
		return
	}
	a.mu.Unlock()

	tree, err := acc.AccessibilityTree()
	if err != nil {
		return
	}
	job := a11yJob{tree: tree, gen: gen}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}
	// The queue holds one slot and the newest tree must win: drop a stale
	// queued snapshot, then queue this one. a.mu serializes producers, so
	// the send cannot block after the drain.
	select {
	case <-a.jobs:
		if a.queued > 0 {
			a.queued--
		}
	default:
	}
	select {
	case a.jobs <- job:
		a.queued++
	default:
		return
	}
	if !a.running {
		a.running = true
		a.wg.Add(1)
		go s.workA11y()
	}
}

// workA11y publishes queued snapshots and exits once the queue is empty.
// A shell that is never closed therefore does not leave an idle worker
// behind, which keeps the window test leak check honest.
func (s *shell) workA11y() {
	a := &s.a11y
	defer a.wg.Done()
	for {
		select {
		case <-a.done:
			a.mu.Lock()
			a.running = false
			a.mu.Unlock()
			return
		case job := <-a.jobs:
			s.applyA11y(job)
		}
		a.mu.Lock()
		if len(a.jobs) == 0 {
			a.running = false
			a.mu.Unlock()
			return
		}
		a.mu.Unlock()
	}
}

func (s *shell) applyA11y(job a11yJob) {
	a := &s.a11y
	a.mu.Lock()
	closed := a.closed
	a.mu.Unlock()
	if closed {
		return
	}
	if !a.pub.Ready() {
		s.recordA11y(job.gen, false)
		return
	}
	err := a.pub.Publish(job.tree)
	s.recordA11y(job.gen, err == nil)
}

func (s *shell) recordA11y(gen uint64, ok bool) {
	a := &s.a11y
	a.mu.Lock()
	if a.queued > 0 {
		a.queued--
	}
	if ok {
		a.retry = false
		a.published = true
		a.gen = gen
	} else {
		a.retry = true
	}
	a.mu.Unlock()
}
