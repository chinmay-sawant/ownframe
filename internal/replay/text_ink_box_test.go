package replay

import (
	"context"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestTextInkBoxMatchesTheShapedRun checks the run-local ink box against the
// shaped advance and the face metrics, and that a run with no glyph image
// reports no ink.
func TestTextInkBoxMatchesTheShapedRun(t *testing.T) {
	t.Parallel()

	d, err := render.DisplayList(context.Background(),
		`<html><body style="margin:0;font-size:16px"><p>Hamburgefonstiv</p></body></html>`,
		240, 80)
	if err != nil {
		t.Fatal(err)
	}

	op := firstTextOp(d)
	if op == nil {
		t.Fatal("no text op")
	}

	face := textFace(op)
	if face == nil {
		t.Fatal("no face")
	}

	plain, ok := textInkBox(face, "Hamburgefonstiv", 0)
	if !ok {
		t.Fatal("a plain run has no ink box")
	}

	ascent := face.Metrics().HAscent
	width, _ := text.Measure("Hamburgefonstiv", face, 0)
	if plain.Min.X < -2 || float64(plain.Max.X) > width+2 {
		t.Fatalf("ink box %v is wider than the advance %.2f", plain, width)
	}
	if float64(plain.Min.Y) >= ascent || float64(plain.Max.Y) <= ascent {
		t.Fatalf("ink box %v does not straddle the baseline at %.2f", plain, ascent)
	}

	spaced, ok := textInkBox(face, "Hamburgefonstiv", 6)
	if !ok || spaced.Dx() <= plain.Dx()+60 {
		t.Fatalf("letter spacing did not widen the box: %v against %v", spaced, plain)
	}

	if _, ok := textInkBox(face, "", 0); ok {
		t.Fatal("an empty run has an ink box")
	}
}
