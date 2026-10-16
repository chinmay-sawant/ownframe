package eyes

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestSpiralImagesAreBound(t *testing.T) {
	a, c := newT(t)
	a.selectMood(spiralMood, c.at)
	for range 20 {
		tk(t, a, c, time.Second/60)
	}
	if a.pm.spiralL == nil || a.pm.spiralR == nil {
		t.Fatal("spiral image ops are not bound")
	}
	if a.pm.spiralL == a.pm.spiralR {
		t.Fatal("both spiral slots bound to the same image op")
	}
	if a.pm.left.Opacity() > 0.05 || a.pm.right.Opacity() > 0.05 {
		t.Fatalf("white eyes should be faded: opacity %.3f / %.3f",
			a.pm.left.Opacity(), a.pm.right.Opacity())
	}
	for _, op := range []*ownframe.DisplayOp{a.pm.spiralL, a.pm.spiralR} {
		data, width, height := op.ImageBytes()
		if len(data) == 0 || width <= 0 || height <= 0 || op.Alpha < 0.9 || op.Opacity() < 0.9 {
			t.Fatalf("spiral image is not visible: %dx%d, %d bytes, alpha %.2f opacity %.2f",
				width, height, len(data), op.Alpha, op.Opacity())
		}
	}
}
