//go:build !windows

package main

import "flag"

// addWSLFlags registers nothing outside Windows; the WSL handoff only runs
// on Windows, so other platforms take no -wsl or -wsl-bin flags.
func addWSLFlags(_ *flag.FlagSet) {}

// maybeReexecInWSL never hands off outside Windows.
func maybeReexecInWSL() bool {
	return false
}
