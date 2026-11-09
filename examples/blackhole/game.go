package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"hash/fnv"
	"image"
	"image/jpeg"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// sceneSrcNames identifies the two page images the frames alternate
// between. Each tick renders into one name and points the template at it,
// so the display list can never reuse a stale image payload: the name (and
// the bytes behind it) is always new.
const (
	sceneSrcNameA = "blackhole-scene-a"
	sceneSrcNameB = "blackhole-scene-b"
)

// sceneView is the template data for the page: just which scene image the
// img tag shows this frame.
type sceneView struct {
	SceneSrc string
}

// staticSceneJPEGBytes is the fallback frame for environments without an
// Ebiten game loop (the -web preview, headless tests): ReadPixels panics
// there, so the tick serves this still instead of dying. It is a CPU render
// of the same Kage math; regenerate it from a desktop BLACKHOLE_DUMP-style
// capture if the shaders change.
var (
	//go:embed scene_fallback.jpg
	staticSceneJPEGBytes []byte
)

func staticSceneJPEG() []byte { return staticSceneJPEGBytes }

// sceneJPEGQuality balances the per-frame encode cost against banding in
// the dark gradients. The dynamic resolution scaling already shrinks the
// frame when the loop runs slow, which shrinks the encode with it.
const sceneJPEGQuality = 85

// maxTickSeconds caps one tick's dt so a stalled loop does not jump the
// camera or the FPS average by a whole stall at once.
const maxTickSeconds = 0.1

// debugLogName is the fixed file this example appends debug traces to, and
// debugLogMode keeps the trace readable only by its owner.
const (
	debugLogName = "blackhole-debug.log"
	debugLogMode = 0o600
)

// sceneRenderer raymarches the black hole offscreen and pushes each frame
// into the page as a JPEG. It runs from the page tick, so desktop Run and
// web Serve share one path: render, SetImage, Redraw.
type sceneRenderer struct {
	app     *App
	shaders *shaderPipeline

	scene *ebiten.Image
	small *ebiten.Image
	glow  *ebiten.Image
	wide  *ebiten.Image
	out   *ebiten.Image

	// prev holds the previous pushed frame. ReadPixels can race
	// still-executing draws, so the tick renders into out but always reads
	// back prev: a full frame behind, guaranteed complete. The pair swaps
	// on every push.
	prev *ebiten.Image

	// renderW/renderH are the applied buffer sizes; see resize hysteresis
	// in render.
	renderW, renderH int

	// lastPush throttles scene pushes so the GPU upload queue is never
	// asked for more than ~10 fresh full-screen textures per second.
	lastPush time.Time

	// frames counts ticks; the scene image alternates every frame.
	frames uint64

	pix     []byte
	jpegBuf bytes.Buffer

	staticOnce bool

	// Consecutive tick failures. A single bad frame is skipped; a sustained
	// failure latches to the static frame instead of killing the window.
	failures int

	// resizeCooldown ticks down after every applied size change so DRS
	// flip-flopping between two qualities cannot thrash GPU buffers.
	resizeCooldown int

	lastQuality float64
	lastYaw     float64
	lastDist    float64

	lastOpLog time.Time

	last time.Time
}

// debugLogf appends one line to blackhole-debug.log next to the executable
// (or the temp dir when that is not writable). It is best-effort: logging
// must never break the frame loop.
func debugLogf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	debugLogOnce.Do(func() {
		dir := os.TempDir()
		if exe, err := os.Executable(); err == nil {
			if d := filepath.Dir(exe); d != "" {
				dir = d
			}
		}
		if f, err := os.OpenFile(filepath.Join(dir, debugLogName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, debugLogMode); err == nil { //nolint:gosec // fixed log name under the executable dir or os.TempDir; no user input reaches the path
			debugLogFile = f
			log.Printf("blackhole debug log: %s", f.Name())
		}
	})
	if debugLogFile == nil {
		return
	}
	_, _ = fmt.Fprintf(debugLogFile, "%s %s\n", time.Now().Format("15:04:05.000"), msg)
}

