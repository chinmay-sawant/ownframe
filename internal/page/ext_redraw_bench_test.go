package page_test

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"html/template"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
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
	calls := uint64(0)
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
	wantRedraws := before.Redraws + calls
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

// heavyHTML builds a page with n CSS rules so the sheet collection is a
// visible share of the Redraw cost.
func heavyHTML(n int) string {
	var b strings.Builder
	b.WriteString("<style>")

	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, ".r%d{color:#%06x;font-size:%dpx;margin:%dpx}", i, i*7919&0xFFFFFF, 10+i%8, i%5)
	}

	b.WriteString("</style>")
	b.WriteString(`<p class="r7">heavy stylesheet</p><div id="box"></div>`)

	return b.String()
}

// BenchmarkRedrawHeavyWarm measures repeated Redraws of a page with 200 CSS
// rules and unchanged data.
func BenchmarkRedrawHeavyWarm(b *testing.B) {
	ctx := context.Background()

	screen, err := page.New(page.Config{HTML: heavyHTML(200), Width: benchWidth, Height: benchHeight})
	if err != nil {
		b.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if err := screen.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// benchInvalidateRows is the box count for BenchmarkInvalidateMany: enough
// that a per-lookup walk over every box would dominate the loop.
const benchInvalidateRows = 3000

func invalidateFixture(b *testing.B) (*page.Page, []string) {
	b.Helper()

	var src strings.Builder
	src.WriteString(`<style>div{height:12px}</style>`)

	ids := make([]string, 0, benchInvalidateRows)
	for i := 0; i < benchInvalidateRows; i++ {
		id := "row-" + strconv.Itoa(i)
		ids = append(ids, id)
		src.WriteString(`<div id="` + id + `"></div>`)
	}

	screen, err := page.New(page.Config{HTML: src.String(), Width: 640, Height: 480})
	if err != nil {
		b.Fatal(err)
	}

	if err := screen.Redraw(context.Background()); err != nil {
		b.Fatal(err)
	}

	if got := len(screen.Boxes()); got < benchInvalidateRows {
		b.Fatalf("boxes = %d, want at least %d", got, benchInvalidateRows)
	}

	return screen, ids
}

// BenchmarkInvalidateMany marks every id on a page with thousands of boxes
// once per iteration and consumes the dirty state the way a frame does.
func BenchmarkInvalidateMany(b *testing.B) {
	screen, ids := invalidateFixture(b)

	for b.Loop() {
		for _, id := range ids {
			screen.Invalidate(id)
		}

		screen.TakeDirty()
	}
}

// BenchmarkRedrawStages breaks Redraw into its pipeline stages so the cost of
// each is readable on its own.
func BenchmarkRedrawStages(b *testing.B) {
	ctx := context.Background()
	tpl := template.Must(template.New("bench").Parse(benchHTML))

	execute := func() string {
		var body bytes.Buffer
		if err := tpl.Execute(&body, nil); err != nil {
			b.Fatal(err)
		}

		return body.String()
	}

	parse := func(source string) *html.Document {
		doc, err := html.Parse([]byte(source))
		if err != nil {
			b.Fatal(err)
		}

		return doc
	}

	apply := func(doc *html.Document) *css.Document {
		styled, err := css.Apply(ctx, doc, css.Options{
			WidthPx:  benchWidth,
			HeightPx: benchHeight,
			Media:    "screen",
		})
		if err != nil {
			b.Fatal(err)
		}

		return styled
	}

	b.Run("TemplateExecute", func(b *testing.B) {
		for b.Loop() {
			execute()
		}
	})

	b.Run("HTMLParse", func(b *testing.B) {
		source := execute()

		for b.Loop() {
			parse(source)
		}
	})

	b.Run("CSSApply", func(b *testing.B) {
		doc := parse(execute())

		for b.Loop() {
			apply(doc)
		}
	})

	b.Run("DisplayList", func(b *testing.B) {
		styled := apply(parse(execute()))

		for b.Loop() {
			if _, err := layout.DisplayListOptions(ctx, styled, layout.Options{}); err != nil {
				b.Fatal(err)
			}
		}
	})

}

// benchHTML is the fixture for the Redraw stage benchmarks: one block fill, a
// text run, and a few rules, laid out at 640x480.
const benchHTML = `<style>` +
	`#box{width:200px;height:80px;background:#3366cc}` +
	`p{color:#222222;font-size:14px;line-height:20px}` +
	`</style><div id="box"></div>` +
	`<p>Resize me. Resize me. Resize me. Resize me. Resize me.</p>`

const benchWidth, benchHeight = 640, 480

// BenchmarkRedrawWarm measures a Redraw whose data never changes: after the
// cache lands this is a relayout plus a layout, before it is the whole
// pipeline.
func BenchmarkRedrawWarm(b *testing.B) {
	ctx := context.Background()

	screen, err := page.New(page.Config{HTML: benchHTML, Width: benchWidth, Height: benchHeight})
	if err != nil {
		b.Fatal(err)
	}

	if err := screen.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if err := screen.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRedrawCold changes executed HTML each iteration, forcing parse and cascade.
func BenchmarkRedrawCold(b *testing.B) {
	ctx := context.Background()
	data := struct{ N int }{}

	screen, err := page.New(page.Config{HTML: benchHTML + `<span>{{.N}}</span>`, Width: benchWidth, Height: benchHeight})
	if err != nil {
		b.Fatal(err)
	}

	screen.SetData(data)
	if err := screen.Redraw(ctx); err != nil {
		b.Fatal(err)
	}

	parses := screen.Stats().Parses
	for i := 0; b.Loop(); i++ {
		data.N = i + 1
		screen.SetData(data)

		if err := screen.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
	parsed := screen.Stats().Parses - parses
	b.ReportMetric(float64(parsed)/float64(b.N), "parses/op")
	if parsed != uint64(max(b.N, 0)) {
		b.Fatalf("cold redraw parsed %d times for %d iterations", parsed, b.N)
	}
}

func TestSetDataChangingOutputRebuildsCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>{{.N}}</p>`)
	p.SetData(struct{ N int }{N: 1})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.SetData(struct{ N int }{N: 2})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Parses; got != 2 {
		t.Fatalf("parses = %d, want 2", got)
	}
}

func TestSetThemeInvalidatesCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p id="a">static</p>`)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.SetTheme(`#a{color:#123456}`); err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Parses; got != 2 {
		t.Fatalf("parses = %d, want 2", got)
	}
}

func TestNavigationInvalidatesCache(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>one</p>`)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Load(ctx, `<p>two</p>`); err != nil {
		t.Fatal(err)
	}

	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Parses; got != 4 {
		t.Fatalf("parses = %d, want 4", got)
	}
}

// Perf stays off unless requested: an untracked Redraw leaves the opt-in
// Stats fields at zero, so end users pay nothing by default.
func TestPerfOffByDefault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>hi</p>`)
	if p.Perf() {
		t.Fatal("perf must default to off")
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s := p.Stats()
	if s.LastTemplate != 0 || s.StyleTime != 0 || s.LayoutTime != 0 || s.PaintTime != 0 {
		t.Fatalf("stages tracked while off: %+v", s)
	}
	if s.AllocFrame != 0 || s.DirtyOps != 0 || s.DirtyRegions != 0 || s.ChangedOps != 0 {
		t.Fatalf("counters tracked while off: %+v", s)
	}
	if s.Redraws != 1 {
		t.Fatalf("redraws = %d, want 1", s.Redraws)
	}
}

// First Redraw fills every stage and dirties the frame; TakeDirty keeps
// the reported counters.
func TestPerfFirstRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>hi</p>`)
	p.SetPerf(true)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s := p.Stats()
	if s.LastTemplate < 0 || s.StyleTime < 0 || s.LayoutTime < 0 || s.PaintTime < 0 {
		t.Fatalf("stages %v %v %v %v", s.LastTemplate, s.StyleTime, s.LayoutTime, s.PaintTime)
	}
	if s.LastRedraw <= 0 || s.AllocFrame == 0 {
		t.Fatalf("redraw %v alloc %d", s.LastRedraw, s.AllocFrame)
	}
	if s.DirtyRegions != 1 || s.DirtyOps <= 0 || s.ChangedOps <= 0 {
		t.Fatalf("dirty %d/%d changed %d", s.DirtyOps, s.DirtyRegions, s.ChangedOps)
	}

	p.TakeDirty()
	if n := p.Stats(); n.DirtyRegions != 1 || n.ChangedOps != s.ChangedOps {
		t.Fatalf("take cleared %d/%d", n.DirtyRegions, n.ChangedOps)
	}
}

// An edit reports a dirty region and a resize reports the whole frame.
func TestPerfDirtyCounts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := page.New(page.Config{HTML: `<p>{{.V}}</p>`, Width: 320, Height: 200, Perf: true})
	if err != nil {
		t.Fatal(err)
	}

	redraw := func(v string) {
		t.Helper()
		p.SetData(struct{ V string }{V: v})
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}

	redraw("a")
	redraw("b")
	if s := p.Stats(); s.DirtyRegions != 1 || s.ChangedOps < 1 || s.DirtyOps < 1 {
		t.Fatalf("edit %d/%d/%d", s.DirtyRegions, s.DirtyOps, s.ChangedOps)
	}

	p.SetSize(640, 400)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if s := p.Stats(); s.DirtyRegions != 1 || s.ChangedOps <= 0 {
		t.Fatalf("resize %d/%d", s.DirtyRegions, s.ChangedOps)
	}
}

func TestPerfStageAccounting(t *testing.T) {
	p := newCachePage(t, `<p>stage accounting</p>`)
	p.SetPerf(true)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	s := p.Stats()
	sum := s.LastTemplate + s.StyleTime + s.LayoutTime + s.PaintTime + s.OtherTime
	if sum != s.LastRedraw {
		t.Fatalf("stage sum %s != redraw %s", sum, s.LastRedraw)
	}
	if s.StyleTime <= 0 || s.LayoutTime <= 0 {
		t.Fatal("style and placement must be measured independently")
	}

	// An identical frame skips the style and placement stages, so they read
	// zero while the stage sum still matches the redraw.
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	s = p.Stats()
	sum = s.LastTemplate + s.StyleTime + s.LayoutTime + s.PaintTime + s.OtherTime
	if sum != s.LastRedraw {
		t.Fatalf("stage sum %s != redraw %s", sum, s.LastRedraw)
	}
	if s.StyleTime != 0 || s.LayoutTime != 0 {
		t.Fatalf("skipped style %s layout %s, want both 0", s.StyleTime, s.LayoutTime)
	}
}

// TestRedrawBatchNestedDrawsOnce checks that a nested batch draws once: the
// inner EndRedrawBatch leaves the draw to the outer one. The resize changes
// the frame so the outer flush has a layout to commit.
func TestRedrawBatchNestedDrawsOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts

	p.SetSize(400, 300)
	p.BeginRedrawBatch()
	p.BeginRedrawBatch()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before {
		t.Fatalf("layouts = %d, want no draw from the inner batch", got-before)
	}

	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want one draw from the outer batch", got-before)
	}
}

// TestRedrawBatchQueuesAFailedDraw checks that a draw that fails leaves the
// redraw queued, so the next batch commits it and the frame catches up.
func TestRedrawBatchQueuesAFailedDraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts

	dead, cancel := context.WithCancel(ctx)
	cancel()

	p.BeginRedrawBatch()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(dead); err == nil {
		t.Fatal("the final draw should fail with a dead context")
	}
	if got := p.Stats().Layouts; got != before {
		t.Fatalf("layouts = %d, want no draw", got-before)
	}

	// The failed draw stays queued; a resize gives the retry a changed frame
	// so the next batch commits it.
	p.SetSize(400, 300)
	p.BeginRedrawBatch()
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want the queued draw", got-before)
	}
}

// TestRedrawBatchSkipsUnchangedIfChanged checks that an IfChanged-only batch
// still honors the unchanged check at flush time.
func TestRedrawBatchSkipsUnchangedIfChanged(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Layouts

	p.BeginRedrawBatch()
	if err := p.RedrawIfChanged(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Layouts; got != before {
		t.Fatalf("layouts = %d, want no draw", got-before)
	}
}

// TestRedrawBatchDrawsFocusAndClickOnce checks the pointer batch: a focus
// change and the click that follows draw one frame, and the frame keeps the
// focus.
func TestRedrawBatchDrawsFocusAndClickOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<style>#go:focus{color:#ff0000}</style><button id="go">Start</button>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	x, y := boxCenter(t, p, "go")
	before := p.Stats().Layouts

	p.BeginRedrawBatch()
	if err := p.Focus(ctx, "go"); err != nil {
		t.Fatal(err)
	}
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats().Layouts; got != before+1 {
		t.Fatalf("layouts = %d, want one batch draw", got-before)
	}
	if got := p.FocusID(); got != "go" {
		t.Fatalf("focus = %q, want go", got)
	}
}

// TestRedrawBatchSkipsUnchangedFlush checks that the flush runs the redraw
// sequence once but lays nothing out when the source and state are unchanged,
// whether the batch held a plain Redraw or only RedrawIfChanged calls.
func TestRedrawBatchSkipsUnchangedFlush(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>same</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats()

	p.BeginRedrawBatch()
	if err := p.RedrawIfChanged(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.EndRedrawBatch(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats(); got.Layouts != before.Layouts || got.Repaints != before.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want no draw", got.Layouts-before.Layouts, got.Repaints-before.Repaints)
	}
	if got := p.Stats().Redraws; got != before.Redraws+1 {
		t.Fatalf("redraws = %d, want one flush", got-before.Redraws)
	}
}

// TestRedrawSkipsIdenticalFrame checks that a second Redraw with no input
// change counts the redraw but lays nothing out again: no new layout,
// repaint, parse, or dirty region, the same retained display list, and no
// layout time.
func TestRedrawSkipsIdenticalFrame(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>stable</p>`)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	first := p.Display()
	if first == nil {
		t.Fatal("no display after the first redraw")
	}
	p.TakeDirtyRects()
	before := p.Stats()
	gen := p.Generation()

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	after := p.Stats()
	if after.Redraws != before.Redraws+1 {
		t.Fatalf("redraws = %d, want %d", after.Redraws, before.Redraws+1)
	}
	if after.Layouts != before.Layouts || after.Repaints != before.Repaints {
		t.Fatalf("layouts = %d, repaints = %d, want no new work",
			after.Layouts-before.Layouts, after.Repaints-before.Repaints)
	}
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("parses = %d, cascades = %d, want no reparse",
			after.Parses-before.Parses, after.Cascades-before.Cascades)
	}
	if after.StyleTime != 0 || after.LayoutTime != 0 {
		t.Fatalf("style = %v, layout = %v, want both 0", after.StyleTime, after.LayoutTime)
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

// TestRedrawDrawsChangedData checks the other side of the skip: data that
// changes the executed source still commits a new frame.
func TestRedrawDrawsChangedData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>{{.V}}</p>`)
	p.SetData(struct{ V string }{V: "a"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats()
	gen := p.Generation()

	p.SetData(struct{ V string }{V: "b"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats(); got.Layouts != before.Layouts+1 || got.Repaints != before.Repaints+1 {
		t.Fatalf("layouts = %d, repaints = %d, want one new frame",
			got.Layouts-before.Layouts, got.Repaints-before.Repaints)
	}
	if p.Generation() == gen {
		t.Fatal("changed data did not bump the generation")
	}
}

func TestSetDataKeepsCacheWhenOutputIsUnchanged(t *testing.T) {
	p, err := page.New(page.Config{HTML: `<p>stable</p>`, Width: 320, Height: 200, Perf: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	parses := p.Stats().Parses
	p.SetData(struct{ Unused string }{Unused: "new value"})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Stats().Parses; got != parses {
		t.Fatalf("parses = %d after unchanged output, want %d", got, parses)
	}
}

func TestSetDataConcurrentWithRedraw(t *testing.T) {
	t.Parallel()
	p, err := page.New(page.Config{HTML: `<p>{{.Text}}</p>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			p.SetData(map[string]string{"Text": "hello"})
		}
		close(done)
	}()
	for {
		select {
		case <-done:
			return
		default:
			_ = p.Redraw(ctx)
		}
	}
}

