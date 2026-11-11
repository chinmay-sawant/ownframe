package perfbench_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/perf-benchmarks/benchutil"
)

const (
	listWidth     = 800
	listHeight    = 600
	listScroll    = 4000
	listEditIndex = 3
)

func newListWorkload(tb testing.TB, perf, windowed bool) *benchutil.ListWorkload {
	tb.Helper()
	w, err := benchutil.NewListWorkload(benchutil.ListOptions{Perf: perf, Windowed: windowed})
	if err != nil {
		tb.Fatal(err)
	}
	if err := w.Page.Redraw(tb.Context()); err != nil {
		tb.Fatal(err)
	}
	if perf != (w.Page.Probe().Stats().StyleTime > 0) {
		tb.Fatal("instrumentation mode did not reach page timing")
	}
	return w
}

func BenchmarkListWorkloads(b *testing.B) {
	for _, perf := range []bool{false, true} {
		mode := "PerfOff"
		if perf {
			mode = "PerfOn"
		}
		b.Run(mode, func(b *testing.B) {
			for _, name := range []string{"Unchanged", "TextUpdate", "ScrollOffset", "Resize", "WindowedScroll"} {
				b.Run(name, func(b *testing.B) {
					benchmarkListWorkload(b, perf, name)
				})
			}
		})
	}
}

func benchmarkListWorkload(b *testing.B, perf bool, name string) {
	w := newListWorkload(b, perf, name == "WindowedScroll")
	ctx := b.Context()
	generation := w.Page.Probe().Generation()
	var last int
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; b.Loop(); i++ {
		last = i
		if err := stepListWorkload(ctx, w, name, i); err != nil {
			b.Fatal(err)
		}
	}
	checkListWorkload(b, w, name, last, generation)
}

func stepListWorkload(ctx context.Context, w *benchutil.ListWorkload, name string, i int) error {
	switch name {
	case "TextUpdate":
		text := "changed body A"
		if i%2 != 0 {
			text = "changed body B"
		}
		if err := w.SetText(listEditIndex, text); err != nil {
			return err
		}
	case "Resize":
		w.Page.SetSize(listWidth+(i+1)%2, listHeight)
	case "ScrollOffset", "WindowedScroll":
		return w.Scroll(ctx, listScroll*(1+i%2))
	}
	return w.Page.Redraw(ctx)
}

func checkListWorkload(tb testing.TB, w *benchutil.ListWorkload, name string, last int, generation uint64) {
	tb.Helper()
	if last < 0 {
		tb.Fatal("no workload iterations")
		return
	}
	expected := generation
	if name == "TextUpdate" || name == "Resize" || name == "WindowedScroll" {
		expected += uint64(last) + 1
	}
	if got := w.Page.Probe().Generation(); got != expected {
		tb.Fatalf("%s generation = %d, want %d; a timed operation was skipped or committed twice", name, got, expected)
	}
	switch name {
	case "Unchanged":
		if w.Page.Probe().Generation() != generation {
			tb.Fatal("unchanged redraw committed a new generation")
		}
	case "TextUpdate":
		want := "changed body A"
		if last%2 != 0 {
			want = "changed body B"
		}
		found := false
		for _, op := range w.Page.Probe().Display().Ops {
			found = found || op.Text == want
		}
		if !found || w.Page.Probe().Generation() == generation {
			tb.Fatal("text update did not reach retained paint")
		}
	case "Resize":
		width, height := w.Page.Size()
		display := w.Page.Probe().Display()
		if width != listWidth+(last+1)%2 || height != listHeight || display.Width != width {
			tb.Fatal("resize did not reach the display")
		}
		if w.Page.Probe().Generation() == generation {
			tb.Fatal("resize did not commit")
		}
	case "ScrollOffset", "WindowedScroll":
		_, offset := w.Page.Scroll().ScrollOffset()
		if offset != listScroll*(1+last%2) {
			tb.Fatal("scroll offset was not submitted")
		}
		first := (offset-benchutil.ListHeadHeight)/benchutil.ListRowHeight + 1
		id := fmt.Sprintf("row-%d", first)
		box := workloadBox(tb, w.Page, id)
		if box.Y > float64(offset) || box.Y+box.H <= float64(offset) {
			tb.Fatalf("first visible row %s has wrong geometry: %+v", id, box)
		}
		if name == "WindowedScroll" && w.Page.Probe().Generation() == generation {
			tb.Fatal("windowed scroll did not rebuild the visible rows")
		}
	}
}

func workloadBox(tb testing.TB, page *ownframe.Page, id string) ownframe.Box {
	tb.Helper()
	for _, box := range page.Boxes() {
		if box.ID == id {
			return box
		}
	}
	tb.Fatalf("missing row %s", id)
	return ownframe.Box{}
}

func TestListWorkloads(t *testing.T) {
	for _, perf := range []bool{false, true} {
		for _, name := range []string{"Unchanged", "TextUpdate", "ScrollOffset", "Resize", "WindowedScroll"} {
			t.Run(fmt.Sprintf("%t/%s", perf, name), func(t *testing.T) {
				w := newListWorkload(t, perf, name == "WindowedScroll")
				generation := w.Page.Probe().Generation()
				for i := 0; i < 2; i++ {
					if err := stepListWorkload(t.Context(), w, name, i); err != nil {
						t.Fatal(err)
					}
					checkListWorkload(t, w, name, i, generation)
				}
			})
		}
	}
}

func TestListWindowGeometryMatchesFull(t *testing.T) {
	full := newListWorkload(t, false, false)
	windowed := newListWorkload(t, false, true)
	for _, offset := range []int{0, listScroll, 2 * listScroll, benchutil.ListHeadHeight + benchutil.ListRows*benchutil.ListRowHeight - listHeight, 0} {
		if err := windowed.Scroll(t.Context(), offset); err != nil {
			t.Fatal(err)
		}
		visible := 0
		for _, want := range full.Page.Boxes() {
			if want.ID == "" || want.ID == "head" || want.Y+want.H <= float64(offset) || want.Y >= float64(offset+listHeight) {
				continue
			}
			got := workloadBox(t, windowed.Page, want.ID)
			if got.X != want.X || got.Y != want.Y || got.W != want.W || got.H != want.H {
				t.Fatalf("offset %d: windowed row %+v differs from full row %+v", offset, got, want)
			}
			visible++
		}
		if visible == 0 {
			t.Fatalf("offset %d has no visible rows", offset)
		}
	}
}
