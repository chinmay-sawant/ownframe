//go:build linux && !android

package filepick

import (
	"errors"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestPickRejectsNilContext(t *testing.T) {
	t.Parallel()
	if !errors.Is(ErrNilContext, fetch.ErrNilContext) {
		t.Fatalf("alias = %v", ErrNilContext)
	}
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	if got, err := Pick(nil, "Open"); !errors.Is(err, ErrNilContext) || got != "" {
		t.Fatalf("Pick = %q %v", got, err)
	}
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	if got, err := pickLinux(nil, "Open"); !errors.Is(err, ErrNilContext) || got != "" {
		t.Fatalf("pickLinux = %q %v", got, err)
	}
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	if got, err := run(nil, "/bin/true", nil); !errors.Is(err, ErrNilContext) || got != "" {
		t.Fatalf("run = %q %v", got, err)
	}
}
