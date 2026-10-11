package page

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

// DispatchDeepLink validates a host-delivered URL and calls the app handler.
// Call it on the page's owning event loop. It does not draw.
func (p *Page) DispatchDeepLink(ctx context.Context, target string) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if p.handlers.DeepLink == nil {
		return ErrNoDeepLinkHandler
	}
	link, err := parseDeepLink(target, p.deepLinkSchemes)
	if err != nil {
		return err
	}
	return p.handlers.DeepLink(ctx, link)
}

func parseDeepLink(target string, schemes map[string]struct{}) (*url.URL, error) {
	if target == "" || strings.TrimSpace(target) != target || strings.ContainsAny(target, "\r\n\t ") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	link, err := url.Parse(target)
	if err != nil || !link.IsAbs() || link.Scheme == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	link.Scheme = strings.ToLower(link.Scheme)
	if blockedLinkScheme(link.Scheme) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	if _, ok := schemes[link.Scheme]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	if (link.Scheme == "http" || link.Scheme == "https") && link.Hostname() == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	if link.Scheme == "mailto" && link.Opaque == "" && link.Path == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	if link.Opaque == "" && link.Host == "" && link.Path == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	return link, nil
}
