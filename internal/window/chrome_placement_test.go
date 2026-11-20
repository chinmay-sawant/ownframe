package window

import (
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// closeStage runs the named update stage on a bare shell.
func closeStage(t *testing.T) func(*shell) error {
	t.Helper()

	for i := range stages {
		if stages[i].name == "close" {
			return stages[i].run
		}
	}

	t.Fatal("close stage missing from the update pipeline")

	return nil
}

func TestCloseRequestRoundTrip(t *testing.T) {
	if takeCloseRequest(false) {
		t.Fatal("stale close request leaked into the test")
	}

	RequestClose()

	if !takeCloseRequest(false) {
		t.Fatal("RequestClose did not arm a close")
	}

	if takeCloseRequest(false) {
		t.Fatal("close request was not consumed")
	}
}

func TestCloseStageEndsLoop(t *testing.T) {
	run := closeStage(t)
	s := &shell{app: &fakeScreen{width: 100, height: 100}, ctx: context.Background()}

	if err := run(s); err != nil {
		t.Fatalf("idle close stage = %v, want nil", err)
	}

	RequestClose()

	if err := run(s); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("armed close stage = %v, want ebiten.Termination", err)
	}

	if takeCloseRequest(false) {
		t.Fatal("close request survived the stage")
	}
}

func TestIsWSLVersion(t *testing.T) {
	cases := []struct {
		version string
		want    bool
	}{
		{"Linux version 5.15.153.1-microsoft-standard-WSL2 (root@9417f6fb4c0e) #1 SMP ...", true},
		{"Linux version 4.4.0-19041-Microsoft (Microsoft@Microsoft.com) ...", true},
		{"Linux version 6.8.0-52-generic (buildd@lcy02-amd64-101) ...", false},
		{"", false},
	}

	for _, tc := range cases {
		if got := isWSLVersion(tc.version); got != tc.want {
			t.Fatalf("isWSLVersion(%q) = %v, want %v", tc.version, got, tc.want)
		}
	}
}

func TestManualMaximizeAbandonsStaleProbe(t *testing.T) {
	manualMax.learning = true
	manualMax.restoring = false
	manualMax.until = time.Now().Add(-time.Second)

	completeManualMaximize(1920, 1032)

	if manualMax.learning {
		t.Fatal("stale probe still armed after the deadline")
	}

	if manualMax.active || manualMax.known {
		t.Fatal("stale probe changed maximize state")
	}
}

func TestManualMaximizeIgnoresIdle(t *testing.T) {
	manualMax.learning = false
	manualMax.restoring = false
	manualMax.active = false
	manualMax.known = false

	completeManualMaximize(1920, 1032)

	if manualMax.learning || manualMax.restoring || manualMax.active || manualMax.known {
		t.Fatal("idle complete changed maximize state")
	}
}

func TestManualMaximizeAbandonsStaleRestore(t *testing.T) {
	manualMax.learning = false
	manualMax.restoring = true
	manualMax.active = false
	manualMax.until = time.Now().Add(-time.Second)

	completeManualMaximize(800, 600)

	if manualMax.restoring {
		t.Fatal("stale restore still armed after the deadline")
	}

	if manualMax.active {
		t.Fatal("stale restore activated the maximize")
	}
}

func TestCloseWaitsForMouseRelease(t *testing.T) {
	RequestClose()
	t.Cleanup(func() { takeCloseRequest(false) })
	if takeCloseRequest(true) {
		t.Fatal("destroyed the window before WSLg could deliver mouse release")
	}
	if !takeCloseRequest(false) {
		t.Fatal("lost close request while waiting for release")
	}
	if takeCloseRequest(false) {
		t.Fatal("close request was consumed more than once")
	}
}

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

func TestF11TogglesFullscreen(t *testing.T) {
	t.Parallel()

	full := false
	var applied []bool

	s := &shell{}
	s.readFullscreen = func() bool { return full }
	s.applyFullscreen = func(on bool) { applied = append(applied, on) }

	if !s.f11(true) {
		t.Fatal("f11 was not consumed")
	}

	if len(applied) != 1 || !applied[0] {
		t.Fatalf("applied = %v", applied)
	}

	if !s.f11(false) {
		t.Fatal("f11 release was not consumed")
	}

	full = true
	if !s.f11(true) {
		t.Fatal("second f11 was not consumed")
	}

	if len(applied) != 2 || applied[1] {
		t.Fatalf("applied = %v", applied)
	}
}

func TestParseXrandrMonitors(t *testing.T) {
	t.Parallel()

	out := "Monitors: 2\n" +
		" 0: +XWAYLAND0 2560/677x1600/423+1920+0  XWAYLAND0\n" +
		" 1: +XWAYLAND1 1920/527x1080/296+0+320  XWAYLAND1\n"
	rects := parseXrandrMonitors(out)
	if len(rects) != 2 {
		t.Fatalf("parsed %d monitors, want 2", len(rects))
	}
	if rects[0] != (monitorRect{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600}) {
		t.Fatalf("first rect = %+v", rects[0])
	}
	if rects[1] != (monitorRect{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080}) {
		t.Fatalf("second rect = %+v", rects[1])
	}
}

