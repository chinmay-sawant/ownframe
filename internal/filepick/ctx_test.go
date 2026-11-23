package filepick

import (
	"context"
	"errors"
	"testing"
)

func TestNilContext(t *testing.T) {
	t.Parallel()
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	if got, err := Pick(nil, "Open"); !errors.Is(err, ErrNilContext) || got != "" {
		t.Fatalf("Pick = %q %v", got, err)
	}
}

func TestCanceledContextIsErrCanceledAndContextError(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Pick(ctx, "Open"); !errors.Is(err, ErrCanceled) || !errors.Is(err, context.Canceled) {
		t.Fatalf("Pick on canceled ctx = %v, want ErrCanceled wrapping context.Canceled", err)
	}
}