// debugLogOnce and debugLogFile hold one process-wide best-effort log
// handle: the first debugLogf call opens the file and every renderer
// shares it, so a per-renderer copy would open extra descriptors.
var (
	//nolint:gochecknoglobals // one process-wide debug log handle, opened once and shared by every renderer
	debugLogOnce sync.Once
	//nolint:gochecknoglobals // one process-wide debug log handle, opened once and shared by every renderer
	debugLogFile *os.File
)

func (r *sceneRenderer) tick(ctx context.Context) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			r.noteFailure(fmt.Sprintf("panic: %v\n%s", rec, debug.Stack()))
			err = nil
		}
	}()

	now := time.Now()
	dt := 0.0
	if !r.last.IsZero() {
		dt = min(now.Sub(r.last).Seconds(), maxTickSeconds)
		if gap := now.Sub(r.last); gap > 100*time.Millisecond {
			debugLogf("tick gap %.0fms (loop stall, not scene)", float64(gap.Milliseconds()))
		}
	}
	r.last = now

	r.app.step(ctx, dt, false)
	// Alternate the image identity every push: the template re-executes
	// with a new src, so each commit builds a fresh display list around
	// fresh bytes instead of reusing a possibly stale payload. Pushes are
	// paced to at most every 100ms: a fresh full-screen texture every tick
	// overwhelms the GPU upload queue on some drivers, and the painted
	// frame then lags or blanks while the HUD keeps ticking at full rate.
	if now.Sub(r.lastPush) >= 100*time.Millisecond {
		r.lastPush = now
		src := sceneSrcNameA
		if r.frames&1 == 1 {
			src = sceneSrcNameB
		}
		r.frames++
		if !r.staticOnce {
			r.render()
		}

		// ReadPixels needs the game loop; the headless unit test covers
		// render instead and checks the push plumbing with synthetic pixels.
		// Push reads prev (last tick's finished frame) so the read never
		// races still-executing draws; out and prev swap every push.
		if err := r.pushImage(ctx, src); err != nil {
			r.noteFailure(err.Error())
		} else {
			r.app.Page().SetData(sceneView{SceneSrc: src})
			r.noteCamera()
		}
		r.out, r.prev = r.prev, r.out
	}

	// Redraw picks up the new scene image; HUD patches ride the same
	// commit. (RedrawIfChanged is avoided here: its unchanged-path skip
	// can leave a fresh scene stranded off the display list.)
	if err := r.app.Page().Redraw(ctx); err != nil {
		r.noteFailure("redraw: " + err.Error())
	}
	if time.Since(r.lastOpLog) >= time.Second {
		r.lastOpLog = time.Now()
		r.logSceneOp()
	}

	return nil
}

// bytesPerMiB converts the byte counters from runtime.MemStats to mebibytes
// for the debug log line.
const bytesPerMiB = 1 << 20

// runtime.MemStats.PauseNs is a ring of pauseRing entries whose next write
// slot is NumGC, so the last completed pause sits at the mask one behind it.
const (
	pauseRing     = 256
	pauseRingMask = pauseRing - 1
)

// logSceneOp records the live display's scene image op bounds and payload
// hash plus heap and GC pause stats once a second. Consecutive differing
// hashes prove fresh bytes reach the display; heap growth or long pauses
// coinciding with a freeze point at memory pressure instead.
func (r *sceneRenderer) logSceneOp() {
	d := r.app.Page().Probe().Display()
	if d == nil {
		debugLogf("display: nil")
		return
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	debugLogf("mem heap=%.1fMB total=%.1fMB gc=%d pausems=%.2f",
		float64(m.HeapAlloc)/bytesPerMiB, float64(m.TotalAlloc)/bytesPerMiB,
		m.NumGC, float64(m.PauseNs[(m.NumGC+pauseRingMask)%pauseRing])/float64(time.Millisecond))
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.IsJPEG {
			data, _, _ := op.ImageBytes()
			h := fnv.New64a()
			_, _ = h.Write(data)
			debugLogf("scene op x=%.0f y=%.0f w=%.0f h=%.0f bytes=%d hash=%x",
				op.X, op.Y, op.W, op.H, len(data), h.Sum64())
			return
		}
	}
	debugLogf("scene op: none")
}

