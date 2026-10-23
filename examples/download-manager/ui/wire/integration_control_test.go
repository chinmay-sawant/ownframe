package wire

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
)

// TestControlFlowFixture drives pause, resume, and cancel through the page
// against the slow fixture route, then finds the cancelled row in history.
func TestControlFlowFixture(t *testing.T) {
	ctx := context.Background()
	backend := newBackend(t, Config{DataDir: t.TempDir(), Fixture: true})

	app, err := ui.New(ctx, backend)
	if err != nil {
		_ = backend.Close()
		t.Fatalf("ui.New: %v", err)
	}

	defer func() {
		if err := app.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	if backend.fixture == nil {
		t.Fatal("the fixture server did not start")
	}

	// A 512 KiB body at 1 KiB per 40 ms keeps the job running for about
	// twenty seconds, long enough to drive the controls.
	backend.fixture.Size = 512 << 10
	backend.fixture.SlowPause = 40 * time.Millisecond

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	url := backend.fixture.URL + fixture.PathSlow
	app.Page().Form().SetFormValue("url", url)
	clickVisible(t, app, "add")

	id := waitState(t, app, url, ui.StateRunning, 5*time.Second)

	clickVisible(t, app, "pause-"+id)
	waitState(t, app, url, ui.StatePaused, 5*time.Second)

	clickVisible(t, app, "resume-"+id)
	waitState(t, app, url, ui.StateRunning, 5*time.Second)

	clickVisible(t, app, "cancel-"+id)
	waitState(t, app, url, ui.StateCancelled, 10*time.Second)
}
