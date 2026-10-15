//go:build linux && !android

package filepick

import (
	"context"
	"os/exec"
	"strings"
)

func fromWindowsPath(ctx context.Context, picked string) (string, bool) {
	if !safeWindowsPath(picked) {
		return "", false
	}

	if linux, ok := uncToWSL(picked); ok {
		return linux, true
	}

	out, err := exec.CommandContext(ctx, wslpathBin(), "-u", picked).Output()
	if err != nil {
		debugf("wslpath -u %q failed: %v", picked, err)

		return "", false
	}

	linux := strings.TrimRight(string(out), "\r\n")
	if linux == "" {
		return "", false
	}

	return linux, true
}

// wslpathBin finds wslpath, falling back to its standard location.
func wslpathBin() string {
	if path, err := exec.LookPath("wslpath"); err == nil {
		return path
	}

	return "/usr/bin/wslpath"
}

// uncToWSL turns \\wsl.localhost\Distro\path and \\wsl$\Distro\path
// into the Linux path under the root.
func uncToWSL(picked string) (string, bool) {
	if !safeWindowsPath(picked) {
		return "", false
	}

	for _, prefix := range []string{`\\wsl.localhost\`, `\\wsl$\`} {
		if !strings.HasPrefix(picked, prefix) {
			continue
		}

		rest := picked[len(prefix):]
		cut := indexOfSep(rest, '\\')
		if cut < 0 {
			return "/", true
		}

		out := "/" + strings.ReplaceAll(rest[cut+1:], `\`, "/")
		if hasDotSegment(out) {
			return "", false
		}

		return out, true
	}

	return "", false
}
