package depthalert

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/depth-alert/alert"
)

func TestDisplayFilterRejectsDistanceOutlier(t *testing.T) {
	var d displayFilter
	for _, mm := range []int32{2100, 2200, 3900, 2100, 2200} {
		d.observe(mm, 50, true, alert.Stop)
	}
	if got := d.view(true).Distance; got != "2.1 m" {
		t.Fatalf("distance = %q, want stable value 2.1 m", got)
	}
}

func TestDisplayFilterHoldsBriefMissingFrames(t *testing.T) {
	var d displayFilter
	for range 3 {
		d.observe(4000, 50, true, alert.Clear)
	}
	d.observe(0, 0, true, alert.Unknown)
	d.observe(0, 0, true, alert.Unknown)
	if got := d.view(true); got.State != "CLEAR AHEAD" || got.Distance == "—" {
		t.Fatalf("brief gap changed view: %+v", got)
	}
	d.observe(0, 0, true, alert.Unknown)
	if got := d.view(true); got.State != "DEPTH UNKNOWN" || got.Distance != "—" {
		t.Fatalf("stale view = %+v, want unknown without stale metrics", got)
	}
	d.observe(1000, 50, true, alert.StopNow)
	if got := d.view(true); got.State != "STOP NOW" || got.Distance != "1.0 m" {
		t.Fatalf("reacquired view = %+v", got)
	}
}

func TestDisplayFilterUnavailableClearsImmediately(t *testing.T) {
	var d displayFilter
	d.observe(4000, 50, true, alert.Clear)
	d.unavailable()
	if got := d.view(true); got.State != "DEPTH UNKNOWN" || got.Distance != "—" {
		t.Fatalf("unavailable view = %+v", got)
	}
}

func TestDisplayFilterDebouncesStateDowngrade(t *testing.T) {
	var d displayFilter
	d.observe(2000, 50, true, alert.Stop)
	d.observe(4000, 50, true, alert.Clear)
	d.observe(4000, 50, true, alert.Clear)
	if got := d.view(true).State; got != "STOP" {
		t.Fatalf("state after two clear samples = %q", got)
	}
	d.observe(4000, 50, true, alert.Clear)
	if got := d.view(true).State; got != "CLEAR AHEAD" {
		t.Fatalf("state after third clear sample = %q", got)
	}
}
