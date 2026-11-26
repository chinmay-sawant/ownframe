package ui

import (
	"bytes"
	"fmt"
	"image/color"
	"image/png"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// bigGrid loads a 1000-row result so the grid overflows.
func (r *rig) bigGrid() {
	r.t.Helper()
	r.must(r.a.RunSQL(r.ctx, "WITH RECURSIVE c(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM c WHERE n < 1000) SELECT n AS id, 'row ' || n AS name FROM c"))
}

func TestScrollbarThumbTracksGridPosition(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.bigGrid()
	tab := r.a.activeTabPtr()
	for step := 0; step < 12; step++ {
		body, thumb := r.box("grid-body"), r.box("grid-thumb")
		content, view := len(tab.grid.rows)*gridRowHeight, int(body.H)
		want := chadcn.ScrollMetrics{Content: content, Viewport: view, Offset: int(tab.grid.scrollY)}
		top, h, vis := want.Thumb(view)
		if !vis {
			t.Fatal("thumb should be visible for 1000 rows")
		}
		if got := thumb.Y - body.Y; got < float64(top)-1 || got > float64(top)+1 {
			t.Fatalf("step %d: thumb top %.1f, want %d (scroll %v)", step, got, top, tab.grid.scrollY)
		}
		if thumb.H < float64(h)-1 || thumb.H > float64(h)+1 {
			t.Fatalf("step %d: thumb height %.1f, want %d", step, thumb.H, h)
		}
		if thumb.H >= body.H {
			t.Fatal("thumb must be shorter than the track when content overflows")
		}
		_, err := r.a.onWheel(r.ctx, body.X+60, body.Y+60, 0, -6)
		r.must(err)
	}
	// At the very end the thumb touches the bottom of the track.
	tab.grid.scrollY = 1e9
	tab.grid.clampScroll(r.a.gridViewW(tab), r.a.gridViewH(tab), 0)
	r.must(r.a.redraw(r.ctx))
	body, thumb := r.box("grid-body"), r.box("grid-thumb")
	if d := body.Y + body.H - (thumb.Y + thumb.H); d < -1 || d > 1 {
		t.Fatalf("thumb should rest on the track end, gap %.1f", d)
	}
	// A short result has no bar at all.
	r.must(r.a.RunSQL(r.ctx, "SELECT 1 AS one"))
	if r.has("grid-thumb") {
		t.Fatal("no scrollbar while the rows fit")
	}
}

// runsAt returns the vertical runs of pixels equal to want in column x
// between y0 and y1, decoding the frame once.
func runsAt(r *rig, want color.RGBA, x, y0, y1 int) [][2]int {
	img, err := png.Decode(bytes.NewReader(r.a.Page().Probe().PNGWithContext(r.ctx)))
	if err != nil {
		r.t.Fatal(err)
	}
	var runs [][2]int
	start := -1
	for y := y0; y < y1; y++ {
		c, _ := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
		if c == want {
			if start < 0 {
				start = y
			}
		} else if start >= 0 {
			runs = append(runs, [2]int{start, y})
			start = -1
		}
	}
	if start >= 0 {
		runs = append(runs, [2]int{start, y1})
	}
	return runs
}

// TestScrollbarThumbStaysOnePiece: through repeated scrolling the thumb is one
// vertical run of thumb-coloured pixels whose length is the thumb height.
func TestScrollbarThumbStaysOnePiece(t *testing.T) {
	for _, dark := range []bool{true, false} {
		r := newRig(t, 1440, 960, "")
		if !dark {
			r.tap("tb-theme")
		}
		thumbColor := color.RGBA{0x4a, 0x4a, 0x4a, 0xff}
		if !dark {
			thumbColor = color.RGBA{0xc4, 0xc4, 0xc4, 0xff}
		}
		r.bigGrid()
		for step := 0; step < 8; step++ {
			body, thumb := r.box("grid-body"), r.box("grid-thumb")
			x := int(thumb.X + thumb.W/2)
			runs := runsAt(r, thumbColor, x, int(body.Y), int(body.Y+body.H))
			long := 0
			for _, run := range runs {
				if run[1]-run[0] >= 20 {
					long++
				}
			}
			if long != 1 {
				t.Fatalf("dark=%v step %d: thumb split into %d long runs %v", dark, step, long, runs)
			}
			dy := -3.0
			if step >= 4 {
				dy = 3
			}
			_, err := r.a.onWheel(r.ctx, body.X+60, body.Y+60, 0, dy)
			r.must(err)
		}
	}
}

func TestScrollbarDragAndTrackPress(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.bigGrid()
	tab := r.a.activeTabPtr()
	thumb := r.box("grid-thumb")
	claimed, err := r.a.onDragStart(r.ctx, thumb)
	r.must(err)
	if !claimed {
		t.Fatal("the thumb must claim the drag")
	}
	body := r.box("grid-body")
	m := r.a.regions["grid"].st.Metrics()
	_, th, _ := m.Thumb(m.Viewport)
	travel := float64(m.Viewport - th)
	r.must(r.a.onDragMove(r.ctx, 0, travel/2))
	want := float64(m.MaxOffset()) / 2
	if d := tab.grid.scrollY - want; d < -2*gridRowHeight || d > 2*gridRowHeight {
		t.Fatalf("dragging half the travel scrolled %v, want about %v", tab.grid.scrollY, want)
	}
	r.must(r.a.onDragEnd(r.ctx))
	moved := r.box("grid-thumb")
	if moved.Y <= thumb.Y {
		t.Fatal("thumb should have followed the drag")
	}
	// Pressing the track below the thumb jumps there and claims the press.
	tab.grid.scrollY = 0
	r.must(r.a.redraw(r.ctx))
	before := tab.grid.scrollY
	after := r.box("grid-bar-after")
	claim, err := r.a.onPress(r.ctx, after, after.X+2, body.Y+body.H*0.8)
	r.must(err)
	if !claim || tab.grid.scrollY <= before {
		t.Fatalf("track press: claim=%v scroll %v -> %v", claim, before, tab.grid.scrollY)
	}
	// Dragging past either end clamps.
	r.must(r.a.onDragEnd(r.ctx))
	_, _ = r.a.onDragStart(r.ctx, r.box("grid-thumb"))
	r.must(r.a.onDragMove(r.ctx, 0, 1e6))
	if int(tab.grid.scrollY) > m.MaxOffset() {
		t.Fatal("drag overshoot must clamp")
	}
	r.must(r.a.onDragMove(r.ctx, 0, -2e6))
	if tab.grid.scrollY != 0 {
		t.Fatalf("drag back overshoot -> %v", tab.grid.scrollY)
	}
	r.must(r.a.onDragEnd(r.ctx))
}

// TestScrollbarsInEveryScrollableArea: each area that can overflow draws a bar,
// and the bar vanishes when the content fits.
func TestScrollbarsInEveryScrollableArea(t *testing.T) {
	// Editor: more lines than rows.
	r := newRig(t, 1440, 960, "")
	r.a.activeTabPtr().ed = newEditor(strings.Repeat("SELECT 1;\n", 60))
	r.must(r.a.redraw(r.ctx))
	if !r.has("editor-thumb") {
		t.Error("editor lacks a scrollbar")
	}
	r.a.activeTabPtr().ed = newEditor("SELECT 1")
	r.must(r.a.redraw(r.ctx))
	if r.has("editor-thumb") {
		t.Error("editor bar should hide when everything fits")
	}
	// Monitor.
	r.openMonitor()
	if !r.has("monitor-thumb") {
		t.Error("monitor lacks a scrollbar")
	}
	// Palette: 12 entries, 10 rows.
	r.chord("k", false)
	if !r.has("pal-thumb") {
		t.Error("palette lacks a scrollbar")
	}
	r.a.pal.query = "zzzz"
	r.must(r.a.redraw(r.ctx))
	if r.has("pal-thumb") {
		t.Error("palette bar should hide for an empty result")
	}
	r.key("escape")

	// Sidebar tree: the mock PostgreSQL has three schemas.
	p := newRig(t, 1280, 700, "mock:postgres")
	for _, s := range p.a.schema {
		for _, g := range s.Children {
			p.a.expanded["group:"+s.Name+":"+g.Name] = true
		}
	}
	p.must(p.a.redraw(p.ctx))
	if !p.has("tree-thumb") {
		t.Error("expanded schema tree lacks a scrollbar")
	}
	body := p.box("tree-vp")
	before := p.box("tree-thumb").Y
	_, err := p.a.onWheel(p.ctx, body.X+20, body.Y+40, 0, -5)
	p.must(err)
	if p.box("tree-thumb").Y <= before {
		t.Error("wheel over the tree should move its thumb")
	}

	// History list.
	h := newRig(t, 1440, 960, "")
	for i := 0; i < 30; i++ {
		_, _ = h.a.prefs.AddHistory(h.ctx, store.HistoryEntry{ConnectionID: h.a.cfg.ID, SQL: fmt.Sprintf("SELECT %d", i), Rows: 1})
	}
	h.tap("rail-history")
	if !h.has("hist-thumb") {
		t.Error("history list lacks a scrollbar")
	}
	hv := h.box("hist-vp")
	if hv.H > float64(histPaneH)+1 {
		t.Errorf("history pane is %v tall, want at most %d", hv.H, histPaneH)
	}
	_, err = h.a.onWheel(h.ctx, hv.X+30, hv.Y+30, 0, -4)
	h.must(err)
	if h.a.regions["hist"].st.Offset() == 0 {
		t.Error("wheel over the history list should scroll it")
	}

	// Inspector with long JSON.
	i := newRig(t, 1280, 700, "")
	i.openTableByName("audit_events")
	i.tap("tbl-inspector")
	long := "{" + strings.Repeat(`"k":"v",`, 80) + `"z":1}`
	i.a.activeTabPtr().grid.rows[0][3] = jsonCell(long)
	i.a.activeTabPtr().grid.ar, i.a.activeTabPtr().grid.ac, i.a.activeTabPtr().grid.fr, i.a.activeTabPtr().grid.fc = 0, 3, 0, 3
	i.must(i.a.redraw(i.ctx))
	if !i.has("insp-thumb") {
		t.Error("inspector lacks a scrollbar for long JSON")
	}
}

func jsonCell(s string) driver.CellValue { return driver.JSON(s) }
