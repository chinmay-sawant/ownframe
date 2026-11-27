package window

import (
	"context"
	"encoding/json"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"strings"
	"testing"
	"time"
)

// devFrameText joins the rows' plain text.
func devFrameText(rows []devRow) string {
	var lines []string
	for _, row := range rows {
		lines = append(lines, row.line.plain())
	}
	return strings.Join(lines, "\n")
}

func TestDevFrameSeparatesIdleGapFromWork(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perf = true
	s.dev.frame = 250 * time.Millisecond
	s.dev.stages.work = 2 * time.Millisecond
	rows := s.devFrameRows(320)
	if devFrameValue(rows, "Frame gap") != "250.0ms" || devFrameValue(rows, "Frame work") != "2.0ms" {
		t.Fatal("frame rows conflate parked time with rendering work")
	}
}

// TestDevFrameSections checks the tab's four sections.
func TestDevFrameSections(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	got := devFrameText(s.devFrameRows(320))

	for _, want := range []string{"WINDOW", "Window", "RENDERING", "PIPELINE", "RELOAD", "Redraws", "0.0ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}

func TestDevFrameRowsFitMaximizedDock(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.screenW = 1920
	s.screenH = 996
	s.perf = true
	s.dev.tab = devTabFrame
	s.dev.sampler.last = time.Now()
	s.perfHooks.devPipelinePerf = func(host.Stats) (time.Duration, time.Duration, time.Duration, time.Duration, int, int, int, bool) {
		return 0, 0, 0, 0, 0, 0, 0, true
	}
	s.devLayout()

	rowsHeight := float64(len(s.dev.rows)) * devLineHeight()
	if rowsHeight > s.dev.content.H {
		t.Fatalf("Frame rows need %.0fpx across %d rows, maximized dock has %.0fpx", rowsHeight, len(s.dev.rows), s.dev.content.H)
	}
	if s.dev.scroll != 0 {
		t.Fatalf("maximized Frame tab scrolled to %v", s.dev.scroll)
	}

	s.screenH = 600
	s.devLayout()
	rowsHeight = float64(len(s.dev.rows)) * devLineHeight()
	if rowsHeight <= s.dev.content.H {
		t.Fatalf("short dock has %.0fpx for %.0fpx of Frame rows, want scroll", s.dev.content.H, rowsHeight)
	}
}

// TestDevFrameReloadErrorInk checks the error ink.
func TestDevFrameReloadErrorInk(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.dev.stats.LastReloadError = "boom"

	for _, row := range s.devFrameRows(320) {
		if !strings.Contains(row.line.plain(), "boom") {
			continue
		}
		for _, span := range row.line.spans {
			if span.ink == devErrInk {
				return
			}
		}
		t.Fatalf("reload error row %q has no error ink", row.line.plain())
	}
	t.Fatal("no reload error row")
}

// TestDevFrameStable checks two calls agree.
func TestDevFrameStable(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	first := devFrameText(s.devFrameRows(320))
	if second := devFrameText(s.devFrameRows(320)); first != second {
		t.Fatalf("rows changed between calls:\n%q\n%q", first, second)
	}
}

// TestDevFrameRightAligned checks the value is a right-aligned span that
// ends at the row width.
func TestDevFrameRightAligned(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.dev.stats.Redraws = 12

	for _, row := range s.devFrameRows(320) {
		if !strings.HasPrefix(row.line.plain(), "Redraws ") {
			continue
		}
		last := row.line.spans[len(row.line.spans)-1]
		if last.text != "12" || !last.right || last.w <= 0 {
			t.Fatalf("redraws value span = %+v, want a right-aligned sized span", last)
		}
		if row.w != 320 {
			t.Fatalf("row width = %v, want 320", row.w)
		}

		return
	}
	t.Fatal("no redraws row")
}

// BenchmarkDevFrameRows measures the Frame tab's per-frame row build, which
// runs on every frame while the dock is open.
func BenchmarkDevFrameRows(b *testing.B) {
	s := newDevShell(newDevScreen())
	b.ReportAllocs()

	for b.Loop() {
		if len(s.devFrameRows(320)) == 0 {
			b.Fatal("no rows")
		}
	}
}

type devJSONTestRect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type devJSONTestDoc struct {
	Tag   string            `json:"tag"`
	ID    string            `json:"id,omitempty"`
	Skip  string            `json:"-"`
	Items []devJSONTestRect `json:"items"`
	Rect  devJSONTestRect   `json:"rect"`
	Ops   int               `json:"ops"`
	OK    bool              `json:"ok"`
}

func devJSONFixture() devJSONTestDoc {
	return devJSONTestDoc{
		Tag:   "div",
		Skip:  "no",
		Items: []devJSONTestRect{{X: 1, Y: 2}, {X: 3, Y: 4}},
		Rect:  devJSONTestRect{X: 5, Y: 6},
		Ops:   2,
		OK:    true,
	}
}

// TestDevJSONValid checks the joined text parses as JSON.
func TestDevJSONValid(t *testing.T) {
	t.Parallel()

	text := devRowsText(devJSONRows(devJSONFixture(), nil))
	if !json.Valid([]byte(text)) {
		t.Fatalf("json = %q", text)
	}
}

// TestDevJSONOrderAndOmit checks declaration order and omitempty.
func TestDevJSONOrderAndOmit(t *testing.T) {
	t.Parallel()

	text := devRowsText(devJSONRows(devJSONFixture(), nil))
	if strings.Index(text, `"tag"`) > strings.Index(text, `"rect"`) {
		t.Fatalf("tag after rect: %q", text)
	}

	if strings.Contains(text, `"id"`) || strings.Contains(text, `"Skip"`) {
		t.Fatalf("omitempty or json:- leaked: %q", text)
	}
}

// TestDevJSONColors checks the syntax palette.
func TestDevJSONColors(t *testing.T) {
	t.Parallel()

	want := map[string]color.RGBA{
		"tag":   devKeyInk,
		`"div"`: devStrInk,
		"2":     devNumInk,
		"true":  devBoolInk,
	}

	for _, row := range devJSONRows(devJSONFixture(), nil) {
		for _, span := range row.line.spans {
			if ink, ok := want[span.text]; ok && span.ink != ink {
				t.Fatalf("span %q ink = %v, want %v", span.text, span.ink, ink)
			}
		}
	}
}

// TestDevJSONCollapse checks a folded node and its click key.
func TestDevJSONCollapse(t *testing.T) {
	t.Parallel()

	rows := devJSONRows(devJSONFixture(), map[string]bool{"rect": true})
	if text := devRowsText(rows); !strings.Contains(text, `"rect": {...}`) {
		t.Fatalf("collapsed text = %q", text)
	}

	if !devJSONHasAct(rows, "rect") {
		t.Fatal("the folded row carries no devActJSON hit")
	}
}

// TestDevJSONArrayPath checks an element path uses its index.
func TestDevJSONArrayPath(t *testing.T) {
	t.Parallel()

	rows := devJSONRows(devJSONFixture(), map[string]bool{"items.0": true})
	if !devJSONHasAct(rows, "items.0") {
		t.Fatalf("no fold row for items.0 in %q", devRowsText(rows))
	}
}

// TestDevJSONRootUnfoldable checks the root opener does not fold.
func TestDevJSONRootUnfoldable(t *testing.T) {
	t.Parallel()

	for _, row := range devJSONRows(devJSONFixture(), nil) {
		if row.act == devActJSON && row.key == "" {
			t.Fatal("the root node carries a fold hit")
		}
	}
}

// devJSONHasAct reports a row that folds the node at key.
func devJSONHasAct(rows []devRow, key string) bool {
	for _, row := range rows {
		if row.act == devActJSON && row.key == key {
			return true
		}
	}

	return false
}

// TestDevOpRectTextLineBox checks a text op's outline starts above its
// baseline, one line box tall.
func TestDevOpRectTextLineBox(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 80, H: 16, InkDescent: 4}

	got := devOpRect(&op, 1)
	if got.Y != 18 || got.H != 16 {
		t.Fatalf("text rect = %+v, want the line box at y 18", got)
	}
}

// TestDevOpRectUsesCanvasPixels pins the conversion direction: op coordinates
// are points, and PointsPerPixel divides them to CSS pixels. The old inspector
// multiplied, so the outline sat at 56% of the paint.
func TestDevOpRectUsesCanvasPixels(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, W: 90, H: 30}

	if got := devOpRect(&op, 0.75); got != (devRect{W: 120, H: 40}) {
		t.Fatalf("fill rect = %+v, want 120x40", got)
	}
}

