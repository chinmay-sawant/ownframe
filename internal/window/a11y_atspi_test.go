//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestLinuxA11yHasPublisher(t *testing.T) {
	s := newShell(context.Background(), &fakeScreen{width: 10, height: 10})
	if s.a11y.pub == nil {
		t.Fatal("publisher")
	}
}

type accessibleFakeScreen struct {
	fakeScreen
	tree host.AccessibilityTree
}

func (a *accessibleFakeScreen) AccessibilityTree() (host.AccessibilityTree, error) {
	return a.tree, nil
}

func (a *accessibleFakeScreen) AccessibilityPerform(context.Context, host.AccessibilityRequest) error {
	return nil
}

// waitA11yWorker waits for every queued snapshot to finish publishing, so
// the test reads the fake bus after the worker stopped writing it.
func waitA11yWorker(t *testing.T, s *shell) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		s.a11y.mu.Lock()
		queued := s.a11y.queued
		s.a11y.mu.Unlock()
		if queued == 0 {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("a11y worker did not drain")
		}
		time.Sleep(time.Millisecond)
	}
}

func TestLinuxA11yPublishesWhenLive(t *testing.T) {
	fakeBus := atspi.NewFakeBus()
	scr := &accessibleFakeScreen{
		fakeScreen: fakeScreen{width: 10, height: 10},
		tree: host.AccessibilityTree{
			Document: 1,
			Revision: 1,
			Root:     "id:root",
			Nodes: []host.AccessibilityNode{{
				ID:   "id:root",
				Role: host.AccessibilityButton,
			}},
		},
	}
	s := newShell(context.Background(), scr)
	s.a11yOrigin = func() (int, int) { return 0, 0 }
	s.a11y.pub = atspi.NewPublisher(fakeBus)
	if !s.a11y.pub.Live() {
		t.Fatal("expected publisher to be live")
	}
	s.syncA11y()
	waitA11yWorker(t, s)
	if len(fakeBus.Apps()) != 1 {
		t.Fatalf("expected 1 published app, got %d", len(fakeBus.Apps()))
	}
	s.syncA11y()
	waitA11yWorker(t, s)
	if len(fakeBus.Apps()) != 1 {
		t.Fatalf("unchanged generation republished: %d", len(fakeBus.Apps()))
	}
	scr.tree.Revision++
	scr.redraws++
	s.syncA11y()
	waitA11yWorker(t, s)
	if len(fakeBus.Apps()) != 2 {
		t.Fatalf("changed generation did not publish: %d", len(fakeBus.Apps()))
	}
	s.closeA11y()
	if s.a11y.pub.Live() {
		t.Fatal("publisher stayed live after close")
	}
	s.syncA11y()
	if len(fakeBus.Apps()) != 2 {
		t.Fatalf("closed publisher published again: %d", len(fakeBus.Apps()))
	}
}

func TestLinuxA11yWatcherClosesPublisher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	s := newShell(ctx, &fakeScreen{width: 10, height: 10})
	bus := atspi.NewFakeBus()
	s.a11y.pub = atspi.NewPublisher(bus)
	cancel()

	deadline := time.Now().Add(2 * time.Second)
	for {
		s.a11y.mu.Lock()
		shutdown := s.a11y.shutdown
		s.a11y.mu.Unlock()
		if shutdown {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("cancelled shell did not close the publisher")
		}
		time.Sleep(time.Millisecond)
	}
	if bus.Live() {
		t.Fatal("publisher stayed live after shell cancellation")
	}
}

func TestLinuxA11yCloseIsIdempotent(t *testing.T) {
	s := newShell(context.Background(), &fakeScreen{width: 10, height: 10})
	bus := atspi.NewFakeBus()
	s.a11y.pub = atspi.NewPublisher(bus)
	s.closeA11y()
	s.closeA11y()
}

type blockingA11yBus struct {
	started   chan struct{}
	release   chan struct{}
	mu        sync.Mutex
	publishes int
}