func TestMonitorForWindowPicksHeaderMonitor(t *testing.T) {
	t.Parallel()

	layout := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	// Double-tap header at the center of the 1120x760 frame on the 1920 panel.
	m, ok := monitorForWindow(100, 400, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND1" {
		t.Fatalf("1920 window attributed to %+v, ok=%v", m, ok)
	}
	// Same frame moved to the 2560 panel.
	m, ok = monitorForWindow(2500, 400, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND0" {
		t.Fatalf("2560 window attributed to %+v, ok=%v", m, ok)
	}
	// The stuck strip at x=-960 keeps 160px on the 1920 panel: it overlaps
	// that panel, but less than half its area, so saves must skip it.
	m, ok = monitorForWindow(-960, 425, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND1" {
		t.Fatalf("strip attributed to %+v, ok=%v", m, ok)
	}
	if substantiallyVisibleInLayout(-960, 425, 1120, 760, layout) {
		t.Fatal("off-screen strip reported substantially visible")
	}
	if !substantiallyVisibleInLayout(100, 400, 1120, 760, layout) {
		t.Fatal("sane window reported insubstantial")
	}
}

func TestClampDragTargetInUnion(t *testing.T) {
	t.Parallel()

	minX, minY, maxX, maxY := 0, 0, 4480, 1600
	// Sane position passes through.
	if x, y := clampDragTargetInUnion(2500, 400, 1120, 760, minX, minY, maxX, maxY); x != 2500 || y != 400 {
		t.Fatalf("visible clamped to %d,%d", x, y)
	}
	// Far right pulls back to keep a grabbable strip.
	if x, _ := clampDragTargetInUnion(9000, 400, 1120, 760, minX, minY, maxX, maxY); x != maxX-dragMinVisibleX {
		t.Fatalf("right clamped to %d", x)
	}
	// Below the union (summed-height bound allowed 2648) now clamps.
	if _, y := clampDragTargetInUnion(200, 9000, 1120, 760, minX, minY, maxX, maxY); y != maxY-dragMinVisibleY {
		t.Fatalf("bottom clamped to %d", y)
	}
}

func TestPlacementSubstantiallyVisible(t *testing.T) {
	t.Parallel()

	if !substantiallyVisibleInLayout(100, 400, 1120, 760, []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}) {
		t.Fatal("sane window reported insubstantial")
	}
}

func TestOriginMustBeInsideUnion(t *testing.T) {
	t.Parallel()

	layout := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	// Fresh cascade at y=149 sits in the dead zone above the 1920 panel:
	// area overlaps, but the header is unreachable.
	if originInLayout(128, 149, layout) {
		t.Fatal("dead-zone origin reported visible")
	}
	if !originInLayout(0, 320, layout) {
		t.Fatal("1920 top-left reported invisible")
	}
	if !originInLayout(1920, 0, layout) {
		t.Fatal("2560 top-left reported invisible")
	}
}

func TestTargetMonitorStartsAtTopLeft(t *testing.T) {
	t.Parallel()

	layout := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	// Smallest panel that fits 1120x760 is the 1920 one.
	saved := "XWAYLAND1"
	if target, ok := targetMonitorInLayout(layout, saved, 1120, 760); !ok || target.name != "XWAYLAND1" || target.x != 0 || target.y != 320 {
		t.Fatalf("target = %+v, ok=%v", target, ok)
	}
}

func TestValidatePlacementID(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"code-view", "app_1", "A-0_z"} {
		if err := validatePlacementID(id); err != nil {
			t.Fatalf("validatePlacementID(%q) = %v, want nil", id, err)
		}
	}

	for _, id := range []string{"", "  ", "has space", "a/b", "a\\b", "..", "a..b", "tab\there", "semi;colon", "dot.name", "slash/name"} {
		if err := validatePlacementID(id); err == nil {
			t.Fatalf("validatePlacementID(%q) = nil, want error", id)
		}
	}

	long := make([]byte, placementMaxIDLen+1)
	for i := range long {
		long[i] = 'a'
	}

	if err := validatePlacementID(string(long)); err == nil {
		t.Fatal("overlong placement id accepted")
	}
}

func TestPlacementRoundTrip(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	want := windowPlacement{X: 12, Y: 24, W: 1500, H: 900, Maximized: true, MaxX: 0, MaxY: 0, MaxW: 1920, MaxH: 1032, Monitor: "XWAYLAND1"}

	if err := savePlacement("code-view", want); err != nil {
		t.Fatalf("savePlacement = %v", err)
	}

	got, ok := loadPlacement("code-view")
	if !ok {
		t.Fatal("loadPlacement missed a saved file")
	}

	want.Version = placementVersion
	if got != want {
		t.Fatalf("placement = %+v, want %+v", got, want)
	}

	path := filepath.Join(config, "ownframe", "placement", "code-view.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat placement file = %v", err)
	}

	if info.Mode().Perm() != placementFileMode {
		t.Fatalf("placement mode = %o, want %o", info.Mode().Perm(), placementFileMode)
	}
}

func TestWSLMaximizePlan(t *testing.T) {
	t.Parallel()

	applyNow, retry := wslMaximizePlan(false, true)
	if applyNow || !retry {
		t.Fatalf("unknown monitor: apply=%v retry=%v, want deferred", applyNow, retry)
	}

	applyNow, retry = wslMaximizePlan(true, true)
	if !applyNow || !retry {
		t.Fatalf("fitting monitor: apply=%v retry=%v, want both", applyNow, retry)
	}

	applyNow, retry = wslMaximizePlan(true, false)
	if applyNow || retry {
		t.Fatalf("oversized rect: apply=%v retry=%v, want neither", applyNow, retry)
	}
}

func TestPlacementLoadMissingFallsBack(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	if _, ok := loadPlacement("nothing-saved"); ok {
		t.Fatal("missing placement loaded")
	}

	if _, err := placementPath("bad/id"); err == nil {
		t.Fatal("bad placement id resolved a path")
	}
}

func TestPlacementRejectsBadRect(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	if err := savePlacement("code-view", windowPlacement{W: 0, H: 900}); err != nil {
		t.Fatalf("savePlacement = %v", err)
	}

	if _, ok := loadPlacement("code-view"); ok {
		t.Fatal("zero-width placement loaded")
	}
}

