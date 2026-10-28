package web

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// tickScreen is a screen that reports damage on every tick.
type tickScreen struct {
	host.Screen
	ticks   int
	redraws int
	ticked  chan struct{}
	redrew  chan struct{}
}

func (f *tickScreen) Tick(context.Context) error {
	f.ticks++
	select {
	case f.ticked <- struct{}{}:
	default:
	}

	return nil
}

func (f *tickScreen) FrameDirty() bool { return true }

func (f *tickScreen) Redraw(context.Context) error {
	f.redraws++
	select {
	case f.redrew <- struct{}{}:
	default:
	}

	return nil
}

func (f *tickScreen) Generation() uint64 { return 0 }

func TestTickLoopRedrawsDirtyScreen(t *testing.T) {
	t.Parallel()

	app := &tickScreen{ticked: make(chan struct{}, 1), redrew: make(chan struct{}, 1)}
	srv := &server{app: app}
	ctx, cancel := context.WithCancel(t.Context())
	stop := srv.startTick(ctx)
	t.Cleanup(func() {
		cancel()
		stop()
	})

	select {
	case <-app.redrew:
	case <-time.After(2 * time.Second):
		t.Fatal("tick loop never redrew a dirty screen")
	}

	cancel()
	stop()

	if app.ticks == 0 || app.redraws == 0 {
		t.Fatalf("ticks = %d, redraws = %d", app.ticks, app.redraws)
	}
	if srv.ticking {
		t.Fatal("server still reports ticking after stop")
	}
	if srv.frameGen == 0 {
		t.Fatal("dirty tick did not advance the frame generation")
	}
}

type panicTickScreen struct {
	host.Screen
	panicked chan struct{}
}

func (p *panicTickScreen) Tick(context.Context) error {
	select {
	case p.panicked <- struct{}{}:
	default:
	}
	panic("tick failure test")
}

func (p *panicTickScreen) Generation() uint64 { return 0 }
func (p *panicTickScreen) Redraw(context.Context) error { return nil }

func TestTickLoopPanicRecovery(t *testing.T) {
	t.Parallel()

	app := &panicTickScreen{panicked: make(chan struct{}, 1)}
	srv := &server{app: app}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	stop := srv.startTick(ctx)
	defer stop()

	select {
	case <-app.panicked:
	case <-time.After(2 * time.Second):
		t.Fatal("tick never ran")
	}

	time.Sleep(50 * time.Millisecond)

	srv.mu.Lock()
	ticking := srv.ticking
	srv.mu.Unlock()

	if ticking {
		t.Fatal("server still reports ticking after panic")
	}
}