// noteFailure counts a bad frame and latches to the static frame when the
// failure sustains, so a transient GPU hiccup blanks at most a few frames
// instead of closing the window.
func (r *sceneRenderer) noteFailure(reason string) {
	r.failures++
	debugLogf("frame failure %d: %s", r.failures, firstLine(reason))
	if r.failures == 30 && !r.staticOnce {
		r.staticOnce = true
		r.app.Page().SetImage(sceneSrcNameA, staticSceneJPEG())
		r.app.Page().SetImage(sceneSrcNameB, staticSceneJPEG())
		r.app.Page().SetData(sceneView{SceneSrc: sceneSrcNameA})
		debugLogf("latched to static scene after 30 consecutive failures")
	}
}

// noteCamera resets the failure count on a good frame and logs camera and
// quality changes, so a drag/zoom-correlated failure shows up in the log.
func (r *sceneRenderer) noteCamera() {
	if r.failures != 0 {
		debugLogf("recovered after %d failures", r.failures)
	}
	r.failures = 0
	q, yaw, dist := r.app.quality, r.app.cam.yaw, r.app.cam.dist
	if q != r.lastQuality || dist != r.lastDist || math.Abs(yaw-r.lastYaw) > 0.05 {
		debugLogf("camera quality=%.2f yaw=%.3f dist=%.2f", q, yaw, dist)
		r.lastQuality, r.lastYaw, r.lastDist = q, yaw, dist
	}
}

func firstLine(s string) string {
	for i, c := range s {
		if c == '\n' {
			return s[:i]
		}
	}
	return s
}

// render raymarches the frame into r.out at the current quality scale.
// Recording the passes validates the shader sources against their
// rectangles even without a display.
func (r *sceneRenderer) render() {
	sw, sh := r.frameSize()
	if sw <= 0 || sh <= 0 {
		return
	}

	w := max(1, int(math.Round(float64(sw)*r.app.quality)))
	h := max(1, int(math.Round(float64(sh)*r.app.quality)))

	// Resize hysteresis: GPU buffers are expensive to replace, so a new
	// size only applies after it survives 60 renders (~4s at the throttled
	// scene rate) or differs by more than a quarter. This stops DRS
	// flip-flopping from reallocating every frame.
	if r.resizeCooldown > 0 {
		r.resizeCooldown--
	}
	if r.renderW == 0 || r.renderH == 0 {
		r.renderW, r.renderH, r.resizeCooldown = w, h, 0
	} else if w != r.renderW || h != r.renderH {
		big := w >= r.renderW*5/4 || w*5/4 <= r.renderW || h >= r.renderH*5/4 || h*5/4 <= r.renderH
		if r.resizeCooldown <= 0 || big {
			r.renderW, r.renderH = w, h
			r.resizeCooldown = 60
			debugLogf("resize scene to %dx%d", w, h)
		}
	}
	w, h = r.renderW, r.renderH

	r.scene = ensureImageSize(r.scene, w, h)
	gw, gh := max(1, w/glowScale), max(1, h/glowScale)
	r.glow = ensureImageSize(r.glow, gw, gh)
	r.small = ensureImageSize(r.small, gw, gh)
	r.wide = ensureImageSize(r.wide, w, h)
	r.out = ensureImageSize(r.out, w, h)
	r.prev = ensureImageSize(r.prev, w, h)

	eye, right, up, forward := r.app.cam.axes()

	// Pass 1: Relativistic Raymarching into r.scene
	holeOpts := &ebiten.DrawRectShaderOptions{
		Uniforms: map[string]any{
			"Time":    float32(r.app.time),
			"Focal":   float32(focal),
			"Eye":     eye.f32(),
			"Right":   right.f32(),
			"Up":      up.f32(),
			"Forward": forward.f32(),
		},
	}
	r.scene.DrawRectShader(w, h, r.shaders.hole, holeOpts)

	// Downsample the scene to glow size: the bloom shader's source must
	// match its rectangle.
	downOpt := &ebiten.DrawImageOptions{}
	downOpt.GeoM.Scale(float64(gw)/float64(w), float64(gh)/float64(h))
	downOpt.Filter = ebiten.FilterLinear
	r.small.DrawImage(r.scene, downOpt)

	// Pass 2: Downsampled Bloom Extraction & Vogel Spiral Blur into r.glow
	bloomOpts := &ebiten.DrawRectShaderOptions{
		Images: [4]*ebiten.Image{r.small},
	}
	r.glow.DrawRectShader(gw, gh, r.shaders.bloom, bloomOpts)

	// Upsample the glow to scene size: the post shader's sources must
	// match the output rectangle. The shader taps are scaled by glowScale
	// in post.kage to keep the original quarter-res spread.
	upOpt := &ebiten.DrawImageOptions{}
	upOpt.GeoM.Scale(float64(w)/float64(gw), float64(h)/float64(gh))
	upOpt.Filter = ebiten.FilterLinear
	r.wide.DrawImage(r.glow, upOpt)

	// Pass 3: Post-processing Composition into r.out
	postOpts := &ebiten.DrawRectShaderOptions{
		Images: [4]*ebiten.Image{r.scene, r.wide},
		Uniforms: map[string]any{
			"Time": float32(r.app.time),
		},
	}
	r.out.DrawRectShader(w, h, r.shaders.post, postOpts)
}

