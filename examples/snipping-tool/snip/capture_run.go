package snip

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// tempPNG reserves a temp path for a screenshot file.
func tempPNG() (string, func(), error) {
	f, err := os.CreateTemp("", "snip-*.png")
	if err != nil {
		return "", nil, err
	}
	path := f.Name()
	_ = f.Close()
	return path, func() { _ = os.Remove(path) }, nil
}

// runOut runs name and returns stdout; stderr joins the error.
func runOut(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // reason: screenshot tool is a fixed allowlist, no shell.
	out, err := cmd.Output()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return nil, fmt.Errorf("%s: %w: %s", name, err, ee.Stderr)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}

// shotToFile runs name with args plus a temp PNG path, then reads it back.
func shotToFile(ctx context.Context, name string, args ...string) ([]byte, error) {
	path, cleanup, err := tempPNG()
	if err != nil {
		return nil, err
	}
	defer cleanup()
	cmd := exec.CommandContext(ctx, name, append(args, path)...) //nolint:gosec // reason: temp path appended to fixed tool args, local file.
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%s: %w: %s", name, err, out)
	}
	raw, err := os.ReadFile(path) //nolint:gosec // reason: temp PNG from CreateTemp, local file.
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return raw, nil
}
