package main

import "testing"

func TestBandHeights(t *testing.T) {
	a, _ := newTestApp(t)
	want := map[string]float64{
		"xl-titlebar":    titleBarHeight,
		"xl-ribbon-slot": ribbonHeightFull,
		"xl-fbar":        formulaBarHeight,
		"xl-tabs":        sheetBarHeight,
		"xl-status":      statusBarHeight,
	}
	for id, h := range want {
		if b := mustBox(t, a, id); b.H != h {
			t.Errorf("%s height %v, want %v", id, b.H, h)
		}
	}
}
