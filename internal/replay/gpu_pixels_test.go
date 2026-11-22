package replay

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"testing"
)

// checkTextClip pins partial-repaint clipping of baked and direct runs.
func checkTextClip() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}

	display := &layout.Display{
		Width: 320, Height: 240, PointsPerPixel: pxPerPt,
		Ops: ops, Order: []int{0, 1, 2, 3, 4, 5},
	}
	rect := image.Rect(8, 40, 220, 232)

	direct := newParityImage(320, 240)
	cached := newParityImage(320, 240)
	defer direct.Deallocate()
	defer cached.Deallocate()

	clip, ok := direct.SubImage(rect).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("sub-image")
	}
	for i := range ops {
		drawTextDirect(&ops[i], clip)
	}
	DrawRect(cached, display, rect, 0, 0)

	a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("clipped text raster differs in %d channels", countDiff(a, b))
	}

	return nil
}

// drawTextDirect draws one text op through the uncached path only.
func drawTextDirect(op *layout.DisplayOp, dst *ebiten.Image) {
	face := textFace(op)
	if face == nil || op.Text == "" {
		return
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	drawTextRun(dst, face, content, op.X*pxPerPt, op.Y*pxPerPt-face.Metrics().HAscent, op)
}

// phase is the run's keyed sub-pixel offset, built the way cachedTextRun
// builds it from the op and the face.
func phase(op *layout.DisplayOp, face *text.GoTextFace) [2]float64 {
	x := op.X * pxPerPt
	y := op.Y*pxPerPt - face.Metrics().HAscent

	return [2]float64{textPhase(x - math.Floor(x)), textPhase(y - math.Floor(y))}
}

// textParityOps are the runs drawn both ways: plain, spaced, bold, dim, a
// whole-pixel shifted share, and one too wide to cache.
func textParityOps() ([]layout.DisplayOp, error) {
	d, err := render.DisplayList(context.Background(),
		`<html><body style="margin:0;font-size:14px"><p>parity</p></body></html>`, 320, 240)
	if err != nil {
		return nil, err
	}

	found := firstTextOp(d)
	if found == nil {
		return nil, fmt.Errorf("no text op in the fixture")
	}

	plain := *found
	plain.Text, plain.Size = "Raster run 128", 14
	plain.Bold, plain.LetterSpacing, plain.Alpha = false, 0, 0
	plain.X, plain.Y = 3, 60
	plain.R, plain.G, plain.B = 0.1, 0.2, 0.3

	spaced := plain
	spaced.Text, spaced.LetterSpacing, spaced.Y = "spaced run", 3.5, 78

	fakeBold := plain
	fakeBold.Text, fakeBold.Bold, fakeBold.Y = "Fake bold 17", true, 96

	dim := plain
	dim.Text, dim.Alpha, dim.Y = "Half alpha 9", 0.5, 114

	shifted := plain
	shifted.X += 90
	shifted.Y += 72

	wide := plain
	wide.Text, wide.Y = strings.Repeat("wide ", 40), 150

	ops := []layout.DisplayOp{plain, spaced, fakeBold, dim, shifted, wide}

	return ops, setRunBounds(ops, &plain)
}

// setRunBounds fills the advance and ink metrics the dirty-rect filter reads,
// so the clipped parity check draws the runs the way a real page would.
func setRunBounds(ops []layout.DisplayOp, base *layout.DisplayOp) error {
	face := textFace(base)
	if face == nil {
		return fmt.Errorf("the parity runs have no face")
	}

	m := face.Metrics()
	for i := range ops {
		width, height := text.Measure(ops[i].Text, face, 0)
		ops[i].W, ops[i].H = width/pxPerPt, height/pxPerPt
		ops[i].InkDescent = m.HDescent / pxPerPt
	}

	return nil
}

// textRasterImages counts the cached rasters that hold an image, so a
// remembered decline does not read as a bake.
func textRasterImages() int {
	textRuns.mu.Lock()
	defer textRuns.mu.Unlock()

	count := 0
	for _, entry := range textRuns.entries {
		if entry.img != nil {
			count++
		}
	}

	return count
}

// checkTextRaster draws the text parity runs through the direct path and
// through the raster cache and requires identical pixels over an opaque
// backdrop. The second pass must hit every run, the two positions that differ
// by whole pixels must have baked one raster, and the run past the size cap
// must have baked none.
func checkTextRaster() error {
	ops, err := textParityOps()
	if err != nil {
		return err
	}

	face := textFace(&ops[0])
	if face == nil {
		return fmt.Errorf("the parity run has no face")
	}
	if first, last := phase(&ops[0], face), phase(&ops[4], face); first != last {
		return fmt.Errorf("parity positions %v and %v do not share a key", first, last)
	}

	direct := newParityImage(320, 240)
	cached := newParityImage(320, 240)
	defer direct.Deallocate()
	defer cached.Deallocate()

	before := textRasterImages()
	for pass := 0; pass < 2; pass++ {
		direct.Fill(color.White)
		cached.Fill(color.White)

		for i := range ops {
			drawTextDirect(&ops[i], direct)
			drawOp(cached, &ops[i], 0, 0)
		}

		a, b := make([]byte, 320*240*4), make([]byte, 320*240*4)
		direct.ReadPixels(a)
		cached.ReadPixels(b)
		if !bytes.Equal(a, b) {
			return fmt.Errorf("text raster pass %d differs in %d channels", pass, countDiff(a, b))
		}

		if got := textRasterImages() - before; got != 4 {
			return fmt.Errorf("pass %d baked %d rasters, want 4", pass, got)
		}
	}

	return nil
}

// checkRasterCache requires identical pixels; four rasters bake, not five.
func checkRasterCache() error {
	direct := newParityImage(256, 200)
	cached := newParityImage(256, 200)
	defer direct.Deallocate()
	defer cached.Deallocate()

	before := rasterEntries()
	ops := rasterParityOps()
	for i := range ops {
		drawDirect(&ops[i], direct)
		drawOp(cached, &ops[i], 0, 0)
	}

	a, b := make([]byte, 256*200*4), make([]byte, 256*200*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("raster cache differs in %d channels", countDiff(a, b))
	}

	if got := rasterEntries() - before; got != 4 {
		return fmt.Errorf("raster cache added %d entries, want 4", got)
	}

	return nil
}

// rasterEntries counts cached rasters.
func rasterEntries() int {
	rasters.mu.Lock()
	defer rasters.mu.Unlock()

	return len(rasters.entries)
}

// countDiff counts differing bytes.
func countDiff(a, b []byte) int {
	count := 0
	for i := range a {
		if a[i] != b[i] {
			count++
		}
	}

	return count
}

// checkRasterClip pins that a partial repaint clips like the direct draw.
func checkRasterClip() error {
	ops := rasterParityOps()
	display := &layout.Display{
		Width: 256, Height: 200, PointsPerPixel: 0.75,
		Ops: ops, Order: []int{0, 1, 2, 3, 4},
	}
	rect := image.Rect(12, 16, 160, 120)

	direct := newParityImage(256, 200)
	cached := newParityImage(256, 200)
	defer direct.Deallocate()
	defer cached.Deallocate()

	clip, ok := direct.SubImage(rect).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("sub-image")
	}
	for i := range ops {
		drawDirect(&ops[i], clip)
	}
	DrawRect(cached, display, rect, 0, 0)

	a, b := make([]byte, 256*200*4), make([]byte, 256*200*4)
	direct.ReadPixels(a)
	cached.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("clipped raster differs in %d channels", countDiff(a, b))
	}

	return nil
}

