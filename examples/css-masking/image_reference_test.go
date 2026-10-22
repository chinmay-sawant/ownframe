package main

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

const referenceHTML = `<!doctype html><html><head><style>
body{margin:0}.photo{position:absolute;top:0;width:160px;height:100px;background-size:cover;background-position:center;background-repeat:no-repeat}
#flowers{left:0;background-image:url("assets/flowers.jpg")}
#coast{left:160px;background-image:url("assets/coast.jpg")}
</style></head><body><div class="photo" id="flowers"></div><div class="photo" id="coast"></div></body></html>`

func imageReference(t *testing.T) image.Image {
	t.Helper()
	p, err := ownframe.New(ownframe.Config{Title: "image reference", HTML: referenceHTML, Width: 320, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"flowers.jpg", "coast.jpg"} {
		data, err := files.ReadFile("assets/" + name)
		if err != nil {
			t.Fatal(err)
		}
		p.SetImage("assets/"+name, data)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(p.Probe().PNG()))
	if err != nil {
		t.Fatal(err)
	}
	return img
}
