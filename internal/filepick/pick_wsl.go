//go:build linux && !android

package filepick

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

// inWSL reports that this Linux program runs under Windows Subsystem for Linux.
func inWSL() bool {
	if os.Getenv("WSL_DISTRO_NAME") != "" || os.Getenv("WSL_INTEROP") != "" {
		return true
	}
	release, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil {
		return false
	}
	name := strings.ToLower(string(release))

	return strings.Contains(name, "microsoft") || strings.Contains(name, "wsl")
}

// pickWindows asks the Windows open dialog through interop and converts
// the chosen path back to Linux. ErrNoDialog means no Windows shell ran,
// so the Linux tools still get a chance.
func pickWindows(ctx context.Context, title string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	shells := shellPaths()
	if len(shells) == 0 {
		debugf("no Windows PowerShell found")

		return "", ErrNoDialog
	}

	return pickWindowsAt(ctx, title, shells)
}

// pickWindowsAt tries each shell until one runs the dialog to a decision.
func pickWindowsAt(ctx context.Context, title string, shells []string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	script := windowScript(title)
	for _, shell := range shells {
		out, err := exec.CommandContext(ctx, shell, "-NoProfile", "-STA", "-Command", script).Output() //nolint:gosec // reason: fixed shell binary plus static flags, script built from title only, picked path flows out and is validated (safeWindowsPath/hasDotSegment).
		if err != nil {
			debugf("%s failed: %v", shell, err)

			continue
		}
		picked := strings.TrimRight(string(out), "\r\n")
		if picked == "" {
			debugf("%s: canceled", shell)

			return "", ErrCanceled
		}
		linux, err := fromWindowsPath(ctx, picked)
		if err != nil {
			debugf("cannot convert Windows path %q", picked)

			return "", err
		}

		return linux, nil
	}

	return "", ErrNoDialog
}
