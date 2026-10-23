package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// checkGroupClip requires a partial DrawRect over a group to match the full
// draw inside the rect, byte for byte.
func checkGroupClip() error {
	source := `<body style="margin:0;height:140px;background:#e8e8e8">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;background:#3060c0;` +
		`mix-blend-mode:multiply"><div style="position:absolute;left:30px;top:20px;width:60px;height:40px;` +
		`background:#c05030"></div></div></body>`

	d, err := render.DisplayList(context.Background(), source, 200, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("clipped group should replay")
	}

	full, part := newParityImage(200, 140), newParityImage(200, 140)
	defer full.Deallocate()
	defer part.Deallocate()

	DrawVisible(full, d, 0, 0)
	rect := image.Rect(30, 30, 120, 90)
	DrawRect(part, d, rect, 0, 0)

	a, b := make([]byte, 200*140*4), make([]byte, 200*140*4)
	full.ReadPixels(a)
	part.ReadPixels(b)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		x0, x1 := (y*200+rect.Min.X)*4, (y*200+rect.Max.X)*4
		if !bytes.Equal(a[x0:x1], b[x0:x1]) {
			return fmt.Errorf("clipped group differs on row %d in %d channels", y, countDiff(a[x0:x1], b[x0:x1]))
		}
	}

	return nil
}

func checkGroupMask() error {
	data, err := os.ReadFile("testdata/alpha-half.png")
	if err != nil {
		return err
	}
	source := `<body style="margin:0;height:140px;background:#159447">` +
		`<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;` +
		`background:#df3348;mask-image:url(mask.png);mask-mode:alpha;` +
		`mask-size:100% 100%;mask-repeat:no-repeat"></div></body>`
	d, err := render.DisplayListState(context.Background(), source, 200, 140,
		render.State{Images: func(string) ([]byte, error) { return data, nil }})
	if err != nil {
		return err
	}
	if !render.Replayable(d) {
		return fmt.Errorf("masked group should replay")
	}
	if err := compareGroupPixels(d, "alpha mask", [][2]int{{50, 50}, {150, 50}, {5, 5}}); err != nil {
		return err
	}
	full, part := newParityImage(200, 140), newParityImage(200, 140)
	defer full.Deallocate()
	defer part.Deallocate()
	DrawVisible(full, d, 0, 0)
	rect := image.Rect(35, 30, 125, 90)
	DrawRect(part, d, rect, 0, 0)
	a, b := make([]byte, 200*140*4), make([]byte, 200*140*4)
	full.ReadPixels(a)
	part.ReadPixels(b)
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		lo, hi := (y*200+rect.Min.X)*4, (y*200+rect.Max.X)*4
		if !bytes.Equal(a[lo:hi], b[lo:hi]) {
			return fmt.Errorf("masked partial replay differs on row %d", y)
		}
	}
	return nil
}

// checkGroupNested compares a multiply element holding a screen child, so the
// child's layer buffers inside the parent's and both blend once.
func checkGroupNested() error {
	source := `<body style="margin:0;height:140px;background:#d2d2d2">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;background:#3060c0;` +
		`mix-blend-mode:multiply"><div style="position:absolute;left:30px;top:20px;width:60px;height:40px;` +
		`background:#c05030;mix-blend-mode:screen"></div></div></body>`

	d, err := render.DisplayList(context.Background(), source, 200, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("nested groups should replay")
	}

	return compareGroupPixels(d, "nested multiply/screen", [][2]int{{60, 50}, {25, 25}, {5, 5}})
}

func checkPhotoMaskGPU() error {
	files := map[string]string{
		"mask.png":           "testdata/alpha-half.png",
		"assets/flowers.jpg": "testdata/flowers.jpg",
		"assets/coast.jpg":   "testdata/coast.jpg",
	}
	images := make(map[string][]byte, len(files))
	for src, path := range files {
		data, err := os.ReadFile(path) //nolint:gosec // reason: test fixture bundled in the repo.
		if err != nil {
			return err
		}
		images[src] = data
	}
	state := render.State{Images: func(src string) ([]byte, error) { return images[src], nil }}
	masked, err := render.DisplayListState(context.Background(), photoMaskHTML(true, true), 200, 140, state)
	if err != nil {
		return err
	}
	if !render.Replayable(masked) {
		return fmt.Errorf("photo mask should replay")
	}
	plain, err := render.DisplayListState(context.Background(), photoMaskHTML(true, false), 200, 140, state)
	if err != nil {
		return err
	}
	backdrop, err := render.DisplayListState(context.Background(), photoMaskHTML(false, false), 200, 140, state)
	if err != nil {
		return err
	}
	got, front, back := gpuDisplayPixels(masked), gpuDisplayPixels(plain), gpuDisplayPixels(backdrop)
	if err := compareGPUPixel(got, back, 50, 50, "transparent photo mask"); err != nil {
		return err
	}
	if err := compareGPUPixel(got, front, 150, 50, "opaque photo mask"); err != nil {
		return err
	}
	return compareGPUPixel(got, back, 5, 5, "photo outside mask")
}