// cardParityPage is the chunk parity fixture: translucent, rounded, plain cards.
func cardParityPage() (*layout.Display, error) {
	source := `<body style="margin:0;width:640px;height:200px;background:#ddeeff">` +
		`<div style="position:absolute;left:20px;top:20px;width:160px;height:100px;` +
		`background:rgba(51,102,204,0.5)"></div>` +
		`<div style="position:absolute;left:220px;top:20px;width:160px;height:100px;` +
		`background:#ffddaa;border:3px solid #aa6622;border-radius:8px;padding:6px;font-size:13px">Card B detail</div>` +
		`<div style="position:absolute;left:420px;top:20px;width:160px;height:100px;` +
		`background:#ccffcc;border:1px solid #338833"></div>` +
		`</body>`

	d, err := render.DisplayList(context.Background(), source, 640, 200)
	if err != nil {
		return nil, err
	}

	if !render.Replayable(d) {
		return nil, fmt.Errorf("the card fixture should replay")
	}

	return d, nil
}

// chunkPartialParity repaints one card's rect into a fresh image and requires
// it to match the full draw inside that rect, byte for byte.
func chunkPartialParity(d *layout.Display, full []byte) error {
	part := newParityImage(640, 200)
	defer part.Deallocate()

	rect := image.Rect(30, 30, 130, 90)
	DrawRect(part, d, rect, 0, 0)
	pixels := make([]byte, 640*200*4)
	part.ReadPixels(pixels)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		x0, x1 := (y*640+rect.Min.X)*4, (y*640+rect.Max.X)*4
		if !bytes.Equal(full[x0:x1], pixels[x0:x1]) {
			return fmt.Errorf("chunked partial draw differs on row %d in %d channels",
				y, countDiff(full[x0:x1], pixels[x0:x1]))
		}
	}

	return nil
}