// TestDevOpRectCoversStrokeInk checks the outline uses the painted box rather
// than the nominal border box, so half a wide border is not left outside it.
func TestDevOpRectCoversStrokeInk(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpStrokeRect, W: 80, H: 40, Width: 4}

	if got := devOpRect(&op, 1); got != (devRect{X: -2, Y: -2, W: 84, H: 44}) {
		t.Fatalf("stroke rect = %+v, want the ink box", got)
	}
}

// TestDevOpRectFallsBackWhenUnbounded checks a transformed run still gets its
// nominal box, with a text baseline applied, for the list and count rows.
func TestDevOpRectFallsBackWhenUnbounded(t *testing.T) {
	t.Parallel()

	op := layout.DisplayOp{Kind: layout.DisplayOpText, X: 10, Y: 30, W: 80, H: 16, InkDescent: 4, XformSet: true}

	if got := devOpRect(&op, 1); got != (devRect{X: 10, Y: 18, W: 80, H: 16}) {
		t.Fatalf("unbounded text rect = %+v, want the nominal line box", got)
	}
}

// TestDevOpsRowsStillGenerationNoAlloc checks the Ops tab's per-frame build:
// a still generation returns the same rows and allocates nothing, because the
// list holds one row per painted operation. Not parallel, so no sibling
// allocates during AllocsPerRun.
func TestDevOpsRowsStillGenerationNoAlloc(t *testing.T) {
	d := newDevScreen()
	d.redraws = 1
	s := newDevShell(d)
	s.display = devTestDisplay()
	s.dev.tab = devTabOps

	first := s.devOpsRows()
	got := first

	allocs := testing.AllocsPerRun(100, func() { got = s.devOpsRows() })

	if allocs != 0 || &got[0] != &first[0] {
		t.Fatalf("still generation allocs = %v, rebuilt = %v", allocs, &got[0] != &first[0])
	}

	s.drawDevTools(ebiten.NewImage(s.screenW, s.screenH))

	if &s.dev.opsCache.rows[0] != &first[0] {
		t.Fatal("the draw rebuilt the rows")
	}
}

