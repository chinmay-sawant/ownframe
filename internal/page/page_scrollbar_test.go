package page

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestScrollbarModeConfig(t *testing.T) {
	p, err := New(Config{
		HTML: `<p>scroll</p>`, Width: 100, Height: 100,
		Scrollbar: host.ScrollbarHidden,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := p.ScrollbarMode(); got != host.ScrollbarHidden {
		t.Fatalf("ScrollbarMode() = %v", got)
	}
}
