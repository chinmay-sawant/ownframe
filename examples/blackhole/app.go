package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const (
	defaultWidth  = 1200
	defaultHeight = 760

	minQuality  = 0.35
	qualityStep = 0.125
	slowFrame   = 1.0 / 50.0 // 20ms
	fastFrame   = 1.0 / 90.0 // 11.1ms
	glowScale   = 4
)

const appHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; user-select: none; }
  html, body {
    width: 100%; height: 100%;
    background-color: #000;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, Helvetica, Arial, sans-serif;
  }
  #scene {
    display: block;
    width: 100%; height: auto;
  }
  .hud-root {
    position: absolute; inset: 0;
    display: flex; flex-direction: column;
    justify-content: space-between;
    padding: 24px 28px 26px 30px;
    pointer-events: none;
  }
  .hud-top {
    display: flex; justify-content: space-between; align-items: flex-start;
  }
  .hud-title-col {
    display: flex; flex-direction: column; gap: 4px;
  }
  .hud-title {
    font-size: 22px; font-weight: 600; color: rgba(255, 255, 255, 0.92);
    letter-spacing: -0.01em;
  }
  .hud-subtitle {
    font-size: 12px; color: rgba(255, 255, 255, 0.5);
  }
  .hud-telemetry {
    pointer-events: auto;
    background: rgba(18, 18, 18, 0.65);
    border: 1px solid rgba(255, 255, 255, 0.12);
    border-radius: 8px;
    padding: 8px 14px;
    display: flex; gap: 16px; align-items: center;
    color: rgba(255, 255, 255, 0.85);
    font-size: 11px;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }
  .metric-item { display: flex; flex-direction: column; gap: 2px; }
  .metric-label { font-size: 9px; text-transform: uppercase; color: rgba(255, 255, 255, 0.4); }
  .metric-val { font-weight: 500; }
  .hud-bottom {
    display: flex; justify-content: space-between; align-items: flex-end;
  }
  .btn-bar {
    pointer-events: auto;
    display: flex; gap: 8px;
  }
  .btn-chip {
    background: rgba(255, 255, 255, 0.08);
    border: 1px solid rgba(255, 255, 255, 0.15);
    color: rgba(255, 255, 255, 0.85);
    border-radius: 6px;
    padding: 6px 12px;
    font-size: 12px;
    cursor: pointer;
  }
  .btn-chip:hover {
    background: rgba(255, 255, 255, 0.16);
  }
  .hint-text {
    font-size: 12px; color: rgba(255, 255, 255, 0.5);
    transition: opacity 0.4s ease;
  }
</style>
</head>
<body>
<img id="scene" src="{{.SceneSrc}}" width="1200" height="760" alt="">
<div class="hud-root">
  <div class="hud-top">
    <div class="hud-title-col">
      <div class="hud-title">Gargantua</div>
      <div class="hud-subtitle">A black hole, its light traced through gravity by shaders written in Go</div>
    </div>
    <div class="hud-telemetry" id="telemetry">
      <div class="metric-item">
        <span class="metric-label">FPS</span>
        <span class="metric-val" id="metric-fps">60</span>
      </div>
      <div class="metric-item">
        <span class="metric-label">Scale</span>
        <span class="metric-val" id="metric-scale">100%</span>
      </div>
      <div class="metric-item">
        <span class="metric-label">Distance</span>
        <span class="metric-val" id="metric-dist">22.0 rs</span>
      </div>
      <div class="metric-item">
        <span class="metric-label">Orbit</span>
        <span class="metric-val" id="metric-orbit">Active</span>
      </div>
    </div>
  </div>
  <div class="hud-bottom">
    <div class="btn-bar">
      <div class="btn-chip" id="btn-reset" data-action="reset">Reset Camera</div>
      <div class="btn-chip" id="btn-orbit" data-action="toggle-orbit">Pause Orbit</div>
      <div class="btn-chip" id="btn-lock" data-action="toggle-quality">Auto Quality</div>
    </div>
    <div class="hint-text" id="hint" style="opacity: 1.0;">Drag to orbit · Scroll to zoom</div>
  </div>
