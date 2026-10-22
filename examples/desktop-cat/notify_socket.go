//go:build !js && !android && !ios

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// notifySocketPerm keeps the local API socket private to the owner.
const notifySocketPerm = 0o600

// notifySocketPath puts the API socket in the runtime dir, or the temp
// dir with the uid in the name when XDG_RUNTIME_DIR is unset.
func notifySocketPath() string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, fmt.Sprintf("ownframe-desktop-cat-%d.sock", os.Getuid()))
}

// listenNotify serves "auto" or a unix: path on a 0600 Unix socket and
// anything else on TCP. Plain TCP stays explicit opt-in (GEN-40).
func listenNotify(ctx context.Context, addr string) (net.Listener, error) {
	if addr == "auto" || addr == "unix" || strings.HasPrefix(addr, "unix:") {
		path := strings.TrimPrefix(strings.TrimPrefix(addr, "unix:"), "unix://")
		if path == "" || path == "auto" || addr == "unix" {
			path = notifySocketPath()
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("notification socket: %w", err)
		}
		listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", path)
		if err != nil {
			return nil, fmt.Errorf("notification socket: %w", err)
		}
		if err := os.Chmod(path, notifySocketPerm); err != nil {
			_ = listener.Close()
			return nil, fmt.Errorf("notification socket: %w", err)
		}
		return listener, nil
	}
	if err := checkNotifyAddr(addr); err != nil {
		return nil, err
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("notification server: %w", err)
	}
	return listener, nil
}
