package ownframe

import "github.com/chinmay-sawant/ownframe/internal/page"

var (
	// ErrNoHistory means Back or Forward has no entry.
	ErrNoHistory = page.ErrNoHistory
	// ErrInvalidLink means a dispatched or clicked URL is malformed.
	ErrInvalidLink = page.ErrInvalidLink
	// ErrUnsupportedLinkScheme means the scheme is outside its allowlist.
	ErrUnsupportedLinkScheme = page.ErrUnsupportedLinkScheme
	// ErrBadLinkScheme means Config contains a blocked or malformed scheme.
	ErrBadLinkScheme = page.ErrBadLinkScheme
	// ErrNoDeepLinkHandler means no incoming URL callback is registered.
	ErrNoDeepLinkHandler = page.ErrNoDeepLinkHandler
)
