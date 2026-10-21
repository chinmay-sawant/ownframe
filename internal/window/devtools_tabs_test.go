package window

import "testing"

// TestDevTabOpensOnFrame checks the dock starts on the Frame tab, since the
// tab order puts Frame first.
func TestDevTabOpensOnFrame(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if s.dev.tab != devTabFrame {
		t.Fatalf("tab = %v, want Frame", s.dev.tab)
	}

	if devTabNames[0] != "Frame" {
		t.Fatalf("first tab = %q, want Frame", devTabNames[0])
	}
}

// TestDevTabAtClamps checks an out-of-range index lands on a real tab.
func TestDevTabAtClamps(t *testing.T) {
	t.Parallel()

	if got := devTabAt(-1); got != devTabFrame {
		t.Fatalf("devTabAt(-1) = %v, want Frame", got)
	}

	if got := devTabAt(len(devTabNames)); got != devTabOps {
		t.Fatalf("devTabAt(%d) = %v, want Ops", len(devTabNames), got)
	}
}