func newBlockingA11yBus() *blockingA11yBus {
	return &blockingA11yBus{started: make(chan struct{}, 1), release: make(chan struct{})}
}

func (b *blockingA11yBus) Claim(string) error { return nil }
func (b *blockingA11yBus) Live() bool         { return true }
func (b *blockingA11yBus) Close() error       { return nil }

func (b *blockingA11yBus) Status() atspi.Status {
	return atspi.Status{Platform: "fake", Live: true}
}

func (b *blockingA11yBus) Publish(atspi.Application) error {
	b.mu.Lock()
	b.publishes++
	b.mu.Unlock()
	select {
	case b.started <- struct{}{}:
	default:
	}
	<-b.release
	return nil
}

func (b *blockingA11yBus) Emit(atspi.Event) error { return nil }

func TestLinuxA11yPublishDoesNotBlockTheLoop(t *testing.T) {
	scr := &accessibleFakeScreen{
		fakeScreen: fakeScreen{width: 10, height: 10},
		tree: host.AccessibilityTree{
			Document: 1,
			Revision: 1,
			Root:     "id:root",
			Nodes: []host.AccessibilityNode{{
				ID: "id:root", Role: host.AccessibilityDocument,
			}},
		},
	}
	bus := newBlockingA11yBus()
	s := newShell(context.Background(), scr)
	s.a11yOrigin = func() (int, int) { return 0, 0 }
	s.a11y.pub = atspi.NewPublisher(bus)

	done := make(chan struct{})
	go func() {
		s.syncA11y()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("syncA11y waited for the bus")
	}
	select {
	case <-bus.started:
	case <-time.After(time.Second):
		t.Fatal("worker never reached Publish")
	}
	close(bus.release)
	s.closeA11y()
}

type retryA11yBus struct {
	mu        sync.Mutex
	ready     bool
	fail      bool
	publishes int
}

func (b *retryA11yBus) Claim(string) error { return nil }
func (b *retryA11yBus) Live() bool         { return true }
func (b *retryA11yBus) Close() error       { return nil }

func (b *retryA11yBus) Status() atspi.Status {
	return atspi.Status{Platform: "fake", Live: true}
}

func (b *retryA11yBus) Ready() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.ready
}

func (b *retryA11yBus) Publish(atspi.Application) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.publishes++
	if b.fail {
		b.fail = false
		b.ready = false
		return errors.New("embed failed")
	}
	return nil
}

func (b *retryA11yBus) Emit(atspi.Event) error { return nil }

func (b *retryA11yBus) setReady(ready bool) {
	b.mu.Lock()
	b.ready = ready
	b.mu.Unlock()
}

func (b *retryA11yBus) count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.publishes
}

func TestLinuxA11yRetriesOnlyWhenBusIsReady(t *testing.T) {
	scr := &accessibleFakeScreen{
		fakeScreen: fakeScreen{width: 10, height: 10},
		tree: host.AccessibilityTree{
			Document: 1,
			Revision: 1,
			Root:     "id:root",
			Nodes: []host.AccessibilityNode{{
				ID: "id:root", Role: host.AccessibilityDocument,
			}},
		},
	}
	bus := &retryA11yBus{ready: true, fail: true}
	s := newShell(context.Background(), scr)
	s.a11yOrigin = func() (int, int) { return 0, 0 }
	s.a11y.pub = atspi.NewPublisher(bus)

	s.syncA11y()
	waitA11yWorker(t, s)
	if bus.count() != 1 {
		t.Fatalf("first publish calls = %d", bus.count())
	}
	s.syncA11y()
	waitA11yWorker(t, s)
	if bus.count() != 1 {
		t.Fatalf("retried during backoff: %d calls", bus.count())
	}
	bus.setReady(true)
	s.syncA11y()
	waitA11yWorker(t, s)
	if bus.count() != 2 {
		t.Fatalf("ready bus did not retry: %d calls", bus.count())
	}
	s.closeA11y()
}
