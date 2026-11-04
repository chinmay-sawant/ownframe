// Command mygo opens the compare thousand-row table in MyGo native UI.
//
// Naive form: all 1000 rows render every frame, so this measures the default
// full-build path against the ownframe naive table.
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
	listWidth    = 800
	listHeight   = 600
	listRows     = 1000
	listSubtitle = "mygo · Go view → GPU, no HTML"
	listBody     = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
)

// listApp is the app state the window shows: which row is selected.
type listApp struct{ selected int }

// view builds the interface from the state, on the main thread, whenever the
// window needs a frame.
func (s *listApp) view(c *ui.Context) {
	t := c.Theme()
	white := ui.Hex("#ffffff")
	ui.Column(c).Fill().Children(func() {
		ui.Column(c).Gap(4).Padding(12).Children(func() {
			ui.Text(c, "Compare Table · 1000 rows").FontSize(18).Bold()
			ui.Text(c, listSubtitle).FontSize(12).TextColor(t.TextMuted)
			ui.Text(c, s.selectedText()).FontSize(14)
		})
		ui.Row(c).Gap(12).Padding(8).Background(ui.Hex("#1c1915")).Children(func() {
			ui.Text(c, "Num").FontSize(13).Bold().Width(64).TextColor(white)
			ui.Text(c, "Title").FontSize(13).Bold().Width(200).TextColor(white)
			ui.Text(c, "Description").FontSize(13).Bold().TextColor(white)
		})
		ui.Scroll(c).Fill().Children(func() {
			for i := range listRows {
				bg := ui.Hex("#ffffff")
				if i%2 == 1 {
					bg = ui.Hex("#f3f4f6")
				}
				if i == s.selected {
					bg = ui.Hex("#dbeafe")
				}
				row := ui.Row(c).Gap(12).Padding(8).Background(bg)
				row.Children(func() {
					ui.Text(c, "#"+fmt.Sprint(i+1)).FontSize(13).Width(64)
					ui.Text(c, "Item "+fmt.Sprint(i+1)).FontSize(14).Bold().Width(200)
					ui.Text(c, listBody).FontSize(13)
				})
				if row.Clicked() {
					s.selected = i
				}
			}
		})
	})
}

func (s *listApp) selectedText() string {
	if s.selected < 0 || s.selected >= listRows {
		return "Selected: none"
	}
	return "Selected: Item " + fmt.Sprint(s.selected+1)
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

	s := &listApp{selected: -1}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:   "Compare Table",
			Width:   listWidth,
			Height:  listHeight,
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
		s := &listApp{selected: 500}
		tt := ui.NewTester(s.view, listWidth, listHeight)
		if !tt.HasText("Item 1000") {
			return fmt.Errorf("missing row text")
		}
		return nil
	})
	if err != nil {
		return err
	}
	rows = append(rows, build)
	s := &listApp{selected: -1}
	tt := ui.NewTester(s.view, listWidth, listHeight)
	click, err := measure("click-select-reframe", 100, func() error {
		return tt.Click("Item 5")
	})
	if err != nil {
		return err
	}
	rows = append(rows, click)
	if s.selected != 4 {
		return fmt.Errorf("selected = %d, want 4", s.selected)
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
