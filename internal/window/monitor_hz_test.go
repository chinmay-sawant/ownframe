package window

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestSanitizeMonitorHz(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   int
		want int
	}{
		{-100, defaultMonitorHz},
		{-1, defaultMonitorHz},
		{0, defaultMonitorHz},
		{1, defaultMonitorHz},
		{24, 24},
		{60, 60},
		{75, 75},
		{120, 120},
		{143, 143},
		{144, 144},
		{165, 165},
		{240, 240},
		{360, 360},
		{480, 480},
		{481, defaultMonitorHz},
		{999, defaultMonitorHz},
	}

	for _, tc := range cases {
		if got := sanitizeMonitorHz(tc.in); got != tc.want {
			t.Errorf("sanitizeMonitorHz(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestRoundMonitorHz(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   float64
		want int
	}{
		{math.NaN(), 0},
		{math.Inf(1), 0},
		{math.Inf(-1), 0},
		{-60, 0},
		{0, 0},
		{59.94, 60},
		{60.0, 60},
		{119.88, 120},
		{143.98, 144},
		{165.01, 165},
		{240.0, 240},
	}

	for _, tc := range cases {
		if got := roundMonitorHz(tc.in); got != tc.want {
			t.Errorf("roundMonitorHz(%v) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestPaceGameDynamicMultiMonitorMove(t *testing.T) {
	t.Parallel()

	sample := monitorSample{posX: 100, posY: 100, winW: 1200, winH: 800}
	probeCalls := 0
	var tpsHistory []int

	// Simulate three monitors arranged horizontally:
	//   x < 1920        -> 165 Hz (laptop display)
	//   1920 <= x < 3840 -> 144 Hz (external display)
	//   x >= 3840       -> 60 Hz  (secondary TV/monitor)
	probe := func(posX, _, winW, _ int) int {
		probeCalls++
		centerX := posX + winW/2
		switch {
		case centerX < 1920:
			return 165
		case centerX < 3840:
			return 144
		default:
			return 60
		}
	}

	g := &paceGame{
		autoMonitorHz:  true,
		maxFPS:         defaultMonitorHz,
		monitorHz:      defaultMonitorHz,
		sampleMonitor:  func() monitorSample { return sample },
		probeMonitorHz: probe,
		setTPS:         func(tps int) { tpsHistory = append(tpsHistory, tps) },
	}

	t0 := time.Unix(1000, 0)

	// First check picks up the 165 Hz laptop monitor.
	g.refreshMonitorHz(t0)
	if g.maxFPS != 165 || g.monitorHz != 165 {
		t.Fatalf("initial monitor = (%d, %d), want (165, 165)", g.maxFPS, g.monitorHz)
	}
	if len(tpsHistory) != 1 || tpsHistory[0] != LoopFPS(165) {
		t.Fatalf("initial TPS = %v, want [%d]", tpsHistory, LoopFPS(165))
	}

	// Within the throttle interval, repeated frames do not re-probe.
	g.refreshMonitorHz(t0.Add(50 * time.Millisecond))
	g.refreshMonitorHz(t0.Add(100 * time.Millisecond))
	if probeCalls != 1 {
		t.Fatalf("throttled probeCalls = %d, want 1", probeCalls)
	}

	// Still on the same monitor after monitorCheckInterval: no redundant OS probe.
	t1 := t0.Add(monitorCheckInterval + 10*time.Millisecond)
	g.refreshMonitorHz(t1)
	if probeCalls != 1 {
		t.Fatalf("unchanged monitor probeCalls = %d, want 1", probeCalls)
	}

	// Drag the window onto the 144 Hz external monitor.
	sample.posX = 2000
	t2 := t1.Add(monitorCheckInterval + 10*time.Millisecond)
	g.lastLoop = t2
	g.refreshMonitorHz(t2)
	if g.maxFPS != 144 || g.monitorHz != 144 {
		t.Fatalf("after drag to 144Hz = (%d, %d), want (144, 144)", g.maxFPS, g.monitorHz)
	}
	if !g.lastLoop.IsZero() {
		t.Fatal("monitor Hz change must reset lastLoop anchor")
	}
	if len(tpsHistory) != 2 || tpsHistory[1] != LoopFPS(144) {
		t.Fatalf("TPS after drag to 144Hz = %v, want [..., %d]", tpsHistory, LoopFPS(144))
	}

	// Drag the window onto the 60 Hz monitor.
	sample.posX = 4000
	t3 := t2.Add(monitorCheckInterval + 10*time.Millisecond)
	g.refreshMonitorHz(t3)
	if g.maxFPS != 60 || g.monitorHz != 60 {
		t.Fatalf("after drag to 60Hz = (%d, %d), want (60, 60)", g.maxFPS, g.monitorHz)
	}
	if len(tpsHistory) != 3 || tpsHistory[2] != LoopFPS(60) {
		t.Fatalf("TPS after drag to 60Hz = %v, want [..., %d]", tpsHistory, LoopFPS(60))
	}
}

func TestPaceGameExplicitCapIgnoresMonitorChanges(t *testing.T) {
	t.Parallel()

	probeCalls := 0
	g := &paceGame{
		autoMonitorHz: false,
		maxFPS:        120,
		sampleMonitor: func() monitorSample { return monitorSample{posX: 2500} },
		probeMonitorHz: func(_, _, _, _ int) int {
			probeCalls++
			return 60
		},
	}

	g.refreshMonitorHz(time.Unix(2000, 0))
	if g.maxFPS != 120 {
		t.Fatalf("explicit cap changed to %d, want 120", g.maxFPS)
	}
	if probeCalls != 0 {
		t.Fatalf("explicit cap called monitor probe %d times, want 0", probeCalls)
	}
}

func TestPaceGameReprobeCatchesInPlaceRefreshSwitch(t *testing.T) {
	t.Parallel()

	// Laptop stays at the same coordinates, but switches from 165 Hz to 60 Hz
	// (e.g., battery saver / Display Settings change).
	currentHz := 165
	var tpsHistory []int
	g := &paceGame{
		autoMonitorHz:  true,
		maxFPS:         defaultMonitorHz,
		monitorHz:      defaultMonitorHz,
		sampleMonitor:  func() monitorSample { return monitorSample{posX: 0, posY: 0, winW: 800, winH: 600} },
		probeMonitorHz: func(_, _, _, _ int) int { return currentHz },
		setTPS:         func(tps int) { tpsHistory = append(tpsHistory, tps) },
	}

	t0 := time.Unix(3000, 0)
	g.refreshMonitorHz(t0)
	if g.maxFPS != 165 {
		t.Fatalf("initial maxFPS = %d, want 165", g.maxFPS)
	}

	currentHz = 60
	g.refreshMonitorHz(t0.Add(monitorReprobeInterval + time.Millisecond))
	if g.maxFPS != 60 {
		t.Fatalf("after in-place switch maxFPS = %d, want 60", g.maxFPS)
	}
	if len(tpsHistory) != 2 || tpsHistory[1] != LoopFPS(60) {
		t.Fatalf("TPS history = %v, want [%d %d]", tpsHistory, LoopFPS(165), LoopFPS(60))
	}
}

func TestPaceGameActiveCapAndIdleParkCoexist(t *testing.T) {
	t.Parallel()

	// Verify that an active animation is capped to the active monitor Hz
	// (e.g., 144 Hz -> ~6.94 ms interval, not uncapped) while an idle window
	// still transitions cleanly to ToIdle.
	const hz = 144
	interval := LoopInterval(hz)
	if interval <= 6*time.Millisecond || interval >= 8*time.Millisecond {
		t.Fatalf("144Hz LoopInterval = %v, want ~6.94ms", interval)
	}

	t0 := time.Unix(4000, 0)
	// Active animation: requesting another frame immediately keeps loopWait
	// bounding repaints to the monitor's refresh interval.
	if wait := loopWait(hz, t0, t0.Add(1*time.Millisecond)); wait != interval-1*time.Millisecond {
		t.Fatalf("active 144Hz wait = %v, want %v", wait, interval-1*time.Millisecond)
	}

	// When the animation stops (clean frames, no input, no deadline), policy
	// parks cleanly to ToIdle.
	policy := pacing.New(t0)
	at := t0
	parked := false
	for range 40 {
		at = at.Add(interval)
		if policy.Step(at, false, false, time.Time{}) == pacing.ToIdle {
			parked = true
			break
		}
	}
	if !parked || !policy.Snapshot().Idle {
		t.Fatal("monitor-paced loop failed to park cleanly when idle")
	}
}

func TestPaceWindowDefaultAutoMonitorHz(t *testing.T) {
	saved := ebiten.TPS()
	defer ebiten.SetTPS(saved)

	sh := &shell{ctx: context.Background(), screenW: 800, screenH: 600}
	out, stop := paceWindow(context.Background(), sh, Options{})
	defer stop()

	pg, ok := out.(*paceGame)
	if !ok {
		t.Fatalf("default Options{} wrapped %T, want *paceGame", out)
	}
	if idlePaceSupported {
		if !pg.autoMonitorHz {
			t.Fatal("default desktop Options{} must enable autoMonitorHz")
		}
		if pg.maxFPS < minValidMonitorHz || pg.maxFPS > maxValidMonitorHz {
			t.Fatalf("default desktop maxFPS = %d, want valid monitor Hz", pg.maxFPS)
		}
		if got := ebiten.TPS(); got != LoopFPS(pg.maxFPS) {
			t.Fatalf("default desktop TPS = %d, want LoopFPS(%d)=%d", got, pg.maxFPS, LoopFPS(pg.maxFPS))
		}
	}
}
