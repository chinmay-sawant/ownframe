package ui

import (
	"image/color"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// TestModalScrollbarsArePainted: thumbs inside the palette and dialogs are
// really on screen (the engine skips absolutely positioned children of the
// fixed overlay, so these bars are laid out in flow).
func TestModalScrollbarsArePainted(t *testing.T) {
	for _, dark := range []bool{true, false} {
		want := color.RGBA{0x4a, 0x4a, 0x4a, 0xff}
		if !dark {
			want = color.RGBA{0xc4, 0xc4, 0xc4, 0xff}
		}
		r := newRig(t, 1440, 960, "")
		if !dark {
			r.tap("tb-theme")
		}
		r.chord("k", false)
		th := r.box("pal-thumb")
		if got := pixelAt(r, int(th.X+th.W/2), int(th.Y+th.H/2)); got != want {
			t.Fatalf("dark=%v palette thumb pixel %v, want %v", dark, got, want)
		}
		r.key("escape")
		for i := 0; i < 30; i++ {
			_, _ = r.a.prefs.AddHistory(r.ctx, store.HistoryEntry{ConnectionID: r.a.cfg.ID, SQL: "SELECT 1", Rows: 1})
		}
		r.tap("rail-history")
		th = r.box("hist-thumb")
		if got := pixelAt(r, int(th.X+th.W/2), int(th.Y+th.H/2)); got != want {
			t.Fatalf("dark=%v history thumb pixel %v, want %v", dark, got, want)
		}
	}
}
