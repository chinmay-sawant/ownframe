package main

import (
	"context"
	"math"
	"testing"
)

// These tests read the page the window paints, not the readout text. The
// readout comes from a separate layout, so a bug that only the painted page
// has (an image the page never registered, a heading the demo never drew)
// passes every readout check.

// TestPaintedImageThresholdMovesText checks the painted shape-image-threshold
// page: the 0.5 box starts its text 40px in and the 0 box 80px in.
func TestPaintedImageThresholdMovesText(t *testing.T) {
	app := newTestApp(t)

	page, err := app.NewPage(context.Background(), "shapes", 1000, 800)
	if err != nil {
		t.Fatal(err)
	}

	d := page.Probe().Display()
	ppp := d.PointsPerPixel

	for id, want := range map[string]float64{"si-low": 80, "si-high": 40} {
		got := shapesTextX(d, id)/ppp - shapesCellLeft(d, id)
		if math.Abs(got-want) > 1 {
			t.Errorf("painted %s text starts %.1fpx from the float's left edge, want %.0f", id, got, want)
		}
	}
}

// TestPaintedCopyIntoDrawsHeading checks that the copy-into row paints the
// heading its description says still paints, above the running header.
func TestPaintedCopyIntoDrawsHeading(t *testing.T) {
	app := newTestApp(t)

	page, err := app.NewPage(context.Background(), "frag", 1000, 800)
	if err != nil {
		t.Fatal(err)
	}

	var heading, header float64

	for _, b := range page.Probe().Display().Boxes {
		switch b.ID {
		case "frag-heading":
			heading = b.Y
			if b.H <= 0 {
				t.Error("copy-into heading has no height")
			}
		case "frag-running-header":
			header = b.Y
		}
	}

	if heading == 0 || header == 0 {
		t.Fatalf("painted page lacks the heading (y=%v) or the running header (y=%v)", heading, header)
	}

	if heading >= header {
		t.Errorf("heading at y=%v is not above the running header at y=%v", heading, header)
	}
}
