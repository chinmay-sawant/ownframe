package replay

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"image"
	"math"
	"testing"
)

func TestOpBoundsTextUsesAscent(t *testing.T) {
	t.Parallel()

	display, err := render.DisplayListState(context.Background(),
		`<html><body><p>Hello</p></body></html>`, 200, 100, render.State{})
	if err != nil {
		t.Fatal(err)
	}

	op := firstTextOp(display)
	if op == nil {
		t.Fatal("no text op")
	}

	ppt := display.PointsPerPixel

	box, ok := opBounds(op, ppt)
	if !ok {
		t.Fatal("text op unbounded")
	}

	if box.Min.Y >= int(op.Y/ppt) {
		t.Fatalf("box %v starts at or below the baseline", box)
	}

	below := image.Rect(box.Min.X, box.Max.Y+2, box.Max.X, box.Max.Y+9)
	if opTouches(op, ppt, below) {
		t.Fatal("a rect below the line box touched the run")
	}

	above := image.Rect(box.Min.X, box.Min.Y-5, box.Max.X, box.Min.Y+1)
	if !opTouches(op, ppt, above) {
		t.Fatal("a rect on the ascent did not touch the run")
	}
}

// TestOpBoundsTextWithoutPayload checks that a hand-built text op, which never
// attached a payload, still bounds as its line box. The scan reads the matrix
// field directly, so the payload must be bound before that read.
func TestOpBoundsTextWithoutPayload(t *testing.T) {
	t.Parallel()

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpText, X: 20, Y: 40, W: 60, H: 16, InkDescent: 4, Size: 12,
	}

	got, ok := opBounds(op, 1)
	if !ok {
		t.Fatal("text op unbounded")
	}

	want := image.Rect(19, 27, 81, 45)
	if got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
}

// TestOpBoundsProjectedTextMoves checks that the projection branch still wins
// for a run whose baseline the engine already projected: the matrix box is far
// from the plain line box, and the plain box is culled.
func TestOpBoundsProjectedTextMoves(t *testing.T) {
	t.Parallel()

	plain := &layout.DisplayOp{
		Kind: layout.DisplayOpText, X: 20, Y: 40, W: 60, H: 16, InkDescent: 4, Size: 12,
	}
	matrix := layout.Matrix2D{A: 2, D: 2}

	projected := *plain
	projected.SetXform(matrix)
	projected.XformSet = false
	projected.X, projected.Y = matrix.Apply(plain.X, plain.Y)

	if box, ok := opBounds(&projected, 1); !ok || box != image.Rect(39, 55, 161, 89) {
		t.Fatalf("projected bounds = %v, %v", box, ok)
	}

	if opTouches(&projected, 1, image.Rect(21, 27, 79, 43)) {
		t.Fatal("the plain line box touched a projected run")
	}

	if !opTouches(&projected, 1, image.Rect(41, 56, 119, 88)) {
		t.Fatal("the matrix box did not touch the projected run")
	}

	if box, ok := opBounds(plain, 1); !ok || box != image.Rect(19, 27, 81, 45) {
		t.Fatalf("plain bounds = %v, %v", box, ok)
	}
}

func firstTextOp(display *layout.Display) *layout.DisplayOp {
	for i := range display.Ops {
		if display.Ops[i].Kind == layout.DisplayOpText && display.Ops[i].Font != nil {
			return &display.Ops[i]
		}
	}

	return nil
}

func TestTextFacesStayBoundedAcrossSizeChanges(t *testing.T) {
	d, err := render.DisplayList(context.Background(), `<p>resize</p>`, 200, 200)
	if err != nil {
		t.Fatal(err)
	}
	var op layout.DisplayOp
	for _, candidate := range d.Ops {
		if candidate.Font != nil {
			op = candidate
			break
		}
	}
	if op.Font == nil {
		t.Fatal("no text font")
	}
	for i := range 400 {
		op.Size = 10 + float64(i)/10
		if textFace(&op) == nil {
			t.Fatal("face did not load")
		}
	}
	fonts.mu.Lock()
	defer fonts.mu.Unlock()
	if len(fonts.faces) > 256 || len(fonts.sources) > 64 {
		t.Fatal("unbounded font cache")
	}
}

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

