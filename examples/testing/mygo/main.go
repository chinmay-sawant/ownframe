package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	bench := flag.Bool("bench", false, "run headless benchmarks instead of opening a window")
	benchFile := flag.String("bench-file", "", "write benchmark results to this file as well as stdout")
	invoice := flag.Bool("invoice", false, "run previous invoice table benchmark instead of black hole")
	flag.Parse()

	if *invoice {
		if *bench {
			if err := runInvoiceBench(*benchFile); err != nil {
				log.Fatal(err)
			}
			return
		}
		runInvoice()
		return
	}

	if *bench {
		if err := runBlackHoleBench(*benchFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	a := newApp()
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:           "Black Hole",
			Width:           1200,
			Height:          760,
			TitleBarStyle:   mygo.TitleBarHiddenInset,
			BackgroundColor: "#000000",
			Content:         ui.View(a.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

type mygoBenchResult struct {
	name       string
	ops        int
	nsPerOp    float64
	bytesPerOp uint64
	allocsOp   uint64
}

func measureMyGo(name string, n int, fn func() error) (mygoBenchResult, error) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	for i := 0; i < n; i++ {
		if err := fn(); err != nil {
			return mygoBenchResult{}, err
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	return mygoBenchResult{
		name:       name,
		ops:        n,
		nsPerOp:    float64(elapsed.Nanoseconds()) / float64(n),
		bytesPerOp: (after.TotalAlloc - before.TotalAlloc) / uint64(n),
		allocsOp:   (after.Mallocs - before.Mallocs) / uint64(n),
	}, nil
}

func runBlackHoleBench(benchFile string) error {
	a := newApp()

	rows := make([]mygoBenchResult, 0, 5)

	// 1. Camera step physics
	r1, err := measureMyGo("camera-step-physics", 10000, func() error {
		a.step(1.0/60.0, false)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r1)

	// 2. Camera axes calculation
	r2, err := measureMyGo("camera-axes-calc", 10000, func() error {
		a.cam.yaw += 0.001
		_, _, _, _ = a.cam.axes()
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r2)

	// 3. Adaptive DRS throttle
	r3, err := measureMyGo("adaptive-drs-throttle", 600, func() error {
		a.step(1.0/20.0, false)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r3)

	// 4. Adaptive DRS recovery
	r4, err := measureMyGo("adaptive-drs-recover", 2000, func() error {
		a.step(1.0/120.0, false)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r4)

	// 5. Headless tester frame
	tt := ui.NewTester(a.view, 1200, 760)
	r5, err := measureMyGo("headless-tester-frame", 200, func() error {
		tt.Frame()
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r5)

	out := "\n=== MyGo Black Hole Headless Benchmark Results ===\n"
	for _, r := range rows {
		fps := 0.0
		if r.nsPerOp > 0 {
			fps = 1e9 / r.nsPerOp
		}
		out += fmt.Sprintf("%-26s ops=%-6d ns/op=%-10.1f fps=%-8.1f B/op=%-8d allocs/op=%d\n",
			r.name, r.ops, r.nsPerOp, fps, r.bytesPerOp, r.allocsOp)
	}
	fmt.Print(out)

	if benchFile != "" {
		if err := os.WriteFile(benchFile, []byte(out), 0o600); err != nil {
			return err
		}
	}
	return nil
}