// chunkEntries reports the number of cached chunk images.
func chunkEntries() int {
	chunks.mu.Lock()
	defer chunks.mu.Unlock()

	return len(chunks.entries)
}

// checkChunkParity requires byte-identical direct and chunked draws.
// Translucent ops stay direct.
func checkChunkParity() error {
	d, err := cardParityPage()
	if err != nil {
		return err
	}

	p := chunkPlanFor(d)
	if p == nil {
		return fmt.Errorf("the card page has no chunk plan")
	}

	if _, runs, _ := planStats(p); runs == 0 {
		return fmt.Errorf("the card page folded no run")
	}

	if runHolds(p, true) {
		return fmt.Errorf("a run holds a translucent op")
	}

	direct, chunked := newParityImage(640, 200), newParityImage(640, 200)
	defer direct.Deallocate()
	defer chunked.Deallocate()

	for _, i := range d.Order {
		drawOp(direct, &d.Ops[i], 0, 0)
	}

	before := chunkEntries()
	DrawVisible(chunked, d, 0, 0)
	baked := chunkEntries()
	if baked == before {
		return fmt.Errorf("the chunked draw baked no chunk")
	}

	DrawVisible(chunked, d, 0, 0)
	if chunkEntries() != baked {
		return fmt.Errorf("the second chunked draw re-baked a chunk")
	}

	a, b := make([]byte, 640*200*4), make([]byte, 640*200*4)
	direct.ReadPixels(a)
	chunked.ReadPixels(b)
	if !bytes.Equal(a, b) {
		return fmt.Errorf("chunked visible draw differs in %d channels: %s",
			countDiff(a, b), diffPixels(a, b, 640))
	}

	return chunkPartialParity(d, a)
}

// runHolds reports whether a run holds an op with opacity under one.
func runHolds(p *chunkPlan, want bool) bool {
	for i := range p.items {
		if p.items[i].run == nil {
			continue
		}

		for _, op := range p.items[i].run.ops {
			if (op.Opacity() < 1) == want {
				return true
			}
		}
	}

	return false
}

// diffPixels describes the first differing pixels for the failure message.
func diffPixels(a, b []byte, width int) string {
	out := ""

	for i := 0; i+4 <= len(a) && len(out) < 400; i += 4 {
		if bytes.Equal(a[i:i+4], b[i:i+4]) {
			continue
		}

		p := i / 4
		out += fmt.Sprintf("(%d,%d) %v vs %v ", p%width, p/width, a[i:i+4], b[i:i+4])
	}

	return out
}

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

func TestGPUVisibleParity(t *testing.T) {
	if os.Getenv("OWNFRAME_REPLAY_GPU_TEST") != "1" && os.Getenv("GPUI_REPLAY_GPU_TEST") != "1" {
		t.Skip("requires an isolated graphics session")
	}
	source, err := os.ReadFile("testdata/layout.html")
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]map[string]any, 480)
	for i := range rows {
		rows[i] = map[string]any{"ID": i, "Region": i % 8, "Name": "service"}
	}
	p, err := page.New(page.Config{HTML: string(source), Width: 1908, Height: 999})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(map[string]any{"Revision": 0, "Top": 0, "Bottom": 0, "Nav": []string{"Overview", "Services", "Regions", "Deployments", "Incidents", "Capacity"}, "Rows": rows})
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	d.Order = append(d.Order, len(d.Ops))
	d.Ops = append(d.Ops, layout.DisplayOp{Kind: layout.DisplayOpFillRect, Fixed: true, X: 10, Y: 10, W: 20, H: 20, R: 1})
	g := &parityGame{display: d}
	defer g.renderer.Close()
	ebiten.SetWindowSize(1908, 999)
	if err = ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
}

