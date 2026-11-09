package main

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"runtime"
	"testing"
	"weak"
)

func BenchmarkLoglensResize(b *testing.B) {
	a, err := newApp()
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.page.SetSize(windowWidth-(i%2)*160, windowHeight)
		if err := a.page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

func TestNavigationRetainedHeapPlateaus(t *testing.T) {
	a, err := newApp()
	if err != nil {
		t.Fatal(err)
	}
	a.page.Window().SetAnimations(false)
	a.page.Window().SetTransitions(false)
	ctx := t.Context()
	probes := []weak.Pointer[layout.Display]{}
	cycle := func() {
		for _, tab := range []string{"logs", "traces", "incident", "overview"} {
			if d := a.page.Probe().Display(); d != nil {
				probes = append(probes, weak.Make(d))
			}
			if err := a.setTab(ctx, tab); err != nil {
				t.Fatal(err)
			}
			a.page.Probe().TakeDirtyRects()
		}
	}
	for range 3 {
		cycle()
	}
	for range 3 {
		runtime.GC()
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for range 12 {
		cycle()
	}
	for range 3 {
		runtime.GC()
	}
	runtime.ReadMemStats(&after)
	alive := 0
	for _, probe := range probes {
		if probe.Value() != nil {
			alive++
		}
	}
	t.Logf("old displays still alive: %d/%d", alive, len(probes))
	if alive != 0 {
		t.Fatal("navigation retained obsolete displays")
	}
	runtime.KeepAlive(a)
	const allowance = 8 * 1024 * 1024
	t.Logf("post-GC heap: warm=%d after=%d bytes", before.HeapAlloc, after.HeapAlloc)
	if after.HeapAlloc > before.HeapAlloc+allowance {
		t.Fatalf("repeated navigation retained %d additional bytes; limit %d", after.HeapAlloc-before.HeapAlloc, allowance)
	}
}
