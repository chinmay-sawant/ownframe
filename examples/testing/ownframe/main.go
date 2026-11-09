// Command ownframe opens the thousand-row invoice table built with the UI builder.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8147", "listen address for -web")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	debugAddr := flag.String("debug-addr", "", "serve window profiling on this loopback address (off by default; needs -tags ownframe_debug)")
	bench := flag.Bool("bench", false, "run headless benchmarks instead of opening a window")
	benchFile := flag.String("bench-file", "", "write benchmark results to this file as well as stdout")
	flag.Parse()

	if *bench {
		if err := runBench(*benchFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	app, err := newInvoiceApp()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	app.Page().Window().SetDevTools(*dev)
	app.Page().Window().SetPerf(*perf)
	if err := app.Page().Redraw(ctx); err != nil {
		log.Fatal(err)
	}

	if *web {
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *debugAddr != "" {
		if err := ownframe.RunWithOptions(ctx, app.Page(), ownframe.WindowOptions{DebugAddr: *debugAddr}); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

// benchCounts balances stable timing against a quick headless run.
// benchFilePerm restricts a written results file to owner read/write.
const (
	benchRowCap       = 14
	benchRedrawOps    = 200
	benchRebuildOps   = 50
	benchClickOps     = 30
	benchThemeOps     = 30
	benchPatchOps     = 30
	benchStyleOps     = 10
	benchColdOps      = 10
	benchResizeOps    = 10
	benchTypingOps    = 20
	benchSelectRow    = 500
	benchResizeWidth  = 1200
	benchResizeHeight = 800
	benchFilePerm     = 0o600
)

// Nanosecond conversions for the derived per-frame rates.
const (
	nanosPerSecond      = 1e9
	nanosPerMillisecond = 1e6
)

// benchResult is one headless measurement row: wall time per op (headless, so
// wall time is CPU time), derived FPS, memory deltas, and retained heap.
type benchResult struct {
	name       string
	ops        int
	nsPerOp    float64
	fps        float64
	msPerFrame float64
	bytesPerOp uint64
	allocsOp   uint64
}

// measure runs fn n times and reports wall time plus MemStats deltas.
// Headless with no vsync or GPU wait, wall time per op is CPU cost per frame.
func measure(name string, n int, fn func() error) (benchResult, error) {
	if n <= 0 {
		return benchResult{}, fmt.Errorf("measure %s: ops must be positive, got %d", name, n)
	}
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	for i := 0; i < n; i++ {
		if err := fn(); err != nil {
			return benchResult{}, err
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	count := uint64(n)
	ns := float64(elapsed.Nanoseconds()) / float64(n)
	return benchResult{
		name:       name,
		ops:        n,
		nsPerOp:    ns,
		fps:        nanosPerSecond / ns,
		msPerFrame: ns / nanosPerMillisecond,
		bytesPerOp: (after.TotalAlloc - before.TotalAlloc) / count,
		allocsOp:   (after.Mallocs - before.Mallocs) / count,
	}, nil
}

// memoryDelta returns after-before as an int64. runtime.MemStats counters are
// byte and object counts, so they stay below math.MaxInt64 on any real
// process; the saturation keeps the conversion checked instead of assuming it.
func memoryDelta(after, before uint64) int64 {
	if after >= before {
		return clampCount(after - before)
	}
	return -clampCount(before - after)
}

// clampCount narrows a non-negative counter delta to int64, saturating at
// math.MaxInt64.
func clampCount(v uint64) int64 {
	if v > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(v)
}

// runBench executes the headless benchmark set without opening a window and
// prints one line per row, also writing benchFile when set.
func runBench(benchFile string) error {
	ctx := context.Background()
	app, err := newInvoiceApp()
	if err != nil {
		return err
	}
	app.patch = false
	app.Page().Window().SetPerf(true)
	if err := app.Page().Redraw(ctx); err != nil {
		return err
	}
	rows := make([]benchResult, 0, benchRowCap)
	render, err := measure("render-html-only", benchRebuildOps, func() error {
		_, err := ui.Render(app.screen())
		return err
	})
	if err != nil {
		return err
	}
	rows = append(rows, render)
	redraw, err := measure("redraw-same-data", benchRedrawOps, func() error {
		return app.Page().Redraw(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, redraw)
	rebuild, err := measure("rebuild-same-data", benchRebuildOps, func() error {
		return app.refresh(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, rebuild)
	click, err := measure("click-select-rebuild", benchClickOps, func() error {
		if app.selected == benchSelectRow {
			return pagetest.ClickAction(ctx, app.Page(), "row:501")
		}
		return pagetest.ClickAction(ctx, app.Page(), "row:500")
	})
	if err != nil {
		return err
	}
	rows = append(rows, click)
	theme, err := measure("theme-toggle-rebuild", benchThemeOps, func() error {
		return pagetest.ClickAction(ctx, app.Page(), "theme")
	})
	if err != nil {
		return err
	}
	rows = append(rows, theme)
	patchApp, err := newInvoiceApp()
	if err != nil {
		return err
	}
	patchApp.patch = true
	if err := patchApp.Page().Redraw(ctx); err != nil {
		return err
	}
	clickPatch, err := measure("click-select-patch", benchPatchOps, func() error {
		if patchApp.selected == benchSelectRow {
			return pagetest.ClickAction(ctx, patchApp.Page(), "row:501")
		}
		return pagetest.ClickAction(ctx, patchApp.Page(), "row:500")
	})
	if err != nil {
		return err
	}
	rows = append(rows, clickPatch)
	themePatch, err := measure("theme-toggle-patch", benchPatchOps, func() error {
		return pagetest.ClickAction(ctx, patchApp.Page(), "theme")
	})
	if err != nil {
		return err
	}
	rows = append(rows, themePatch)
	styleSwap, err := measure("theme-toggle-style", benchStyleOps, func() error {
		return app.styleThemeSwap(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, styleSwap)
	windowApp, err := newInvoiceApp()
	if err != nil {
		return err
	}
	windowApp.setWindowed(0)
	if err := windowApp.Page().Redraw(ctx); err != nil {
		return err
	}
	if err := windowApp.refresh(ctx); err != nil {
		return err
	}
	windowed, err := measure("windowed-rebuild", benchRebuildOps, func() error {
		return windowApp.refresh(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, windowed)
	// skipClean runs two refresh calls back to back with no mutation between
	// them. The second call is a no-change refresh, so it covers the
	// render-skip path once app.go provides it.
	skipClean, err := measure("skip-clean-refresh", benchRebuildOps, func() error {
		if err := app.refresh(ctx); err != nil {
			return err
		}
		return app.refresh(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, skipClean)
	cold, err := measure("cold-build", benchColdOps, func() error {
		coldApp, err := newInvoiceApp()
		if err != nil {
			return err
		}
		return coldApp.Page().Redraw(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, cold)
	resize, err := measure("resize", benchResizeOps, func() error {
		app.Page().SetSize(benchResizeWidth, benchResizeHeight)
		if err := app.Page().Redraw(ctx); err != nil {
			return err
		}
		app.Page().SetSize(invoiceWidth, invoiceHeight)
		return app.Page().Redraw(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, resize)
	if err := pagetest.Click(ctx, app.Page(), "invoice-filter"); err != nil {
		return err
	}
	typed := false
	typing, err := measure("typing", benchTypingOps, func() error {
		if !typed {
			typed = true
			return pagetest.Type(ctx, app.Page(), "a")
		}
		typed = false
		return app.Page().Edit().Backspace(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, typing)
	runtime.GC()
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := 0; i < 10; i++ {
		if err := pagetest.ClickAction(ctx, app.Page(), "row:500"); err != nil {
			return err
		}
		if err := pagetest.ClickAction(ctx, app.Page(), "row:501"); err != nil {
			return err
		}
	}
	runtime.GC()
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	st := app.Page().Probe().Stats()
	htmlBytes := len(app.Page().HTML())
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%-20s ops=%-4d ns/op=%-12.1f ms/frame=%-8.3f fps=%-10.1f B/op=%-10d allocs/op=%d\n",
			r.name, r.ops, r.nsPerOp, r.msPerFrame, r.fps, r.bytesPerOp, r.allocsOp)
	}
	out += fmt.Sprintf("stages template=%s style=%s layout=%s paint=%s other=%s last-redraw=%s\n",
		st.LastTemplate, st.StyleTime, st.LayoutTime, st.PaintTime, st.OtherTime, st.LastRedraw)
	out += fmt.Sprintf("sizes rows=%d grand-total=%s html-bytes=%d boxes=%d ops=%d heap-alloc=%d heap-sys=%d\n",
		invoiceRows, formatCents(app.grandTotal()), htmlBytes, st.Boxes, st.Ops, mem.HeapAlloc, mem.HeapSys)
	out += fmt.Sprintf("heap rows=10-rebuilds heap-retained=%d heap-objects=%d heap-growth=%d growth-objects=%d\n",
		after.HeapAlloc, after.HeapObjects,
		memoryDelta(after.HeapAlloc, before.HeapAlloc), memoryDelta(after.HeapObjects, before.HeapObjects))
	fmt.Print(out)
	if benchFile != "" {
		if err := os.WriteFile(benchFile, []byte(out), benchFilePerm); err != nil {
			return err
		}
	}
	return nil
}