type parityGame struct {
	renderer      Renderer
	display       *layout.Display
	before, after *ebiten.Image
	step          int
	err           error
}

func (g *parityGame) Layout(int, int) (int, int) { return 1908, 999 }
func (g *parityGame) Update() error {
	if g.step >= 5 || g.err != nil {
		return ebiten.Termination
	}
	return nil
}
func (g *parityGame) Draw(dst *ebiten.Image) {
	if g.step >= 5 {
		return
	}
	if g.step == 0 {
		if err := checkSleepyEyeRenders(); err != nil {
			g.err = err
			return
		}
		if err := checkTranslucentPixels(); err != nil {
			g.err = err
			return
		}
		if err := checkGPUOpacity(); err != nil {
			g.err = err
			return
		}
		if err := checkRasterCache(); err != nil {
			g.err = err
			return
		}
		if err := checkRasterClip(); err != nil {
			g.err = err
			return
		}
		if err := checkTextRaster(); err != nil {
			g.err = err
			return
		}
		if err := checkTextClip(); err != nil {
			g.err = err
			return
		}
		if err := checkXformParity(); err != nil {
			g.err = err
			return
		}
		if err := checkProjectedText(); err != nil {
			g.err = err
			return
		}
		if err := checkQuarterTurnText(); err != nil {
			g.err = err
			return
		}
		if err := checkGroupComposite(); err != nil {
			g.err = err
			return
		}
		if err := checkGroupClip(); err != nil {
			g.err = err
			return
		}
		if err := checkBackdropFilter(); err != nil {
			g.err = err
			return
		}
		if err := checkChunkParity(); err != nil {
			g.err = err
			return
		}
		if err := checkPathGolden(); err != nil {
			g.err = err
			return
		}
		if err := checkPathXform(); err != nil {
			g.err = err
			return
		}
		if err := checkPathGroup(); err != nil {
			g.err = err
			return
		}
	}
	if g.before == nil {
		g.before = newParityImage(1908, 999)
		g.after = newParityImage(1908, 999)
	}
	g.before.Clear()
	g.after.Clear()
	dy := -[]float64{0, 500.25, 1000, 20000, 36000}[g.step]
	for _, i := range g.display.Order {
		op := &g.display.Ops[i]
		drawOp(g.before, op, 0, dy)
	}
	g.renderer.DrawVisible(g.after, g.display, 0, dy)
	a, b := make([]byte, 1908*999*4), make([]byte, 1908*999*4)
	g.before.ReadPixels(a)
	g.after.ReadPixels(b)
	if !bytes.Equal(a, b) {
		count := 0
		for i := range a {
			if a[i] != b[i] {
				count++
			}
		}
		g.err = fmt.Errorf("scroll %.2f differs in %d channels", -dy, count)
	}
	dst.DrawImage(g.after, nil)
	g.step++
}

// Standalone targets keep the same pixel origin for both sides of a comparison.
// Atlas placement can change interpolation rounding by one color level.
func newParityImage(width, height int) *ebiten.Image {
	return ebiten.NewImageWithOptions(image.Rect(0, 0, width, height), &ebiten.NewImageOptions{Unmanaged: true})
}

// The corner-shape pages below are 200 by 140 CSS pixels. The box is a 128 by
// 88 CSS pixel border box at (20, 20), a 24 px bevel or scoop corner, and a
// 4 px border. The body carries the height so the page fill reaches the
// bottom edge and every empty sample reads as the page colour.
const (
	pathPageWidth  = 200
	pathPageHeight = 140
	pathChannels   = 4
	pathSlack      = 3
)