func TestPlacementFitsUnknownMonitorKeepsOldBehavior(t *testing.T) {
	t.Parallel()

	if !placementFits("no-such-monitor-xyz", 2560, 1528) {
		t.Fatal("unknown monitor should not discard the rect")
	}
}

func TestSavePlacementNowSkipsWorkareaDance(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	prevLearning, prevRestoring := manualMax.learning, manualMax.restoring
	manualMax.learning = true
	manualMax.restoring = false
	t.Cleanup(func() {
		manualMax.learning, manualMax.restoring = prevLearning, prevRestoring
	})

	savePlacementNow("dance-guard")

	path := filepath.Join(config, "ownframe", "placement", "dance-guard.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("dance save wrote %q during learning, err=%v", path, err)
	}

	manualMax.learning = false
	manualMax.restoring = true

	savePlacementNow("dance-guard")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("dance save wrote %q during restoring, err=%v", path, err)
	}
}

func TestRestoreManualForDragClearsActive(t *testing.T) {
	prevActive := manualMax.active
	prevW, prevH := manualMax.sw, manualMax.sh
	t.Cleanup(func() {
		manualMax.active = prevActive
		manualMax.sw, manualMax.sh = prevW, prevH
	})

	manualMax.active = true
	manualMax.sw, manualMax.sh = 1500, 900

	sw, sh, restored := restoreManualForDrag()
	if !restored {
		t.Fatal("drag restore did not trigger while manually maximized")
	}

	if manualMax.active {
		t.Fatal("drag restore left the manual maximize active")
	}

	if sw <= 0 || sh <= 0 {
		t.Fatalf("drag restore size = %dx%d, want the windowed size", sw, sh)
	}

	if _, _, restored := restoreManualForDrag(); restored {
		t.Fatal("second drag restore triggered without a maximize")
	}
}

func TestApplyWorkareaKeysByMonitorName(t *testing.T) {
	prevPending, prevUntil := pendingManualRestore, pendingManualUntil
	t.Cleanup(func() { pendingManualRestore, pendingManualUntil = prevPending, prevUntil })
	prevActive, prevKnown := manualMax.active, manualMax.known
	prevMon := manualMax.learnedMon
	prevAreas := manualMax.areas
	t.Cleanup(func() {
		manualMax.active, manualMax.known = prevActive, prevKnown
		manualMax.learnedMon = prevMon
		manualMax.areas = prevAreas
	})

	manualMax.areas = nil
	manualMax.sx, manualMax.sy = 100, 100
	manualMax.sw, manualMax.sh = 1500, 900

	applyWorkarea("XWAYLAND1", workarea{x: 0, y: 320, w: 1920, h: 1040})
	if !pendingManualRestore || !pendingManualUntil.IsZero() {
		t.Fatal("workarea must await post-resize position acknowledgement")
	}

	if !manualMax.active || !manualMax.known {
		t.Fatal("applyWorkarea did not mark the maximize active")
	}

	if manualMax.learnedMon != "XWAYLAND1" {
		t.Fatalf("learned monitor = %q, want %q", manualMax.learnedMon, "XWAYLAND1")
	}

	got, ok := manualMax.areas["XWAYLAND1"]
	if !ok {
		t.Fatal("workarea not cached under the monitor name")
	}

	if got.w != 1920 || got.h != 1040 || got.x != 0 || got.y != 320 {
		t.Fatalf("cached workarea = %+v, want 0,320 1920x1040", got)
	}
}

func TestWorkareaMatchesMonitorUnknownMonitorKeepsOldBehavior(t *testing.T) {
	t.Parallel()

	if !workareaMatchesMonitor("no-such-monitor-xyz", workarea{w: 2560, h: 1528}) {
		t.Fatal("unknown monitor should not discard the rect")
	}
}

func TestClampDragTargetKeepsStripVisible(t *testing.T) {
	t.Parallel()

	// Far off-screen right/bottom pulls back to the bound edge.
	if x, y := clampDragTarget(9000, 9000, 1500, 900, 4480, 2680); x != 4480-dragMinVisibleX || y != 2680-dragMinVisibleY {
		t.Fatalf("clamped = %d,%d, want %d,%d", x, y, 4480-dragMinVisibleX, 2680-dragMinVisibleY)
	}

	// Far off-screen left/top pulls back so a strip stays reachable.
	if x, y := clampDragTarget(-9000, -9000, 1500, 900, 4480, 2680); x != -(1500-dragMinVisibleX) || y != -(900-dragMinVisibleY) {
		t.Fatalf("clamped = %d,%d, want %d,%d", x, y, -(1500 - dragMinVisibleX), -(900 - dragMinVisibleY))
	}

	// A sane on-screen position passes through untouched.
	if x, y := clampDragTarget(2000, 400, 1500, 900, 4480, 2680); x != 2000 || y != 400 {
		t.Fatalf("clamped = %d,%d, want 2000,400", x, y)
	}
}

func TestClampDragTargetSkipsUnknownBound(t *testing.T) {
	t.Parallel()

	if x, y := clampDragTarget(-6368, 9274, 2560, 1528, 0, 0); x != -6368 || y != 9274 {
		t.Fatalf("unknown bound clamped to %d,%d, want untouched", x, y)
	}
}

