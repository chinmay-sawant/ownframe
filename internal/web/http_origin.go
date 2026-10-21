package web

import (
	"net/http"
	"net/url"
	"strings"
)

// reqScheme is http unless the request arrived over TLS.
func reqScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

// sameOrigin allows a missing Origin for safe GETs. When Origin is
// present it must match the request scheme and host.
func sameOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if !strings.EqualFold(u.Scheme, reqScheme(r)) {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}

// requireOrigin is for mutating routes. A missing Origin fails closed,
// so a plain form or curl without Origin cannot change page state.
func requireOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil || u.Host == "" {
		return false
	}
	if !strings.EqualFold(u.Scheme, reqScheme(r)) {
		return false
	}
	return strings.EqualFold(u.Host, r.Host)
}
