package web

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type ctxFake struct{ host.Screen }

func testListener(t *testing.T) net.Listener {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	return ln
}

func waitUp(t *testing.T, addr string, done <-chan error) {
	t.Helper()
	for i := 0; i < 50; i++ {
		if done != nil {
			select {
			case err := <-done:
				t.Fatalf("serve exited unexpectedly: %v", err)
			default:
			}
		}
		c, err := (&net.Dialer{Timeout: 50 * time.Millisecond}).DialContext(t.Context(), "tcp", addr)
		if err == nil {
			_ = c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("listener never came up")
}

func TestCancelStopsListener(t *testing.T) {
	t.Parallel()
	ln := testListener(t)
	addr := ln.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- ServeWithOptionsContext(ctx, &ctxFake{}, addr, Options{Listener: ln}) }()
	waitUp(t, addr, done)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serve did not stop after cancel")
	}
	if _, err := (&net.Dialer{Timeout: 200 * time.Millisecond}).DialContext(t.Context(), "tcp", addr); err == nil {
		t.Fatal("listener still open after cancel")
	}
}