// pathBoxStyle is the shared shaped box: a background, a uniform border, a
// 24 px radius, and one corner shape, over the given page colour.
func pathBoxStyle(backdrop, shape, extra string) string {
	return `<body style="margin:0;height:` + strconv.Itoa(pathPageHeight) + `px;background:` + backdrop + `">` +
		`<div style="position:absolute;left:20px;top:20px;width:120px;height:80px;` +
		`background:#3060c0;border:4px solid #102030;border-radius:24px;corner-shape:` +
		shape + `;` + extra + `"></div></body>`
}

// checkPathGolden is the shaped-box golden: a bevel box must paint its fill,
// its border, the cut corners, and the page around it at exact colours. The
// samples are picked away from the antialiased edges.
func checkPathGolden() error {
	display, err := render.DisplayList(context.Background(), pathBoxStyle("#ffffff", "bevel", ""), pathPageWidth, pathPageHeight)
	if err != nil {
		return err
	}

	if !render.Replayable(display) {
		return fmt.Errorf("golden corner-shape page should replay")
	}

	img := newParityImage(pathPageWidth, pathPageHeight)
	defer img.Deallocate()

	DrawVisible(img, display, 0, 0)

	pixels := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	img.ReadPixels(pixels)

	samples := []struct {
		name string
		x, y int
		want [4]uint8
	}{
		{"fill centre", 84, 64, [4]uint8{48, 96, 192, 255}},
		{"fill under the top border", 84, 26, [4]uint8{48, 96, 192, 255}},
		{"top border", 84, 20, [4]uint8{16, 32, 48, 255}},
		{"left border", 20, 84, [4]uint8{16, 32, 48, 255}},
		{"bevel stroke", 32, 32, [4]uint8{16, 32, 48, 255}},
		{"inside the bevel", 36, 36, [4]uint8{48, 96, 192, 255}},
		{"outside the bevel", 28, 28, [4]uint8{255, 255, 255, 255}},
		{"right of the box", 160, 64, [4]uint8{255, 255, 255, 255}},
		{"below the box", 84, 120, [4]uint8{255, 255, 255, 255}},
		{"page corner", 4, 4, [4]uint8{255, 255, 255, 255}},
	}

	for _, sample := range samples {
		if err := checkPathPixel(pixels, sample.x, sample.y, sample.want); err != nil {
			return fmt.Errorf("%s: %w", sample.name, err)
		}
	}

	return nil
}

// checkPathXform draws a rotated scoop box twice, requires identical pixels,
// then requires the ink to leave the upright box (the matrix applied) and to
// stay inside the painted bounds (the matrix bounded).
func checkPathXform() error {
	display, err := render.DisplayList(context.Background(), pathBoxStyle("#ffffff", "scoop", "transform:rotate(15deg)"), pathPageWidth, pathPageHeight)
	if err != nil {
		return err
	}

	if !render.Replayable(display) {
		return fmt.Errorf("transformed corner-shape page should replay")
	}

	before, after := newParityImage(pathPageWidth, pathPageHeight), newParityImage(pathPageWidth, pathPageHeight)
	defer before.Deallocate()
	defer after.Deallocate()

	for _, index := range display.Order {
		drawOp(before, &display.Ops[index], 0, 0)
	}

	DrawVisible(after, display, 0, 0)

	one := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	two := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	before.ReadPixels(one)
	after.ReadPixels(two)

	if !bytes.Equal(one, two) {
		return fmt.Errorf("transformed path differs in %d channels: %s",
			countDiff(one, two), diffPixels(one, two, pathPageWidth))
	}

	op := findPathOp(display, false)
	if op == nil {
		return fmt.Errorf("transformed page has no fill path")
	}

	ppt := display.PointsPerPixel
	upright := image.Rect(
		int(math.Round(op.X/ppt)), int(math.Round(op.Y/ppt)),
		int(math.Round((op.X+op.W)/ppt)), int(math.Round((op.Y+op.H)/ppt)),
	)

	limit, ok := PaintBounds(op, ppt)
	if !ok || limit.Empty() {
		return fmt.Errorf("transformed path is unbounded")
	}

	limit = limit.Inset(-pathSlack)

	ink, outside, stray := 0, 0, 0

	for y := range pathPageHeight {
		for x := range pathPageWidth {
			at := (y*pathPageWidth + x) * pathChannels
			if two[at] == 255 && two[at+1] == 255 && two[at+2] == 255 {
				continue
			}

			ink++

			point := image.Pt(x, y)
			if !point.In(upright) {
				outside++
			}

			if !point.In(limit) {
				stray++
			}
		}
	}

	if ink == 0 {
		return fmt.Errorf("transformed path painted no ink")
	}

	if outside == 0 {
		return fmt.Errorf("a 15 degree rotation left no ink outside the upright box; the matrix was not applied")
	}

	if stray > 0 {
		return fmt.Errorf("transformed path painted %d pixels outside %v", stray, limit)
	}

	return nil
}

