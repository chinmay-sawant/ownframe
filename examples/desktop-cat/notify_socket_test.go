//go:build !js && !android && !ios

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestListenNotifySocket(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_RUNTIME_DIR", dir)
	got := notifySocketPath()
	if filepath.Dir(got) != dir {
		t.Fatalf("runtime dir: %s", got)
	}
	path := filepath.Join(dir, "test.sock")
	ln, err := listenNotify(context.Background(), "unix:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	if ln.Addr().Network() != "unix" {
		t.Fatalf("network: %s", ln.Addr().Network())
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("mode: %v", info.Mode().Perm())
	}
	_ = ln.Close()
	ln, err = listenNotify(context.Background(), "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	if ln.Addr().Network() != "tcp" {
		t.Fatalf("tcp: %s", ln.Addr().Network())
	}
}
