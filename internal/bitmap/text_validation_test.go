package bitmap_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestBitmapRejectsInvalidAndExcessiveText(t *testing.T) {
	d, err := render.DisplayList(context.Background(), `<body>sample</body>`, 100, 100)
	if err != nil {
		t.Fatal(err)
	}
	var text layout.DisplayOp
	for _, op := range d.Ops {
		if op.Kind == layout.DisplayOpText {
			text = op
			break
		}
	}
	if text.Font == nil {
		t.Fatal("fixture has no text")
	}
	for _, invalid := range []bool{true, false} {
		op := text
		if invalid {
			op.Size = -1
		} else {
			op.W = 1 << 25
			op.RotateDeg = 30
		}
		view := &layout.Display{Width: 100, Height: 100, Ops: []layout.DisplayOp{op}, Order: []int{0}}
		if _, err := bitmap.Paint(view); err == nil {
			t.Fatal("invalid text silently omitted")
		}
	}
}