// frameSize prefers the live window size and falls back to the page size
// when there is no window, such as the headless tests and the web preview.
func (r *sceneRenderer) frameSize() (int, int) {
	if w, h := ebiten.WindowSize(); w > 0 && h > 0 {
		return w, h
	}
	return r.app.Page().Size()
}

// bytesPerPixel is the RGBA byte width of one readback pixel.
const bytesPerPixel = 4

// pushImage reads the composed frame back and registers it under src.
// The pixel and JPEG buffers are reused across frames; SetImage clones the
// bytes, so reuse is safe. Without an Ebiten game loop (the -web preview and
// headless tests) ReadPixels panics; push recovers once and falls back to a
// still frame so the tick - and the HUD - stay live.
func (r *sceneRenderer) pushImage(ctx context.Context, src string) (err error) {
	defer func() {
		if rec := recover(); rec != nil && !r.staticOnce {
			r.staticOnce = true
			r.app.Page().SetImage(sceneSrcNameA, staticSceneJPEG())
			r.app.Page().SetImage(sceneSrcNameB, staticSceneJPEG())
			r.app.Page().SetData(sceneView{SceneSrc: sceneSrcNameA})
			err = nil
		}
	}()
	if r.staticOnce {
		return nil
	}
	_ = ctx
	b := r.prev.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= 0 || h <= 0 {
		return nil
	}
	if need := bytesPerPixel * w * h; len(r.pix) != need {
		r.pix = make([]byte, need)
	}
	r.prev.ReadPixels(r.pix)
	if blankFrame(r.pix) {
		// The readback came back empty (no GL error is raised for this),
		// so keep the last good frame instead of pushing black.
		debugLogf("blank readback skipped buf=%p size=%dx%d", r.prev, w, h)
		r.noteFailure("blank readback skipped")
		return nil
	}
	return r.pushPixels(w, h, src)
}

// blankFrame reports a fully zero pixel buffer. A real raymarched frame
// always carries post grain, so all-zero means the GPU readback failed
// rather than the scene went dark.
func blankFrame(pix []byte) bool {
	for i := 0; i < len(pix); i += 4096 {
		if pix[i] != 0 {
			return false
		}
	}
	return true
}

// pushPixels JPEG-encodes the pixels in r.pix and registers them under src.
// It is split from push so the headless test can cover the encode,
// register, and redraw path with synthetic pixels; ReadPixels needs the
// game loop.
func (r *sceneRenderer) pushPixels(w, h int, src string) error {
	img := &image.RGBA{
		Pix:    r.pix,
		Stride: bytesPerPixel * w,
		Rect:   image.Rect(0, 0, w, h),
	}
	r.jpegBuf.Reset()
	if err := jpeg.Encode(&r.jpegBuf, img, &jpeg.Options{Quality: sceneJPEGQuality}); err != nil {
		return err
	}
	r.app.Page().SetImage(src, r.jpegBuf.Bytes())
	return nil
}

func ensureImageSize(img *ebiten.Image, w, h int) *ebiten.Image {
	if img != nil {
		b := img.Bounds()
		if b.Dx() == w && b.Dy() == h {
			return img
		}
	}
	return ebiten.NewImage(w, h)
}