func gpuDisplayPixels(d *render.Display) []byte {
	img := newParityImage(200, 140)
	defer img.Deallocate()
	DrawVisible(img, d, 0, 0)
	pixels := make([]byte, 200*140*4)
	img.ReadPixels(pixels)
	return pixels
}

func compareGPUPixel(got, want []byte, x, y int, name string) error {
	at := (y*200 + x) * 4
	for i := 0; i < 4; i++ {
		delta := int(got[at+i]) - int(want[at+i])
		if delta < -3 || delta > 3 {
			return fmt.Errorf("%s pixel (%d,%d): got %v, want %v", name, x, y, got[at:at+4], want[at:at+4])
		}
	}
	return nil
}

func photoMaskHTML(overlay, masked bool) string {
	html := `<body style="margin:0;height:140px;background-image:url(assets/coast.jpg);background-size:cover;background-position:center">`
	if overlay {
		html += `<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;background-image:url(assets/flowers.jpg);background-size:cover;background-position:center;`
		if masked {
			html += `mask-image:url(mask.png);mask-mode:alpha;mask-size:100% 100%;mask-repeat:no-repeat;`
		}
		html += `"></div>`
	}
	return html + `</body>`
}

// compareGroupPixels draws the display with the replay and the bitmap
// painters and requires the sampled pixels to agree within rounding.
func compareGroupPixels(d *render.Display, name string, points [][2]int) error {
	want, err := bitmap.Paint(d)
	if err != nil {
		return err
	}

	img := newParityImage(200, 140)
	defer img.Deallocate()

	DrawVisible(img, d, 0, 0)

	pixels := make([]byte, 200*140*4)
	img.ReadPixels(pixels)

	for _, p := range points {
		if err := closePixel(pixels, want, 200, p[0], p[1]); err != nil {
			return fmt.Errorf("%s pixel (%d,%d): %w", name, p[0], p[1], err)
		}
	}

	return nil
}

// closePixel compares one rendered pixel with the bitmap painter's, within
// three channel steps.
func closePixel(pixels []byte, want image.Image, stride, x, y int) error {
	at := (y*stride + x) * 4
	c, ok := color.NRGBAModel.Convert(want.At(x, y)).(color.NRGBA)
	if !ok {
		return fmt.Errorf("pixel model changed for %T", want.At(x, y))
	}
	wantRGBA := []uint8{c.R, c.G, c.B, c.A}

	for i, v := range wantRGBA {
		if delta := int(pixels[at+i]) - int(v); delta < -3 || delta > 3 {
			return fmt.Errorf("got %v, want %v", pixels[at:at+4], wantRGBA)
		}
	}

	return nil
}

// groupModes are the separable mix-blend-modes the compositor accepts.
var groupModes = []string{"normal", "multiply", "screen", "overlay", "darken", "lighten",
	"difference", "exclusion", "hard-light", "soft-light", "color-dodge", "color-burn"}

// checkGroupComposite compares each accepted mode's composited group with the
// bitmap painter inside and outside the blend overlap, then checks nesting.
func checkGroupComposite() error {
	for _, mode := range groupModes {
		if err := checkGroupMode(mode); err != nil {
			return err
		}
	}

	if err := checkGroupNested(); err != nil {
		return err
	}
	if err := checkGroupMask(); err != nil {
		return err
	}
	return checkPhotoMaskGPU()
}

func checkGroupMode(mode string) error {
	source := `<body style="margin:0;height:140px;background:#d2d2d2">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;background:#3060c0">` +
		`<div style="position:absolute;left:30px;top:20px;width:60px;height:40px;background:#c05030;` +
		`mix-blend-mode:` + mode + `"></div></div></body>`

	d, err := render.DisplayList(context.Background(), source, 200, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("mode %s should replay", mode)
	}

	return compareGroupPixels(d, mode, [][2]int{{60, 50}, {25, 25}, {5, 5}})
}