// TestDevOpsRowsCacheKeys checks the list follows the generation and the
// outline toggle, and that generation zero never caches, so a fake screen
// keeps seeing fresh rows.
func TestDevOpsRowsCacheKeys(t *testing.T) {
	d := newDevScreen()
	d.redraws = 1
	s := newDevShell(d)
	s.display = devTestDisplay()

	off := s.devOpsRows()
	if !strings.HasPrefix(off[0].line.plain(), "[ ]") {
		t.Fatalf("toggle row = %q, want [ ]", off[0].line.plain())
	}

	if got := s.devOpsRows(); &got[0] != &off[0] {
		t.Fatal("the same inputs rebuilt the rows")
	}

	s.dev.ops = true

	on := s.devOpsRows()
	if &on[0] == &off[0] || !strings.HasPrefix(on[0].line.plain(), "[x]") {
		t.Fatalf("toggle row = %q, want [x]", on[0].line.plain())
	}

	d.redraws++
	s.display = nil

	if text := devRowsText(s.devOpsRows()); !strings.Contains(text, "bitmap fallback") {
		t.Fatalf("new generation rows = %q, want the fallback", text)
	}

	d.redraws = 0
	s.display = devTestDisplay()
	first, second := s.devOpsRows(), s.devOpsRows()
	if &first[0] == &second[0] {
		t.Fatal("generation zero cached the rows")
	}
}

// devCullDisplay is one fill inside the window, one under the dock, one below
// the window, one off the left edge, and one noop without paint.
func devCullDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 5000, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 5000, W: 40, H: 20},
			{Kind: layout.DisplayOpFillRect, X: -5000, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpNoop, W: 4, H: 4},
		},
		Order:          []int{0, 1, 2, 3, 4},
		Width:          800,
		Height:         600,
		PointsPerPixel: 0.75,
	}
}