func TestPositionVisibleInBound(t *testing.T) {
	t.Parallel()

	// 1920 + 2560 side by side, heights summed generously.
	const bw, bh = 4480, 2680

	if !positionVisibleInBound(89, 537, 1120, 760, bw, bh) {
		t.Fatal("sane ui-builder origin reported off-screen")
	}
	if !positionVisibleInBound(2000, 400, 1120, 760, bw, bh) {
		t.Fatal("visible right-monitor origin reported off-screen")
	}
	if positionVisibleInBound(9000, 9000, 1120, 760, bw, bh) {
		t.Fatal("far right origin reported visible")
	}
	if positionVisibleInBound(-9000, -9000, 1120, 760, bw, bh) {
		t.Fatal("far left origin reported visible")
	}
	if !positionVisibleInBound(9000, 9000, 1120, 760, 0, 0) {
		t.Fatal("unknown bound should keep the old behavior")
	}
}

func TestExpandedResolutionMustFillMonitor(t *testing.T) {
	t.Parallel()

	// A 1920 workarea is not an expanded resolution on the 2560 panel.
	if areaMatchesMonitorSize(2560, 1600, workarea{w: 1920, h: 1040}) {
		t.Fatal("1920 workarea matched the 2560 panel")
	}
	// A 2560 workarea is not an expanded resolution on the 1920 panel.
	if areaMatchesMonitorSize(1920, 1080, workarea{w: 2560, h: 1528}) {
		t.Fatal("2560 workarea matched the 1920 panel")
	}
	// The right workarea still matches its own panel, minus the taskbar gap.
	if !areaMatchesMonitorSize(2560, 1600, workarea{w: 2560, h: 1528}) {
		t.Fatal("2560 workarea did not match the 2560 panel")
	}
	if !areaMatchesMonitorSize(1920, 1080, workarea{w: 1920, h: 1040}) {
		t.Fatal("1920 workarea did not match the 1920 panel")
	}
}

func TestLiveMonitorResolutionNeverPanics(t *testing.T) {
	t.Parallel()

	w, h, ok := liveMonitorResolution()
	if !ok {
		return
	}
	if w <= 0 || h <= 0 {
		t.Fatalf("live resolution = %dx%d, want positive", w, h)
	}
}

func TestPlacementSamplingThrottlesUnchangedGeometry(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := &shell{placementID: "sampling-test", lastPlacementValid: true, lastPlacement: currentPlacement()}
	s.syncPlacement()
	sampled := s.lastPlacementSave
	if sampled.IsZero() {
		t.Fatal("unchanged geometry did not advance the sample timer")
	}
	s.syncPlacement()
	if !s.lastPlacementSave.Equal(sampled) {
		t.Fatal("sampled geometry again before the interval elapsed")
	}
}

func BenchmarkPlacementSamplingUnchanged(b *testing.B) {
	b.Setenv("XDG_CONFIG_HOME", b.TempDir())
	s := &shell{placementID: "sampling-bench", lastPlacementValid: true, lastPlacement: currentPlacement()}
	s.syncPlacement()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.syncPlacement()
	}
}

func TestRestoreRejectsCascadedWorkarea(t *testing.T) {
	want := workarea{x: -1920, y: 320, w: 1920, h: 1032}
	for _, offset := range []int{1, 40, 48, 50} {
		actual := want
		actual.y += offset
		if restoreRectMatches(actual, want) {
			t.Fatalf("accepted %d-pixel top gap", offset)
		}
	}
	if !restoreRectMatches(want, want) {
		t.Fatal("correct saved rectangle rejected")
	}
}

func TestPartialWindowChoosesLargestVisibleArea(t *testing.T) {
	layout := []monitorRect{{name: "left", x: -1920, y: 320, w: 1920, h: 1080}, {name: "right", w: 2560, h: 1600}}
	for _, percent := range []int{10, 20, 40, 45, 50, 55, 60, 80, 90} {
		// A 1000px-wide window with this percentage on the right panel.
		m, ok := monitorForWindow(percent*10-1000, 400, 1000, 600, layout)
		want := "left"
		if percent >= 50 {
			want = "right"
		}
		if !ok || m.name != want {
			t.Fatalf("%d%% on right: got %q, want %q", percent, m.name, want)
		}
	}
	// The center is in the left monitor, but most visible pixels are right.
	m, ok := monitorForWindow(-600, 0, 1000, 600, layout)
	if !ok || m.name != "right" {
		t.Fatalf("vertical offset selected %+v", m)
	}
}

func TestWorkareaProbeWaitsForFinalPosition(t *testing.T) {
	var p workareaProbe
	now := time.Now()
	partial := workarea{x: -1920, y: 360, w: 1920, h: 1032}
	final := partial
	final.y = 320
	if p.observe(partial, now) {
		t.Fatal("accepted first configure")
	}
	if p.observe(final, now.Add(50*time.Millisecond)) {
		t.Fatal("accepted moving workarea")
	}
	if p.observe(final, now.Add(100*time.Millisecond)) {
		t.Fatal("accepted before settlement")
	}
	if !p.observe(final, now.Add(200*time.Millisecond)) {
		t.Fatal("final workarea did not settle")
	}
}

func TestHostWorkareaUsesSelectedMonitor(t *testing.T) {
	screens := []hostWorkarea{
		{Bounds: desktopRect{X: 0, Y: 0, W: 2560, H: 1600}, Work: desktopRect{X: 0, Y: 0, W: 2560, H: 1528}},
		{Bounds: desktopRect{X: -1920, Y: 320, W: 1920, H: 1080}, Work: desktopRect{X: -1920, Y: 320, W: 1920, H: 1032}},
	}
	area, ok := matchHostWorkarea(monitorRect{x: -1920, y: 320, w: 1920, h: 1080}, screens)
	if !ok || area != (workarea{x: -1920, y: 320, w: 1920, h: 1032}) {
		t.Fatalf("left workarea=%+v ok=%v", area, ok)
	}
}