// TestStatsCountRedraws checks the Phase 1 exit.
func TestStatsCountRedraws(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p id="known">hello</p>`)

	for range 2 {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}

	stats := p.Stats()
	if stats.Redraws != 2 || stats.Parses == 0 {
		t.Fatalf("redraws = %d, parses = %d, want 2 and a parse", stats.Redraws, stats.Parses)
	}

	// The second Redraw sees an identical frame and skips the layout, so
	// it counts the redraw but no layout or repaint, and no layout time.
	if stats.Layouts != 1 || stats.Repaints != 1 {
		t.Fatalf("layouts = %d, repaints = %d, want 1 each", stats.Layouts, stats.Repaints)
	}
	if stats.LayoutTime != 0 {
		t.Fatalf("layout time = %v, want 0 on the skipped frame", stats.LayoutTime)
	}

	if stats.Boxes == 0 || stats.Ops == 0 {
		t.Fatalf("boxes = %d, ops = %d", stats.Boxes, stats.Ops)
	}

	if stats.LastRedraw <= 0 || stats.LastDraw <= 0 {
		t.Fatalf("last redraw = %v, last draw = %v", stats.LastRedraw, stats.LastDraw)
	}
}

func TestStatsClickReportsBoxID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<div id="known">pick me</div>`)

	got := ""
	p.Handle(page.Handlers{
		Click: func(_ context.Context, box page.Box) error {
			got = box.ID

			return nil
		},
	})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	x, y := boxCenter(t, p, "known")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if got != "known" {
		t.Fatalf("click id = %q, want known", got)
	}
}

// TestSetDrawTimeWinsLastDraw checks the window hook: once a shell records
// its draw time, LastDraw is that time, not the page's own paint stage.
func TestSetDrawTimeWinsLastDraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := newCachePage(t, `<p>frame</p>`)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.SetDrawTime(7 * time.Millisecond)
	if got := p.Stats().LastDraw; got != 7*time.Millisecond {
		t.Fatalf("last draw = %v, want 7ms", got)
	}
}
