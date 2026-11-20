//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"errors"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/atspi"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

func TestA11yActionsUsePageLoopAndRejectRemovedTargets(t *testing.T) {
	ctx := t.Context()
	p, err := page.New(page.Config{Width: 320, Height: 200,
		HTML: `<input id="name"><button id="go">Go</button><p>{{.}}</p>`})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	p.SetData(count)
	p.Handle(page.Handlers{Click: func(_ context.Context, box page.Box) error {
		if box.ID == "go" {
			count++
			p.SetData(count)
		}
		return nil
	}})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	s := newShell(ctx, p)
	t.Cleanup(s.cancel)
	t.Cleanup(s.closeA11y)
	wake := make(chan struct{}, 1)
	s.a11y.actions.Close()
	s.a11y.actions = atspi.NewActions(ctx, func() { wake <- struct{}{} })
	enqueue := func(id host.AccessibilityNodeID, kind host.AccessibilityActionKind) <-chan error {
		tree, err := p.AccessibilityTree()
		if err != nil {
			t.Fatal(err)
		}
		result := make(chan error, 1)
		go func() {
			result <- s.a11y.actions.Request(ctx, host.AccessibilityRequest{Document: tree.Document, Node: id, Kind: kind})
		}()
		select {
		case <-wake:
		case <-time.After(time.Second):
			t.Fatal("request did not wake the host")
		}
		return result
	}
	focus := enqueue("id:name", host.AccessibilityFocus)
	s.runA11yActions()
	if err := <-focus; err != nil || p.FocusID() != "name" {
		t.Fatal("focus action did not reach the page")
	}
	activate := enqueue("id:go", host.AccessibilityActivate)
	s.runA11yActions()
	if err := <-activate; err != nil || count != 1 {
		t.Fatal("activation did not reach the application handler")
	}
	removed := enqueue("id:go", host.AccessibilityActivate)
	p.SetData(count)
	if err := p.Load(ctx, `<p>Removed controls</p>`); err != nil {
		t.Fatal(err)
	}
	s.runA11yActions()
	if err := <-removed; !errors.Is(err, host.ErrAccessibilityStale) || count != 1 {
		t.Fatal("removed target reached another control")
	}
}

func TestWindowAccessibilityActionWake(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for a graphics session")
	}
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	p, err := page.New(page.Config{Width: 128, Height: 96,
		HTML: `<body style="margin:0"><button id="go" aria-label="Activate" style="width:128px;height:96px;border:0;background:{{.}}"></button></body>`})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("red")
	p.Handle(page.Handlers{Click: func(_ context.Context, _ page.Box) error { p.SetData("blue"); return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	defer sh.closeA11y()
	game, stop := paceWindow(ctx, sh, Options{})
	defer stop()
	pace, ok := game.(*paceGame)
	if !ok {
		t.Fatal("missing desktop pacing wrapper")
	}
	g := &a11yWakeGame{pace: pace, page: p, ctx: ctx, result: make(chan error, 1)}
	p.SetTick(g.tick)
	watchdog := time.AfterFunc(5*time.Second, func() {
		g.expired.Store(true)
		//nolint:staticcheck // reason: watchdog wakes an otherwise stuck test only to report failure.
		ebiten.ScheduleFrame()
	})
	defer watchdog.Stop()
	ebiten.SetWindowSize(128, 96)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil || !g.done {
		t.Fatalf("idle action did not paint: %v", g.err)
	}
	select {
	case err := <-g.result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("painted action did not complete its reply")
	}
}

type a11yWakeGame struct {
	pace      *paceGame
	page      *page.Page
	ctx       context.Context
	deadline  time.Time
	armed     bool
	submitted bool
	done      bool
	expired   atomic.Bool
	result    chan error
	err       error
}

func (g *a11yWakeGame) Layout(w, h int) (int, int) { return g.pace.Layout(w, h) }

func (g *a11yWakeGame) Update() error {
	if g.expired.Load() {
		return fmt.Errorf("accessibility action failed to wake idle window")
	}
	if g.err != nil || g.done {
		return ebiten.Termination
	}
	return g.pace.Update()
}

func (g *a11yWakeGame) tick(context.Context) error {
	if !g.deadline.IsZero() {
		g.page.WantFrameAt(g.deadline)
		g.armed = true
	}
	return nil
}

func (g *a11yWakeGame) Draw(screen *ebiten.Image) {
	g.pace.Draw(screen)
	if g.done || g.err != nil {
		return
	}
	if g.deadline.IsZero() {
		if g.pace.idle {
			g.deadline = time.Now().Add(2 * time.Second)
		}
		return
	}
	if g.armed && !g.submitted {
		if !g.pace.idle || idleSessionPixel(screen) != [4]uint32{0xffff, 0, 0, 0xffff} {
			g.err = fmt.Errorf("action fixture was not red and idle before submission")
			return
		}
		g.submitted = true
		tree, err := g.page.AccessibilityTree()
		if err != nil {
			g.err = err
			return
		}
		go func() {
			timer := time.NewTimer(100 * time.Millisecond)
			defer timer.Stop()
			select {
			case <-g.ctx.Done():
				return
			case <-timer.C:
			}
			g.result <- g.pace.shell.a11y.actions.Request(g.ctx, host.AccessibilityRequest{
				Document: tree.Document, Node: "id:go", Kind: host.AccessibilityActivate,
			})
		}()
		return
	}
	if g.submitted && idleSessionPixel(screen) == [4]uint32{0, 0, 0xffff, 0xffff} {
		if !time.Now().Before(g.deadline) {
			g.err = fmt.Errorf("action relied on the later page deadline to wake")
			return
		}
		g.done = true
	}
}
