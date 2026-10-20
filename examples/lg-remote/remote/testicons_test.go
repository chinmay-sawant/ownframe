package remote

import (
	"bytes"
	"image"
	"image/png"
)

// testIcon is a transparent 1x1 PNG. Tests replace the SVG icons with it:
// the layout still reserves the image box, but blinkless does not rasterize
// an SVG and encode a PNG on every redraw.
var testIcon = func() []byte {
	var b bytes.Buffer
	_ = png.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	return b.Bytes()
}()

func useTestIcons(a *App) {
	for _, name := range []string{
		"icon-play", "icon-pause", "icon-stop",
		"logo-prime", "logo-hotstar", "logo-youtube",
		"arrow-n", "arrow-s", "arrow-w", "arrow-e",
	} {
		a.page.SetImage(name, testIcon)
	}
}
