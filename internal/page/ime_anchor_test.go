package page

import (
	"context"
	"testing"
)

func TestIMEAnchorTracksGlyphCaret(t *testing.T) {
	t.Parallel()
	p := imeField(t, "WWWW")
	p.setCaret(1, false)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	first, ok := p.IMEAnchor()
	if !ok {
		t.Fatal("missing caret anchor")
	}
	p.setCaret(3, false)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	second, ok := p.IMEAnchor()
	if !ok || second.X <= first.X {
		t.Fatalf("caret x positions = %v then %v", first.X, second.X)
	}
	field := p.boxByID("e")
	if first.X <= field.X || second.X >= field.X+field.W {
		t.Fatalf("caret anchors outside field: first=%+v second=%+v field=%+v", first, second, field)
	}
}
