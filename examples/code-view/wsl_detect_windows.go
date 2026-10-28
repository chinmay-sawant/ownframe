//go:build windows

package main

import "os/exec"

// wslAvailable reports whether the WSL launcher is present. The handoff
// shells out to wsl.exe, so resolving it on PATH is enough to attempt it.
func wslAvailable() bool {
	_, err := exec.LookPath("wsl.exe")

	return err == nil
}
