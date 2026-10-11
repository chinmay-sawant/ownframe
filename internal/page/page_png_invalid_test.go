package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestPNGRejectsUnsupportedRetainedPaint(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body style="background:red">text</body>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	if d == nil || len(d.Ops) == 0 {
		t.Fatal("fixture has no replay operations")
	}
	d.Ops[0].Kind = layout.DisplayKind(100)
	if png := p.PNG(); png != nil {
		t.Fatal("unsupported retained paint was replaced by source repaint")
	}
}