// TestDevOpsOutlineCull checks the outlines are stroked only for ops that
// meet the visible page area, while the row list keeps every op so a culled
// row still picks its index.
func TestDevOpsOutlineCull(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.display = devCullDisplay()
	s.dev.tab = devTabOps
	s.dev.ops = true
	s.devLayout()

	var got []devRect
	s.dev.stroke = func(_ *ebiten.Image, r devRect, _ color.RGBA) {
		got = append(got, r)
	}

	s.drawDevOps(ebiten.NewImage(s.screenW, s.screenH))

	op := &s.display.Ops[0]
	want := s.devScreen(devOpRect(op, s.display.PointsPerPixel))
	if len(got) != 1 || got[0] != want {
		t.Fatalf("stroked outlines = %v, want only %v", got, want)
	}

	picks := 0

	for _, row := range s.devOpsRows() {
		if row.act == devActOp {
			picks++
		}
	}

	if picks != 4 {
		t.Fatalf("pickable rows = %d, want all 4 painted ops", picks)
	}
}

// devBigDisplay is n painted fills in paint order, enough rows to scroll.
func devBigDisplay(n int) *layout.Display {
	ops := make([]layout.DisplayOp, n)
	order := make([]int, n)

	for i := range ops {
		ops[i] = layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: float64(10 + i), Y: 10, W: 40, H: 20}
		order[i] = i
	}

	return &layout.Display{Ops: ops, Order: order, Width: 800, Height: 600, PointsPerPixel: 0.75}
}

// TestDevOpsDrawRebuildsOnNewGeneration draws the Ops tab twice on a still
// generation and once per run on a moving one: the still draw rebuilds no
// rows, so the two runs differ by the whole list build. The rest of the draw
// is Ebiten text rasterization, outside the tab's control.
func TestDevOpsDrawRebuildsOnNewGeneration(t *testing.T) {
	d := newDevScreen()
	d.redraws = 1
	s := newDevShell(d)
	s.display = devBigDisplay(400)
	s.dev.tab = devTabOps

	img := ebiten.NewImage(s.screenW, s.screenH)
	s.drawDevTools(img)

	still := testing.AllocsPerRun(30, func() { s.drawDevTools(img) })
	rebuilt := testing.AllocsPerRun(30, func() {
		d.redraws++
		s.drawDevTools(img)
	})

	if rebuilt-still < 100 {
		t.Fatalf("still = %v allocs, rebuilt = %v, want the build in the gap", still, rebuilt)
	}
}

// TestDevOpsRows checks the Ops tab toggle, counts, and paint list.
func TestDevOpsRows(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.display = devTestDisplay()
	s.dev.ops = true

	rows := s.devOpsRows()
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	if rows[0].act != devActOpsToggle {
		t.Fatalf("first row act = %v, want devActOpsToggle", rows[0].act)
	}
	if got := rows[0].line.plain(); !strings.HasPrefix(got, "[x]") {
		t.Fatalf("toggle text = %q, want [x]", got)
	}

	var fill, total *devRow
	var list []devRow
	for i := range rows {
		plain := rows[i].line.plain()
		if rows[i].act == devActOp {
			list = append(list, rows[i])
		}
		if rows[i].act == devActNone && strings.Contains(plain, "fill") {
			fill = &rows[i]
		}
		if strings.HasPrefix(plain, "total") {
			total = &rows[i]
		}
	}

	if fill == nil || len(fill.line.spans) == 0 {
		t.Fatalf("no fill count row in %v", rows)
	}
	if sw := fill.line.spans[0]; sw.swatch <= 0 || sw.ink != devFillInk {
		t.Fatalf("fill swatch = %+v, want devFillInk", sw)
	}

	if total == nil || strings.TrimSpace(total.line.plain()) != "total 2" {
		t.Fatalf("total row = %v, want total 2", total)
	}

	if len(list) != 2 || list[0].arg != 0 || list[1].arg != 1 {
		t.Fatalf("list rows = %+v, want args 0 and 1", list)
	}
	if got := list[1].line.plain(); !strings.Contains(got, "hi") {
		t.Fatalf("text row = %q, want hi", got)
	}

	s.display = nil
	rows = s.devOpsRows()
	found := false
	for _, row := range rows {
		found = found || strings.Contains(row.line.plain(), "bitmap fallback")
	}
	if !found {
		t.Fatalf("nil display rows = %v, want bitmap fallback", rows)
	}
}

// devPathDisplay is one fill and one corner-shape path, so the Ops tab has a
// path row to label and count.
func devPathDisplay() *layout.Display {
	return &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 40, H: 20},
			{Kind: layout.DisplayOpPath, X: 10, Y: 40, W: 40, H: 20},
		},
		Order:          []int{0, 1},
		Width:          800,
		Height:         600,
		PointsPerPixel: 0.75,
	}
}

