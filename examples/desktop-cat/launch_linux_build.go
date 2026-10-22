//go:build linux && !android

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// buildWindowsExe compiles this example for Windows into dir.
func buildWindowsExe(ctx context.Context, dir string) (string, error) {
	_, source, _, _ := runtime.Caller(0)
	exe := filepath.Join(dir, "desktop-cat.exe")
	build := exec.CommandContext(ctx, "go", "build", "-p", "1", "-o", exe, "./examples/desktop-cat") //nolint:gosec // reason: fixed go build with static flags, no shell.
	build.Dir = filepath.Dir(filepath.Dir(filepath.Dir(source)))
	build.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	log.Print("desktop cat: WSL detected; building and opening a native Windows overlay")
	if err := build.Run(); err != nil {
		return "", fmt.Errorf("desktop cat: Windows build: %w", err)
	}
	return exe, nil
}
