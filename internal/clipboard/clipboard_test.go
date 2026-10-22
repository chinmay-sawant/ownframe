package clipboard

import (
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/leakcheck"
)

func TestMain(m *testing.M) {
	UseMemory(true)
	os.Exit(leakcheck.Verify(m))
}

func TestMemoryRoundTrip(t *testing.T) {
	UseMemory(true)

	Write("alpha")

	if got, err := Read(); err != nil || got != "alpha" {
		t.Fatalf("Read() = %q, %v", got, err)
	}

	Write("beta\nline")

	if got, err := Read(); err != nil || got != "beta\nline" {
		t.Fatalf("Read() = %q, %v", got, err)
	}
}

func TestReadOSFailureHasNoText(t *testing.T) {
	UseMemory(true)
	Write("stale")

	UseMemory(false)
	t.Setenv("WAYLAND_DISPLAY", "test-unavailable")
	defer UseMemory(true)

	if got, err := Read(); err == nil || got != "" {
		t.Fatalf("Read() = %q, %v", got, err)
	}
}

func TestWriteOutsideMemoryLeavesMemory(t *testing.T) {
	UseMemory(true)
	Write("kept")

	UseMemory(false)
	t.Setenv("WAYLAND_DISPLAY", "test-unavailable")
	defer UseMemory(true)
	Write("dropped")

	UseMemory(true)

	if got, err := Read(); err != nil || got != "kept" {
		t.Fatalf("Read() = %q, %v", got, err)
	}
}

func TestUTF16Clipboard(t *testing.T) {
	got := utf16Clipboard("A€")
	want := []byte{
		'A', 0,
		0xac, 0x20,
		0, 0,
	}

	if len(got) != len(want) {
		t.Fatalf("len %d", len(got))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("byte %d = %d", i, got[i])
		}
	}
}
