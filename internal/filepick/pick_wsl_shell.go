//go:build linux && !android

package filepick

import (
	"os/exec"
	"path/filepath"
)

// shellPaths lists Windows PowerShell programs to try, best first.
// PATH is not enough: appendWindowsPath=false in /etc/wsl.conf keeps the
// Windows directories out of the environment, so the standard install
// locations under /mnt are probed too.
func shellPaths() []string {
	var paths []string

	for _, name := range []string{"powershell.exe", "pwsh.exe"} {
		if path, err := exec.LookPath(name); err == nil {
			paths = append(paths, path)
		}
	}

	for _, pattern := range []string{
		"/mnt/*/Windows/System32/WindowsPowerShell/v1.0/powershell.exe",
		"/mnt/*/Program Files/PowerShell/*/pwsh.exe",
	} {
		found, _ := filepath.Glob(pattern)
		paths = append(paths, found...)
	}

	return paths
}
