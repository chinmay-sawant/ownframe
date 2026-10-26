package player

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// eqMatrix returns the transform the CSS driver wrote to one equalizer bar.
func eqMatrix(t *testing.T, app *App, id string) [6]float64 {
	t.Helper()

	box, ok := findBox(app.Boxes(), id)
	if !ok {
		t.Fatalf("no box %s", id)
	}

	fill := ownframe.Fill(app.Page().Probe().Display(), box, accent)
	if fill == nil {
		t.Fatalf("no accent fill for %s", id)
	}

	m := fill.Transform()

	return [6]float64{m.A, m.B, m.C, m.D, m.E, m.F}
}

func TestEqualizerBindsCSSMotion(t *testing.T) {
	app, _ := newFakeApp(t)
	redraw(t, app)

	stats := app.Page().Probe().Motion()
	if stats.Bound < 4 {
		t.Fatalf("Bound = %d, want at least four equalizer players", stats.Bound)
	}

	if stats.LastError != "" {
		t.Fatalf("LastError = %q", stats.LastError)
	}
}

func TestEqualizerMovesFromCSSWhilePlaying(t *testing.T) {
	ctx := context.Background()
	app, _ := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("first tick: %v", err)
	}

	before := eqMatrix(t, app, "eq-0")

	time.Sleep(40 * time.Millisecond)

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("second tick: %v", err)
	}

	if after := eqMatrix(t, app, "eq-0"); after == before {
		t.Fatalf("equalizer transform did not move: %v", before)
	}

	if stats := app.Page().Probe().Motion(); stats.Active == 0 {
		t.Fatal("no CSS player reported work while playing")
	}
}

func TestEqualizerPausesWithPlayback(t *testing.T) {
	ctx := context.Background()
	app, _ := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("tick: %v", err)
	}

	click(t, app, "play")

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("paused tick: %v", err)
	}

	before := eqMatrix(t, app, "eq-0")

	time.Sleep(40 * time.Millisecond)

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("second paused tick: %v", err)
	}

	if after := eqMatrix(t, app, "eq-0"); after != before {
		t.Fatalf("equalizer moved while paused: %v -> %v", before, after)
	}
}

// TestSeekHoverLeavesTheAudioFill guards the seek bar against hover paint.
// The bar is ten clickable zones; a hover rule on a zone would draw a green
// strip away from the playhead, and a hover rule on the fill would hide the
// accent operation the tick positions.
func TestSeekHoverLeavesTheAudioFill(t *testing.T) {
	ctx := context.Background()
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	voice.dur = 3 * time.Minute
	voice.pos = 10 * time.Second

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("tick: %v", err)
	}

	zone, ok := findBox(app.Boxes(), "seek-6")
	if !ok {
		t.Fatal("no seek-6 box")
	}

	if err := app.Page().Edit().Hover(ctx, zone.X+zone.W/2, zone.Y+zone.H/2); err != nil {
		t.Fatalf("hover: %v", err)
	}

	// The window sends a tick after the hover, and the tick restores the
	// audio-driven fill.
	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("tick after hover: %v", err)
	}

	d := app.Page().Probe().Display()
	seek, ok := findBox(app.Boxes(), "seek")
	if !ok {
		t.Fatal("no seek box")
	}

	fills := ownframe.Fills(d, seek, accent)
	if len(fills) != 1 {
		t.Fatalf("hover painted %d accent fills, want the audio fill only", len(fills))
	}

	_, _, width, _ := ownframe.BoxUnits(d, seek)
	want := width * float64(voice.pos) / float64(voice.dur)

	if math.Abs(fills[0].W-want) > 0.01 {
		t.Fatalf("seek fill width = %v, want %v", fills[0].W, want)
	}
}

func TestHoverCapturesCSSTransition(t *testing.T) {
	ctx := context.Background()
	app, _ := newFakeApp(t)
	redraw(t, app)

	box, ok := findBox(app.Boxes(), "card-0")
	if !ok {
		t.Fatal("no card-0 box")
	}

	if err := app.Page().Edit().Hover(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("hover: %v", err)
	}

	stats := app.Page().Probe().Motion()
	if stats.TransitionsBound == 0 {
		t.Fatalf("hover bound no CSS transition (skip %q, error %q)",
			stats.TransitionSkipReason, stats.TransitionError)
	}
}
