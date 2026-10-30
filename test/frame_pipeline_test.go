package test_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// ----------------------------------------------------------------------------
// Category 1: Frontend Pipeline & Microsecond Redraw Stages
// ----------------------------------------------------------------------------

const cardHTML = `
<style>
  .card { width: 300px; padding: 16px; background: #ffffff; border-radius: 8px; border: 1px solid #e0e0e0; }
  .title { font-size: 18px; font-weight: bold; margin-bottom: 8px; color: #333333; }
  .body { font-size: 14px; color: #666666; line-height: 1.4; margin-bottom: 12px; }
  .btn { display: inline-block; padding: 6px 12px; background: #007acc; color: #ffffff; border-radius: 4px; }
</style>
<div class="card">
  <div class="title">System Performance Card</div>
  <div class="body">This is an empirical frame latency and pipeline breakdown test for ownframe.</div>
  <button id="btn" class="btn" data-action="click-test">Execute</button>
</div>
`

const dashboardLayoutHTML = `
<style>
  body { margin: 0; padding: 0; background: #1e1e1e; color: #d4d4d4; font-family: sans-serif; }
  .layout { display: flex; width: 800px; height: 600px; }
  .sidebar { width: 200px; background: #252526; padding: 10px; display: flex; flex-direction: column; gap: 8px; }
  .nav-item { padding: 8px 12px; border-radius: 4px; background: #2d2d2d; color: #cccccc; }
  .content { flex: 1; padding: 20px; display: flex; flex-direction: column; gap: 16px; }
  .header { font-size: 20px; font-weight: bold; color: #ffffff; }
  .metrics-grid { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; }
  .metric-box { background: #2d2d2d; border: 1px solid #3c3c3c; padding: 14px; border-radius: 6px; }
  .metric-val { font-size: 24px; font-weight: bold; color: #4ec9b0; margin-top: 6px; }
</style>
<div class="layout">
  <div class="sidebar">
    <div class="nav-item">Dashboard</div>
    <div class="nav-item">Explorer</div>
    <div class="nav-item">Source Control</div>
    <div class="nav-item">Settings</div>
  </div>
  <div class="content">
    <div class="header">Operational Telemetry</div>
    <div class="metrics-grid">
      <div class="metric-box"><div>Frame Latency</div><div class="metric-val">1.2 ms</div></div>
      <div class="metric-box"><div>Memory Alloc</div><div class="metric-val">34 MB</div></div>
      <div class="metric-box"><div>Process Count</div><div class="metric-val">1</div></div>
    </div>
  </div>
</div>
`

func generateScaledTableHTML(rows int) string {
	var sb strings.Builder
	sb.WriteString(`<style>
	  .table { width: 700px; display: flex; flex-direction: column; background: #1e1e1e; font-family: monospace; }
	  .row { display: flex; height: 20px; border-bottom: 1px solid #2d2d2d; padding: 0 8px; align-items: center; }
	  .cell-id { width: 60px; color: #858585; }
	  .cell-name { width: 240px; color: #9cdcfe; }
	  .cell-status { width: 120px; color: #4ec9b0; }
	  .cell-time { width: 100px; color: #b5cea8; }
	</style><div class="table">`)
	for i := 0; i < rows; i++ {
		fmt.Fprintf(&sb, `<div class="row" id="row-%d"><span class="cell-id">#%04d</span><span class="cell-name">metric_worker_process_%d</span><span class="cell-status">RUNNING</span><span class="cell-time">%d ms</span></div>`, i, i, i, (i%15)+1)
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

func TestPipeline_StageBreakdown_Card(t *testing.T) {
	ctx := context.Background()
	p, err := ownframe.NewWithOptions(
		ownframe.Config{HTML: cardHTML, Width: 400, Height: 300},
		ownframe.WithPerf(true),
	)
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}

	t0 := time.Now()
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("Cold Redraw failed: %v", err)
	}
	coldTotal := time.Since(t0)
	statsCold := p.Probe().Stats()

	t1 := time.Now()
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("Warm Redraw failed: %v", err)
	}
	warmTotal := time.Since(t1)
	statsWarm := p.Probe().Stats()

	t.Logf("=== Simple Card UI (Ops: %d, Boxes: %d) ===", statsWarm.Ops, statsWarm.Boxes)
	t.Logf("Cold Redraw Wall Time: %v (Pipeline: Template=%v, Style=%v, Layout=%v, Paint=%v, Other=%v)",
		coldTotal, statsCold.LastTemplate, statsCold.StyleTime, statsCold.LayoutTime, statsCold.PaintTime, statsCold.OtherTime)
	t.Logf("Warm Redraw Wall Time: %v (Pipeline: Template=%v, Style=%v, Layout=%v, Paint=%v, Other=%v)",
		warmTotal, statsWarm.LastTemplate, statsWarm.StyleTime, statsWarm.LayoutTime, statsWarm.PaintTime, statsWarm.OtherTime)
	t.Logf("Allocated Memory in Redraw: %d bytes (%.2f KB)", statsWarm.AllocFrame, float64(statsWarm.AllocFrame)/1024.0)

	if warmTotal > 16*time.Millisecond {
		t.Errorf("Warm redraw %v exceeded 60 FPS frame budget (16.6ms)", warmTotal)
	}
}

