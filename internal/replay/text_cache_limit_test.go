package replay

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestTextFacesStayBoundedAcrossSizeChanges(t *testing.T) {
	d, err := render.DisplayList(context.Background(), `<p>resize</p>`, 200, 200)
	if err != nil {
		t.Fatal(err)
	}
	var op layout.DisplayOp
	for _, candidate := range d.Ops {
		if candidate.Font != nil {
			op = candidate
			break
		}
	}
	if op.Font == nil {
		t.Fatal("no text font")
	}
	for i := range 400 {
		op.Size = 10 + float64(i)/10
		if textFace(&op) == nil {
			t.Fatal("face did not load")
		}
	}
	fonts.mu.Lock()
	defer fonts.mu.Unlock()
	if len(fonts.faces) > 256 || len(fonts.sources) > 64 {
		t.Fatal("unbounded font cache")
	}
}
