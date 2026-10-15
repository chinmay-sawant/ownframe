//go:build linux && !android

package filepick

import (
	"context"
	"os/exec"
	"strings"
)

// tool is one dialog program and how to ask it for a file.
type tool struct {
	name string
	args func(title string) []string
}

var tools = []tool{
	{"zenity", zenityArgs},
	{"qarma", zenityArgs},
	{"matedialog", zenityArgs},
	{"kdialog", func(title string) []string {
		return []string{"--getopenfilename", "--title", title}
	}},
}

func zenityArgs(title string) []string {
	return []string{"--file-selection", "--title=" + title}
}

// Pick runs the Windows dialog under WSL, then the first dialog program on PATH.
// ok is false when the person canceled or no program is installed.
func Pick(ctx context.Context, title string) (string, bool) {
	if inWSL() {
		debugf("WSL detected")

		if path, ok, handled := pickWindows(ctx, title); handled {
			return path, ok
		}

		debugf("falling back to Linux dialogs")
	}

	return pickLinux(ctx, title)
}

// pickLinux runs the first dialog program on PATH.
func pickLinux(ctx context.Context, title string) (string, bool) {
	for _, t := range tools {
		path, err := exec.LookPath(t.name)
		if err != nil {
			continue
		}

		return run(ctx, path, t.args(title))
	}

	return "", false
}

func run(ctx context.Context, path string, args []string) (string, bool) {
	out, err := exec.CommandContext(ctx, path, args...).Output()
	if err != nil {
		debugf("%s %v failed: %v", path, args, err)

		return "", false
	}

	picked := strings.TrimRight(string(out), "\r\n")
	if picked == "" {
		return "", false
	}

	return picked, true
}
