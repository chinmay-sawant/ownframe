package page_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestDesktopMinimumSurvivesMobilePrepareError(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<body>desktop</body>`, Width: 420, Height: 934, MinWidth: 320, MinHeight: 480})
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Prepare(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(nil, p); err == nil {
		t.Fatal("nil context accepted")
	}
	p.SetSize(100, 100)
	if w, h := p.Size(); w != 320 || h != 480 {
		t.Fatalf("desktop minimum became %dx%d", w, h)
	}
}

func TestMobilePrepareRejectsNilPage(t *testing.T) {
	if err := page.PrepareMobile(context.Background(), nil); err != page.ErrNilPage {
		t.Fatalf("nil page error = %v", err)
	}
}
