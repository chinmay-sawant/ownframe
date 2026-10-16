package page

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// errRelativeLink marks a target without an absolute scheme. Outbound links
// treat it as in-page fallthrough; deep links reject it as invalid.
var errRelativeLink = errors.New("ownframe: relative link")

// classifyLink parses target once and enforces the shared trust policy:
// surrounding or inner whitespace is invalid, blocked schemes are denied,
// empty-host http is invalid, and empty mailto is invalid. Each caller keeps
// only its allowlist line.
func classifyLink(target string) (*url.URL, error) {
	if strings.TrimSpace(target) != target || strings.ContainsAny(target, "\r\n\t ") {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	link, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	if !link.IsAbs() || link.Scheme == "" {
		return nil, fmt.Errorf("%w: %q", errRelativeLink, target)
	}
	link.Scheme = strings.ToLower(link.Scheme)
	if blockedLinkScheme(link.Scheme) {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedLinkScheme, link.Scheme)
	}
	switch link.Scheme {
	case "http", "https":
		if link.Hostname() == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
	case "mailto":
		if link.Opaque == "" && link.Path == "" {
			return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
		}
	}
	if link.Opaque == "" && link.Host == "" && link.Path == "" {
		return nil, fmt.Errorf("%w: %q", ErrInvalidLink, target)
	}
	return link, nil
}
