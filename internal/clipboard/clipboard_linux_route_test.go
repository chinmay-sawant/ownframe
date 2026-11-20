//go:build linux && !android

package clipboard

import "testing"

func TestRouteLinux(t *testing.T) {
	none := func(string) (string, bool) { return "", false }
	have := func(string) (string, bool) { return "/bin/tool", true }
	onlyCopy := func(n string) (string, bool) { return n, n == "wl-copy" }
	onlyPaste := func(n string) (string, bool) { return n, n == "wl-paste" }
	tests := []struct {
		name, wayland, display string
		look                   func(string) (string, bool)
		want                   clipHost
	}{
		{"tools", "w", ":0", have, hostWayland},
		{"no tools display", "w", ":0", none, hostX11},
		{"no tools no display", "w", "", none, hostNone},
		{"x11 only", "", ":0", none, hostX11},
		{"neither", "", "", none, hostNone},
		{"x11 ignores tools", "", ":0", have, hostX11},
		{"only copy", "w", ":0", onlyCopy, hostX11},
		{"only paste", "w", ":0", onlyPaste, hostX11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routeLinux(tt.wayland, tt.display, tt.look)
			if got != tt.want {
				t.Fatalf("got %s want %s", got, tt.want)
			}
		})
	}
}

func TestDiagnoseLinux(t *testing.T) {
	none := func(string) (string, bool) { return "", false }
	have := func(string) (string, bool) { return "/bin/tool", true }
	onlyCopy := func(n string) (string, bool) { return n, n == "wl-copy" }

	tests := []struct {
		name, wayland, display string
		look                   func(string) (string, bool)
		wantBackend            string
		wantAvail              bool
		wantReason             string
	}{
		{"wayland tools", "wayland-0", ":0", have, "wayland", true, ""},
		{"wayland fallback x11", "wayland-0", ":0", none, "x11", true, "WAYLAND_DISPLAY set without wl-copy/wl-paste on PATH; using DISPLAY"},
		{"wayland missing tools", "wayland-0", "", none, "none", false, ErrWaylandToolsMissing.Error()},
		{"wayland partial tools", "wayland-0", "", onlyCopy, "none", false, ErrWaylandToolsMissing.Error()},
		{"x11 only", "", ":0", none, "x11", true, ""},
		{"headless", "", "", none, "none", false, "neither WAYLAND_DISPLAY nor DISPLAY is set"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := diagnoseLinux(tt.wayland, tt.display, tt.look)
			if got.Backend != tt.wantBackend || got.Available != tt.wantAvail || got.Reason != tt.wantReason {
				t.Fatalf("got %+v want backend=%q avail=%v reason=%q", got, tt.wantBackend, tt.wantAvail, tt.wantReason)
			}
		})
	}
}

func TestDiagnoseMemoryMode(t *testing.T) {
	UseMemory(true)

	got := Diagnose()
	if got.Backend != "memory" || !got.Available || got.Reason != "" {
		t.Fatalf("Diagnose() = %+v", got)
	}
}