// checkPathGroup puts the shaped box inside a multiply group and requires the
// full and partial draws to agree inside the repainted rect, the composite to
// show the multiplied fill, and the cut corner to leave the backdrop alone.
func checkPathGroup() error {
	display, err := render.DisplayList(context.Background(), pathBoxStyle("#d2d2d2", "bevel", "mix-blend-mode:multiply"), pathPageWidth, pathPageHeight)
	if err != nil {
		return err
	}

	if !render.Replayable(display) {
		return fmt.Errorf("corner-shape group should replay")
	}

	full, part := newParityImage(pathPageWidth, pathPageHeight), newParityImage(pathPageWidth, pathPageHeight)
	defer full.Deallocate()
	defer part.Deallocate()

	DrawVisible(full, display, 0, 0)

	rect := image.Rect(40, 40, 120, 100)
	DrawRect(part, display, rect, 0, 0)

	one := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	two := make([]byte, pathPageWidth*pathPageHeight*pathChannels)
	full.ReadPixels(one)
	part.ReadPixels(two)

	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		from, to := (y*pathPageWidth+rect.Min.X)*pathChannels, (y*pathPageWidth+rect.Max.X)*pathChannels
		if !bytes.Equal(one[from:to], two[from:to]) {
			return fmt.Errorf("shaped group partial repaint differs on row %d in %d channels",
				y, countDiff(one[from:to], two[from:to]))
		}
	}

	// multiply: #3060c0 over #d2d2d2, channel by channel.
	if err := checkPathPixel(one, 84, 64, [4]uint8{40, 79, 158, 255}); err != nil {
		return fmt.Errorf("multiplied fill: %w", err)
	}

	// The bevel cut is outside the shaped layer, so the group leaves the
	// backdrop untouched there.
	if err := checkPathPixel(one, 28, 28, [4]uint8{210, 210, 210, 255}); err != nil {
		return fmt.Errorf("backdrop under the cut corner: %w", err)
	}

	return checkPathPixel(one, 4, 4, [4]uint8{210, 210, 210, 255})
}

// checkPathPixel compares one pixel against a wanted colour within the test
// slack, which covers one step of shader rounding.
func checkPathPixel(pixels []byte, x, y int, want [4]uint8) error {
	at := (y*pathPageWidth + x) * pathChannels

	for index, channel := range want {
		if delta := int(pixels[at+index]) - int(channel); delta < -pathSlack || delta > pathSlack {
			return fmt.Errorf("pixel (%d,%d): got %v, want %v", x, y, pixels[at:at+pathChannels], want)
		}
	}

	return nil
}

// findPathOp returns the first corner path op of a display, by fill or stroke,
// or nil when the display carries none.
func findPathOp(display *layout.Display, stroke bool) *layout.DisplayOp {
	for index := range display.Ops {
		op := &display.Ops[index]
		if op.Kind != layout.DisplayOpPath {
			continue
		}

		if corner, ok := op.CornerPathValue(); ok && corner != nil && corner.Stroke == stroke {
			return op
		}
	}

	return nil
}

// checkXformBitmap samples the mapped center of each transformed fill against
// the bitmap painter, away from the antialiased edges.
func checkXformBitmap(d *layout.Display, img *ebiten.Image) error {
	want, err := bitmap.Paint(d)
	if err != nil {
		return err
	}

	pixels := make([]byte, 240*140*4)
	img.ReadPixels(pixels)
	ppt := d.PixelPerPoint

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != layout.DisplayOpFillRect || !op.XformSet {
			continue
		}

		m := op.Transform()
		cx, cy := m.Apply(op.X+op.W/2, op.Y+op.H/2)
		x, y := int(cx*ppt), int(cy*ppt)

		if err := closePixel(pixels, want, 240, x, y); err != nil {
			return fmt.Errorf("transformed fill center: %w", err)
		}
	}

	return closePixel(pixels, want, 240, 232, 6)
}