func phaseOf(x float64) float64 {
	return textPhase(x - math.Floor(x))
}

// TestTextPhaseAbsorbsFloatNoise locks the sub-pixel grid. Adding whole
// pixels to an origin rounds its low bits, and a key that treated that as a
// new offset would rebake a scrolled list on every step.
func TestTextPhaseAbsorbsFloatNoise(t *testing.T) {
	t.Parallel()

	base := 12.25 * pxPerPt
	for shift := 1; shift <= 64; shift++ {
		if phaseOf(base) != phaseOf(base+float64(shift)) {
			t.Fatalf("phase moved by %d whole pixels", shift)
		}
	}
}

// TestTextRasterLayoutPadsAndCaps checks the image a bake would allocate: the
// ink box grows by the pad on every side, and a run past a cap is declined.
func TestTextRasterLayoutPadsAndCaps(t *testing.T) {
	t.Parallel()

	size, dx, dy, ok := textRasterLayout(image.Rect(20, 30, 120, 50))
	if !ok || dx != 16 || dy != 26 || size != image.Pt(108, 28) {
		t.Fatalf("layout = %v %d %d %v", size, dx, dy, ok)
	}

	if _, _, _, ok := textRasterLayout(image.Rect(0, 0, textRasterMaxSide+1, 20)); ok {
		t.Fatal("an oversize wide run was cacheable")
	}
	if _, _, _, ok := textRasterLayout(image.Rect(0, 0, 20, textRasterMaxSide+1)); ok {
		t.Fatal("an oversize tall run was cacheable")
	}
}

// Compare the engine's projected-baseline contract with a known local run
// transformed through the same matrix. This runs inside the GPU parity loop.
func checkProjectedText() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}
	var renderer Renderer
	defer renderer.Close()
	actual, expected := newParityImage(320, 240), newParityImage(320, 240)
	defer actual.Deallocate()
	defer expected.Deallocate()
	for _, matrix := range []layout.Matrix2D{
		{A: 0.5, D: 1, E: 10},
		{A: 1, D: 0.5, F: 10},
		{A: 1.2, D: 1.2},
	} {
		local := ops[0]
		local.X, local.Y = 30, 40
		local.SetXform(matrix)
		local.XformSet = true
		projected := local
		projected.X, projected.Y = matrix.Apply(local.X, local.Y)
		projected.XformSet = false
		actual.Clear()
		expected.Clear()
		renderer.drawOp(actual, &projected, 0, 0)
		if projected.Transform() != matrix {
			return fmt.Errorf("replaying projected text mutated the display matrix")
		}
		renderer.drawOp(expected, &local, 0, 0)
		a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
		actual.ReadPixels(a)
		expected.ReadPixels(b)
		if !bytes.Equal(a, b) {
			return fmt.Errorf("projected text differs from transformed local text for %+v", matrix)
		}
		got, _ := opBounds(&projected, 1/pxPerPt)
		want, _ := opBounds(&local, 1/pxPerPt)
		if got != want {
			return fmt.Errorf("projected text bounds = %v, want %v", got, want)
		}
	}
	return nil
}

func TestValidateUserFont(t *testing.T) {
	t.Parallel()
	if err := ValidateUserFont(nil); err == nil {
		t.Fatal("empty allowed")
	}
	if err := ValidateUserFont(make([]byte, MaxUserFontBytes+1)); err == nil {
		t.Fatal("oversize allowed")
	}
	if err := ValidateUserFont([]byte("not a font")); err == nil {
		t.Fatal("garbage allowed")
	}
}

