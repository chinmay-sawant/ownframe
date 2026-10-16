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
	if got, ok := Pick(nil, "Open"); ok || got != "" {
		t.Fatalf("Pick = %q %v", got, ok)
	}
	if got, ok := pickLinux(nil, "Open"); ok || got != "" {
		t.Fatalf("pickLinux = %q %v", got, ok)
	}
	if got, ok := run(nil, "/bin/true", nil); ok || got != "" {
		t.Fatalf("run = %q %v", got, ok)
	}
}
