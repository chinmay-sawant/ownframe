package window

import "testing"

func TestMobileMetricsLayoutAndCollapse(t *testing.T) {
	for _, size := range [][2]int{{420, 934}, {934, 420}, {200, 60}} {
		s := newDevShell(newDevScreen())
		s.screenW, s.screenH = size[0], size[1]
		s.dev.on, s.dev.mobile = true, true
		s.devLayout()
		collapsed := s.dev.panel
		width, height := s.viewOf(size[0], size[1])
		if width != size[0] || height < 1 || collapsed.X != 0 || collapsed.W != float64(size[0]) {
			t.Fatalf("mobile viewport narrowed: %dx%d panel=%+v", width, height, collapsed)
		}
		if collapsed.Y < 0 || collapsed.Y+collapsed.H > float64(size[1]) {
			t.Fatalf("panel outside screen: %+v", collapsed)
		}
		if size[1] > 80 && collapsed.H != devMobileHeader {
			t.Fatalf("collapsed touch target is not 48dp: %+v", collapsed)
		}
		s.devRun(devHit{act: devActMobileExpand})
		s.devLayout()
		if !s.dev.mobileExpanded || s.dev.panel.H < collapsed.H {
			t.Fatal("mobile panel did not expand")
		}
		if size[1] > 120 && s.dev.panel.H > float64(size[1])*devMobileFraction {
			t.Fatalf("panel occupies over 40%%: %+v", s.dev.panel)
		}
		if s.dockClipX(1000) != 1000 {
			t.Fatal("mobile panel incorrectly clipped page width")
		}
		s.devRun(devHit{act: devActMobileExpand})
		s.devLayout()
		if s.dev.panel != collapsed {
			t.Fatalf("collapse did not restore geometry: %+v", s.dev.panel)
		}
	}
}

func TestMobileMetricsTouchHeaderDoesNotReachPage(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.screenW, s.screenH = 420, 934
	s.dev.on, s.dev.mobile = true, true
	s.devLayout()
	touch := touchPos{id: 1, x: 20, y: int(s.dev.panel.Y) + 20}
	content, taps, _ := devPanelStep([]touchPos{touch}, s.dev.panel, &s.dev.panelTrack)
	if len(content) != 0 || len(taps) != 0 {
		t.Fatal("metrics press leaked to application")
	}
	_, taps, _ = devPanelStep(nil, s.dev.panel, &s.dev.panelTrack)
	if len(taps) != 1 {
		t.Fatal("header lift did not produce a tap")
	}
	s.devPanelPress(taps[0].x, taps[0].y)
	if !s.dev.mobileExpanded {
		t.Fatal("header did not expand panel")
	}
	content, _, _ = devPanelStep([]touchPos{{id: 2, x: 20, y: 20}}, s.dev.panel, &s.dev.panelTrack)
	if len(content) != 1 {
		t.Fatal("application touch was swallowed outside mobile panel")
	}
}

func TestDesktopMetricsLayoutRemainsSideDock(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.screenW, s.screenH = 420, 934
	s.dev.on = true
	s.dev.mobile = false
	s.devLayout()
	if s.dev.panel != (devRect{X: 80, W: 340, H: 934}) {
		t.Fatalf("desktop dock changed: %+v", s.dev.panel)
	}
	if width, height := s.viewOf(420, 934); width != 80 || height != 934 {
		t.Fatalf("desktop viewport changed: %dx%d", width, height)
	}
	s.devRun(devHit{act: devActMobileExpand})
	if s.dev.mobileExpanded {
		t.Fatal("mobile action changed desktop state")
	}
}

func TestMobileMetricsCloseTouchRestoresViewport(t *testing.T) {
	for _, expanded := range []bool{false, true} {
		d := newDevScreen()
		s := newDevShell(d)
		s.screenW, s.screenH = 420, 934
		s.dev.mobile, s.dev.mobileExpanded = true, expanded
		s.perf = true
		s.devLayout()
		close := s.dev.hits[0]
		if close.act != devActClose || close.rect.W != 48 || close.rect.H != 48 {
			t.Fatalf("close button lacks 48dp target: %+v", close)
		}
		touch := touchPos{id: 1, x: int(close.rect.X) + 24, y: int(close.rect.Y) + 24}
		content, _, _ := devPanelStep([]touchPos{touch}, s.dev.panel, &s.dev.panelTrack)
		if len(content) != 0 {
			t.Fatal("close press reached page")
		}
		_, taps, _ := devPanelStep(nil, s.dev.panel, &s.dev.panelTrack)
		if len(taps) != 1 {
			t.Fatal("close touch did not produce tap")
		}
		s.devPanelPress(taps[0].x, taps[0].y)
		if err := s.devSync(); err != nil {
			t.Fatal(err)
		}
		if s.dev.on || d.DevTools() {
			t.Fatal("X left devtools open")
		}
		if !s.perf {
			t.Fatal("closing devtools disabled opted-in metrics")
		}
		if w, h := s.viewSize(); w != 420 || h != 934 {
			t.Fatalf("viewport not restored: %dx%d", w, h)
		}
		if d.clicked != 0 || len(s.dev.panelTrack.fingers) != 0 {
			t.Fatal("close leaked input or retained fingers")
		}
	}
}
