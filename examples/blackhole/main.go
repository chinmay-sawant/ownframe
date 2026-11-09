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
)

func main() {
	web := flag.Bool("web", false, "serve the picture in a browser on -addr")
	addr := flag.String("addr", "127.0.0.1:8155", "listen address for -web")
	dev := flag.Bool("dev", false, "open the developer tools dock at startup")
	perf := flag.Bool("perf", false, "enable perf sampling for the devtools Frame tab")
	bench := flag.Bool("bench", false, "run headless benchmarks instead of opening a window")
	benchFile := flag.String("bench-file", "", "write benchmark results to this file as well as stdout")
	flag.Parse()

	if *bench {
		if err := runBench(*benchFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	app, err := newApp(*dev, *perf)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	shaders, err := compileShaders()
	if err != nil {
		log.Fatalf("failed to compile shaders: %v", err)
	}

	ctx := context.Background()
	renderer := &sceneRenderer{app: app, shaders: shaders}
	app.Page().SetTick(renderer.tick)
	debugLogf("blackhole start (shaders ok)")

	if *web {
		log.Printf("Serving Black Hole web view on http://%s", *addr)
		if err := ownframe.Serve(ctx, app.Page(), *addr); err != nil {
			log.Fatal(err)
		}
		return
	}

	if err := ownframe.Run(ctx, app.Page()); err != nil {
		log.Fatal(err)
	}
}

type benchResult struct {
	name       string
	ops        uint64
	nsPerOp    float64
	bytesPerOp uint64
	allocsOp   uint64
}

// Headless benchmark sizes: op counts per case, the frame each simulated
// step advances by, the results table's capacity, and the mode of the
// optional results file.
const (
	benchRows        = 6
	benchCameraOps   = 10000
	benchThrottleOps = 600
	benchRecoverOps  = 2000
	benchHUDPatches  = 200
	benchPageTicks   = 200

	throttleFrame = 1.0 / 20.0  // slow frame that drives DRS step-down
	recoveryFrame = 1.0 / 120.0 // fast frame that drives DRS recovery

	benchFileMode os.FileMode = 0o600
)

func measure(name string, n uint64, fn func() error) (benchResult, error) {
	runtime.GC()
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	start := time.Now()
	for i := uint64(0); i < n; i++ {
		if err := fn(); err != nil {
			return benchResult{}, err
		}
	}
	elapsed := time.Since(start)
	runtime.ReadMemStats(&after)
	return benchResult{
		name:       name,
		ops:        n,
		nsPerOp:    float64(elapsed.Nanoseconds()) / float64(n),
		bytesPerOp: (after.TotalAlloc - before.TotalAlloc) / n,
		allocsOp:   (after.Mallocs - before.Mallocs) / n,
	}, nil
}

func runBench(benchFile string) error {
	app, err := newApp(false, true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	_ = app.Page().Redraw(ctx)

	rows := make([]benchResult, 0, benchRows)

	// 1. Camera step physics
	r1, err := measure("camera-step-physics", benchCameraOps, func() error {
		app.cam.step(initialFrame, false, false)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r1)

	// 2. Camera orthonormal basis
	r2, err := measure("camera-axes-calc", benchCameraOps, func() error {
		app.cam.yaw += 0.001
		_, _, _, _ = app.cam.axes()
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r2)

	// 3. Dynamic quality scaling (step-down)
	r3, err := measure("adaptive-drs-throttle", benchThrottleOps, func() error {
		app.step(ctx, throttleFrame, false)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r3)

	// 4. Dynamic quality scaling (recovery)
	r4, err := measure("adaptive-drs-recover", benchRecoverOps, func() error {
		app.step(ctx, recoveryFrame, false)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r4)

	// 5. HUD Telemetry element patching
	r5, err := measure("hud-telemetry-patch", benchHUDPatches, func() error {
		app.updateHUDMetrics(ctx)
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, r5)

	// 6. Ownframe page tick
	r6, err := measure("page-testing-tick", benchPageTicks, func() error {
		return app.Page().Testing().Tick(ctx)
	})
	if err != nil {
		return err
	}
	rows = append(rows, r6)

	out := "\n=== Ownframe Black Hole Headless Benchmark Results ===\n"
	for _, r := range rows {
		fps := 0.0
		if r.nsPerOp > 0 {
			fps = float64(time.Second) / r.nsPerOp
		}
		out += fmt.Sprintf("%-26s ops=%-6d ns/op=%-10.1f fps=%-8.1f B/op=%-8d allocs/op=%d\n",
			r.name, r.ops, r.nsPerOp, fps, r.bytesPerOp, r.allocsOp)
	}
	fmt.Print(out)

	if benchFile != "" {
		if err := os.WriteFile(benchFile, []byte(out), benchFileMode); err != nil {
			return err
		}
	}
	return nil
}
