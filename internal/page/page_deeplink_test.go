package page

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

func TestDispatchDeepLink(t *testing.T) {
	p := deepLinkPage(t, []string{"myapp"})
	var got *url.URL
	p.Handle(Handlers{DeepLink: func(_ context.Context, link *url.URL) error {
		got = link
		return nil
	}})
	if err := p.DispatchDeepLink(context.Background(), "MyApp://open/chat?id=3"); err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Scheme != "myapp" || got.Host != "open" || got.Path != "/chat" {
		t.Fatalf("link = %#v", got)
	}
}

func TestDispatchDeepLinkRejectsBadTargets(t *testing.T) {
	p := deepLinkPage(t, []string{"myapp"})
	p.Handle(Handlers{DeepLink: func(context.Context, *url.URL) error { return nil }})
	for _, test := range []struct {
		target string
		want   error
	}{
		{"other://open", ErrUnsupportedLinkScheme},
		{"javascript:alert(1)", ErrUnsupportedLinkScheme},
		{"settings", ErrInvalidLink},
		{"myapp://", ErrInvalidLink},
	} {
		if err := p.DispatchDeepLink(context.Background(), test.target); !errors.Is(err, test.want) {
			t.Errorf("DispatchDeepLink(%q) = %v, want %v", test.target, err, test.want)
		}
	}
}

func TestDeepLinkConfigAndMissingHandler(t *testing.T) {
	if _, err := New(Config{HTML: `<p>x</p>`, Width: 100, Height: 100, DeepLinkSchemes: []string{"javascript"}}); !errors.Is(err, ErrBadLinkScheme) {
		t.Fatalf("bad scheme error = %v", err)
	}
	p := deepLinkPage(t, []string{"myapp"})
	if err := p.DispatchDeepLink(context.Background(), "myapp://open"); !errors.Is(err, ErrNoDeepLinkHandler) {
		t.Fatalf("missing handler error = %v", err)
	}
}

func deepLinkPage(t *testing.T, schemes []string) *Page {
	t.Helper()
	p, err := New(Config{HTML: `<p>Page</p>`, Width: 100, Height: 100, DeepLinkSchemes: schemes})
	if err != nil {
		t.Fatal(err)
	}
	return p
}
