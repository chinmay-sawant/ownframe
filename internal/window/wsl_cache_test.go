package window

import (
	"os"
	"runtime"
	"sync"
	"testing"
)

// TestIsWSLReadsProcVersionOnce pins PERF-WSL-01: repeated isWSL calls, as the
// per-frame close stage makes, hit the filesystem at most once.
func TestIsWSLReadsProcVersionOnce(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("WSL detection reads /proc/version on Linux only")
	}
	// isWSL uses LookupEnv, so the variables must be unset, not empty.
	// t.Setenv registers the restore; Unsetenv then removes them for the test.
	for _, name := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP"} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}

	oldOnce, oldRead := isWSLOnce, readProcVersion
	t.Cleanup(func() { isWSLOnce, readProcVersion = oldOnce, oldRead })

	reads := 0
	readProcVersion = func(string) ([]byte, error) {
		reads++
		return []byte("Linux version 6.1 (plain)"), nil
	}
	isWSLOnce = sync.OnceValue(detectWSL)

	for range 1000 {
		if isWSL() {
			t.Fatal("plain kernel reported as WSL")
		}
	}
	if reads != 1 {
		t.Fatalf("/proc/version reads = %d, want 1", reads)
	}
}
