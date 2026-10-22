//go:build linux && !android

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/cat"
	"github.com/chinmay-sawant/ownframe/internal/safego"
)

// WSLg does not carry this window's alpha and input regions to Windows.
// Run the same example as a Windows window instead.
func launchNative(ctx context.Context, inbox *cat.Inbox, addr string) (bool, error) {
	version, err := os.ReadFile("/proc/sys/kernel/osrelease")
	if err != nil || !strings.Contains(strings.ToLower(string(version)), "microsoft") {
		return false, nil
	}

	if _, err := os.Stat("/proc/sys/fs/binfmt_misc/WSLInterop"); err != nil {
		return true, fmt.Errorf("desktop cat: Windows interop is required for the WSL overlay: %w", err)
	}

	dir, err := os.MkdirTemp("", "ownframe-cat-")
	if err != nil {
		return true, err
	}
	defer func() { _ = os.RemoveAll(dir) }()

	exe, err := buildWindowsExe(ctx, dir)
	if err != nil {
		return true, err
	}

	args := append(append([]string{}, os.Args[1:]...), "-notify-addr=off", "-stdin-notifications", "-media-endpoint="+nativeEndpoint(addr))
	child := exec.CommandContext(ctx, exe, args...) //nolint:gosec // reason: runs the exe just built with fixed args, no shell.
	child.Stdout, child.Stderr = os.Stdout, os.Stderr
	pipe, err := child.StdinPipe()
	if err != nil {
		return true, err
	}
	defer func() { _ = pipe.Close() }()
	safego.Go("notify forward", func() { forwardNotifications(ctx, inbox, pipe) })
	err = child.Run()
	if ctx.Err() != nil {
		return true, ctx.Err()
	}

	return true, err
}
