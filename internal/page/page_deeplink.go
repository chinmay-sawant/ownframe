package page

import (
	"context"
	"errors"
	"fmt"
	"net/url"
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
	link, err := classifyLink(target)
	if err != nil {
		if errors.Is(err, errRelativeLink) {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
		return nil, err
	}
	if _, ok := schemes[link.Scheme]; !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	return link, nil
}