func TestPipeline_StageBreakdown_Dashboard(t *testing.T) {
	ctx := context.Background()
	p, err := ownframe.NewWithOptions(
		ownframe.Config{HTML: dashboardLayoutHTML, Width: 800, Height: 600},
		ownframe.WithPerf(true),
	)
	if err != nil {
		t.Fatalf("NewWithOptions failed: %v", err)
	}

	t0 := time.Now()
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("Cold Redraw failed: %v", err)
	}
	coldTotal := time.Since(t0)
	statsCold := p.Probe().Stats()

	t1 := time.Now()
	if err := p.Redraw(ctx); err != nil {
		t.Fatalf("Warm Redraw failed: %v", err)
	}
	warmTotal := time.Since(t1)
	statsWarm := p.Probe().Stats()

	t.Logf("=== Complex Dashboard UI (Ops: %d, Boxes: %d) ===", statsWarm.Ops, statsWarm.Boxes)
	t.Logf("Cold Redraw Wall Time: %v (Template=%v, Style=%v, Layout=%v, Paint=%v)",
		coldTotal, statsCold.LastTemplate, statsCold.StyleTime, statsCold.LayoutTime, statsCold.PaintTime)
	t.Logf("Warm Redraw Wall Time: %v (Template=%v, Style=%v, Layout=%v, Paint=%v)",
		warmTotal, statsWarm.LastTemplate, statsWarm.StyleTime, statsWarm.LayoutTime, statsWarm.PaintTime)

	if warmTotal > 16*time.Millisecond {
		t.Errorf("Dashboard warm redraw %v exceeded 16.6ms frame budget", warmTotal)
	}
}

func TestPipeline_Scaling_Rows(t *testing.T) {
	ctx := context.Background()
	rowCounts := []int{10, 50, 100, 250}

	for _, count := range rowCounts {
		t.Run(fmt.Sprintf("%d_rows", count), func(t *testing.T) {
			html := generateScaledTableHTML(count)
			p, err := ownframe.NewWithOptions(
				ownframe.Config{HTML: html, Width: 800, Height: 600},
				ownframe.WithPerf(true),
			)
			if err != nil {
				t.Fatalf("NewWithOptions failed: %v", err)
			}

			t0 := time.Now()
			if err := p.Redraw(ctx); err != nil {
				t.Fatalf("Cold redraw: %v", err)
			}
			coldDur := time.Since(t0)
			coldStats := p.Probe().Stats()

			t1 := time.Now()
			if err := p.Redraw(ctx); err != nil {
				t.Fatalf("Warm redraw: %v", err)
			}
			warmDur := time.Since(t1)
			warmStats := p.Probe().Stats()

			t.Logf("[%3d rows] Ops=%4d Boxes=%4d | Cold=%8v (Style=%6v, Layout=%6v) | Warm=%8v (Style=%6v, Layout=%6v)",
				count, warmStats.Ops, warmStats.Boxes,
				coldDur, coldStats.StyleTime, coldStats.LayoutTime,
				warmDur, warmStats.StyleTime, warmStats.LayoutTime,
			)
		})
	}
}