func TestTitlebarRespectsAppOwnedChrome(t *testing.T) {
	if runtime.GOOS == "js" || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		t.Skip("desktop titlebar")
	}
	if !usesTitlebar(Options{}) {
		t.Fatal("default desktop window omitted shared titlebar")
	}
	for _, o := range []Options{
		{CustomHeader: true}, {Borderless: true}, {Transparent: true},
		{MousePassthrough: true}, {Interactive: func(int, int) bool { return true }}, {NativeHeader: true},
	} {
		if usesTitlebar(o) {
			t.Fatalf("titlebar replaced explicitly configured chrome: %+v", o)
		}
	}
}

func TestHeaderTitleUsesProportionalFontAndFits(t *testing.T) {
	face, err := newHeaderFont()
	if err != nil {
		t.Fatal(err)
	}
	wide, _ := text.Measure("WWW", face, 0)
	narrow, _ := text.Measure("III", face, 0)
	if wide <= narrow {
		t.Fatal("header title still uses a fixed-width face")
	}
	const title = "Workspace — a long window title"
	got := fitHeaderTitle(title, face, 90)
	width, _ := text.Measure(got, face, 0)
	if !strings.HasSuffix(got, "…") || width > 90 {
		t.Fatalf("fitted title=%q width=%f", got, width)
	}
	if got := fitHeaderTitle("Workspace", face, 200); got != "Workspace" {
		t.Fatalf("short title changed: %q", got)
	}
}

func TestTitlebarKeepsPageViewport(t *testing.T) {
	s := &shell{titlebar: titlebarState{enabled: true}}
	w, h := s.Layout(1000, 836)
	if w != 1000 || h != 836 || s.pendingH != 800 || s.screenH != 800 {
		t.Fatalf("canvas=%dx%d content=%dx%d screenH=%d", w, h, s.pendingW, s.pendingH, s.screenH)
	}
	s.titlebar.enabled = false
	s.Layout(1000, 800)
	if s.pendingH != 800 {
		t.Fatalf("custom chrome viewport changed: %d", s.pendingH)
	}
}

func TestCustomHeaderHeightKeepsPageViewport(t *testing.T) {
	s := &shell{titlebar: titlebarState{enabled: true, style: headerStyle{height: headerHeight(Options{HeaderHeight: 48})}}}
	w, h := s.Layout(900, 648)
	if w != 900 || h != 648 || s.pendingH != 600 || s.titlebarInset() != 48 {
		t.Fatalf("canvas=%dx%d contentHeight=%d headerHeight=%d", w, h, s.pendingH, s.titlebarInset())
	}
	if headerHeight(Options{}) != 36 || headerHeight(Options{HeaderHeight: 1}) != 24 {
		t.Fatal("header default or control minimum changed")
	}
}

func TestTitlebarControlHitBounds(t *testing.T) {
	for _, tc := range []struct{ x, want int }{{0, 0}, {861, 0}, {862, 1}, {907, 1}, {908, 2}, {953, 2}, {954, 3}, {999, 3}, {1000, 0}} {
		if got := titlebarButton(tc.x, 1000); got != tc.want {
			t.Fatalf("x=%d: control=%d want=%d", tc.x, got, tc.want)
		}
	}
}

func TestWindowDragNeedsUserPress(t *testing.T) {
	var d windowDrag
	handled, move, _, _, _ := d.step(200, 100, false, true)
	if handled || move {
		t.Fatal("hover moved window")
	}
	handled, move, _, _, _ = d.step(200, 100, true, false)
	if handled || move {
		t.Fatal("transparent pixels began drag")
	}
	d.step(200, 100, false, false)
	d.step(200, 100, true, true)
	_, move, _, dx, dy := d.step(230, 120, true, false)
	if !move || dx != 30 || dy != 20 {
		t.Fatal("drag delta incorrect")
	}
	_, move, _, _, _ = d.step(200, 100, true, true)
	if move {
		t.Fatal("settled grab point moved again")
	}
	_, _, click, _, _ := d.step(200, 100, false, true)
	if click || d.active {
		t.Fatal("drag released as click")
	}
	d.step(200, 100, true, true)
	_, _, click, _, _ = d.step(201, 100, false, true)
	if !click {
		t.Fatal("small gesture did not click")
	}
}

func TestRecoverMotionCancelsOwnWindowShift(t *testing.T) {
	t.Parallel()

	// Steady hand while the window slides under it: the cursor mirror
	// cancels the window shift and the hand reads zero.
	if dx, dy := recoverMotion(900, 100, 1000, 100, 100, 50, 0, 50); dx != 0 || dy != 0 {
		t.Fatalf("steady hand reads %d,%d, want 0,0", dx, dy)
	}

	// Real hand motion passes through untouched.
	if dx, dy := recoverMotion(1010, 120, 1000, 100, 100, 50, 100, 50); dx != 10 || dy != 20 {
		t.Fatalf("hand motion reads %d,%d, want 10,20", dx, dy)
	}
}

func TestWindowRebaseIgnoresMonitorOrigin(t *testing.T) {
	t.Parallel()

	// Landed request, still in flight, and a partial step are all ours.
	if windowRebased(-1283, 384, -1290, 382, -1290, 382) {
		t.Fatal("landed request looked like a new origin")
	}

	if windowRebased(-1283, 384, -1290, 382, -1283, 384) {
		t.Fatal("unapplied request looked like a new origin")
	}

	if windowRebased(-1283, 384, -1290, 382, -1286, 383) {
		t.Fatal("partial apply looked like a new origin")
	}

	// drag.log line 189: the window origin jumped to the other monitor
	// while the cursor stayed in the titlebar.
	if !windowRebased(-1290, 382, -1290, 382, 630, 62) {
		t.Fatal("monitor origin jump was treated as our request")
	}
}

