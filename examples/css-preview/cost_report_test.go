package main

import (
	"runtime"
	"sort"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// costReport prints the sorted numbers of one measurement.
func costReport(tb testing.TB, label string, values []float64, unit string) {
	tb.Helper()
	if len(values) == 0 {
		tb.Logf("%s: no samples", label)

		return
	}
	sort.Float64s(values)
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	tb.Logf("%s: n=%d p50=%.3f p95=%.3f max=%.3f mean=%.3f %s",
		label, len(values), values[len(values)/2],
		values[len(values)*95/100], values[len(values)-1],
		sum/float64(len(values)), unit)
}

// costAlloc reads the bytes the process allocated so far.
func costAlloc() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m.TotalAlloc
}

// costActive counts the previews that still run: the page's automatic
// players plus the manual transition samples.
func costActive(p *ownframe.Page) int {
	return p.Probe().Motion().Active
}

// costManual counts the manual samples that still run.
func costManual(d *demo) int {
	n := 0
	for _, s := range d.manual {
		if s.active {
			n++
		}
	}

	return n
}
