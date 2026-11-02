package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
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

func checkBackdropFilter() error {
	if err := checkBackdropSourceOrigin(); err != nil {
		return err
	}
	source := `<body style="margin:0;height:100px;background:#ffffff">` +
		`<div style="position:absolute;left:0;top:0;width:100px;height:100px;background:#ff0000"></div>` +
		`<div style="position:absolute;left:100px;top:0;width:100px;height:100px;background:#0000ff"></div>` +
		`<div style="position:absolute;left:80px;top:0;width:40px;height:100px;` +
		`backdrop-filter:%s;background:rgba(255,255,255,0.1)"></div></body>`
	plain, err := render.DisplayList(context.Background(), fmt.Sprintf(source, "none"), 200, 100)
	if err != nil {
		return err
	}
	filtered, err := render.DisplayList(context.Background(), fmt.Sprintf(source, "blur(10px)"), 200, 100)
	if err != nil {
		return err
	}
	groupFound := false
	for i := range filtered.Ops {
		if group := filtered.Ops[i].Group(); group != nil && group.BackdropFilter == "blur(10px)" {
			groupFound = true
			break
		}
	}
	if !groupFound {
		return fmt.Errorf("display list did not preserve backdrop-filter")
	}
	if _, err := backdropFilterShader(); err != nil {
		return fmt.Errorf("compile backdrop-filter shader: %w", err)
	}

	plainImage, filteredImage := newParityImage(200, 100), newParityImage(200, 100)
	defer plainImage.Deallocate()
	defer filteredImage.Deallocate()
	DrawVisible(plainImage, plain, 0, 0)
	DrawVisible(filteredImage, filtered, 0, 0)

	plainPixels, filteredPixels := make([]byte, 200*100*4), make([]byte, 200*100*4)
	plainImage.ReadPixels(plainPixels)
	filteredImage.ReadPixels(filteredPixels)
	at := (40*200 + 99) * 4
	if int(filteredPixels[at+2])-int(plainPixels[at+2]) < 30 {
		return fmt.Errorf("blurred backdrop pixel %v did not mix the blue backdrop into plain pixel %v",
			filteredPixels[at:at+4], plainPixels[at:at+4])
	}
	for _, x := range []int{70, 79, 120, 130} {
		outside := (40*200 + x) * 4
		if !bytes.Equal(plainPixels[outside:outside+4], filteredPixels[outside:outside+4]) {
			return fmt.Errorf("backdrop filter changed neighboring pixel x=%d from %v to %v",
				x, plainPixels[outside:outside+4], filteredPixels[outside:outside+4])
		}
	}

	refracted, err := render.DisplayList(context.Background(), `<body style="margin:0;height:60px;background:#ffffff">`+
		`<div style="position:absolute;left:0;top:0;width:42px;height:60px;background:#ff0000"></div>`+
		`<div style="position:absolute;left:42px;top:0;width:58px;height:60px;background:#0000ff"></div>`+
		`<div style="position:absolute;left:40px;top:0;width:40px;height:60px;`+
		`backdrop-filter:refract(5px);background:transparent"></div></body>`, 100, 60)
	if err != nil {
		return err
	}
	refractedImage := newParityImage(100, 60)
	defer refractedImage.Deallocate()
	DrawVisible(refractedImage, refracted, 0, 0)
	refractedPixels := make([]byte, 100*60*4)
	refractedImage.ReadPixels(refractedPixels)
	refractedAt := (30*100 + 40) * 4
	if refractedPixels[refractedAt+2] < 240 || refractedPixels[refractedAt] > 20 {
		return fmt.Errorf("refracted edge pixel %v did not bend the blue backdrop inward", refractedPixels[refractedAt:refractedAt+4])
	}

	empty, err := render.DisplayList(context.Background(), `<body style="margin:0">`+
		`<div style="position:absolute;left:0;top:0;width:100px;height:100px;background:#804020"></div>`+
		`<div style="position:absolute;left:20px;top:20px;width:60px;height:60px;`+
		`backdrop-filter:saturate(0);background:transparent"></div></body>`, 100, 100)
	if err != nil {
		return err
	}
	emptyFiltered := newParityImage(100, 100)
	defer emptyFiltered.Deallocate()
	DrawVisible(emptyFiltered, empty, 0, 0)
	emptyPixels := make([]byte, 100*100*4)
	emptyFiltered.ReadPixels(emptyPixels)
	emptyAt := (50*100 + 50) * 4
	if emptyPixels[emptyAt+3] < 250 ||
		max(emptyPixels[emptyAt], emptyPixels[emptyAt+1], emptyPixels[emptyAt+2])-
			min(emptyPixels[emptyAt], emptyPixels[emptyAt+1], emptyPixels[emptyAt+2]) > 2 ||
		emptyPixels[emptyAt] < 70 || emptyPixels[emptyAt] > 80 {
		return fmt.Errorf("empty saturated filter pixel %v is not opaque grayscale near luma 75", emptyPixels[emptyAt:emptyAt+4])
	}
	emptyOutsideAt := (10*100 + 10) * 4
	if emptyPixels[emptyOutsideAt] != 128 || emptyPixels[emptyOutsideAt+1] != 64 ||
		emptyPixels[emptyOutsideAt+2] != 32 || emptyPixels[emptyOutsideAt+3] < 250 {
		return fmt.Errorf("pixel outside empty filter is %v, want original backdrop color", emptyPixels[emptyOutsideAt:emptyOutsideAt+4])
	}

	translucentSource := `<body style="margin:0">` +
		`<div style="position:absolute;left:0;top:0;width:100px;height:100px;background:rgba(255,0,0,.5)"></div>%s</body>`
	plainTranslucent, err := render.DisplayList(context.Background(), fmt.Sprintf(translucentSource, ""), 100, 100)
	if err != nil {
		return err
	}
	filteredTranslucent, err := render.DisplayList(context.Background(), fmt.Sprintf(translucentSource,
		`<div style="position:absolute;left:20px;top:20px;width:60px;height:60px;`+
			`backdrop-filter:saturate(0);background:transparent"></div>`), 100, 100)
	if err != nil {
		return err
	}
	plainAlphaImage, filteredAlphaImage := newParityImage(100, 100), newParityImage(100, 100)
	defer plainAlphaImage.Deallocate()
	defer filteredAlphaImage.Deallocate()
	DrawVisible(plainAlphaImage, plainTranslucent, 0, 0)
	DrawVisible(filteredAlphaImage, filteredTranslucent, 0, 0)
	plainAlphaPixels, filteredAlphaPixels := make([]byte, 100*100*4), make([]byte, 100*100*4)
	plainAlphaImage.ReadPixels(plainAlphaPixels)
	filteredAlphaImage.ReadPixels(filteredAlphaPixels)
	alphaAt := (50*100 + 50) * 4
	filteredAlpha := filteredAlphaPixels[alphaAt : alphaAt+4]
	if int(filteredAlpha[3])-int(plainAlphaPixels[alphaAt+3]) > 2 ||
		int(plainAlphaPixels[alphaAt+3])-int(filteredAlpha[3]) > 2 ||
		max(filteredAlpha[0], filteredAlpha[1], filteredAlpha[2])-min(filteredAlpha[0], filteredAlpha[1], filteredAlpha[2]) > 2 ||
		max(filteredAlpha[0], filteredAlpha[1], filteredAlpha[2]) < 20 {
		return fmt.Errorf("translucent saturated pixel %v differs in alpha or is not visible grayscale from plain pixel %v",
			filteredAlpha, plainAlphaPixels[alphaAt:alphaAt+4])
	}
	translucentOutsideAt := (10*100 + 10) * 4
	if !bytes.Equal(plainAlphaPixels[translucentOutsideAt:translucentOutsideAt+4],
		filteredAlphaPixels[translucentOutsideAt:translucentOutsideAt+4]) {
		return fmt.Errorf("pixel outside translucent filter changed from %v to %v",
			plainAlphaPixels[translucentOutsideAt:translucentOutsideAt+4],
			filteredAlphaPixels[translucentOutsideAt:translucentOutsideAt+4])
	}

	if err := checkFixedBackdropGroupOffset(); err != nil {
		return err
	}

	return nil
}

