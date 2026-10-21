//go:build linux && !android

package filepick

import (
	"fmt"
	"os"
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

// debugf writes to stderr when OWNFRAME_FILEPICK_DEBUG is not empty.
// Values are sanitized to one line. Keep this flag off in shipped builds.
func debugf(format string, args ...any) {
	if os.Getenv("OWNFRAME_FILEPICK_DEBUG") == "" && os.Getenv("GPUI_FILEPICK_DEBUG") == "" {
		return
	}

	fmt.Fprintf(os.Stderr, "filepick: "+format+"\n", cleanArgs(args)...)
}
