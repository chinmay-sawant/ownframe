package host

import (
	"context"
	"testing"
)

// batchFake is a screen that batches its draws. The end counts the draws and
// can fail or panic, so the batch contract is observable.
type batchFake struct {
	begins   int
	ends     int
	draws    int
	endErr   error
	endPanic bool
}

func (b *batchFake) BeginRedrawBatch() { b.begins++ }

func (b *batchFake) EndRedrawBatch(ctx context.Context) error {
	b.ends++
	if b.endPanic {
		panic("the end panicked")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	b.draws++

	return b.endErr
}

// TestBatchDrawsOnceAtTheEnd checks that the batch draws once, after fn.
func TestBatchDrawsOnceAtTheEnd(t *testing.T) {
	t.Parallel()

	fake := &batchFake{}
	err := Batch(context.Background(), fake, func() error {
		if fake.draws != 0 {
			t.Fatal("the batch drew before the end")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if fake.begins != 1 || fake.ends != 1 || fake.draws != 1 {
		t.Fatalf("begins %d ends %d draws %d, want one of each",
			fake.begins, fake.ends, fake.draws)
	}
}

// TestBatchWithoutTheRoleRunsFn checks that a screen without the batch role
// runs fn on its own.
func TestBatchWithoutTheRoleRunsFn(t *testing.T) {
	t.Parallel()

	ran := false
	err := Batch(context.Background(), pointerFake{}, func() error {
		ran = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !ran {
		t.Fatal("fn did not run")
	}
}
