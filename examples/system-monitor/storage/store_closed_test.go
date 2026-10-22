package storage

import (
	"context"
	"errors"
	"testing"
)

// TestCanceledAndClosed checks the two ways a call can be refused.
func TestCanceledAndClosed(t *testing.T) {
	st := openTest(t)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := st.SetSetting(ctx, "k", "v"); err == nil {
		t.Fatal("canceled write succeeded")
	}
	if _, err := st.History(ctx, Query{SessionID: 1}); err == nil {
		t.Fatal("canceled read succeeded")
	}

	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	if err := st.SetSetting(t.Context(), "k", "v"); !errors.Is(err, ErrClosed) {
		t.Fatalf("err = %v, want ErrClosed", err)
	}
}