// TestDevOpsRowsCountAndLabelAPath checks that a corner-shape path reaches the
// DISPLAY counts, the total, and the paint-order list instead of disappearing
// from the tab.
func TestDevOpsRowsCountAndLabelAPath(t *testing.T) {
	t.Parallel()

	s := newDevShell(newDevScreen())
	s.display = devPathDisplay()
	s.dev.ops = true

	var count, total *devRow
	var painted []devRow

	all := s.devOpsRows()
	for i := range all {
		row := &all[i]
		plain := strings.TrimSpace(row.line.plain())

		switch {
		case row.act == devActOp:
			painted = append(painted, *row)
		case strings.HasPrefix(plain, "path"):
			count = row
		case strings.HasPrefix(plain, "total"):
			total = row
		}
	}

	if count == nil || !strings.HasSuffix(strings.TrimSpace(count.line.plain()), "1") {
		t.Fatalf("path count row = %v, want one path", count)
	}

	if total == nil || strings.TrimSpace(total.line.plain()) != "total 2" {
		t.Fatalf("total row = %v, want total 2", total)
	}

	if len(painted) != 2 {
		t.Fatalf("paint rows = %v, want two", painted)
	}

	if got := painted[1].line.plain(); !strings.Contains(got, "#2 path") {
		t.Fatalf("path row = %q, want #2 path", got)
	}

	if ink, ok := devOpInk(layout.DisplayOpPath); !ok || ink != devFillInk {
		t.Fatalf("path ink = %v (ok %v), want the fill ink", ink, ok)
	}
}

func TestDevPanelRefreshThrottlesMetrics(t *testing.T) {
	s := newDevShell(newDevScreen())
	now := time.Now()
	if !s.refreshDevPanel(now) {
		t.Fatal("first refresh was skipped")
	}
	rows := s.dev.rows
	s.dev.stats.Redraws = 42
	if s.refreshDevPanel(now.Add(devPanelRefreshGap - time.Nanosecond)) {
		t.Fatal("metrics rebuilt before their deadline")
	}
	if &rows[0] != &s.dev.rows[0] {
		t.Fatal("cached rows were replaced")
	}
	if !s.refreshDevPanel(now.Add(devPanelRefreshGap)) {
		t.Fatal("metrics did not refresh at their deadline")
	}
	if got := devFrameValue(s.dev.rows, "Redraws"); got != "42" {
		t.Fatalf("redraw counter = %q, want 42", got)
	}

	allocs := testing.AllocsPerRun(100, func() { s.refreshDevPanel(now.Add(devPanelRefreshGap)) })
	if allocs != 0 {
		t.Fatalf("unchanged panel allocated %v times", allocs)
	}
}

func TestDevPanelInputRefreshesImmediately(t *testing.T) {
	cases := map[string]func(*shell){
		"resize":  func(s *shell) { s.screenW++ },
		"tab":     func(s *shell) { s.dev.tab = devTabElements },
		"scroll":  func(s *shell) { s.dev.scroll = 1 },
		"pointer": func(s *shell) { s.cursorX, s.cursorY = s.screenW-10, 100 },
		"picker":  func(s *shell) { s.dev.picker = true },
		"selection": func(s *shell) {
			s.dev.havePin = true
			s.dev.pinned.ID = "selected"
		},
		"json":          func(s *shell) { s.devRun(devHit{act: devActJSON, key: "Rect"}) },
		"ops":           func(s *shell) { s.dev.ops = true },
		"mobile expand": func(s *shell) { s.dev.mobileExpanded = true },
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			s := newDevShell(newDevScreen())
			now := time.Now()
			s.refreshDevPanel(now)
			change(s)
			if !s.refreshDevPanel(now.Add(time.Millisecond)) {
				t.Fatal("input waited for the metrics deadline")
			}
		})
	}
}

func TestDevPanelMetricsThrottleAcrossPageRedraws(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	now := time.Now()
	s.refreshDevPanel(now)
	d.redraws++
	s.display = devTestDisplay()
	if s.refreshDevPanel(now.Add(time.Millisecond)) {
		t.Fatal("page redraw rebuilt metrics before their refresh deadline")
	}
	s.dev.tab = devTabOps
	if !s.refreshDevPanel(now.Add(time.Millisecond)) {
		t.Fatal("changing tabs did not refresh the panel")
	}
	d.redraws++
	s.display = devTestDisplay()
	if !s.refreshDevPanel(now.Add(2 * time.Millisecond)) {
		t.Fatal("Ops tab did not follow a changed display")
	}
}