func TestDragAppliesFastHandInFull(t *testing.T) {
	t.Parallel()

	// drag.log line 26: the cursor moved 190px in one sample. The window
	// has to take the same step or it trails the pointer.
	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 1640, 138
	d.lsx, d.lsy = 1640, 138
	d.lwx, d.lwy = -452, 244

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 1450, rawY: 174,
		sx: 1450, sy: 174,
		wx: -452, wy: 244,
	})
	if !ok || dx != -190 || dy != 36 {
		t.Fatalf("fast hand = %d,%d ok=%v, want -190,36", dx, dy, ok)
	}
}

func TestDragCursorJumpDoesNotReverse(t *testing.T) {
	t.Parallel()

	// drag.log lines 39-40. The raw cursor jumps by about one monitor
	// width while the window stays where we put it. The first sample is
	// refused and the smoother rewound. The second sample, still in the
	// new place, adopts that coordinate without moving the window.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 1187, 219
	d.lrx, d.lry = 1187, 219
	d.lsx, d.lsy = 1187, 219
	d.lwx, d.lwy = -1096, 354
	d.reqx, d.reqy = -1096, 354
	d.haveReq = true

	_, _, ok := d.prepareDrag(dragSample{
		rawX: 1187, rawY: 219,
		sx: 1187, sy: 219,
		wx: -1096, wy: 354,
	})
	if ok {
		t.Fatal("held cursor moved the window")
	}

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 3115, rawY: 64,
		sx: 2151, sy: 64,
		wx: -1096, wy: 354,
	})
	if ok {
		t.Fatalf("cursor jump applied %d,%d", dx, dy)
	}

	if !d.jumpDrop || d.jumpSnap || d.lrx != 1187 || d.fx != 1187 {
		t.Fatalf("first jump snapped, drop=%v snap=%v cursor=%d fx=%v", d.jumpDrop, d.jumpSnap, d.lrx, d.fx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 3115, rawY: 64,
		sx: 2151, sy: 64,
		wx: -1096, wy: 354,
	})
	if ok {
		t.Fatalf("stable coordinate moved the window by %d,%d", dx, dy)
	}

	if !d.jumpSnap || d.lrx != 3115 || d.lry != 64 {
		t.Fatalf("stable coordinate not adopted, snap=%v cursor=%d,%d", d.jumpSnap, d.lrx, d.lry)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 3145, rawY: 64,
		sx: 3145, sy: 64,
		wx: -1096, wy: 354,
	})
	if !ok || dx != 30 || dy != 0 {
		t.Fatalf("motion after adopt = %d,%d ok=%v, want 30,0", dx, dy, ok)
	}
}

func TestDragKeepsRealHand(t *testing.T) {
	t.Parallel()

	// Motion under the jump threshold moves the window by the full hand
	// delta, including a step that crosses a monitor.
	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 1640, 138
	d.lsx, d.lsy = 1640, 138
	d.lwx, d.lwy = -452, 244
	d.haveReq = true
	d.reqx, d.reqy = -452, 244

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 1450, rawY: 174,
		sx: 1450, sy: 174,
		wx: -452, wy: 244,
	})
	if !ok || dx != -190 || dy != 36 {
		t.Fatalf("hand = %d,%d ok=%v, want -190,36", dx, dy, ok)
	}
}

func TestDragSeamWindowRebaseStaysWithCursor(t *testing.T) {
	t.Parallel()

	// drag.log lines 188-189. The presented origin jumps by the monitor
	// rebase. The cursor only moved a few pixels in the titlebar.
	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 630, 281
	d.lsx, d.lsy = 630, 281
	d.lwx, d.lwy = -1290, 382
	d.reqx, d.reqy = -1290, 382
	d.haveReq = true

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 605, rawY: 275,
		sx: 605, sy: 275,
		wx: 630, wy: 62,
	})
	if !ok || dx != -25 || dy != -6 {
		t.Fatalf("rebased hand = %d,%d ok=%v, want -25,-6", dx, dy, ok)
	}

	if !d.rebased || d.rebaseX != -1290 || d.rebaseY != 382 {
		t.Fatalf("rebase from %d,%d flagged %v", d.rebaseX, d.rebaseY, d.rebased)
	}
}

func TestDragDoesNotOscillateAfterApply(t *testing.T) {
	t.Parallel()

	var d windowDrag
	_, _, ok := d.prepareDrag(dragSample{
		rawX: 100, rawY: 50,
		sx: 100, sy: 50,
		wx: 0, wy: 0,
	})
	if ok {
		t.Fatal("anchor frame moved the window")
	}

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 200, rawY: 50,
		sx: 200, sy: 50,
		wx: 0, wy: 0,
	})
	if !ok || dx != 100 || dy != 0 {
		t.Fatalf("hand = %d,%d ok=%v, want 100,0", dx, dy, ok)
	}

	d.commitDrag(0, 0, dx, dy, false)

	// The window caught the hand and the window-relative cursor mirrored
	// back. A steady hand must read zero, not the inverse of the move.
	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 100, rawY: 50,
		sx: 100, sy: 50,
		wx: 100, wy: 0,
	})
	if ok {
		t.Fatalf("settled hand moved %d,%d", dx, dy)
	}
}

