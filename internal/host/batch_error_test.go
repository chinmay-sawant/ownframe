package host

import (
	"context"
	"errors"
	"testing"
)

// TestBatchReturnsTheEndError checks that the end error reaches the caller.
func TestBatchReturnsTheEndError(t *testing.T) {
	t.Parallel()

	want := errors.New("end failed")
	fake := &batchFake{endErr: want}
	err := Batch(context.Background(), fake, func() error { return nil })
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
}

// TestBatchKeepsTheFnError checks that the fn error wins over the end error.
func TestBatchKeepsTheFnError(t *testing.T) {
	t.Parallel()

	want := errors.New("fn failed")
	fake := &batchFake{endErr: errors.New("end failed")}
	err := Batch(context.Background(), fake, func() error { return want })
	if !errors.Is(err, want) {
		t.Fatalf("err = %v, want %v", err, want)
	}
	if fake.draws != 1 {
		t.Fatalf("draws = %d, want one draw", fake.draws)
	}
}
