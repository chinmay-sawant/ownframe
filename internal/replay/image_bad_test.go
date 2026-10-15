package replay

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestImageBadPayloadSkips(t *testing.T) {
	src := `<body style="margin:0"><img src="bad.png" style="width:20px;height:20px"></body>`
	st := render.State{Images: func(s string) ([]byte, error) { return []byte("not an image"), nil }}
	d, err := render.DisplayListState(context.Background(), src, 100, 100, st)
	if err != nil {
		t.Fatal(err)
	}
	op := imageKind(d)
	if op == nil {
		t.Fatal("no image op")
	}
	if _, _, skip, err := masking.DecodeImage(op); skip || err == nil {
		t.Fatalf("skip %v err %v", skip, err)
	}
	if decodedImage(op) != nil {
		t.Fatal("bad payload decoded")
	}
}
