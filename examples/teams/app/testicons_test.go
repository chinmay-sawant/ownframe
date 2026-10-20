package app

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"testing"
)

// testIcon is a transparent 1x1 PNG. Tests replace every SVG with it so a
// redraw does not rasterize an icon and encode a PNG for each image box.
var testIcon = func() []byte {
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	return b.Bytes()
}()

func TestMain(m *testing.M) {
	assetData = func(string) []byte { return testIcon }
	os.Exit(m.Run())
}