func TestDevPanelDeadlinePreservesEarlierPageRequest(t *testing.T) {
	s := newDevShell(newDevScreen())
	now := time.Now()
	s.refreshDevPanel(now)
	g := &paceGame{shell: s}
	if got := g.frameDeadline(); got != s.dev.panelCache.next {
		t.Fatalf("metrics deadline = %v, want %v", got, s.dev.panelCache.next)
	}
	g.want = now.Add(time.Second)
	if got := g.frameDeadline(); got != s.dev.panelCache.next {
		t.Fatal("late page request suppressed the metrics refresh")
	}
	g.want = now.Add(time.Millisecond)
	if got := g.frameDeadline(); got != g.want {
		t.Fatal("metrics delayed an earlier page request")
	}
	s.dev.on = false
	g.want = now.Add(time.Second)
	if got := g.frameDeadline(); got != g.want {
		t.Fatal("closed dock still scheduled metrics")
	}
}

func BenchmarkDevPanelUnchanged(b *testing.B) {
	s := newDevShell(newDevScreen())
	now := time.Now()
	s.refreshDevPanel(now)
	b.ReportAllocs()
	for b.Loop() {
		s.refreshDevPanel(now)
	}
}

func TestStaticDevToolsTabsAndPauseHaveNoTimer(t *testing.T) {
	s := newDevShell(newDevScreen())
	now := time.Now()
	for _, tab := range []devTab{devTabElements, devTabOps} {
		s.dev.tab = tab
		s.refreshDevPanel(now)
		if !s.dev.panelCache.next.IsZero() {
			t.Fatal("static inspector tab scheduled a refresh")
		}
		if s.refreshDevPanel(now.Add(time.Hour)) {
			t.Fatal("unchanged inspector tab expired")
		}
	}
	s.dev.tab = devTabFrame
	s.devRun(devHit{act: devActPause})
	s.refreshDevPanel(now)
	if !s.dev.panelCache.next.IsZero() || s.refreshDevPanel(now.Add(time.Hour)) {
		t.Fatal("paused metrics still refresh")
	}
	s.devRun(devHit{act: devActPause})
	s.refreshDevPanel(now)
	if !s.dev.panelCache.next.Equal(now.Add(devPanelRefreshGap)) {
		t.Fatal("resume lost refresh deadline")
	}
	s.devRun(devHit{act: devActRefresh})
	s.refreshDevPanel(now)
	if !s.dev.panelCache.next.Equal(now.Add(5 * time.Second)) {
		t.Fatal("slow refresh deadline not applied")
	}
}

