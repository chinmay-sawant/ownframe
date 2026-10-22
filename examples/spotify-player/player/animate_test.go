package player

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// eqHeights returns the equalizer bar heights in display units.
func eqHeights(app *App) []float64 {
	d := app.Page().Testing().Display()
	box, ok := findBox(app.Boxes(), "eq")
	if !ok {
		return nil
	}

	bars := ownframe.Fills(d, box, accent)
	out := make([]float64, len(bars))

	for i, op := range bars {
		out[i] = op.H
	}

	return out
}

func TestTickAnimatesDisplay(t *testing.T) {
	app, voice := newFakeApp(t)
	redraw(t, app)

	laidOut := eqHeights(app)

	click(t, app, "play")
	waitAudio(t, app)

	voice.pos = voice.dur / 2

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick: %v", err)
	}

	d := app.Page().Testing().Display()
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

	heights := eqHeights(app)
	changed := false

	for i := range heights {
		if i < len(laidOut) && heights[i] != laidOut[i] {
			changed = true
		}
	}

	if !changed {
		t.Fatalf("eq heights = %v laid out = %v", heights, laidOut)
	}
}
