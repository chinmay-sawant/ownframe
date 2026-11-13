package player

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestProgressTickIsPartial(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	voice.pos = 30 * time.Second

	before := app.Page().Probe().Stats().Redraws
	beforeView := app.View()
	app.Page().Probe().TakeDirtyRects()

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	if got := app.Page().Probe().Stats().Redraws; got != before {
		t.Fatalf("progress tick redraws = %d, want %d", got, before)
	}

	// Template truth stays untouched by the tick; the retained ops carry
	// the new paint until the next real commit.
	after := app.View()
	if after.Progress != beforeView.Progress || after.Elapsed != beforeView.Elapsed || after.Remaining != beforeView.Remaining || after.Credit != beforeView.Credit {
		t.Fatalf("tick mutated view progress: %+v, want %+v", after, beforeView)
	}

	d := app.Page().Probe().Display()
	elapsed, ok := findBox(app.Boxes(), "elapsed")
	if !ok {
		t.Fatal("no elapsed box")
	}

	if op := ownframe.Text(d, elapsed); op == nil || op.Text != formatSeconds(30) {
		t.Fatalf("elapsed op = %+v, want %q", op, formatSeconds(30))
	}

	rects, ok := app.Page().Probe().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("progress tick left no dirty rects")
	}

	// A second tick at the same position is a no-op: no new damage.
	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("second Tick: %v", err)
	}

	if rects, ok := app.Page().Probe().TakeDirtyRects(); ok {
		t.Fatalf("steady tick left dirty rects: %v", rects)
	}
}

func TestSeekClickUpdatesProgressThroughRedraw(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	before := app.Page().Probe().Stats().Redraws

	click(t, app, "seek-6")

	if app.View().Progress != 70 {
		t.Fatalf("Progress = %d, want 70", app.View().Progress)
	}

	if want := time.Duration(0.7 * float64(voice.dur)); voice.pos != want {
		t.Fatalf("pos = %v, want %v", voice.pos, want)
	}

	if got := app.Page().Probe().Stats().Redraws; got != before+1 {
		t.Fatalf("seek click redraws = %d, want %d", got, before+1)
	}

	rects, ok := app.Page().Probe().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatal("seek click left no dirty rects")
	}
}

func TestNavKeepsAudioDerivedProgress(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	voice.pos = 90 * time.Second

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	// Template truth is still the pre-tick commit; retained ops moved.
	preNav := app.View()
	d := app.Page().Probe().Display()
	elapsedBox, ok := findBox(app.Boxes(), "elapsed")
	if !ok {
		t.Fatal("no elapsed box")
	}
	if op := ownframe.Text(d, elapsedBox); op == nil {
		t.Fatal("no elapsed op")
	}

	click(t, app, "nav-search")
	redraw(t, app)
	click(t, app, "nav-home")
	redraw(t, app)

	// Nav preserves view playback state, like the profile screen does.
	after := app.View()
	if after.Progress != preNav.Progress || after.Elapsed != preNav.Elapsed || after.Remaining != preNav.Remaining {
		t.Fatalf("nav changed progress: %+v, want %+v", after, preNav)
	}

	// The round trip commits the preserved values, so a follow-up Redraw
	// does not move the retained fill: template and retained ops agree.
	d = app.Page().Probe().Display()
	seek, ok := findBox(app.Boxes(), "seek")
	if !ok {
		t.Fatal("no seek box")
	}

	fill := ownframe.Fill(d, seek, accent)
	if fill == nil {
		t.Fatal("no seek fill")
	}

	beforeW := fill.W
	redraw(t, app)

	d = app.Page().Probe().Display()
	seek, ok = findBox(app.Boxes(), "seek")
	if !ok {
		t.Fatal("no seek box after redraw")
	}

	fill = ownframe.Fill(d, seek, accent)
	if fill == nil {
		t.Fatal("no seek fill after redraw")
	}

	if math.Abs(fill.W-beforeW) > 0.01 {
		t.Fatalf("redraw moved seek fill: %v -> %v", beforeW, fill.W)
	}
}
