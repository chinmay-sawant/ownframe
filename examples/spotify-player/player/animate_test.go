package player

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTickAnimatesDisplay(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	click(t, app, "play")
	waitAudio(t, app)

	voice.pos = voice.dur / 2

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	d := app.Page().Probe().Display()
	seek, ok := findBox(app.Boxes(), "seek")
	if !ok {
		t.Fatal("no seek box")
	}

	fill := ownframe.Fill(d, seek, accent)
	if fill == nil || fill.W <= 0 {
		t.Fatalf("seek fill = %+v", fill)
	}

	_, _, w, _ := ownframe.BoxUnits(d, seek)
	if want := w / 2; math.Abs(fill.W-want) > 0.01 {
		t.Fatalf("fill.W = %v want %v", fill.W, want)
	}

	elapsed, ok := findBox(app.Boxes(), "elapsed")
	if !ok {
		t.Fatal("no elapsed box")
	}

	if op := ownframe.Text(d, elapsed); op == nil || op.Text != formatSeconds(int(voice.pos.Seconds())) {
		t.Fatalf("elapsed = %+v", op)
	}
}
