package window

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
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
	if takeCloseRequest() {
		t.Fatal("stale close request leaked into the test")
	}

	RequestClose()

	if !takeCloseRequest() {
		t.Fatal("RequestClose did not arm a close")
	}

	if takeCloseRequest() {
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

	if takeCloseRequest() {
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
	manualMax.active = false
	manualMax.known = false

	completeManualMaximize(1920, 1032)

	if manualMax.learning || manualMax.active || manualMax.known {
		t.Fatal("idle complete changed maximize state")
	}
}
