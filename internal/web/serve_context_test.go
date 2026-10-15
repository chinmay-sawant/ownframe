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

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	a := ln.Addr().String()
	ln.Close()
	return a
}

func waitUp(t *testing.T, addr string) {
	t.Helper()
	for i := 0; i < 50; i++ {
		c, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("listener never came up")
}

func TestCancelStopsListener(t *testing.T) {
	t.Parallel()
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- ServeWithOptionsContext(ctx, &ctxFake{}, addr, Options{}) }()
	waitUp(t, addr)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("err = %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serve did not stop after cancel")
	}
	if _, err := net.DialTimeout("tcp", addr, 200*time.Millisecond); err == nil {
		t.Fatal("listener still open after cancel")
	}
}