func checkBackdropSourceOrigin() error {
	shader, err := backdropFilterShader()
	if err != nil {
		return err
	}
	source := ebiten.NewImage(128, 128)
	defer source.Deallocate()
	source.Fill(color.RGBA{R: 255, A: 255})
	sub := source.SubImage(image.Rect(20, 20, 80, 80))
	region, ok := sub.(*ebiten.Image)
	if !ok {
		return fmt.Errorf("backdrop subimage has unexpected type %T", sub)
	}
	output := newParityImage(60, 60)
	defer output.Deallocate()
	op := &ebiten.DrawRectShaderOptions{Blend: ebiten.BlendCopy}
	op.Images[0], op.Images[1] = region, region
	op.Uniforms = map[string]any{
		"Radius": float32(0), "Axis": float32(1), "Saturation": float32(0),
		"Refraction": float32(0), "BoxX": float32(0), "BoxY": float32(0),
		"BoxW": float32(60), "BoxH": float32(60), "CornerRadius": float32(0),
	}
	output.DrawRectShader(60, 60, shader, op)
	pixels := make([]byte, 60*60*4)
	output.ReadPixels(pixels)
	at := (50*60 + 50) * 4
	pixel := pixels[at : at+4]
	if pixel[0] < 50 || pixel[0] > 58 || pixel[1] != pixel[0] || pixel[2] != pixel[0] || pixel[3] != 255 {
		return fmt.Errorf("offset backdrop source left the far corner unfiltered: %v", pixel)
	}
	return nil
}

