package draw

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

func TestStyleDocument(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var layout time.Duration
	styled, err := StyleDocument(ctx, false, &layout, nil, "", Viewport{}, render.State{})
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if styled != (Styled{}) {
		t.Fatal("styled")
	}
	if layout != 0 {
		t.Fatal("untracked")
	}

	layout = -time.Nanosecond
	styled, err = StyleDocument(ctx, true, &layout, nil, "", Viewport{}, render.State{})
	if !errors.Is(err, context.Canceled) || styled != (Styled{}) {
		t.Fatal(err)
	}
	if layout < 0 {
		t.Fatal("unnoted")
	}

	styled, err = StyleDocument(ctx, true, nil, nil, "", Viewport{}, render.State{})
	if !errors.Is(err, context.Canceled) || styled != (Styled{}) {
		t.Fatal(err)
	}
}
