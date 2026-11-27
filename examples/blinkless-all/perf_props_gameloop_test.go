package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

type gameLoopPropRow struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Sample      string `json:"sample"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	ColdDrawNS  int64  `json:"cold_draw_ns"`
	SteadyNS    int64  `json:"steady_draw_ns"`
	SteadyBytes uint64 `json:"steady_alloc_bytes"`
	Ops         int    `json:"ops"`
	Boxes       int    `json:"boxes"`
	DirtyOps    int    `json:"dirty_ops"`
	DirtyArea   int    `json:"dirty_area"`
	FullLayouts uint64 `json:"full_layouts"`
	Repaints    uint64 `json:"repaints"`
	Err         string `json:"err,omitempty"`
}

// TestPropPerfGameLoopDump sweeps all 688 properties (reusing genProps and
// perf_props_harness_test.go fixtures) through ownframe.NewGame's Ebiten
// Draw/Update loop at target monitor resolutions (default 1920x1080; override
// with OWNFRAME_PROP_GAMELOOP_W / OWNFRAME_PROP_GAMELOOP_H) and writes
// gameloop-summary.md + gameloop.json into OWNFRAME_PROP_DUMP_DIR.
func TestPropPerfGameLoopDump(t *testing.T) {
	opts, err := loadPropBenchOpts()
	if err != nil {
		t.Fatal(err)
	}
	if opts.DumpDir == "" {
		t.Skip("set OWNFRAME_PROP_DUMP_DIR to run the Ebiten game-loop property dump")
	}

	w := envInt("OWNFRAME_PROP_GAMELOOP_W", 1920)
	h := envInt("OWNFRAME_PROP_GAMELOOP_H", 1080)
	nodeCount := envInt("OWNFRAME_PROP_GAMELOOP_NODES", 64)
	if err := os.MkdirAll(opts.DumpDir, 0o750); err != nil {
		t.Fatal(err)
	}

	props := genProps()
	if opts.Filter != nil {
		filtered := props[:0]
		for _, r := range props {
			if opts.Filter.MatchString(r.Name) {
				filtered = append(filtered, r)
			}
		}
		props = filtered
	}
	if opts.Limit > 0 && opts.Limit < len(props) {
		props = props[:opts.Limit]
	}

	head := propBenchHead()
	dst := ebiten.NewImage(w, h)
	defer dst.Deallocate()

	start := time.Now()
	rows := make([]gameLoopPropRow, 0, len(props))
	for _, r := range props {
		sample, _ := activeSample(r)
		decl := variantDecl(r.Name, sample)
		htmlDoc := buildUniformScaleDoc(head, decl, nodeCount)
		row := measureGameLoopProp(t.Context(), r, sample, htmlDoc, w, h, dst)
		rows = append(rows, row)
	}
	elapsed := time.Since(start)

	raw, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	jsonPath := filepath.Join(opts.DumpDir, "gameloop.json")
	if err := os.WriteFile(jsonPath, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	mdPath := filepath.Join(opts.DumpDir, "gameloop-summary.md")
	if err := os.WriteFile(mdPath, []byte(renderGameLoopSummary(rows, w, h, nodeCount, elapsed)), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Logf("wrote %d properties (%dx%d, %d nodes) in %v -> %s", len(rows), w, h, nodeCount, elapsed, mdPath)
}

func measureGameLoopProp(
	ctx context.Context,
	r genProp,
	sample, htmlDoc string,
	w, h int,
	dst *ebiten.Image,
) gameLoopPropRow {
	row := gameLoopPropRow{
		Name:   r.Name,
		Group:  r.Group,
		Sample: sample,
		Width:  w,
		Height: h,
	}

	p, err := newBenchPage(htmlDoc, w, h, true)
	if err != nil {
		row.Err = err.Error()
		return row
	}
	p.Window().SetPerf(true)
	if err := p.Redraw(ctx); err != nil {
		row.Err = err.Error()
		return row
	}

	game := ownframe.NewGame(ctx, p)
	game.Layout(w, h)

	t0 := time.Now()
	game.Draw(dst)
	row.ColdDrawNS = time.Since(t0).Nanoseconds()

	// Measure 3 steady-state Ebiten frames (Update + Draw).
	var steady [3]int64
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	for i := range steady {
		ts := time.Now()
		_ = game.Update()
		game.Draw(dst)
		steady[i] = time.Since(ts).Nanoseconds()
	}
	runtime.ReadMemStats(&after)
	slices.Sort(steady[:])
	row.SteadyNS = steady[1]
	if after.TotalAlloc >= before.TotalAlloc {
		row.SteadyBytes = (after.TotalAlloc - before.TotalAlloc) / uint64(len(steady))
	}

	st := p.Probe().Stats()
	row.Ops = st.Ops
	row.Boxes = st.Boxes
	row.DirtyOps = st.DirtyOps
	row.DirtyArea = st.DirtyArea
	row.FullLayouts = st.FullLayouts
	row.Repaints = st.Repaints
	return row
}

func renderGameLoopSummary(rows []gameLoopPropRow, w, h, nodes int, elapsed time.Duration) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Ebiten Game-Loop Property Dump (%dx%d, %d nodes, %d properties in %v)\n\n",
		w, h, nodes, len(rows), elapsed.Round(time.Millisecond))

	bySteady := slices.Clone(rows)
	slices.SortFunc(bySteady, func(a, c gameLoopPropRow) int {
		if a.SteadyNS > c.SteadyNS {
			return -1
		}
		if a.SteadyNS < c.SteadyNS {
			return 1
		}
		return strings.Compare(a.Name, c.Name)
	})

	b.WriteString("## Top 20 by steady-state Ebiten frame time\n\n")
	b.WriteString("| Property | Group | Steady Draw | Cold Draw | Steady Alloc | Ops | DirtyOps | Repaints |\n")
	b.WriteString("|---|---|---:|---:|---:|---:|---:|---:|\n")
	limit := min(20, len(bySteady))
	for _, r := range bySteady[:limit] {
		fmt.Fprintf(&b, "| `%s` | %s | **%.2f ms** | %.2f ms | %.1f KB | %d | %d | %d |\n",
			r.Name, r.Group,
			float64(r.SteadyNS)/1e6, float64(r.ColdDrawNS)/1e6,
			float64(r.SteadyBytes)/1024, r.Ops, r.DirtyOps, r.Repaints)
	}

	byCold := slices.Clone(rows)
	slices.SortFunc(byCold, func(a, c gameLoopPropRow) int {
		if a.ColdDrawNS > c.ColdDrawNS {
			return -1
		}
		if a.ColdDrawNS < c.ColdDrawNS {
			return 1
		}
		return strings.Compare(a.Name, c.Name)
	})

	b.WriteString("\n## Top 20 by cold Ebiten Draw time\n\n")
	b.WriteString("| Property | Group | Cold Draw | Steady Draw | Ops | Boxes |\n")
	b.WriteString("|---|---|---:|---:|---:|---:|\n")
	limit = min(20, len(byCold))
	for _, r := range byCold[:limit] {
		fmt.Fprintf(&b, "| `%s` | %s | **%.2f ms** | %.2f ms | %d | %d |\n",
			r.Name, r.Group,
			float64(r.ColdDrawNS)/1e6, float64(r.SteadyNS)/1e6, r.Ops, r.Boxes)
	}
	return b.String()
}