</div>
</body>
</html>
`

// App manages the black hole simulation state, camera navigation, and HUD overlays.
type App struct {
	page *ownframe.Page
	cam  camera

	time     float64
	lastTime time.Time

	down     bool
	lastX    float64
	lastY    float64
	lastMove time.Time
	touched  bool
	hint     float64

	quality       float64
	frame         float64
	slow          float64
	fast          float64
	qualityLocked bool
	orbitPaused   bool

	fpsCounter int
	currentFPS int
	fpsTimer   time.Time
}

func newApp(devtools, perf bool) (*App, error) {
	app := &App{
		cam:        newCamera(),
		hint:       1.0,
		quality:    1.0,
		frame:      1.0 / 60.0,
		currentFPS: 60,
		fpsTimer:   time.Now(),
	}

	p, err := ownframe.New(ownframe.Config{
		Title:       "Gargantua · Ownframe Black Hole",
		HTML:        appHTML,
		Width:       defaultWidth,
		Height:      defaultHeight,
		DevTools:    devtools,
		Perf:        perf,
		RetainState: true,
	})
	if err != nil {
		return nil, fmt.Errorf("create black hole page: %w", err)
	}

	app.page = p
	p.Handle(ownframe.Handlers{
		DragStart: app.onDragStart,
		DragMove:  app.onDragMove,
		DragEnd:   app.onDragEnd,
		Wheel:     app.onWheel,
		Click:     app.onClick,
	})

	// A 1x1 black frame until the first rendered scene lands. Both names
	// stay registered so the template can alternate between them; the
	// per-frame SetData below swaps which one the img shows.
	p.SetImage(sceneSrcNameA, blackJPEG())
	p.SetImage(sceneSrcNameB, blackJPEG())
	p.SetData(sceneView{SceneSrc: sceneSrcNameA})

	return app, nil
}

// blackJPEG renders a 1x1 black JPEG for the scene image placeholder.
func blackJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.Black)
	var buf bytes.Buffer
	_ = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85})
	return buf.Bytes()
}

func (a *App) Page() *ownframe.Page {
	return a.page
}

func (a *App) onDragStart(ctx context.Context, box ownframe.Box) (bool, error) {
	a.down = true
	a.touched = true
	a.lastMove = time.Now()
	a.cam.spinYaw = 0
	a.cam.spinPitch = 0
	return true, nil
}

func (a *App) onDragMove(ctx context.Context, dx, dy float64) error {
	if !a.down {
		return nil
	}
	now := time.Now()
	dt := max(now.Sub(a.lastMove).Seconds(), 1.0/240.0)

	dyaw := -dx * 0.005
	dpitch := dy * 0.004

	a.cam.yaw += dyaw
	a.cam.pitch = clamp(a.cam.pitch+dpitch, -maxPitch, maxPitch)

	k := min(1, dt/0.06)
	a.cam.spinYaw += (dyaw/dt - a.cam.spinYaw) * k
	a.cam.spinPitch += (dpitch/dt - a.cam.spinPitch) * k

	a.lastMove = now
	return nil
}

func (a *App) onDragEnd(ctx context.Context) error {
	a.down = false
	if time.Since(a.lastMove) > 80*time.Millisecond {
		a.cam.spinYaw = 0
		a.cam.spinPitch = 0
	}
	return nil
}

func (a *App) onWheel(ctx context.Context, x, y, dx, dy float64) (bool, error) {
	a.touched = true
	a.cam.dist = clamp(a.cam.dist*math.Exp(dy*0.003), nearDist, farDist)
	return true, nil
}

func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	switch box.Action {
	case "reset":
		a.cam = newCamera()
		a.quality = 1.0
	case "toggle-orbit":
		a.orbitPaused = !a.orbitPaused
		label := "Pause Orbit"
		if a.orbitPaused {
			label = "Resume Orbit"
		}
		_ = a.page.PatchElement(ctx, "btn-orbit", ownframe.ElementEdit{Text: &label})
	case "toggle-quality":
		a.qualityLocked = !a.qualityLocked
		label := "Auto Quality"
		if a.qualityLocked {
			label = "Locked 100%"
			a.quality = 1.0
		}
		_ = a.page.PatchElement(ctx, "btn-lock", ownframe.ElementEdit{Text: &label})
	}
	return nil
}

// step advances the camera, dampens motion, updates FPS, and adjusts quality scaling.
func (a *App) step(dt float64, reduceMotion bool) {
	if !reduceMotion && !a.orbitPaused {
		a.time += dt
	}
	a.cam.step(dt, a.down, reduceMotion || a.orbitPaused)

	if a.touched && a.hint > 0 {
		a.hint = max(0, a.hint-dt/0.6)
		hintStyle := fmt.Sprintf("opacity: %.2f;", a.hint)
		_ = a.page.PatchElement(context.Background(), "hint", ownframe.ElementEdit{
			Attribute: "style",
			Value:     hintStyle,
		})
	}

	if dt <= 0 {
		return
	}

	// FPS monitoring
	a.fpsCounter++
	if time.Since(a.fpsTimer) >= time.Second {
		a.currentFPS = a.fpsCounter
		a.fpsCounter = 0
		a.fpsTimer = time.Now()
		a.updateHUDMetrics()
	}

	if a.qualityLocked {
		a.quality = 1.0
		return
	}

	a.frame += (dt - a.frame) * 0.1
	if a.frame > slowFrame {
		a.slow += dt
	} else {
		a.slow = 0
	}
	if a.frame < fastFrame {
		a.fast += dt
	} else {
		a.fast = 0
	}

	switch {
	case a.slow > 0.5 && a.quality > minQuality:
		a.quality = max(minQuality, a.quality-qualityStep)
		a.slow, a.frame = 0, fastFrame
		a.updateHUDMetrics()
	case a.fast > 3.0 && a.quality < 1.0:
		a.quality = min(1.0, a.quality+qualityStep)
		a.fast = 0
		a.updateHUDMetrics()
	}
}

func (a *App) updateHUDMetrics() {
	ctx := context.Background()
	fpsText := fmt.Sprintf("%d", a.currentFPS)
	scaleText := fmt.Sprintf("%d%%", int(a.quality*100))
	distText := fmt.Sprintf("%.1f rs", a.cam.dist)
	orbitText := "Active"
	if a.orbitPaused {
		orbitText = "Paused"
	}

	_ = a.page.PatchElements(ctx, []ownframe.ElementPatch{
		{ID: "metric-fps", Edits: []ownframe.ElementEdit{{Text: &fpsText}}},
		{ID: "metric-scale", Edits: []ownframe.ElementEdit{{Text: &scaleText}}},
		{ID: "metric-dist", Edits: []ownframe.ElementEdit{{Text: &distText}}},
		{ID: "metric-orbit", Edits: []ownframe.ElementEdit{{Text: &orbitText}}},
	})
}
