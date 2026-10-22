//go:build linux && !android

package filepick

import (
	"context"
	"errors"
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
func Pick(ctx context.Context, title string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	if inWSL() {
		debugf("WSL detected")
		path, err := pickWindows(ctx, title)
		if !errors.Is(err, ErrNoDialog) {
			return path, err
		}
		debugf("falling back to Linux dialogs")
	}

	return pickLinux(ctx, title)
}

// pickLinux runs the first dialog program on PATH.
func pickLinux(ctx context.Context, title string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	for _, t := range tools {
		path, err := exec.LookPath(t.name)
		if err != nil {
			continue
		}

		return run(ctx, path, t.args(title))
	}

	return "", ErrNoDialog
}

func run(ctx context.Context, path string, args []string) (string, error) {
	if err := requireCtx(ctx); err != nil {
		return "", err
	}
	out, err := exec.CommandContext(ctx, path, args...).Output() //nolint:gosec // reason: binary is a known dialog tool resolved from PATH, fixed arg builders, no shell.
	if err != nil {
		debugf("%s %v failed: %v", path, args, err)

		return "", ErrCanceled
	}
	picked := strings.TrimRight(string(out), "\r\n")
	if picked == "" {
		return "", ErrCanceled
	}

	return picked, nil
}
