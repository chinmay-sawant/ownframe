package page

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"slices"
	"strconv"
	"testing"
	"time"
)

// TestPageSoak drives real page operations. It does not open a native window.
func TestPageSoak(t *testing.T) {
	duration := os.Getenv("OWNFRAME_SOAK_SECONDS")
	if duration == "" {
		t.Skip("run make soak-check for the bounded page workload")
	}
	seconds, err := strconv.Atoi(duration)
	if err != nil || seconds < 1 || seconds > 1800 {
		t.Fatal("OWNFRAME_SOAK_SECONDS must be 1..1800")
	}
	ctx := t.Context()
	p, err := New(Config{Width: 320, Height: 240, Perf: true, HTML: `<body style="margin:0"><div id="action" style="width:100px;height:30px;background:red">{{.Count}}</div><input id="name" value="seed"><div style="height:1600px">{{if .Visible}}<div id="optional">optional</div>{{end}}</div></body>`})
	if err != nil {
		t.Fatal(err)
	}
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
	cycle := func(i int) {
		t.Helper()
		view.Visible = i%2 == 0
		p.SetData(view)
		p.SetSize(320+(i%3)*16, 240)
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
		picture := p.PNGWithContext(ctx)
		if len(picture) == 0 {
			t.Fatal("workload produced no picture")
		}
	}
	for i := 0; i < 10; i++ {
		cycle(i)
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	goroutines := runtime.NumGoroutine()
	start := time.Now()
	deadline := start.Add(time.Duration(seconds) * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var latencies []time.Duration
	for i := 0; time.Now().Before(deadline); i++ {
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-ticker.C:
		}
		at := time.Now()
		cycle(i)
		latencies = append(latencies, time.Since(at))
	}
	runtime.GC()
	runtime.ReadMemStats(&after)
	endGoroutines := runtime.NumGoroutine()
	if after.HeapAlloc > before.HeapAlloc+64*1024*1024 {
		t.Fatalf("retained heap grew over 64 MiB: %d -> %d", before.HeapAlloc, after.HeapAlloc)
	}
	if endGoroutines > goroutines+4 {
		t.Fatalf("goroutines grew: %d -> %d", goroutines, endGoroutines)
	}
	slices.Sort(latencies)
	report := map[string]any{
		"workload":      "page resize/click/type/backspace/scroll/create-remove/PNG",
		"source_commit": os.Getenv("OWNFRAME_SOAK_COMMIT"), "go": runtime.Version(), "os": runtime.GOOS, "arch": runtime.GOARCH,
		"seconds": time.Since(start).Seconds(), "cycles": len(latencies), "cycle_p95_ms": float64(latencies[(len(latencies)-1)*95/100]) / float64(time.Millisecond),
		"heap_before": before.HeapAlloc, "heap_after": after.HeapAlloc, "goroutines_before": goroutines, "goroutines_after": endGoroutines,
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(encoded))
}
