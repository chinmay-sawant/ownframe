//go:build linux && !android

package clipboard

import (
	"context"
	"io"
	"os/exec"
	"strings"
	"time"
)

const wlWait = time.Second

func runWL(ctx context.Context, name string, args []string, stdin string) bool {
	path, ok := lookWL(name)
	if !ok {
		return false
	}

	ctx, cancel := context.WithTimeout(ctx, wlWait)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...) //nolint:gosec // reason: LookPath of a fixed tool name, constant flags, clipboard text only on stdin, no shell.
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}

func outWL(ctx context.Context, name string, args []string) (string, bool) {
	path, ok := lookWL(name)
	if !ok {
		return "", false
	}

	ctx, cancel := context.WithTimeout(ctx, wlWait)
	defer cancel()

	cmd := exec.CommandContext(ctx, path, args...) //nolint:gosec // reason: LookPath of a fixed tool name, constant flags, no shell.
	cmd.Stderr = io.Discard
	out, err := cmd.Output()
	if err != nil {
		return "", false
	}

	return string(out), true
}
