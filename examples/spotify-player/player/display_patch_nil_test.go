package player

import (
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestDisplayPatchHelperNilSafe(t *testing.T) {
	var h *DisplayPatchHelper
	if h.Report().Changed != 0 {
		t.Fatal("nil helper reported changed")
	}
	if h.PatchOp(0, func(*ownframe.DisplayOp) {}) {
		t.Fatal("nil helper PatchOp reported success")
	}
	if r := h.PatchDisplay(nil); r.Changed != 0 || r.Full {
		t.Fatalf("nil helper patch = %+v", r)
	}
	if got := (*App)(nil).PatchReport(); got.Changed != 0 {
		t.Fatalf("nil app report = %+v", got)
	}
	if (*App)(nil).PatchOp(0, func(*ownframe.DisplayOp) {}) {
		t.Fatal("nil app PatchOp reported success")
	}
}
