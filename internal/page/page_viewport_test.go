package page_test

import (
	"context"
	"testing"
)

func TestMediaWidthFollowsTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}` +
		`#a{width:40px;height:30px}` +
		`@media (min-width: 700px){#a{width:120px}}` +
		`</style></head><body><div id="a"></div></body></html>`

	p := newViewportPage(t, source, 800, 600)
	if got := viewportBox(t, p, "a").W; got != 120 {
		t.Fatalf("wide width = %.0f, want 120", got)
	}

	p.SetSize(400, 300)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if got := viewportBox(t, p, "a").W; got != 40 {
		t.Fatalf("narrow width = %.0f, want 40", got)
	}
}

func TestMediaHeightFollowsTheFrame(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>body{margin:0}#a{height:30px}` +
		`@media (min-height:500px){#a{height:90px}}` +
		`@media (max-height:400px){#a{height:60px}}` +
		`</style></head><body><div id="a"></div></body></html>`
	p := newViewportPage(t, source, 800, 499)
	for _, test := range []struct{ height, want int }{{499, 30}, {500, 90}, {501, 90}, {500, 90}, {499, 30}, {401, 30}, {400, 60}, {399, 60}, {400, 60}, {401, 30}, {600, 90}} {
		before := p.Stats()
		p.SetSize(800, test.height)
		if err := p.Redraw(t.Context()); err != nil {
			t.Fatal(err)
		}
		if got := viewportBox(t, p, "a").H; got != float64(test.want) {
			t.Fatalf("viewport height %d: box %g want %d", test.height, got, test.want)
		}
		after := p.Stats()
		if after.Parses != before.Parses || after.Cascades != before.Cascades {
			t.Fatal("height resize reparsed or recascaded document")
		}
	}

}
