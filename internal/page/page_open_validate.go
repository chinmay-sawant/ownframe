package page

import "net/url"

// ValidateOpenTarget allows only http, https, and mailto targets.
// It wraps parseOutboundLink so clicks and helpers share one gate.
// ownframe never invokes an OS opener itself; the app opens the
// returned link after an explicit user action such as a click.
func ValidateOpenTarget(target string) (*url.URL, error) {
	link, err := parseOutboundLink(target)
	if err != nil {
		return nil, err
	}
	if link == nil {
		return nil, ErrInvalidLink
	}
	return link, nil
}
