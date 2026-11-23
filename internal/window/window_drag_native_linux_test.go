//go:build linux && !android && !ios && !js && cgo

package window

import "testing"

func TestSecNativeWindowRectRejectsNonPositive(t *testing.T) {
	t.Parallel()

	for _, c := range [][2]int{{0, 10}, {10, 0}, {-1, 10}, {10, -5}, {0, 0}} {
		if setNativeWindowRect(0, 0, c[0], c[1]) {
			t.Fatalf("setNativeWindowRect(%d, %d) = true", c[0], c[1])
		}
	}
}

// TestSecNativeClientSurvivesBadWindow walks from a window id that does not
// exist. Without the scoped error handler Xlib exits the whole test binary.
func TestSecNativeClientSurvivesBadWindow(t *testing.T) {
	got, ok := nativeClientFrom(0x7ffffff0)
	if !ok {
		t.Skip("no X display")
	}
	if got != 0 {
		t.Fatalf("bad window resolved to %#x", got)
	}
}