func TestDragCursorSpikeRewindsWhenTheHandReturns(t *testing.T) {
	t.Parallel()

	// drag.log line 781. The smoothed cursor was -414 and the raw sample
	// reconstructed from the low-pass is 3414. The next sample is back
	// on the titlebar. Applying the spike is what threw the window from
	// x=-1328 to about 1551.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = -414, 15
	d.lrx, d.lry = -414, 15
	d.lsx, d.lsy = -414, 15
	d.lwx, d.lwy = -1328, 425
	d.haveReq = true
	d.reqx, d.reqy = -1328, 425

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 3414, rawY: -697,
		sx: 1500, sy: -341,
		wx: -1328, wy: 425,
	})
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("spike applied %d,%d ok=%v", dx, dy, ok)
	}

	if d.lrx != -414 || d.lwx != -1328 || d.fx != -414 {
		t.Fatalf("spike moved anchors cursor=%d window=%d fx=%v", d.lrx, d.lwx, d.fx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: -411, rawY: 16,
		sx: -414, sy: 15,
		wx: -1328, wy: 425,
	})
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("return moved the window by %d,%d ok=%v", dx, dy, ok)
	}
}

func TestDragCursorSpikeTwoFramesThenReturn(t *testing.T) {
	t.Parallel()

	// drag.log lines 1141-1143. Two raw samples sit in the bogus
	// coordinate space and are too far apart to be the same rebase.
	// The third sample is back near the titlebar. None of the bogus
	// steps may move the window.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 1001, 39
	d.lrx, d.lry = 1001, 39
	d.lsx, d.lsy = 1001, 39
	d.lwx, d.lwy = -1219, 427

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 2923, rawY: -279,
		sx: 1962, sy: -120,
		wx: -1219, wy: 427,
	})
	if ok || dx != 0 || dy != 0 {
		t.Fatalf("first spike applied %d,%d", dx, dy)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 1900, rawY: -100,
		sx: 1450, sy: -30,
		wx: -1219, wy: 427,
	})
	if ok || dx != 0 || dy != 0 || d.jumpSnap {
		t.Fatalf("second spike applied %d,%d snap=%v", dx, dy, d.jumpSnap)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: 984, rawY: 45,
		sx: 992, sy: 42,
		wx: -1219, wy: 427,
	})
	if !ok || dx != -9 || dy != 3 {
		t.Fatalf("return = %d,%d ok=%v, want -9,3", dx, dy, ok)
	}
}

func TestDragCursorJumpAdoptsStableCoordinates(t *testing.T) {
	t.Parallel()

	// drag.log lines 892-893. The cursor's coordinate space changes and
	// stays there. The window must not take the jump, and a later small
	// move in the new space must apply.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 765, 3
	d.lrx, d.lry = 765, 3
	d.lsx, d.lsy = 765, 3
	d.lwx, d.lwy = -135, -7

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: -735, rawY: 350,
		sx: 15, sy: 176,
		wx: -135, wy: -7,
	})
	if ok || dx != 0 || dy != 0 || d.lrx != 765 {
		t.Fatalf("first jump applied %d,%d cursor=%d", dx, dy, d.lrx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: -700, rawY: 360,
		sx: -700, sy: 360,
		wx: -135, wy: -7,
	})
	if ok || dx != 0 || dy != 0 || !d.jumpSnap || d.lrx != -700 {
		t.Fatalf("adopt = %d,%d ok=%v snap=%v cursor=%d", dx, dy, ok, d.jumpSnap, d.lrx)
	}

	dx, dy, ok = d.prepareDrag(dragSample{
		rawX: -660, rawY: 360,
		sx: -660, sy: 360,
		wx: -135, wy: -7,
	})
	if !ok || dx != 40 || dy != 0 {
		t.Fatalf("motion after adopt = %d,%d ok=%v, want 40,0", dx, dy, ok)
	}
}

func TestDragAppliesThreeHundredPixelStep(t *testing.T) {
	t.Parallel()

	var d windowDrag
	d.tracked = true
	d.lrx, d.lry = 1000, 100
	d.lsx, d.lsy = 1000, 100
	d.lwx, d.lwy = 0, 0

	dx, dy, ok := d.prepareDrag(dragSample{
		rawX: 1300, rawY: 80,
		sx: 1300, sy: 80,
		wx: 0, wy: 0,
	})
	if !ok || dx != 300 || dy != -20 {
		t.Fatalf("step = %d,%d ok=%v, want 300,-20", dx, dy, ok)
	}
}

func TestDragSpikeLeavesInFlightRequestToCancel(t *testing.T) {
	t.Parallel()

	// A position request is in flight. The next sample is a cursor spike
	// while the window is still at the old origin. The sample after that
	// is the request landing and the cursor back on the titlebar. The
	// landing cancels through the cursor mirror.
	var d windowDrag
	d.tracked = true
	d.fx, d.fy = 500, 40
	d.lrx, d.lry = 500, 40
	d.lsx, d.lsy = 500, 40
	d.lwx, d.lwy = 0, 0

	dx, _, ok := d.prepareDrag(dragSample{
		rawX: 600, rawY: 40,
		sx: 600, sy: 40,
		wx: 0, wy: 0,
	})
	if !ok || dx != 100 {
		t.Fatalf("hand = %d ok=%v, want 100", dx, ok)
	}

	d.commitDrag(0, 0, 100, 0, false)
	_, _, ok = d.prepareDrag(dragSample{
		rawX: 2500, rawY: 40,
		sx: 1550, sy: 40,
		wx: 0, wy: 0,
	})
	if ok || d.lwx != 0 || d.lrx != 600 {
		t.Fatalf("spike moved anchors window=%d cursor=%d ok=%v", d.lwx, d.lrx, ok)
	}

	dx, _, ok = d.prepareDrag(dragSample{
		rawX: 500, rawY: 40,
		sx: 500, sy: 40,
		wx: 100, wy: 0,
	})
	if ok {
		t.Fatalf("landed request applied again, dx=%d", dx)
	}
}

