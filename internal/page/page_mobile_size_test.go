package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestMobileViewportUsesLandscapeHeight(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body>mobile</body>`, Width: 420, Height: 934, MinWidth: 320, MinHeight: 480})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	for _, size := range [][2]int{{420, 934}, {885, 420}, {420, 934}} {
		p.SetSize(size[0], size[1])
		w, h := p.Size()
		if w != size[0] || h != size[1] {
			t.Fatalf("native viewport %v became %dx%d", size, w, h)
		}
	}
}

func TestMobileLockedViewKeepsCanvasMinimum(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body>game</body>`, Width: 1280, Height: 720, MinWidth: 1280, MinHeight: 720, LockView: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	p.SetSize(420, 320)
	if w, h := p.Size(); w != 1280 || h != 720 {
		t.Fatalf("locked canvas became %dx%d", w, h)
	}
}
