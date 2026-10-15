package page

import (
	"context"
	"errors"
	"net/url"
	"testing"
)

// TestClassifyLinkSharedCorpus drives one allow-plus-deny table through both
// link entries: parseOutboundLink and DispatchDeepLink share classifyLink,
// so allow, javascript deny, empty-host http, and relative fallthrough must
// agree on both.
func TestClassifyLinkSharedCorpus(t *testing.T) {
	deep := func(target string) (*url.URL, error) {
		p, err := New(Config{HTML: `<p>x</p>`, Width: 10, Height: 10,
			DeepLinkSchemes: []string{"https", "mailto", "myapp"}})
		if err != nil {
			t.Fatal(err)
		}
		var got *url.URL
		p.Handle(Handlers{DeepLink: func(_ context.Context, link *url.URL) error {
			got = link
			return nil
		}})
		err = p.DispatchDeepLink(context.Background(), target)
		return got, err
	}
	for _, tc := range []struct {
		target   string
		outLink  bool
		outErr   error
		deepLink bool
		deepErr  error
	}{
		{"https://example.com/x", true, nil, true, nil},
		{"mailto:a@b.c", true, nil, true, nil},
		{"myapp://open/chat", false, ErrUnsupportedLinkScheme, true, nil},
		{"javascript:alert(1)", false, ErrUnsupportedLinkScheme, false, ErrUnsupportedLinkScheme},
		{"http://", false, ErrInvalidLink, false, ErrInvalidLink},
		{"settings", false, nil, false, ErrInvalidLink},
	} {
		link, err := parseOutboundLink(tc.target)
		if tc.outErr != nil {
			if !errors.Is(err, tc.outErr) || link != nil {
				t.Errorf("outbound(%q) = %v, %v, want %v", tc.target, link, err, tc.outErr)
			}
		} else if err != nil || (link == nil) != !tc.outLink {
			t.Errorf("outbound(%q) = %v, %v, want link %v", tc.target, link, err, tc.outLink)
		}
		got, err := deep(tc.target)
		if tc.deepErr != nil {
			if !errors.Is(err, tc.deepErr) || got != nil {
				t.Errorf("deep(%q) = %v, %v, want %v", tc.target, got, err, tc.deepErr)
			}
		} else if err != nil || (got == nil) != !tc.deepLink {
			t.Errorf("deep(%q) = %v, %v, want link %v", tc.target, got, err, tc.deepLink)
		}
	}
}
