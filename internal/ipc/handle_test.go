package ipc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/ipc"
)

func TestHandleCancelKeepsReplacement(t *testing.T) {
	t.Parallel()

	first := ipc.Handle("request-swap", func(context.Context, string) (string, error) {
		return "old", nil
	})
	ipc.Handle("request-swap", func(context.Context, string) (string, error) {
		return "new", nil
	})
	first()

	got, err := ipc.Request(context.Background(), "request-swap", "")
	if err != nil || got != "new" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestEmptyChannel(t *testing.T) {
	t.Parallel()

	heard := false
	stop := ipc.Listen("", func(string) { heard = true })
	ipc.Send("", "x")
	stop()
	ipc.Handle("", func(context.Context, string) (string, error) {
		heard = true
		return "no", nil
	})

	_, err := ipc.Request(nil, "", "x")
	if !errors.Is(err, ipc.ErrNoHandler) || heard {
		t.Fatalf("err=%v heard=%v", err, heard)
	}
}
