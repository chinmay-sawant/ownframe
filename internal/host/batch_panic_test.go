package host

import (
	"context"
	"testing"
)

// TestBatchKeepsThePanic checks that a panic inside fn keeps travelling and
// that the batch draws nothing on the way out.
func TestBatchKeepsThePanic(t *testing.T) {
	t.Parallel()

	fake := &batchFake{}
	defer func() {
		rec := recover()
		if rec != "boom" {
			t.Fatalf("panic = %v, want boom", rec)
		}
		if fake.draws != 0 {
			t.Fatalf("draws = %d, want no draw", fake.draws)
		}
	}()

	_ = Batch(context.Background(), fake, func() error {
		panic("boom")
	})
}

// TestBatchEndPanicKeepsThePanic checks that a panic inside the end does not
// replace the panic that is travelling.
func TestBatchEndPanicKeepsThePanic(t *testing.T) {
	t.Parallel()

	fake := &batchFake{endPanic: true}
	defer func() {
		rec := recover()
		if rec != "boom" {
			t.Fatalf("panic = %v, want boom", rec)
		}
	}()

	_ = Batch(context.Background(), fake, func() error {
		panic("boom")
	})
}
