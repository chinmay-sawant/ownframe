package wire

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
)

// TestEndToEndFixture runs a real HTTP transfer against the local fixture
// service: add a URL, wait for completion, and find the row in history.
func TestEndToEndFixture(t *testing.T) {
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

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	url := backend.fixture.URL + fixture.PathOK
	app.Page().Form().SetFormValue("url", url)
	clickVisible(t, app, "add")

	waitHistory(t, app, url, ui.StateCompleted, 15*time.Second)
}