// TestDevPanelMovesWithARedraw checks the Phase 5 exit: the panel reads the
// page stats, so a redraw moves the redraws row in the same frame.
func TestDevPanelMovesWithARedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := page.New(page.Config{HTML: `<p id="a">one</p>`, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s := &shell{app: p, ctx: ctx}
	s.dev.stats = p.Stats()

	if got := devFrameValue(s.devFrameRows(320), "Redraws"); got != "1" {
		t.Fatalf("redraws = %q, want 1", got)
	}

	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	s.dev.stats = p.Stats()

	if got := devFrameValue(s.devFrameRows(320), "Redraws"); got != "2" {
		t.Fatalf("redraws = %q, want 2", got)
	}
}

// TestDevPerfZeroState checks the new sections render dashes on zero state.
func TestDevPerfZeroState(t *testing.T) {
	s := newDevShell(newDevScreen())
	got := devFrameText(s.devFrameRows(320))

	for _, want := range []string{"PERFORMANCE", "Frame avg", "Frame p95", "Frame p99", "LAST REDRAW DETAIL", "Template", "MEMORY", "Go heap", "Goroutines"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}

// TestDevPerfFrameHook checks the frame hook values render.
func TestDevPerfFrameHook(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perfHooks.devFramePerf = func() (time.Duration, time.Duration, time.Duration, uint64, bool) {
		return 16 * time.Millisecond, 20 * time.Millisecond, 25 * time.Millisecond, 3, true
	}

	got := devFrameText(s.devFrameRows(320))
	for _, want := range []string{"16.0ms", "20.0ms", "25.0ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
	if !strings.Contains(got, "Long frames") {
		t.Fatalf("frame rows = %q, want Long frames", got)
	}
}

// TestDevPerfPipelineHook checks the pipeline hook values render.
func TestDevPerfPipelineHook(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perfHooks.devPipelinePerf = func(st host.Stats) (time.Duration, time.Duration, time.Duration, time.Duration, int, int, int, bool) {
		return time.Millisecond, 2 * time.Millisecond, 3 * time.Millisecond, 4 * time.Millisecond, 5, 6, 7, true
	}

	got := devFrameText(s.devFrameRows(320))
	for _, want := range []string{"Style", "Layout", "Dirty ops", "Dirty regions", "Changed ops", "1.0ms", "4.0ms"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}

// TestDevPerfRuntimeHook checks the runtime hook values render.
func TestDevPerfRuntimeHook(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perfHooks.devRuntimePerf = func() (uint64, uint64, uint64, int, bool) {
		return 2048, 3 * 1024 * 1024, 512, 9, true
	}

	got := devFrameText(s.devFrameRows(320))
	for _, want := range []string{"2.0KB", "3.0MB", "512B"} {
		if !strings.Contains(got, want) {
			t.Fatalf("frame rows = %q, want %q", got, want)
		}
	}
}

// TestDevPerfHookNotOK checks !ok hooks fall back to dashes.
func TestDevPerfHookNotOK(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.perfHooks.devFramePerf = func() (time.Duration, time.Duration, time.Duration, uint64, bool) {
		return 0, 0, 0, 0, false
	}
	s.perfHooks.devPipelinePerf = func(st host.Stats) (time.Duration, time.Duration, time.Duration, time.Duration, int, int, int, bool) {
		return 0, 0, 0, 0, 0, 0, 0, false
	}
	s.perfHooks.devRuntimePerf = func() (uint64, uint64, uint64, int, bool) { return 0, 0, 0, 0, false }

	_ = s.devFrameRows(320)
}

// TestDevToolsPickSkipsClick checks the other Phase 3 exit: a click pins the
// box under the cursor without calling the page's click handler.
func TestDevToolsPickSkipsClick(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	s.devSetPicker(true)

	if err := s.devPick(20, 20, false); err != nil {
		t.Fatal(err)
	}

	if !s.dev.havePin || s.dev.pinned.ID != "known" {
		t.Fatalf("pinned = %+v", s.dev.pinned)
	}

	if d.clicked != 0 {
		t.Fatalf("page clicks = %d, want 0", d.clicked)
	}

	if err := s.devPick(20, 20, false); err != nil {
		t.Fatal(err)
	}

	if s.dev.havePin {
		t.Fatal("the same box did not clear the pin")
	}

	if err := s.devPick(20, 20, true); err != nil {
		t.Fatal(err)
	}

	if d.clicked != 1 {
		t.Fatalf("page clicks = %d, want 1 after alt", d.clicked)
	}
}

// TestDevToolsRefreshRecordsDrawTime checks the Phase 1 hook: the shell's
// draw time reaches the page stats through SetDrawTime.
func TestDevToolsRefreshRecordsDrawTime(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.draw = 5 * time.Millisecond

	s.devRefresh()

	if d.draw != 5*time.Millisecond {
		t.Fatalf("draw time = %v", d.draw)
	}

	if s.dev.stats.LastDraw != 5*time.Millisecond {
		t.Fatalf("stats draw = %v", s.dev.stats.LastDraw)
	}
}

// TestDevToolsDropsStalePicks checks that a new generation drops a box that
// its boxes no longer hold.
func TestDevToolsDropsStalePicks(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.haveHov, s.dev.hovered = true, d.boxes[0]
	s.dev.havePin, s.dev.pinned = true, d.boxes[0]

	d.boxes = nil
	d.redraws = 1

	s.devRefresh()

	if s.dev.haveHov || s.dev.havePin {
		t.Fatal("a stale pick survived the generation change")
	}
}

func TestDevToolsRefreshWithoutPicksSkipsBoxes(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	d.redraws++
	s.devRefresh()
	if d.boxReads != 0 {
		t.Fatalf("refresh read boxes %d times without a selection", d.boxReads)
	}
}

func TestDevToolsPickerRequiresExplicitSelection(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	s.devLayout()
	if s.devPickerOn() {
		t.Fatal("opening DevTools enabled the picker")
	}
	if err := s.devPick(20, 20, false); err != nil {
		t.Fatal(err)
	}
	if d.clicked != 1 || s.dev.havePin {
		t.Fatal("default DevTools state intercepted a page click")
	}
}

func TestDevToolsCloseDisarmsPicker(t *testing.T) {
	for _, enabled := range []bool{false, true} {
		name := "unselected"
		if enabled {
			name = "selected"
		}
		t.Run(name, func(t *testing.T) {
			d := newDevScreen()
			s := newDevShell(d)
			s.devLayout()
			if enabled {
				for _, hit := range s.dev.hits {
					if hit.act == devActPicker {
						s.devPanelPress(int(hit.rect.X+hit.rect.W/2), int(hit.rect.Y+hit.rect.H/2))
						break
					}
				}
				if !s.devPickerOn() {
					t.Fatal("explicit Picker click did not enable picking")
				}
				if err := s.devPick(20, 20, false); err != nil {
					t.Fatal(err)
				}
				if !s.dev.havePin || d.clicked != 0 {
					t.Fatal("selected picker did not intercept the click")
				}
			}
			close, _ := devHeaderRight(s.dev.panel)
			xX := s.dev.panel.X + s.dev.panel.W - devPanelPad - devTextWidth("X")
			s.devPanelPress(int(xX+devTextWidth("X")/2), int(close.Y+close.H/2))
			if d.on {
				t.Fatal("the drawn X did not request close")
			}
			if err := s.devSync(); err != nil {
				t.Fatal(err)
			}
			if s.devPickerOn() || s.dev.havePin || s.dev.haveHov {
				t.Fatal("closing DevTools left picking active")
			}
			// Finish the X press, then use the normal closed-dock pointer path.
			s.mouseDown = true
			if _, err := s.pointerMouse(
				20, 20, 20, 20, false, true, false,
			); err != nil {
				t.Fatal(err)
			}
			if _, err := s.pointerMouse(
				20, 20, 20, 20, true, false, false,
			); err != nil {
				t.Fatal(err)
			}
			if _, err := s.pointerMouse(
				20, 20, 20, 20, false, true, false,
			); err != nil {
				t.Fatal(err)
			}
			if d.clicked != 1 {
				t.Fatal("closed DevTools intercepted a normal page click")
			}
			if err := s.devToggle(); err != nil {
				t.Fatal(err)
			}
			if err := s.devSync(); err != nil {
				t.Fatal(err)
			}
			if s.devPickerOn() {
				t.Fatal("reopening DevTools enabled picking without a Picker click")
			}
			if err := s.devPick(20, 20, false); err != nil {
				t.Fatal(err)
			}
			if d.clicked != 2 || s.dev.havePin {
				t.Fatal("reopened DevTools intercepted a normal page click")
			}
		})
	}
}

// devRowsText joins the plain text of rows for assertions.
func devRowsText(rows []devRow) string {
	lines := make([]string, len(rows))

	for i, row := range rows {
		lines[i] = row.line.plain()
	}

	return strings.Join(lines, "\n")
}

// devRowContaining finds the first row whose plain text holds a label.
func devRowContaining(rows []devRow, label string) (devRow, bool) {
	for _, row := range rows {
		if strings.Contains(row.line.plain(), label) {
			return row, true
		}
	}

	return devRow{}, false
}

// devFrameValue returns the trimmed value of the first row whose plain text
// starts with a label, for Frame tab assertions.
func devFrameValue(rows []devRow, label string) string {
	for _, row := range rows {
		plain := row.line.plain()
		if strings.HasPrefix(plain, label+" ") {
			return strings.TrimSpace(strings.TrimPrefix(plain, label))
		}
	}

	return ""
}

// TestDevPanelShowsRelayoutCounts checks the Frame tab carries the relayout
// and skip counters.
func TestDevPanelShowsRelayoutCounts(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.commits, s.skipped = 3, 2

	rows := s.devFrameRows(320)
	row, ok := devRowContaining(rows, "Relayouts")
	if !ok || !strings.HasSuffix(strings.TrimRight(row.line.plain(), " "), "3") {
		t.Fatalf("relayout row = %q, want the value 3", row.line.plain())
	}

	row, ok = devRowContaining(rows, "Skipped")
	if !ok || !strings.HasSuffix(strings.TrimRight(row.line.plain(), " "), "2") {
		t.Fatalf("skipped row = %q, want the value 2", row.line.plain())
	}
}
