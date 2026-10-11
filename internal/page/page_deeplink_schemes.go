package page

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidLink means a URL is malformed or relative.
var ErrInvalidLink = errors.New("ownframe: invalid link")

// ErrUnsupportedLinkScheme means the URL scheme is not allowlisted.
var ErrUnsupportedLinkScheme = errors.New("ownframe: unsupported link scheme")

// ErrBadLinkScheme means Config contains an invalid or blocked scheme.
var ErrBadLinkScheme = errors.New("ownframe: bad link scheme")

// ErrNoDeepLinkHandler means no DeepLink callback is registered.
var ErrNoDeepLinkHandler = errors.New("ownframe: no deep link handler")

func acceptedDeepLinkSchemes(input []string) (map[string]struct{}, error) {
	schemes := make(map[string]struct{}, len(input))
	for _, raw := range input {
		scheme := strings.ToLower(raw)
		if !validLinkScheme(scheme) || blockedLinkScheme(scheme) {
			return nil, fmt.Errorf("%w: %q", ErrBadLinkScheme, raw)
		}
		schemes[scheme] = struct{}{}
	}
	return schemes, nil
}

func validLinkScheme(scheme string) bool {
	if len(scheme) == 0 || !letter(scheme[0]) {
		return false
	}
	for i := 1; i < len(scheme); i++ {
		c := scheme[i]
		if !letter(c) && (c < '0' || c > '9') && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return true
}

func letter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func blockedLinkScheme(scheme string) bool {
	switch scheme {
	case "javascript", "data", "vbscript", "file", "blob", "about":
		return true
	default:
		return false
	}
}
