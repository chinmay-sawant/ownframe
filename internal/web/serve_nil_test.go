package web

import (
	"context"
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestServeRejectsNilInputs(t *testing.T) {
	t.Parallel()
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	if err := ServeWithOptionsContext(nil, &ctxFake{}, "127.0.0.1:0", Options{}); !errors.Is(err, ErrNilContext) {
		t.Fatalf("nil ctx err = %v", err)
	}
	//nolint:staticcheck // reason: intentionally passes nil to verify fetch.ErrNilContext alias.
	if err := ServeWithOptionsContext(nil, &ctxFake{}, "127.0.0.1:0", Options{}); !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("nil ctx canon err = %v", err)
	}
	if err := ServeWithOptionsContext(context.Background(), nil, "127.0.0.1:0", Options{}); !errors.Is(err, ErrNilApp) {
		t.Fatalf("nil app err = %v", err)
	}
}