// A fixed element inside a filtered group still needs the group's local
// offset while it paints into the offscreen layer. The composite restores the
// group's page position afterward.
func checkFixedBackdropGroupOffset() error {
	const source = `<body style="margin:0;width:200px;height:120px;background:#182030">` +
		`<div style="position:fixed;left:40px;top:20px;width:120px;height:80px;` +
		`backdrop-filter:blur(10px);background:rgba(255,255,255,.1)">` +
		`<div style="position:absolute;left:0;top:0;width:30px;height:30px;background:#ff0000"></div>` +
		`</div></body>`
	display, err := render.DisplayList(context.Background(), source, 200, 120)
	if err != nil {
		return err
	}
	var fixedMember bool
	for i := range display.Ops {
		op := &display.Ops[i]
		group := op.Group()
		if op.Fixed && op.Kind == layout.DisplayOpFillRect && group != nil &&
			group.BackdropFilter == "blur(10px)" && op.R > 0.9 && op.G < 0.1 && op.B < 0.1 {
			fixedMember = true
			break
		}
	}
	if !fixedMember {
		return fmt.Errorf("filtered panel did not retain its fixed red child")
	}

	image := newParityImage(200, 120)
	defer image.Deallocate()
	DrawVisible(image, display, 0, 0)
	pixels := make([]byte, 200*120*4)
	image.ReadPixels(pixels)
	pixel := func(x, y int) []byte {
		at := (y*200 + x) * 4
		return pixels[at : at+4]
	}
	inside := pixel(55, 35)
	if inside[0] < 220 || inside[1] > 35 || inside[2] > 35 {
		return fmt.Errorf("fixed child painted at %v, want red near its page position", inside)
	}
	shifted := pixel(95, 55)
	if shifted[0] > 220 && shifted[1] < 35 && shifted[2] < 35 {
		return fmt.Errorf("fixed child was painted at its doubled group offset: %v", shifted)
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
