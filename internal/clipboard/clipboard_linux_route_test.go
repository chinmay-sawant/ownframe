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
