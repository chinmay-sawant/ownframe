// Command ownframe opens the compare thousand-row table built with the UI builder.
// Pass -web to serve the picture in a browser instead.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8146", "listen address for -web")
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

	app, err := newListApp()
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

// benchCounts are the headless benchmark sizes: rows collects one result
// per measurement, and each ops count balances stable timing against the
// headless run staying quick. benchFilePerm restricts a written results
// file to owner read/write.
const (
	benchRowCap     = 3
	benchRedrawOps  = 200
	benchRebuildOps = 100
	benchClickOps   = 30
	benchFilePerm   = 0o600
)

// benchResult is one headless measurement row.
type benchResult struct {
	name       string
	ops        int
	nsPerOp    float64
	bytesPerOp uint64
	allocsOp   uint64
}

// measure runs fn n times and reports time plus MemStats deltas.
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
	return benchResult{
		name:       name,
		ops:        n,
		nsPerOp:    float64(elapsed.Nanoseconds()) / float64(n),
		bytesPerOp: (after.TotalAlloc - before.TotalAlloc) / count,
		allocsOp:   (after.Mallocs - before.Mallocs) / count,
	}, nil
}

// runBench executes the headless benchmark set without opening a window and
// prints one line per row, also writing benchFile when set.
func runBench(benchFile string) error {
	ctx := context.Background()
	app, err := newListApp()
	if err != nil {
		return err
	}
	if err := app.Page().Redraw(ctx); err != nil {
		return err
	}
	rows := make([]benchResult, 0, benchRowCap)
	redraw, err := measure("redraw-same-data", benchRedrawOps, func() error {
		return app.Page().Redraw(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, redraw)
	update, err := measure("rebuild-same-data", benchRebuildOps, func() error {
		return app.refresh(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, update)
	click, err := measure("click-select-rebuild", benchClickOps, func() error {
		return pagetest.ClickAction(ctx, app.Page(), "row:500")
	})
	if err != nil {
		return err
	}
	rows = append(rows, click)
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%-20s ops=%-4d ns/op=%-12.1f B/op=%-10d allocs/op=%d\n",
			r.name, r.ops, r.nsPerOp, r.bytesPerOp, r.allocsOp)
	}
	fmt.Print(out)
	if benchFile != "" {
		if err := os.WriteFile(benchFile, []byte(out), benchFilePerm); err != nil {
			return err
		}
	}
	return nil
}
