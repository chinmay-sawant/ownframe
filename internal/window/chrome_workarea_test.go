package window

import "testing"

func testMonitors() []monitorGeometry {
	return []monitorGeometry{
		{name: "XWAYLAND0", w: 2560, h: 1600},
		{name: "XWAYLAND1", w: 1920, h: 1080},
	}
}

func TestAreaMatchesMonitorSize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		monW, monH int
		area       workarea
		want       bool
	}{
		{"exact fit", 1920, 1080, workarea{w: 1920, h: 1080}, true},
		{"taskbar gap fits", 1920, 1080, workarea{w: 1920, h: 1040}, true},
		{"wider than monitor", 1920, 1080, workarea{w: 2560, h: 1040}, false},
		{"taller than monitor", 1920, 1080, workarea{w: 1920, h: 1200}, false},
		{"gap too big is another monitor", 2560, 1600, workarea{w: 1920, h: 1080}, false},
		{"small window is not a workarea", 1920, 1080, workarea{w: 800, h: 600}, false},
	}
	for _, tc := range cases {
		if got := areaMatchesMonitorSize(tc.monW, tc.monH, tc.area); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestAttributeWorkarea(t *testing.T) {
	t.Parallel()

	mons := testMonitors()
	cases := []struct {
		name     string
		area     workarea
		claimant string
		wantMon  string
		wantOK   bool
	}{
		{
			"claimant wins when it fits",
			workarea{x: 0, y: 320, w: 1920, h: 1040}, "XWAYLAND1",
			"XWAYLAND1", true,
		},
		{
			"2560 rect claimed by the 1920 panel moves over",
			workarea{x: 0, y: 0, w: 2560, h: 1528}, "XWAYLAND1",
			"XWAYLAND0", true,
		},
		{
			"no plausible monitor keeps the claimant",
			workarea{x: 100, y: 100, w: 800, h: 600}, "XWAYLAND1",
			"XWAYLAND1", false,
		},
		{
			"unknown claimant still attributes a clear match",
			workarea{x: 1920, y: 0, w: 2560, h: 1528}, "gone",
			"XWAYLAND0", true,
		},
	}
	for _, tc := range cases {
		if mon, ok := attributeWorkarea(tc.area, tc.claimant, mons); mon != tc.wantMon || ok != tc.wantOK {
			t.Errorf("%s: got (%q, %v), want (%q, %v)",
				tc.name, mon, ok, tc.wantMon, tc.wantOK)
		}
	}

	if mon, ok := attributeWorkarea(workarea{w: 1920, h: 1040}, "XWAYLAND1", nil); mon != "XWAYLAND1" || ok {
		t.Errorf("no monitors: got (%q, %v), want claimant without match", mon, ok)
	}
}
