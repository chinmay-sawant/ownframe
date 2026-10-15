package ownframe

import "github.com/chinmay-sawant/ownframe/internal/crash"

// Report writes a local crash file and returns its path.
func Report(title, reason string) (string, error) {
	return crash.Write(title, reason)
}

// SetCrashDir sets the directory Report and a recovered panic use.
// It rejects an empty or unsafe path. An empty value used to
// restore the default; callers should keep a previous value instead.
func SetCrashDir(dir string) error {
	return crash.SetDir(dir)
}
