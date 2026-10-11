package window

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"testing"
)

func TestBitmapViewportReusesGPUBufferUntilResize(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body style="margin:0"><div style="height:400px;background:red;` +
		`border-radius:50% / 25%"></div></body>`, Width: 100, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	s := &shell{app: p, fallback: true, screenW: 100, screenH: 100}
	if err := s.prepareBitmapViewport(); err != nil {
		t.Fatal(err)
	}
	first := s.bitmapView.img
	if first == nil {
		t.Fatal("no viewport bitmap")
	}
	s.scrollY = 50
	if err := s.prepareBitmapViewport(); err != nil {
		t.Fatal(err)
	}
	if s.bitmapView.img != first {
		t.Fatal("scroll recreated GPU image")
	}
	s.screenW = 120
	p.SetSize(120, 100)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := s.prepareBitmapViewport(); err != nil {
		t.Fatal(err)
	}
	if s.bitmapView.img == first {
		t.Fatal("resize retained wrong buffer size")
	}
	s.bitmapView.img.Dispose()
}
