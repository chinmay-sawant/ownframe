package ipc

import (
	"bytes"
	"context"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
)

func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	return &buf
}

// TestRemoveEarZeroesTailSlot pins HYG-IPC-01: removing a listener drops the
// closure from the backing array.
func TestRemoveEarZeroesTailSlot(t *testing.T) {
	const ch = "hardening-remove"
	cancelA := Listen(ch, func(string) {})
	cancelB := Listen(ch, func(string) {})
	mu.Lock()
	backing := ears[ch][:2]
	mu.Unlock()
	cancelA()
	if backing[1].fn != nil {
		t.Fatal("removed listener still referenced from the backing array")
	}
	cancelB()
}

func TestSendAndListenSurfacePanicAndCap(t *testing.T) {
	buf := captureLog(t)
	const ch = "hardening-panic"
	cancel := Listen(ch, func(string) { panic("boom") })
	defer cancel()
	ran := false
	cancel2 := Listen(ch, func(string) { ran = true })
	defer cancel2()
	Send(ch, "secret-payload")
	if !ran {
		t.Fatal("a panicking listener stopped the others")
	}
	if got := buf.String(); !strings.Contains(got, "panicked") || strings.Contains(got, "secret-payload") {
		t.Fatalf("panic log = %q, want a panic note without the payload", got)
	}

	const full = "hardening-cap"
	var cancels []func()
	for range maxListeners {
		cancels = append(cancels, Listen(full, func(string) {}))
	}
	buf.Reset()
	Listen(full, func(string) {})()
	if !strings.Contains(buf.String(), "already has") {
		t.Fatalf("cap log = %q, want a listener-cap note", buf.String())
	}
	for _, c := range cancels {
		c()
	}
}

func TestRequestWrapsPanicValue(t *testing.T) {
	const ch = "hardening-request-panic"
	cancel := Handle(ch, func(context.Context, string) (string, error) { panic("kaboom") })
	defer cancel()
	_, err := Request(context.Background(), ch, "x")
	if !errors.Is(err, ErrHandlerPanicked) || !strings.Contains(err.Error(), "kaboom") {
		t.Fatalf("err = %v, want ErrHandlerPanicked carrying the panic value", err)
	}
}
