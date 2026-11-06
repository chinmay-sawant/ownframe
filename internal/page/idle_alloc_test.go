package page_test

import (
	"context"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// idleHTML is the realistic fixture for the idle regression tests: a styled
// header, a data-driven row list with a selected highlight, and a status
// line, laid out at 640x480 like the redraw benchmarks.
const idleHTML = `<style>` +
	`#app{width:600px;margin:0 auto;font-family:sans-serif}` +
	`.row{padding:8px 12px;border-bottom:1px solid #dddddd}` +
	`.on{background:#eef4ff}` +
	`.muted{color:#666666}` +
	`</style><div id="app"><h1>Inbox</h1>` +
	`{{ $sel := .Selected }}{{range .Rows}}` +
	`<div id="row-{{.ID}}" data-action="select" class="row{{if eq $sel .ID}} on{{end}}">{{.Title}}</div>` +
	`{{end}}<p id="status" class="muted">selected {{$sel}}</p></div>`

const (
	idleWidth         = 640
	idleHeight        = 480
	idleRowCount      = 20
	idleAllocsRuns    = 100
	idleMemIters      = 50
	idleMemDeltaLimit = 32768
	idleWorkers       = 8
	idleWorkerRedraws = 25
)

type idleRow struct {
	ID    int
	Title string
}

type idleView struct {
	Selected int
	Rows     []idleRow
}

func makeIdleView() idleView {
	view := idleView{Selected: -1, Rows: make([]idleRow, idleRowCount)}
	for i := range view.Rows {
		view.Rows[i] = idleRow{ID: i, Title: "Message " + strconv.Itoa(i)}
	}

	return view
}

func newIdlePage(t *testing.T, view idleView) *page.Page {
	t.Helper()

	p, err := page.New(page.Config{HTML: idleHTML, Width: idleWidth, Height: idleHeight})
	if err != nil {
		t.Fatal(err)
	}

	p.SetData(view)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

// TestIdleRedrawAllocatesNothing pins the idle cost of a realistic page: an
// identical Redraw commits no new frame and allocates no heap objects. The
// object count is exact; the byte delta carries headroom because the runtime
// may surprise on TotalAlloc.
func TestIdleRedrawAllocatesNothing(t *testing.T) {
	ctx := context.Background()
	p := newIdlePage(t, makeIdleView())
	p.TakeDirtyRects()

	before := p.Stats()
	gen := p.Generation()
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}

	var runErr error
	calls := 0
	allocs := testing.AllocsPerRun(idleAllocsRuns, func() {
		calls++
		if err := p.Redraw(ctx); err != nil {
			runErr = err
		}
	})
	if runErr != nil {
		t.Fatal(runErr)
	}
	if allocs != 0 {
		t.Fatalf("idle redraw allocated %v objects per run, want 0", allocs)
	}

	runtime.GC()

	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	for range idleMemIters {
		calls++
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}
	runtime.ReadMemStats(&memAfter)

	delta := memAfter.TotalAlloc - memBefore.TotalAlloc
	t.Logf("idle redraw bytes for %d runs: %d", idleMemIters, delta)
	if delta > idleMemDeltaLimit {
		t.Fatalf("idle redraw allocated %d bytes, limit %d", delta, idleMemDeltaLimit)
	}

	after := p.Stats()
	wantRedraws := before.Redraws + uint64(calls)
	t.Logf("idle redraw calls: %d", calls)
	if after.Redraws != wantRedraws {
		t.Fatalf("redraws = %d, want %d", after.Redraws, wantRedraws)
	}
	if after.Layouts != before.Layouts || after.Repaints != before.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want no new frame",
			after.Layouts-before.Layouts, after.Repaints-before.Repaints)
	}
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("parses = %d, cascades = %d, want no reparse",
			after.Parses-before.Parses, after.Cascades-before.Cascades)
	}
	if p.Generation() != gen {
		t.Fatalf("generation = %d, want %d", p.Generation(), gen)
	}
	if p.Display() != first {
		t.Fatal("identical redraw rebuilt the display list")
	}
	if _, ok := p.TakeDirtyRects(); ok {
		t.Fatal("identical redraw reported dirty")
	}
}

// TestIdleClickSameRowSkipsLayout pins that a click which changes nothing
// lays nothing out again. Selecting row 2 commits one frame; re-selecting
// row 2 leaves the layout, repaint, and generation counters flat.
func TestIdleClickSameRowSkipsLayout(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := makeIdleView()
	p, err := page.New(page.Config{HTML: idleHTML, Width: idleWidth, Height: idleHeight})
	if err != nil {
		t.Fatal(err)
	}
	p.Handle(page.Handlers{
		Click: func(_ context.Context, box page.Box) error {
			n, err := strconv.Atoi(strings.TrimPrefix(box.ID, "row-"))
			if err != nil {
				return nil
			}
			view.Selected = n
			p.SetData(view)

			return nil
		},
	})
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, p, "row-2")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	selected := p.Stats()
	selectedGen := p.Generation()
	if selected.Layouts == 0 {
		t.Fatal("first select laid nothing out")
	}
	if boxText(t, p, "status") != "selected 2" {
		t.Fatalf("status = %q, want selected 2", boxText(t, p, "status"))
	}
	p.TakeDirtyRects()

	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	after := p.Stats()
	if after.Layouts != selected.Layouts || after.Repaints != selected.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want no new frame",
			after.Layouts-selected.Layouts, after.Repaints-selected.Repaints)
	}
	if p.Generation() != selectedGen {
		t.Fatalf("generation = %d, want %d", p.Generation(), selectedGen)
	}
	if _, ok := p.TakeDirtyRects(); ok {
		t.Fatal("re-select reported dirty")
	}
}

// TestIdleConcurrentRedrawsRaceClean drives idle redraws from several
// goroutines, each owning its page, so the race detector can catch shared
// state behind the skip path. Every worker must commit exactly its first
// frame and nothing after.
func TestIdleConcurrentRedrawsRaceClean(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pages := make([]*page.Page, idleWorkers)
	layouts := make([]uint64, idleWorkers)
	for i := range pages {
		pages[i] = newIdlePage(t, makeIdleView())
		pages[i].TakeDirtyRects()
		layouts[i] = pages[i].Stats().Layouts
	}

	type workerResult struct {
		err     error
		layouts uint64
	}
	results := make([]workerResult, idleWorkers)
	var wg sync.WaitGroup
	for i := range pages {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for range idleWorkerRedraws {
				if err := pages[idx].Redraw(ctx); err != nil {
					results[idx].err = err

					return
				}
			}
			results[idx].layouts = pages[idx].Stats().Layouts
		}(i)
	}
	wg.Wait()

	for i := range results {
		if results[i].err != nil {
			t.Fatal(results[i].err)
		}
		if results[i].layouts != layouts[i] {
			t.Fatalf("worker %d layouts = %d, want %d", i, results[i].layouts, layouts[i])
		}
	}
}