// checkXformRepeat draws the page twice and requires bit-identical frames, so
// the cached raster blits stay deterministic across draws.
func checkXformRepeat(d *layout.Display) error {
	one, two := newParityImage(240, 140), newParityImage(240, 140)
	defer one.Deallocate()
	defer two.Deallocate()

	for _, img := range []*ebiten.Image{one, two} {
		for i := range d.Order {
			drawOp(img, &d.Ops[i], 0, 0)
		}
	}

	a, b := make([]byte, 240*140*4), make([]byte, 240*140*4)
	one.ReadPixels(a)
	two.ReadPixels(b)

	if !bytes.Equal(a, b) {
		return fmt.Errorf("repeated draw differs in %d channels: %s", countDiff(a, b), diffPixels(a, b, 240))
	}

	return nil
}

// checkXformParity draws a transformed page through drawOp and DrawVisible
// and requires identical pixels, then checks fills against the bitmap and
// the rotated run's placement.
func checkXformParity() error {
	source := `<body style="margin:0;height:140px;background:#f4f4f4">` +
		`<div style="position:absolute;left:20px;top:20px;width:100px;height:50px;background:#2860c8;transform:rotate(25deg)"></div>` +
		`<div style="position:absolute;left:150px;top:20px;width:60px;height:30px;background:#2ca04a;transform:skewX(20deg) scale(1.2,1)"></div>` +
		`<div style="position:absolute;left:20px;top:80px;width:90px;height:36px;border:3px solid #204020;transform:rotate(-15deg)"></div>` +
		`</body>`

	d, err := render.DisplayList(context.Background(), source, 240, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("transformed page should replay")
	}

	if err := checkXformRepeat(d); err != nil {
		return err
	}

	before, after := newParityImage(240, 140), newParityImage(240, 140)
	defer before.Deallocate()
	defer after.Deallocate()

	for i := range d.Order {
		drawOp(before, &d.Ops[i], 0, 0)
	}

	DrawVisible(after, d, 0, 0)

	a, b := make([]byte, 240*140*4), make([]byte, 240*140*4)
	before.ReadPixels(a)
	after.ReadPixels(b)

	if !bytes.Equal(a, b) {
		return fmt.Errorf("transformed draw differs in %d channels", countDiff(a, b))
	}

	if err := checkXformBitmap(d, after); err != nil {
		return err
	}

	return checkXformText()
}

// checkXformText pins a rotated run's placement: its ink must lie inside the
// bounds the matrix maps and nowhere far outside them.
func checkXformText() error {
	source := `<body style="margin:0;height:140px;background:#ffffff">` +
		`<div style="position:absolute;left:40px;top:40px;transform:rotate(20deg)">` +
		`<span style="color:#c00000;font-size:20px">rotated run</span></div></body>`

	d, err := render.DisplayList(context.Background(), source, 240, 140)
	if err != nil {
		return err
	}

	if !render.Replayable(d) {
		return fmt.Errorf("transformed text should replay")
	}

	img := newParityImage(240, 140)
	defer img.Deallocate()

	DrawVisible(img, d, 0, 0)

	pixels := make([]byte, 240*140*4)
	img.ReadPixels(pixels)

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != layout.DisplayOpText || !op.XformSet {
			continue
		}

		box, ok := PaintBounds(op, d.PointsPerPixel)
		if !ok || box.Empty() {
			return fmt.Errorf("transformed text is unbounded")
		}

		outside := box.Inset(-8)
		ink, stray := 0, 0

		for y := range 140 {
			for x := range 240 {
				at := (y*240 + x) * 4
				if pixels[at] == 255 && pixels[at+1] == 255 && pixels[at+2] == 255 {
					continue
				}

				if image.Pt(x, y).In(box) {
					ink++
				} else if !image.Pt(x, y).In(outside) {
					stray++
				}
			}
		}

		if ink == 0 {
			return fmt.Errorf("rotated run painted no ink inside %v", box)
		}
		if stray > 0 {
			return fmt.Errorf("rotated run painted %d pixels outside %v", stray, box)
		}

		return nil
	}

	return fmt.Errorf("no transformed text op")
}