func TestNegligibleDragWithholdsMicroJitter(t *testing.T) {
	t.Parallel()

	if !negligibleDrag(1, -1) {
		t.Fatal("1px jitter not withheld")
	}

	if negligibleDrag(2, 0) {
		t.Fatal("2px move withheld")
	}
}

func TestSmoothCursorHalvesDistance(t *testing.T) {
	t.Parallel()

	if got := smoothCursor(0, 100); got != 50 {
		t.Fatalf("smoothed = %v, want 50", got)
	}

	if got := smoothCursor(100, 100); got != 100 {
		t.Fatalf("settled cursor moved to %v", got)
	}

	prev := 0.0
	for range 20 {
		prev = smoothCursor(prev, 100)
	}

	if prev < 99.9 {
		t.Fatalf("cursor did not converge, stuck at %v", prev)
	}
}

func TestDoubleClickPair(t *testing.T) {
	t.Parallel()

	now := time.Now()
	if !doubleClickPair(now.Add(-300*time.Millisecond), 100, 40, now, 103, 42) {
		t.Fatal("close press in time should pair")
	}
	if doubleClickPair(time.Time{}, 100, 40, now, 100, 40) {
		t.Fatal("zero previous press must not pair")
	}
	if doubleClickPair(now.Add(-600*time.Millisecond), 100, 40, now, 100, 40) {
		t.Fatal("slow second press is two single clicks")
	}
	if doubleClickPair(now.Add(-100*time.Millisecond), 100, 40, now, 200, 40) {
		t.Fatal("far second press is a new drag, not a pair")
	}
}

func TestDragPointControl(t *testing.T) {
	t.Parallel()

	// Settings-card shape: nested bare text under an id-bearing card
	// without an action must stay chrome, while the same nesting under
	// an action button is a control.
	boxes := []layout.Box{
		{ID: "settings-card", Tag: "div", X: 400, Y: 200, W: 580, H: 340, ParentIndex: -1},
		{ID: "settings-panel", Tag: "div", X: 450, Y: 260, W: 480, H: 220, ParentIndex: 0},
		{Tag: "span", Text: "Appearance", X: 470, Y: 300, W: 90, H: 16, ParentIndex: 1},
		{ID: "theme-dark", Tag: "button", Action: "theme:dark", X: 470, Y: 380, W: 440, H: 60, ParentIndex: 1},
		{Tag: "span", Text: "Dark", X: 490, Y: 392, W: 34, H: 16, ParentIndex: 3},
		{ID: "mail-search", Tag: "input", X: 900, Y: 30, W: 240, H: 36, ParentIndex: -1},
	}
	cases := []struct {
		name string
		x, y float64
		want bool
	}{
		{"card title text is chrome", 500, 308, false},
		{"card padding is chrome", 920, 500, false},
		{"option label resolves to its button", 500, 400, true},
		{"search field is a control", 950, 48, true},
		{"outside everything is chrome", 10, 10, false},
	}
	for _, tc := range cases {
		if got := dragPointControl(boxes, tc.x, tc.y); got != tc.want {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestDragPointControlBadParentIndex(t *testing.T) {
	t.Parallel()

	boxes := []layout.Box{
		{Tag: "span", Text: "x", X: 0, Y: 0, W: 10, H: 10, ParentIndex: 9},
	}
	if dragPointControl(boxes, 5, 5) {
		t.Fatal("dangling parent index must not count as a control")
	}
}

func TestDragPointControlIDLessAction(t *testing.T) {
	t.Parallel()

	// Titlebar shape: the close/min/max buttons carry a data-action but no
	// id. They must still count as controls so a press on them never starts
	// a window drag or a double-click maximize.
	boxes := []layout.Box{
		{Tag: "header", X: 0, Y: 0, W: 1120, H: 68, ParentIndex: -1},
		{Tag: "div", X: 900, Y: 10, W: 210, H: 48, ParentIndex: 0},
		{Tag: "button", Action: "win-close", X: 1060, Y: 16, W: 40, H: 36, ParentIndex: 1},
		{Tag: "span", Text: "×", X: 1072, Y: 24, W: 16, H: 20, ParentIndex: 2},
	}
	if !dragPointControl(boxes, 1065, 20) {
		t.Fatal("id-less action button must count as a control")
	}
	if !dragPointControl(boxes, 1080, 34) {
		t.Fatal("nested label must resolve to its id-less action button")
	}
	if dragPointControl(boxes, 500, 34) {
		t.Fatal("bare titlebar padding must stay chrome")
	}
}

func TestWSLgLeftMonitorDragAndMaximize(t *testing.T) {
	raw := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	layout := wslgMonitorLayout(raw)
	minX, minY, maxX, maxY, ok := virtualUnion(layout)
	if !ok {
		t.Fatal("missing monitor bounds")
	}
	// The captured drag was wrongly pinned at -960 by the unshifted union.
	if x, y := clampDragTargetInUnion(-1200, 600, 1120, 760, minX, minY, maxX, maxY); x != -1200 || y != 600 {
		t.Fatalf("left-panel drag clamped to %d,%d", x, y)
	}
	m, ok := monitorForWindow(-1200, 600, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND1" || m.x != -1920 || m.y != 320 {
		t.Fatalf("left maximize target=%+v ok=%v", m, ok)
	}
	m, ok = monitorForWindow(720, 420, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND0" || m.x != 0 || m.y != 0 {
		t.Fatalf("right maximize target=%+v ok=%v", m, ok)
	}
	if raw[1].x != 1920 {
		t.Fatal("source layout mutated")
	}
}
