package main

import (
	"context"
	"math"
	"strings"
	"testing"
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
