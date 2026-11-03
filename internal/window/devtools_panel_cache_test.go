package window

import (
	"testing"
	"time"
)

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
