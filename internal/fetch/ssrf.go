package fetch

import (
	"context"
	"errors"
	"net"
	"net/netip"
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
// It never skips the scheme or userinfo checks. A nil context stays nil, so
// Do reports ErrNilContext instead of this call panicking.
func AllowPrivate(ctx context.Context) context.Context {
	if ctx == nil {
		return nil
	}

	return context.WithValue(ctx, privateKey{}, true)
}

func allowPrivate(ctx context.Context) bool {
	return ctx != nil && ctx.Value(privateKey{}) == true
}

// blockedPrefixes are global-looking ranges that still reach internal or
// special-purpose hosts. net.IP.IsGlobalUnicast accepts all of them.
var blockedPrefixes = mustPrefixes(
	"0.0.0.0/8",       // "this" network
	"100.64.0.0/10",   // RFC 6598 carrier-grade NAT
	"192.0.0.0/24",    // IETF protocol assignments
	"192.0.2.0/24",    // TEST-NET-1
	"198.18.0.0/15",   // RFC 2544 benchmarking
	"198.51.100.0/24", // TEST-NET-2
	"203.0.113.0/24",  // TEST-NET-3
	"240.0.0.0/4",     // reserved and broadcast
	"64:ff9b::/96",    // NAT64 embeds any IPv4, including metadata
	"64:ff9b:1::/48",  // local-use NAT64
	"100::/64",        // discard-only
	"2001::/32",       // Teredo
	"2001:db8::/32",   // documentation
	"2002::/16",       // 6to4 embeds any IPv4
)

func mustPrefixes(list ...string) []netip.Prefix {
	out := make([]netip.Prefix, len(list))
	for i, s := range list {
		out[i] = netip.MustParsePrefix(s)
	}

	return out
}

// blockedAddr reports loopback, private, link-local, multicast, unspecified,
// and special-purpose addresses. An IPv4-mapped IPv6 address is judged as the
// IPv4 address it carries, and a zone is ignored.
func blockedAddr(addr netip.Addr) bool {
	addr = addr.Unmap().WithZone("")
	if !addr.IsValid() || addr.IsUnspecified() || addr.IsLoopback() || addr.IsMulticast() {
		return true
	}
	if addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() || addr.IsInterfaceLocalMulticast() {
		return true
	}
	for _, p := range blockedPrefixes {
		if p.Contains(addr) {
			return true
		}
	}

	return !addr.IsGlobalUnicast()
}

// blockedIP is blockedAddr for a net.IP. A nil or malformed IP is blocked.
func blockedIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)

	return !ok || blockedAddr(addr)
}
