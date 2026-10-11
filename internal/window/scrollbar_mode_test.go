package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

type scrollbarModeScreen struct {
	*fakeScreen
	mode host.ScrollbarMode
}

func (s *scrollbarModeScreen) ScrollbarMode() host.ScrollbarMode { return s.mode }

func TestScrollbarModePlatformDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode host.ScrollbarMode
		goos string
		want bool
	}{
		{name: "desktop default", mode: host.ScrollbarAuto, goos: "windows", want: true},
		{name: "web default", mode: host.ScrollbarAuto, goos: "js", want: true},
		{name: "android default", mode: host.ScrollbarAuto, goos: "android"},
		{name: "ios default", mode: host.ScrollbarAuto, goos: "ios"},
		{name: "show on android", mode: host.ScrollbarVisible, goos: "android", want: true},
		{name: "hide on desktop", mode: host.ScrollbarHidden, goos: "linux"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scrollbarModeEnabled(tc.mode, tc.goos); got != tc.want {
				t.Fatalf("scrollbarModeEnabled(%v, %q) = %t", tc.mode, tc.goos, got)
			}
		})
	}
}

func TestScrollbarGrabHonorsHiddenMode(t *testing.T) {
	t.Parallel()

	s := &shell{
		app: &scrollbarModeScreen{
			fakeScreen: &fakeScreen{
				width: 100, height: 100,
				boxes: []layout.Box{{X: 0, Y: 0, W: 100, H: 200}},
			},
			mode: host.ScrollbarHidden,
		},
		screenW: 100, screenH: 100,
	}

	if _, _, ok := s.scrollbarGrab(95, 10); ok {
		t.Fatal("hidden scrollbar was hit-testable")
	}
}
