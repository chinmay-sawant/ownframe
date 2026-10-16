package page_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// TestLoadRouteRejectOversizedSource checks that Load and Route enforce
// MaxSourceSize before parsing trusted template source.
func TestLoadRouteRejectOversizedSource(t *testing.T) {
	t.Parallel()

	screen := navPage(t, `<div id="a">A</div>`)
	tooBig := strings.Repeat("x", page.MaxSourceSize+1)

	if err := screen.Load(context.Background(), tooBig); !errors.Is(err, page.ErrSourceTooLarge) {
		t.Fatalf("Load err = %v, want %v", err, page.ErrSourceTooLarge)
	}
	if err := screen.Route("huge", tooBig); !errors.Is(err, page.ErrSourceTooLarge) {
		t.Fatalf("Route err = %v, want %v", err, page.ErrSourceTooLarge)
	}
}
