//go:build windows

package main

import (
	"os/exec"
	"syscall"
)

// hideConsoleWindow keeps a child console process from opening its own
// terminal window. A windowsgui parent has no console to share, so without
// this every git or rg call pops a visible console next to the app.
func hideConsoleWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}

	cmd.SysProcAttr.HideWindow = true
}
