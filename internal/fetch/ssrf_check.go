package fetch

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
)

// checkURL rejects empty URLs, userinfo, non-web schemes, and hosts that
// resolve to a blocked address.
func checkURL(ctx context.Context, raw string) error {
	if raw == "" {
		return ErrScheme
	}

	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if u.User != nil {
		return ErrCredentialsInURL
	}
	if !web(u) {
		return ErrScheme
	}

	return checkHost(ctx, u.Hostname())
}

// checkHost rejects literal blocked IPs and hostnames whose DNS answers
// include one. It fails closed when resolution fails.
func checkHost(ctx context.Context, host string) error {
	if host == "" {
		return ErrScheme
	}
	if allowPrivate(ctx) {
		return nil
	}
	if addr, err := netip.ParseAddr(host); err == nil {
		if blockedAddr(addr) {
			return ErrPrivateAddress
		}

		return nil
	}
	if !useSystemProxy.Load() {
		// Without a proxy the connection is made by dialValidated, which
		// resolves the name once and refuses blocked answers. Resolving
		// here as well would only double the lookups. A redirect hop is
		// pinned the same way. Behind a proxy the proxy resolves, so the
		// lookup below stays as the guard.
		return nil
	}
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("ownframe: fetch cannot resolve %s: %w", host, err)
	}
	for _, addr := range addrs {
		if blockedIP(addr.IP) {
			return ErrPrivateAddress
		}
	}

	return nil
}

func redirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return ErrTooManyRedirects
	}
	if req.URL.User != nil {
		return ErrCredentialsInURL
	}
	if !web(req.URL) {
		return ErrScheme
	}
	if len(via) > 0 {
		prev := via[len(via)-1]
		if prev.URL.Scheme == "https" && req.URL.Scheme == "http" {
			return ErrDowngrade
		}
	}

	return checkHost(req.Context(), req.URL.Hostname())
}
