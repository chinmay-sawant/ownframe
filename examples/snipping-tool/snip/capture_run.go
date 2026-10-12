package snip

import (
	"context"
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
	f.Close()
	return path, func() { os.Remove(path) }, nil
}

// runOut runs name and returns stdout; stderr joins the error.
func runOut(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if ee, ok := err.(*exec.ExitError); ok {
		return nil, fmt.Errorf("%s: %v: %s", name, err, ee.Stderr)
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
	cmd := exec.CommandContext(ctx, name, append(args, path)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("%s: %v: %s", name, err, out)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return raw, nil
}
