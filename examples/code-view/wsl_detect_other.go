//go:build !windows

package main

// wslAvailable is always false outside Windows; the WSL handoff only ever
// runs on Windows. Kept as a function (rather than inlining false at the
// call site) so both platform files expose the identical contract.
//
//nolint:unused // used by wsl_detect_windows.go; this stub keeps the non-Windows build compiling.
func wslAvailable() bool {
	return false
}
