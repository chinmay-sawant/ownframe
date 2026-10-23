package replay

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
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
