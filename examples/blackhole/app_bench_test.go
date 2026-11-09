package main

import (
	"context"
	"runtime"
	"testing"
)

func BenchmarkCamera_Step(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.cam.step(1.0/60.0, false, false)
	}
}

func BenchmarkCamera_Axes(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.cam.yaw += 0.001
		_, _, _, _ = app.cam.axes()
	}
}

func BenchmarkAdaptive_Quality(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.step(ctx, 1.0/30.0, false)
	}
}

func BenchmarkHUD_Patch(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	_ = app.Page().Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.updateHUDMetrics(ctx)
	}
}

func BenchmarkPage_Testing_Tick(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	_ = app.Page().Redraw(ctx)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = app.Page().Testing().Tick(ctx)
	}
}

func BenchmarkHeapGrowth(b *testing.B) {
	app, err := newApp(false, false)
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	_ = app.Page().Redraw(ctx)

	runtime.GC()
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.step(ctx, initialFrame, false)
		_ = app.Page().Testing().Tick(ctx)
	}
	b.StopTimer()

	runtime.ReadMemStats(&m1)
	b.ReportMetric(float64(m1.Alloc-m0.Alloc)/float64(b.N), "heap-bytes/op")
}
