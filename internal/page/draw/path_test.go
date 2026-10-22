package draw

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

const unknownOp = layout.DisplayKind(100)

func TestChooseNil(t *testing.T) {
	t.Parallel()

	if Choose(nil) != PathNone {
		t.Fatal("nil")
	}
}

func TestChooseEmptyList(t *testing.T) {
	t.Parallel()

	if Choose(&layout.Display{}) != PathList {
		t.Fatal("empty")
	}
}

func TestChooseUnknownOpIsBitmap(t *testing.T) {
	t.Parallel()

	d := &layout.Display{Ops: []layout.DisplayOp{{Kind: unknownOp}}}
	if Choose(d) != PathBitmap {
		t.Fatal("unknown")
	}
}
