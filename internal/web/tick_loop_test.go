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
