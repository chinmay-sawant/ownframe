package storage

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestRecorderCloseDeadline checks that Close respects its context and can be
// retried once the store frees up.
func TestRecorderCloseDeadline(t *testing.T) {
	st := openTest(t)
	ctx := t.Context()
	id := newTestSession(t, st)

	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := NewRecorder(st, id, 1)
	rec.RecordSample(fixtureSamples()[0])

	short, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	if err := rec.Close(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("close = %v, want deadline", err)
	}

	_ = tx.Rollback()

	long, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rec.Close(long); err != nil {
		t.Fatal(err)
	}
}
