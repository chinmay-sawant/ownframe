package main

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestCamera(t *testing.T) {
	c := camera{yaw: 0.7, pitch: 0.3, dist: 20}
	eye, right, up, forward := c.axes()

	d := math.Sqrt(eye.dot(eye))
	if math.Abs(d-20) > 1e-9 {
		t.Errorf("eye distance = %.3f, want 20", d)
	}

	if math.Abs(forward.dot(eye.norm())+1) > 1e-9 {
		t.Error("camera does not point at the origin (black hole)")
	}

	if math.Abs(right.dot(up)) > 1e-9 || math.Abs(right.dot(forward)) > 1e-9 || up[1] <= 0 {
		t.Errorf("axes not orthonormal: right=%v, up=%v, fwd=%v", right, up, forward)
	}
}

func TestInput(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	yaw0, pitch0 := app.cam.yaw, app.cam.pitch

	// Simulate drag move
	if _, err := app.onDragStart(ctx, ownframe.Box{}); err != nil {
		t.Fatal(err)
	}
	if err := app.onDragMove(ctx, 120, 50); err != nil {
		t.Fatal(err)
	}
	if err := app.onDragEnd(ctx); err != nil {
		t.Fatal(err)
	}

	expectedYaw := yaw0 - 120*0.005
	if math.Abs(app.cam.yaw-expectedYaw) > 0.05 {
		t.Errorf("yaw = %.3f, want about %.3f", app.cam.yaw, expectedYaw)
	}
	if app.cam.pitch <= pitch0 {
		t.Errorf("pitch = %.3f, expected higher than %.3f", app.cam.pitch, pitch0)
	}

	// Step simulation until hint fades
	for i := 0; i < 15; i++ {
		app.step(0.1, false)
	}
	if app.hint != 0 {
		t.Errorf("hint = %.2f, expected faded to 0", app.hint)
	}

	// Scroll in past nearDist
	for i := 0; i < 40; i++ {
		_, _ = app.onWheel(ctx, 600, 380, 0, -200)
	}
	if math.Abs(app.cam.dist-nearDist) > 1e-6 {
		t.Errorf("dist = %.2f, want nearDist (%.2f)", app.cam.dist, nearDist)
	}
}

func TestQualityScaling(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}

	// 600 slow ticks should throttle quality to minQuality
	for i := 0; i < 600; i++ {
		app.step(1.0/20.0, false)
	}
	if app.quality != minQuality {
		t.Fatalf("quality after slow ticks = %.3f, want %.3f", app.quality, minQuality)
	}

	// 2400 fast ticks should recover quality to 1.0
	for i := 0; i < 2400; i++ {
		app.step(1.0/120.0, false)
	}
	if app.quality != 1.0 {
		t.Fatalf("quality after fast ticks = %.3f, want 1.0", app.quality)
	}
}

func TestActions(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Trigger "toggle-orbit"
	if err := pagetest.ClickAction(ctx, app.Page(), "toggle-orbit"); err != nil {
		t.Fatal(err)
	}
	if !app.orbitPaused {
		t.Fatal("expected orbitPaused = true after click")
	}

	// Trigger "toggle-quality"
	if err := pagetest.ClickAction(ctx, app.Page(), "toggle-quality"); err != nil {
		t.Fatal(err)
	}
	if !app.qualityLocked || app.quality != 1.0 {
		t.Fatal("expected qualityLocked = true with quality = 1.0")
	}

	// Move camera and trigger "reset"
	app.cam.dist = 30.0
	if err := pagetest.ClickAction(ctx, app.Page(), "reset"); err != nil {
		t.Fatal(err)
	}
	if app.cam.dist != restDist {
		t.Fatalf("dist after reset = %.2f, want restDist (%.2f)", app.cam.dist, restDist)
	}
}

func TestShadersCompile(t *testing.T) {
	s, err := compileShaders()
	if err != nil {
		t.Fatalf("shader compilation failed: %v", err)
	}
	if s.hole == nil || s.bloom == nil || s.post == nil {
		t.Fatal("nil shader pointer in pipeline")
	}
}

func TestSceneRenderTick(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	s, err := compileShaders()
	if err != nil {
		t.Fatalf("shader compilation failed: %v", err)
	}
	r := &sceneRenderer{app: app, shaders: s}
	ctx := context.Background()

	// Full quality: every DrawRectShader source matches its rectangle,
	// or the pass panics while recording.
	r.render()

	// Minimum quality after slow frames: exercises the downscaled sizes.
	for i := 0; i < 600; i++ {
		app.step(1.0/20.0, false)
	}
	if app.quality != minQuality {
		t.Fatalf("quality after slow ticks = %.3f, want %.3f", app.quality, minQuality)
	}
	r.render()

	// Encode, register, and redraw with synthetic pixels. ReadPixels needs
	// the game loop, so the loopback screenshot covers the real pixels.
	w, h := 64, 48
	r.pix = make([]byte, 4*w*h)
	for i := range r.pix {
		r.pix[i] = 0x80
	}
	if err := r.pushPixels(w, h, sceneSrcNameA); err != nil {
		t.Fatalf("push pixels failed: %v", err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw with scene image failed: %v", err)
	}
}

func TestSceneTickHeadlessFallback(t *testing.T) {
	app, err := newApp(false, false)
	if err != nil {
		t.Fatal(err)
	}
	s, err := compileShaders()
	if err != nil {
		t.Fatal(err)
	}
	r := &sceneRenderer{app: app, shaders: s}
	ctx := context.Background()

	// No Ebiten game loop runs in `go test`, so the first tick must recover
	// from the ReadPixels panic, serve the static frame, and keep ticking.
	if err := r.tick(ctx); err != nil {
		t.Fatalf("headless tick failed: %v", err)
	}
	if !r.staticOnce {
		t.Fatal("expected static fallback after headless tick")
	}
	if len(staticSceneJPEG()) == 0 {
		t.Fatal("static fallback JPEG is empty")
	}
	// A second tick must not panic either.
	if err := r.tick(ctx); err != nil {
		t.Fatalf("second headless tick failed: %v", err)
	}
}
