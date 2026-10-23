package page

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"slices"
	"strconv"
	"testing"
	"time"
)

type soakScenario struct {
	name               string
	page               *Page
	step               func(int)
	cycle, redraw, png []time.Duration
}

func soakP95(samples []time.Duration) float64 {
	if len(samples) == 0 {
		return 0
	}
	sorted := slices.Clone(samples)
	slices.Sort(sorted)
	return float64(sorted[(len(sorted)*95+99)/100-1]) / float64(time.Millisecond)
}

func soakLimit(t *testing.T, name string, fallback float64) float64 {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	limit, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(limit) || math.IsInf(limit, 0) || limit <= 0 || limit > 60000 {
		t.Fatalf("%s must be greater than 0 and at most 60000 ms", name)
	}
	return limit
}

// TestPageSoak drives page operations and software PNG encoding, not GPU replay.
func TestPageSoak(t *testing.T) {
	duration := os.Getenv("OWNFRAME_SOAK_SECONDS")
	if duration == "" {
		t.Skip("run make soak-check for the bounded page workload")
	}
	seconds, err := strconv.Atoi(duration)
	if err != nil || seconds < 1 || seconds > 1800 {
		t.Fatal("OWNFRAME_SOAK_SECONDS must be 1..1800")
	}
	redrawLimit := soakLimit(t, "OWNFRAME_SOAK_REDRAW_P95_MS", 16.67)
	pngLimit := soakLimit(t, "OWNFRAME_SOAK_PNG_P95_MS", 100)
	ctx := t.Context()
	scenarios := makeSoakScenarios(t, ctx)
	cycle := func(s *soakScenario, i int, measure bool) {
		t.Helper()
		start := time.Now()
		s.step(i)
		at := time.Now()
		if err := s.page.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		redraw := time.Since(at)
		at = time.Now()
		if len(s.page.PNGWithContext(ctx)) == 0 {
			t.Fatal("workload produced no picture")
		}
		if measure {
			s.png = append(s.png, time.Since(at))
			s.redraw = append(s.redraw, redraw)
			s.cycle = append(s.cycle, time.Since(start))
		}
	}
	for _, scenario := range scenarios {
		for i := range 10 {
			cycle(scenario, i, false)
		}
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	goroutines := runtime.NumGoroutine()
	start := time.Now()
	t.Cleanup(func() {
		runtime.GC()
		runtime.ReadMemStats(&after)
		commit := os.Getenv("OWNFRAME_SOAK_COMMIT")
		if commit == "" {
			commit = "unknown"
		}
		rows := make([]map[string]any, 0, len(scenarios))
		for _, scenario := range scenarios {
			rows = append(rows, map[string]any{
				"name": scenario.name, "cycles": len(scenario.cycle),
				"cycle_p95_ms": soakP95(scenario.cycle), "redraw_p95_ms": soakP95(scenario.redraw),
				"png_p95_ms": soakP95(scenario.png),
			})
		}
		status := "passed"
		if t.Failed() {
			status = "failed"
		}
		report := map[string]any{
			"status": status, "source_commit": commit, "go": runtime.Version(),
			"os": runtime.GOOS, "arch": runtime.GOARCH, "gomaxprocs": runtime.GOMAXPROCS(0),
			"seconds": time.Since(start).Seconds(), "scenarios": rows,
			"redraw_p95_limit_ms": redrawLimit, "png_p95_limit_ms": pngLimit,
			"heap_before": before.HeapAlloc, "heap_after": after.HeapAlloc,
			"goroutines_before": goroutines, "goroutines_after": runtime.NumGoroutine(),
		}
		encoded, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Error(err)
			return
		}
		t.Log(string(encoded))
	})
	deadline := start.Add(time.Duration(seconds) * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for i := 0; time.Now().Before(deadline); i++ {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
		cycle(scenarios[i%len(scenarios)], i, true)
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	if after.HeapAlloc > before.HeapAlloc+64*1024*1024 {
		t.Errorf("retained heap grew over 64 MiB: %d -> %d", before.HeapAlloc, after.HeapAlloc)
	}
	if got := runtime.NumGoroutine(); got > goroutines+4 {
		t.Errorf("goroutines grew: %d -> %d", goroutines, got)
	}
	for _, scenario := range scenarios {
		if got := soakP95(scenario.redraw); got > redrawLimit {
			t.Errorf("%s redraw p95 %.3f ms exceeds %.3f ms", scenario.name, got, redrawLimit)
		}
		if got := soakP95(scenario.png); got > pngLimit {
			t.Errorf("%s PNG p95 %.3f ms exceeds %.3f ms", scenario.name, got, pngLimit)
		}
	}
}

func makeSoakScenarios(t *testing.T, ctx context.Context) []*soakScenario {
	t.Helper()
	newPage := func(source string) *Page {
		t.Helper()
		p, err := New(Config{Width: 640, Height: 480, Perf: true, HTML: source})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	p := newPage(`<body style="margin:0"><button id="action">{{.Count}}</button><input id="name" value="seed"><div style="height:1600px">{{if .Visible}}<p id="optional">optional</p>{{end}}</div></body>`)
	view := struct {
		Count   int
		Visible bool
	}{}
	p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
		if box.ID == "action" {
			view.Count++
			p.SetData(view)
		}
		return nil
	}})
	interaction := &soakScenario{name: "resize/click/type/backspace/scroll/conditional", page: p}
	interaction.step = func(i int) {
		view.Visible = i%2 == 0
		p.SetData(view)
		p.SetSize(640+(i%3)*16, 480)
		p.SetScrollOffset(0, 0)
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		before := view.Count
		for _, box := range p.Boxes() {
			if box.ID == "action" {
				if err := p.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
					t.Fatal(err)
				}
				break
			}
		}
		if view.Count != before+1 {
			t.Fatal("workload click did not reach its handler")
		}
		if err := p.Focus(ctx, "name"); err != nil {
			t.Fatal(err)
		}
		if err := p.Type(ctx, "x"); err != nil {
			t.Fatal(err)
		}
		if got := p.FormValue("name"); got != "seedx" {
			t.Fatalf("typing produced %q", got)
		}
		if err := p.Backspace(ctx); err != nil {
			t.Fatal(err)
		}
		if got := p.FormValue("name"); got != "seed" {
			t.Fatalf("backspace produced %q", got)
		}
		p.ScrollBy(0, 20)
		p.ScrollBy(0, -20)
	}
	rich := newPage(`<style>p{font-size:18px;width:420px}</style><body><img src="soak-image" width="64" height="64"><p>{{.}} नमस्कार जग مرحبا بالعالم Ελληνικά ligatures office ffi</p></body>`)
	var encoded bytes.Buffer
	picture := image.NewRGBA(image.Rect(0, 0, 64, 64))
	picture.SetRGBA(20, 20, color.RGBA{R: 255, A: 255})
	if err := png.Encode(&encoded, picture); err != nil {
		t.Fatal(err)
	}
	rich.SetImage("soak-image", encoded.Bytes())
	themed := &soakScenario{name: "image/theme/shaped-text/tick", page: rich}
	tick := 0
	rich.SetTick(func(context.Context) error {
		tick++
		rich.SetData(tick)
		return nil
	})
	themed.step = func(i int) {
		themes := [2]string{"body{background:#fff;color:#222}", "body{background:#222;color:#fff}"}
		if err := rich.SetTheme(themes[i%2]); err != nil {
			t.Fatal(err)
		}
		if err := rich.Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}
	nav := newPage(`<p id="first">first</p>`)
	navigation := &soakScenario{name: "navigation", page: nav}
	navigation.step = func(int) {
		if err := nav.Load(ctx, `<p id="second">second</p>`); err != nil {
			t.Fatal(err)
		}
		if err := nav.Back(ctx); err != nil {
			t.Fatal(err)
		}
		if nav.HTML() != `<p id="first">first</p>` {
			t.Fatal("Back did not restore first page")
		}
		if err := nav.Forward(ctx); err != nil {
			t.Fatal(err)
		}
		if nav.HTML() != `<p id="second">second</p>` {
			t.Fatal("Forward did not restore second page")
		}
		if err := nav.Back(ctx); err != nil {
			t.Fatal(err)
		}
	}
	return []*soakScenario{interaction, themed, navigation}
}
