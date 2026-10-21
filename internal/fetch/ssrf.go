package fetch

import (
	"context"
	"errors"
	"net"
)

// ErrPrivateAddress means a fetch URL resolved to a blocked address:
// loopback, private, link-local, multicast, or unspecified, including
// the 169.254.169.254 metadata range.
var ErrPrivateAddress = errors.New("ownframe: fetch blocked private address")

// ErrDowngrade means a redirect stepped from https down to http.
var ErrDowngrade = errors.New("ownframe: fetch redirect downgrades https")

// privateKey marks a context that may reach blocked addresses. Tests
// and trusted local tooling use it; production callers pass a plain
// context and stay behind the block.
type privateKey struct{}

// AllowPrivate returns a context that skips the private-address block.
// It never skips the scheme or userinfo checks.
func AllowPrivate(ctx context.Context) context.Context {
	return context.WithValue(ctx, privateKey{}, true)
}

func allowPrivate(ctx context.Context) bool {
	return ctx != nil && ctx.Value(privateKey{}) == true
}

// blockedIP reports loopback, private, link-local, multicast, and
// unspecified addresses.
func blockedIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() || ip.IsLoopback() || ip.IsMulticast() {
		return true
	}
	if ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	return !ip.IsGlobalUnicast()
}
