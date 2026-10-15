package page_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

func TestOversizeWrapsBothSentinels(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	big := strings.Repeat("x", page.MaxSourceSize+1)
	if _, err := page.New(page.Config{HTML: big, Width: 320, Height: 200}); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("New err = %v", err)
	}
	s := navPage(t, `<div id="a">A</div>`)
	if err := s.Load(ctx, big); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Load err = %v", err)
	}
	if err := s.Route("huge", big); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Route err = %v", err)
	}
	if err := s.SetTheme(big); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("Theme err = %v", err)
	}
	path := writeSource(t, "index.html", `<p id="p">one</p>`)
	p := watchedPage(t, path)
	rewriteSource(t, path, big)
	if _, err := p.PollReload(ctx); !errors.Is(err, page.ErrSourceTooLarge) || !errors.Is(err, page.ErrBadSource) {
		t.Fatalf("poll err = %v", err)
	}
}
