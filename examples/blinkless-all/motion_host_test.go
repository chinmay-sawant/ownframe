package main

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func wantNear(t *testing.T, label string, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 0.6 {
		t.Errorf("%s = %.2f, want %.2f", label, got, want)
	}
}

// TestMotionReadings reads the matrices the engine gave the lanes. Each box is
// 28px, so its center starts 14px from its own corner.
func TestMotionReadings(t *testing.T) {
	r, err := readMotion(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	for _, l := range motionLanes {
		if !r[l.ID].Moved {
			t.Errorf("%s: the engine gave the box no matrix", l.ID)
		}
	}

	d := func(id string) (float64, float64) {
		return r[id].CenterX - r[id].StartedX, r[id].CenterY - r[id].StartedY
	}

	dx0, _ := d("mp-dist-0")
	dx50, dy50 := d("mp-dist-50")
	dx100, _ := d("mp-dist-100")

	wantNear(t, "0% dx", dx0, -14)
	wantNear(t, "50% dx", dx50, 86)
	wantNear(t, "50% dy", dy50, -14)
	wantNear(t, "100% dx", dx100, 186)

	wantNear(t, "auto turn", r["mp-rot-auto"].TurnDeg, 90)
	wantNear(t, "reverse turn", r["mp-rot-reverse"].TurnDeg, -90)
	wantNear(t, "fixed turn", r["mp-rot-fixed"].TurnDeg, 45)

	cx, cy := d("mp-anchor-center")
	kx, ky := d("mp-anchor-corner")

	wantNear(t, "center-anchored dx", cx, 86)
	wantNear(t, "center-anchored dy", cy, -4)
	wantNear(t, "corner-anchored dx", kx, 100)
	wantNear(t, "corner-anchored dy", ky, 10)

	rx, ry := d("mp-ray-origin")
	wantNear(t, "ray dx", rx, 86)
	wantNear(t, "ray dy", ry, 16)

	sx, sy := d("mp-short")
	wantNear(t, "shorthand dx", sx, 50)
	wantNear(t, "shorthand dy", sy, 0)
}

// TestMotionBodyShowsReadout checks the tab section carries the lanes and the
// numbers the host read back.
func TestMotionBodyShowsReadout(t *testing.T) {
	body := ""
	for _, p := range motionTab().Props {
		body += p.Demo()
	}

	for _, l := range motionLanes {
		if !strings.Contains(body, `id="`+l.ID+`"`) || !strings.Contains(body, `id="`+l.ID+`-read"`) {
			t.Errorf("body lacks lane %s or its readout", l.ID)
		}
	}

	if !strings.Contains(body, "turns 90 deg") {
		t.Error("readout does not show the 90 degree turn")
	}
}

// TestMotionLiveBoxMoves: the session steps offset-distance with the clock, so
// the box's matrix carries it further along the line each frame.
func TestMotionLiveBoxMoves(t *testing.T) {
	ctx := context.Background()

	app, err := NewApp(catalogPath())
	if err != nil {
		t.Fatal(err)
	}

	page, err := app.NewPage(ctx, "motion", 900, 900)
	if err != nil {
		t.Fatal(err)
	}

	clock := &manualLiveClock{t: time.Unix(1_700_000_000, 0).Truncate(motionLoop)}
	session := newMotionSession()
	runner := &liveRunner{page: page, sessions: []liveSession{session}, clock: clock,
		src: liveSource{cursor: func() (float64, float64) { return 0, 0 }, click: func() bool { return false },
			down: func() bool { return false }, keys: func() []string { return nil }, text: func() string { return "" }}}
	runner.attach()

	var seen []float64

	for range 3 {
		clock.t = clock.t.Add(motionLoop / 4)

		if err := page.Testing().Tick(ctx); err != nil {
			t.Fatal(err)
		}

		seen = append(seen, liveBoxShift(t, page))
	}

	if !(seen[0] < seen[1] && seen[1] < seen[2]) {
		t.Errorf("box did not move steadily along the line: %v", seen)
	}

	if session.patches == 0 {
		t.Error("session wrote nothing")
	}
}

// liveBoxShift is how far, in CSS pixels, the matrix on the live box's ops
// carries it from where layout put it. The ops hold the painted matrix.
func liveBoxShift(t *testing.T, page *ownframe.Page) float64 {
	t.Helper()

	display := page.Probe().Display()

	for index, op := range display.Ops {
		if index < len(display.OpOwners) && display.OpOwners[index] >= 0 &&
			display.Boxes[display.OpOwners[index]].ID == motionLiveID && op.XformSet {
			return op.Transform().E / display.PointsPerPixel
		}
	}

	t.Fatal("no moved op for the live box")

	return 0
}