func TestGlyphAt(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{
			input: "Hello, world!",
			want:  []string{"H", "e", "l", "l", "o", ",", " ", "w", "o", "r", "l", "d", "!"},
		},
		{
			input: "Привет мир", // Cyrillic multi-byte
			want:  []string{"П", "р", "и", "в", "е", "т", " ", "м", "и", "р"},
		},
		{
			input: "こんにちは", // Japanese 3-byte
			want:  []string{"こ", "ん", "に", "ち", "は"},
		},
		{
			input: "A🚀B", // 4-byte emoji
			want:  []string{"A", "🚀", "B"},
		},
	}

	for _, tc := range cases {
		var got []string
		for i := 0; i < len(tc.input); {
			g, size := glyphAt(tc.input, i)
			got = append(got, g)
			i += size
		}
		if len(got) != len(tc.want) {
			t.Fatalf("for %q: got %d glyphs, want %d", tc.input, len(got), len(tc.want))
		}
		for idx := range got {
			if got[idx] != tc.want[idx] {
				t.Errorf("for %q at %d: got %q, want %q", tc.input, idx, got[idx], tc.want[idx])
			}
		}
	}
}

func TestGlyphAt_ASCIIZeroAllocs(t *testing.T) {
	str := "The quick brown fox jumps over the lazy dog"
	allocs := testing.AllocsPerRun(100, func() {
		for i := 0; i < len(str); {
			g, size := glyphAt(str, i)
			_ = g
			i += size
		}
	})
	if allocs > 0 {
		t.Fatalf("expected 0 allocs for glyphAt, got %f", allocs)
	}
}

// TestQuarterTurnTextBoundsCoverSoftwareInk holds replay's bounds for a
// vertical-writing run to the ink the software painter draws, and checks the
// run now replays instead of forcing the bitmap fallback.
func TestQuarterTurnTextBoundsCoverSoftwareInk(t *testing.T) {
	t.Parallel()
	source := `<body style="margin:0;background:#fff"><div style="position:absolute;left:60px;top:20px;` +
		`width:60px;height:150px;color:#000;font-size:18px;writing-mode:vertical-rl">Vertical text</div></body>`
	d, err := render.DisplayList(context.Background(), source, 240, 200)
	if err != nil {
		t.Fatal(err)
	}
	if !render.Replayable(d) {
		t.Fatal("vertical writing should replay on the display list")
	}
	var run *layout.DisplayOp
	for i := range d.Ops {
		if d.Ops[i].Kind == layout.DisplayOpText && d.Ops[i].RotateDeg != 0 {
			run = &d.Ops[i]
		}
	}
	if run == nil {
		t.Fatal("the fixture emitted no rotated text run")
	}
	box, ok := opBounds(run, d.PointsPerPixel)
	if !ok {
		t.Fatal("a quarter-turned run must be boundable")
	}
	img, err := bitmap.Paint(d)
	if err != nil {
		t.Fatal(err)
	}
	var ink image.Rectangle
	for y := 0; y < 200; y++ {
		for x := 0; x < 240; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			if r < 0xfa00 || g < 0xfa00 || b < 0xfa00 {
				ink = ink.Union(image.Rect(x, y, x+1, y+1))
			}
		}
	}
	if ink.Empty() || !ink.In(box) {
		t.Fatalf("software ink %v falls outside the replay bounds %v", ink, box)
	}
}

// TestQuarterTurnTextRejectsUnsupportedRotation keeps the software painter
// for a rotation the display list cannot replay exactly.
func TestQuarterTurnTextRejectsUnsupportedRotation(t *testing.T) {
	t.Parallel()
	ops, err := textParityOps()
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		name   string
		mutate func(*layout.DisplayOp)
		want   bool
	}{
		{"quarter", func(op *layout.DisplayOp) { op.RotateDeg = 90 }, true},
		{"negative quarter", func(op *layout.DisplayOp) { op.RotateDeg = -90 }, true},
		{"half", func(op *layout.DisplayOp) { op.RotateDeg = 180 }, true},
		{"oblique angle", func(op *layout.DisplayOp) { op.RotateDeg = 30 }, false},
		{"fake oblique", func(op *layout.DisplayOp) { op.RotateDeg = 90; op.FakeOblique = true }, false},
		{"baked transform", func(op *layout.DisplayOp) {
			op.RotateDeg = 90
			op.SetXform(layout.Matrix2D{A: 1, D: 1, E: 3})
			op.XformSet = true
		}, false},
	} {
		op := ops[0]
		c.mutate(&op)
		if _, got := textProjection(&op); got != c.want {
			t.Fatalf("%s: textProjection ok = %t, want %t", c.name, got, c.want)
		}
	}
}
