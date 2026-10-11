package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestNativeLandscapeStillScrollsVertically(t *testing.T) {
	ctx := context.Background()
	p, err := page.New(page.Config{
		HTML:  `<body style="margin:0"><div style="height:1400px;background:red"></div></body>`,
		Width: 420, Height: 934, MinWidth: 320, MinHeight: 480,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(ctx, p); err != nil {
		t.Fatal(err)
	}
	s := NewGame(ctx, p).(*shell)
	s.Layout(885, 420)
	if err := s.resize(); err != nil {
		t.Fatal(err)
	}
	if s.stretched() {
		t.Fatal("native landscape viewport is stretched instead of scrollable")
	}
	w, h := s.frameSize()
	if err := s.touchMove(touchUpdate{dy: -120}, nil, w, h); err != nil {
		t.Fatal(err)
	}
	if s.scrollY != 120 || s.scrollX != 0 {
		t.Fatalf("vertical swipe became offset %d,%d", s.scrollX, s.scrollY)
	}
}
