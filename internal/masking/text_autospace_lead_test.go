package masking_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// spacedTextOp returns a real text op that carries text-autospace, so the
// unexported autospace field is set the way layout sets it.
func spacedTextOp(t *testing.T) layout.DisplayOp {
	t.Helper()
	d, err := render.DisplayList(context.Background(),
		`<body style="margin:0"><div style="font-size:18px;text-autospace:normal">abc</div></body>`, 200, 60)
	if err != nil {
		t.Fatal(err)
	}
	for _, op := range d.Ops {
		if op.Kind == layout.DisplayOpText && op.Text != "" {
			return op
		}
	}
	t.Fatal("no text op")

	return layout.DisplayOp{}
}

// TestLeadingAutospaceGapAtSplitBoundary: layout emits one text op per font,
// so a boundary can fall between two ops. The gap is then owed to the second.
func TestLeadingAutospaceGapAtSplitBoundary(t *testing.T) {
	t.Parallel()
	cur := spacedTextOp(t)
	want := cur.TextAutospaceGap()
	if want <= 0 {
		t.Fatal("op carries no autospace gap")
	}
	prev := cur
	prev.Text = "日本語"
	prev.X, prev.W = 7.5, 40.5
	cur.Text = "abc"
	cur.X = prev.X + prev.W
	cur.Y = prev.Y

	if got := masking.LeadingAutospaceGap(&prev, &cur); got != want {
		t.Fatalf("boundary between ops: gap %v, want %v", got, want)
	}

	apart := cur
	apart.X += 10
	if got := masking.LeadingAutospaceGap(&prev, &apart); got != 0 {
		t.Fatalf("ops that do not abut: gap %v, want 0", got)
	}

	otherLine := cur
	otherLine.Y += 20
	if got := masking.LeadingAutospaceGap(&prev, &otherLine); got != 0 {
		t.Fatalf("ops on different baselines: gap %v, want 0", got)
	}

	latin := prev
	latin.Text = "xyz"
	if got := masking.LeadingAutospaceGap(&latin, &cur); got != 0 {
		t.Fatalf("latin next to latin: gap %v, want 0", got)
	}

	if got := masking.LeadingAutospaceGap(nil, &cur); got != 0 {
		t.Fatalf("no previous op: gap %v, want 0", got)
	}
}
