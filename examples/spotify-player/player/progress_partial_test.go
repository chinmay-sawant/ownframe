package player

import (
	"context"
	"testing"
	"time"
)

func TestProgressTickIsPartial(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	voice.pos = 30 * time.Second

	before := app.Page().Stats().Redraws
	app.Page().Testing().TakeDirtyRects()

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	if got := app.Page().Stats().Redraws; got != before {
		t.Fatalf("progress tick redraws = %d, want %d", got, before)
	}

	if app.View().Elapsed != formatSeconds(30) {
		t.Fatalf("Elapsed = %q, want %q", app.View().Elapsed, formatSeconds(30))
	}

	rects, ok := app.Page().Testing().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("progress tick left no dirty rects")
	}
}

func TestSeekClickUpdatesProgressThroughRedraw(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	before := app.Page().Stats().Redraws

	click(t, app, "seek-6")

	if app.View().Progress != 70 {
		t.Fatalf("Progress = %d, want 70", app.View().Progress)
	}

	if want := time.Duration(0.7 * float64(voice.dur)); voice.pos != want {
		t.Fatalf("pos = %v, want %v", voice.pos, want)
	}

	if got := app.Page().Stats().Redraws; got != before+1 {
		t.Fatalf("seek click redraws = %d, want %d", got, before+1)
	}

	rects, ok := app.Page().Testing().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatal("seek click left no dirty rects")
	}
}
