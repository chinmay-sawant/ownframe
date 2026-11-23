package crash

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateDirRejectsDriveRoots(t *testing.T) {
	t.Parallel()
	for _, dir := range []string{`C:`, `C:\`, `C:/`, `z:\`} {
		if _, err := validateDir(dir); err == nil {
			t.Errorf("validateDir(%q) accepted a drive root", dir)
		}
	}
}

func TestOneLineStripsHeaderForgery(t *testing.T) {
	t.Parallel()
	got := oneLine("a\nreason: forged\r\nb\u2028c")
	if strings.ContainsAny(got, "\n\r\u2028") {
		t.Fatalf("oneLine kept a line break: %q", got)
	}
}

func TestWriteSanitizesTitleAndReason(t *testing.T) {
	dir := t.TempDir()
	if err := SetDir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { mu.Lock(); custom = ""; mu.Unlock() })
	path, err := Write("t\nreason: forged", "r\ngo: forged")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n")[:3] {
		if strings.HasPrefix(line, "reason: forged") || strings.HasPrefix(line, "go: forged") {
			t.Fatalf("forged header line in report: %q", line)
		}
	}
}

func TestTempOwnerIsNeverNegative(t *testing.T) {
	t.Parallel()
	if got := tempOwner(); got == "" || strings.HasPrefix(got, "-") {
		t.Fatalf("tempOwner = %q", got)
	}
	if got := alnum(`DOMAIN\Ada Lovelace-1`); got != "DOMAINAdaLovelace1" {
		t.Fatalf("alnum = %q", got)
	}
}

func TestSecureSharedDirRejectsSymlinkAndTightens(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix modes and symlinks")
	}
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.Mkdir(real, 0o750); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	if err := secureSharedDir(link); err == nil {
		t.Fatal("symlinked crash folder accepted")
	}
	if err := secureSharedDir(real); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(real)
	if err != nil || info.Mode().Perm() != crashDirPerm {
		t.Fatalf("mode = %v err = %v, want 0700", info.Mode().Perm(), err)
	}
	// Calling secureSharedDir on an already-0700 directory must still succeed
	// for the owner and run Chmod unconditionally.
	if err := secureSharedDir(real); err != nil {
		t.Fatalf("repeat secureSharedDir on 0700 dir: %v", err)
	}
}

func TestWriteRejectsSymlinkedCrashDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("unix symlinks")
	}
	base := t.TempDir()
	real := filepath.Join(base, "real")
	if err := os.Mkdir(real, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(base, "crashes")
	if err := os.Symlink(real, link); err != nil {
		t.Fatal(err)
	}
	mu.Lock()
	custom = link
	mu.Unlock()
	t.Cleanup(func() { mu.Lock(); custom = ""; mu.Unlock() })
	if _, err := Write("panic", "boom"); err == nil {
		t.Fatal("Write accepted a symlinked crash directory")
	}
}
