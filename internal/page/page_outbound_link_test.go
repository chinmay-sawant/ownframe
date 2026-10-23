package page

import (
	"context"
	"net/url"
	"testing"
)

func TestClickHandsValidatedOutboundLinkToApp(t *testing.T) {
	p, err := New(Config{HTML: `<a href="https://example.test/path">Open</a>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	var got *url.URL
	p.Handle(Handlers{OpenLink: func(_ context.Context, link *url.URL) error {
		got = link
		return nil
	}})
	for i := range p.draw.Display.Ops {
		op := &p.draw.Display.Ops[i]
		if op.Kind != DisplayOpLinkURI {
			continue
		}
		x := (op.X + op.W/2) / p.draw.Display.PointsPerPixel
		y := (op.Y + op.H/2) / p.draw.Display.PointsPerPixel
		if err = p.Click(context.Background(), x, y); err != nil {
			t.Fatal(err)
		}
		break
	}
	if got == nil || got.Host != "example.test" || got.Path != "/path" {
		t.Fatalf("opened link = %#v", got)
	}
}
