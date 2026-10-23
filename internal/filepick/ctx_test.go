package filepick

import (
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
