package ui

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// TestModalPaneBarSitsBesideContent: in a dialog the bar column never overlaps
// the scrolled content, and the pane keeps its fixed height.
func TestModalPaneBarSitsBesideContent(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	for i := 0; i < 30; i++ {
		_, _ = r.a.prefs.AddHistory(r.ctx, store.HistoryEntry{ConnectionID: r.a.cfg.ID, SQL: "SELECT 1", Rows: 1})
	}
	r.tap("rail-history")
	vp, in, bar := r.box("hist-vp"), r.box("hist-inner"), r.box("hist-bar")
	if in.X+in.W > bar.X+0.5 {
		t.Fatalf("content (right %.0f) overlaps the bar (left %.0f)", in.X+in.W, bar.X)
	}
	if bar.X+bar.W > vp.X+vp.W+0.5 || int(vp.H) != histPaneH {
		t.Fatalf("pane geometry: vp %+v bar %+v", vp, bar)
	}
	if in.H <= vp.H {
		t.Fatal("30 rows must overflow the pane")
	}
}
