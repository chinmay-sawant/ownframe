// Command mygo-inventory opens the Northwind Goods inventory dashboard
// in MyGo native UI: the top bar with search, alerts and account, the
// page header with period and category filters plus the New PO dialog,
// the five KPI cells, the stock health table, the bin utilization grid
// with hover data, the product heat map, inbound orders, and value by
// category. Data mirrors examples/builder/inventory so the two apps stay
// with respect to each other.
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

const (
	winWidth  = 1360
	winHeight = 900
	// testHeight fits the whole scrolled dashboard so the heat cells are
	// clickable: the real window is shorter and scrolls to them.
	testHeight = 2400
)

// heatApp is the dashboard state: filters, overlays, and selections.
type heatApp struct {
	period       string
	category     string
	statusFilter string
	query        string
	picked       string
	poOpen       bool
	notifOpen    bool
	userOpen     bool
	selP         int
	selA         int
}

func newHeatApp() *heatApp {
	return &heatApp{period: "30 days", category: "All categories", selP: -1, selA: -1}
}

// view builds the interface from the state, on the main thread, whenever
// the window needs a frame.
func (s *heatApp) view(c *ui.Context) {
	ui.Column(c).Fill().Children(func() {
		s.topbar(c)
		ui.Scroll(c).Fill().Children(func() {
			ui.Column(c).Gap(12).Padding(12).Children(func() {
				s.pageHeader(c)
				s.kpiPanel(c)
				ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
					ui.Column(c).Gap(12).Grow(1.65).MinWidth(0).Children(func() {
						s.stockPanel(c)
						s.categoryPanel(c)
					})
					ui.Column(c).Gap(12).Grow(1).MinWidth(0).Children(func() {
						s.binsPanel(c)
						s.inboundPanel(c)
					})
				})
				s.heatPanel(c)
			})
		})
	})
}

func main() {
	bench := flag.Bool("bench", false, "run headless benchmarks instead of opening a window")
	benchFile := flag.String("bench-file", "", "write benchmark results to this file as well as stdout")
	flag.Parse()

	if *bench {
		if err := runBench(*benchFile); err != nil {
			log.Fatal(err)
		}
		return
	}

	s := newHeatApp()
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "MyGo-inventory · Inventory health",
			Width:   winWidth,
			Height:  winHeight,
			Content: ui.View(s.view),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

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
	return benchResult{
		name:       name,
		ops:        n,
		nsPerOp:    float64(elapsed.Nanoseconds()) / float64(n),
		bytesPerOp: (after.TotalAlloc - before.TotalAlloc) / uint64(n),
		allocsOp:   (after.Mallocs - before.Mallocs) / uint64(n),
	}, nil
}

// runBench executes the headless benchmark set without opening a window and
// prints one line per row, also writing benchFile when set.
func runBench(benchFile string) error {
	rows := make([]benchResult, 0, 2)
	build, err := measure("build-full-view", 20, func() error {
		s := newHeatApp()
		s.selP, s.selA = 7, 4
		tt := ui.NewTester(s.view, winWidth, testHeight)
		if !tt.HasText("Inventory health") {
			return fmt.Errorf("missing dashboard text")
		}
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, build)
	s := newHeatApp()
	tt := ui.NewTester(s.view, winWidth, testHeight)
	click, err := measure("click-select-reframe", 100, func() error {
		if err := tt.Click(cellLabel(7, 4)); err != nil {
			return err
		}
		return tt.Click(cellLabel(0, 0))
	})
	if err != nil {
		return err
	}
	rows = append(rows, click)
	if s.selP != 0 || s.selA != 0 {
		return fmt.Errorf("selected = %d,%d, want 0,0", s.selP, s.selA)
	}
	out := ""
	for _, r := range rows {
		out += fmt.Sprintf("%-20s ops=%-4d ns/op=%-12.1f B/op=%-10d allocs/op=%d\n",
			r.name, r.ops, r.nsPerOp, r.bytesPerOp, r.allocsOp)
	}
	fmt.Print(out)
	if benchFile != "" {
		if err := os.WriteFile(benchFile, []byte(out), 0o600); err != nil {
			return err
		}
	}
	return nil
}
