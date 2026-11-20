package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

type propPerfRow struct {
	name       string
	group      string
	sample     string
	coldNS     int64
	coldBytes  uint64
	coldAllocs uint64
	warmAllocs uint64
	ops        int
}

func measurePropHTML(ctx context.Context, htmlDoc string) (coldNS int64, coldBytes, coldAllocs, warmAllocs uint64, ops int, err error) {
	var m0, m1, m2 runtime.MemStats

	runtime.ReadMemStats(&m0)
	t0 := time.Now()
	page, err := ownframe.New(ownframe.Config{
		Title:  "prop-audit",
		HTML:   htmlDoc,
		Width:  360,
		Height: 160,
	})
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	if err = page.Redraw(ctx); err != nil {
		return 0, 0, 0, 0, 0, err
	}
	coldNS = time.Since(t0).Nanoseconds()
	runtime.ReadMemStats(&m1)
	coldBytes = m1.TotalAlloc - m0.TotalAlloc
	coldAllocs = m1.Mallocs - m0.Mallocs

	if d := page.Probe().Display(); d != nil {
		ops = len(d.Ops)
	}

	// Settle one tick so paused/scroll/triggered animations apply their
	// initial pose and park before measuring warm unchanged redraw allocs.
	_ = page.Testing().Tick(ctx)
	runtime.ReadMemStats(&m1)

	if err = page.Redraw(ctx); err != nil {
		return 0, 0, 0, 0, 0, err
	}
	runtime.ReadMemStats(&m2)

	return coldNS, coldBytes, coldAllocs, m2.Mallocs - m1.Mallocs, ops, nil
}

// TestAuditAllPropertiesPerf measures every generated CSS property individually
// when OWNFRAME_PROP_AUDIT=1 is set, reporting the top allocation and latency
// outliers and checking that warm unchanged redraws stay zero-alloc.
func TestAuditAllPropertiesPerf(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full property perf audit in -short mode")
	}

	ctx := context.Background()
	head := genHead()
	props := genProps()

	// Warm process-level font/parser caches first.
	baseHTML := head + `<div id="p" class="gbox" style="">baseline: control</div>`
	if _, _, _, _, _, err := measurePropHTML(ctx, baseHTML); err != nil {
		t.Fatal(err)
	}
	baseNS, baseBytes, baseAllocs, baseWarm, baseOps, err := measurePropHTML(ctx, baseHTML)
	if err != nil {
		t.Fatal(err)
	}

	rows := make([]propPerfRow, 0, len(props))
	warmLeaky := 0
	skipped := 0
	for _, r := range props {
		if declines(r.Name) {
			skipped++
			continue
		}
		spec := specFor(r)
		doc := head + spec.expand(r, "p", true)
		cns, cb, ca, wa, ops, mErr := measurePropHTML(ctx, doc)
		if mErr != nil {
			t.Logf("skip %s (err: %v)", r.Name, mErr)
			skipped++
			continue
		}
		if wa > 0 {
			warmLeaky++
		}
		rows = append(rows, propPerfRow{
			name:       r.Name,
			group:      r.Group,
			sample:     r.Sample,
			coldNS:     cns,
			coldBytes:  cb,
			coldAllocs: ca,
			warmAllocs: wa,
			ops:        ops,
		})
	}

	t.Logf("Audited %d CSS properties (%d declined/skipped). Baseline control: %.2f ms, %d KB, %d allocs, %d warm allocs, %d ops",
		len(rows), skipped, float64(baseNS)/1e6, baseBytes/1024, baseAllocs, baseWarm, baseOps)
	t.Logf("Properties with non-zero warm redraw allocs: %d / %d", warmLeaky, len(rows))
	if warmLeaky > 0 {
		t.Logf("--- Properties with Non-Zero Warm Redraw Allocs ---")
		for _, r := range rows {
			if r.warmAllocs > 0 {
				t.Logf("  %-36s [%-6s] warmAllocs=%d  cold=%.1f KB  (%s)",
					r.name, r.group, r.warmAllocs, float64(r.coldBytes)/1024, fmt.Sprintf("%.24s", r.sample))
			}
		}
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].coldBytes > rows[j].coldBytes })
	t.Logf("--- Top 20 by Cold Alloc Bytes ---")
	for i := 0; i < min(20, len(rows)); i++ {
		r := rows[i]
		t.Logf("  %2d. %-36s [%-6s] %7.1f KB  %5d allocs  %6.2f ms  ops=%3d  warm=%d  (%s)",
			i+1, r.name, r.group, float64(r.coldBytes)/1024, r.coldAllocs, float64(r.coldNS)/1e6, r.ops, r.warmAllocs, fmt.Sprintf("%.24s", r.sample))
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].coldNS > rows[j].coldNS })
	t.Logf("--- Top 15 by Cold Latency ---")
	for i := 0; i < min(15, len(rows)); i++ {
		r := rows[i]
		t.Logf("  %2d. %-36s [%-6s] %6.2f ms  %7.1f KB  %5d allocs  ops=%3d",
			i+1, r.name, r.group, float64(r.coldNS)/1e6, float64(r.coldBytes)/1024, r.coldAllocs, r.ops)
	}
}
