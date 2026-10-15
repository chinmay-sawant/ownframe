package ipc_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/ipc"
)

func TestSameChannelCollides(t *testing.T) {
	t.Parallel()
	ch := "gate-collide"
	a := ipc.Handle(ch, func(context.Context, string) (string, error) { return "A", nil })
	b := ipc.Handle(ch, func(context.Context, string) (string, error) { return "B", nil })
	t.Cleanup(b)
	t.Cleanup(a)
	got, err := ipc.Request(context.Background(), ch, "x")
	if err != nil || got != "B" {
		t.Fatalf("got %q err %v, want B", got, err)
	}
}

func TestListenCaps(t *testing.T) {
	t.Parallel()
	ch := "gate-cap"
	var cancels []func()
	t.Cleanup(func() {
		for _, c := range cancels {
			c()
		}
	})
	calls := 0
	for i := 0; i < 300; i++ {
		c := ipc.Listen(ch, func(string) { calls++ })
		cancels = append(cancels, c)
	}
	ipc.Send(ch, "x")
	if calls > 256 {
		t.Fatalf("calls = %d, want <=256", calls)
	}
	if calls != 256 {
		t.Fatalf("calls = %d, want 256", calls)
	}
}
