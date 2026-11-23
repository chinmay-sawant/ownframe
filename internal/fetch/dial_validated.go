package fetch

import (
	"context"
	"net"
	"net/netip"
)

// dialValidated resolves the dial host once and dials a validated IP.
// Literal blocked IPs fail. Hostnames fail when any answer is blocked.
func dialValidated(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	if allowPrivate(ctx) {
		return dialer().DialContext(ctx, network, addr)
	}
	if ip, perr := netip.ParseAddr(host); perr == nil {
		if blockedAddr(ip) {
			return nil, ErrPrivateAddress
		}
		return dialer().DialContext(ctx, network, addr)
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	for _, a := range addrs {
		if blockedIP(a.IP) {
			return nil, ErrPrivateAddress
		}
	}
	var last error
	for _, a := range addrs {
		c, e := dialer().DialContext(ctx, network, net.JoinHostPort(a.IP.String(), port))
		if e == nil {
			return c, nil
		}
		last = e
	}
	if last != nil {
		return nil, last
	}
	return nil, ErrPrivateAddress
}
