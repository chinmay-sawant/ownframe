package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPinnedChromeDoesNotExtendScroll(t *testing.T) {
	_, s := newPinShell()
	// IDs deliberately differ from operation indexes.
	s.display.Ops[0].ID = 19
	s.display.Ops[3].ID = 73
	s.display.Ops[3].Y = 900
	s.app.(*pinScreen).boxes = []layout.Box{
		{Y: 0, H: 400, OperationIDs: []uint64{19}},
		{Y: 900, H: 40, OperationIDs: []uint64{73}},
	}
	_, h := s.contentSize()
	if h != 400 {
		t.Fatalf("pinned chrome extended document to %d, want 400", h)
	}
	s.scrollY = 700
	s.pullScroll()
	if s.scrollY != 200 {
		t.Fatalf("scroll offset %d, want 200", s.scrollY)
	}
}
