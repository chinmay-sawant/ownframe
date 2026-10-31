//go:build linux && !android && !ios && !js

package window

import "github.com/chinmay-sawant/ownframe/internal/host"

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
	select {
	case a.jobs <- job:
		a.queued++
	default:
		// The previous frame is still being published; that snapshot
		// already carries a tree at least as new as the one replaced.
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
