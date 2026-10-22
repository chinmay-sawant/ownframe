package ownframe

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// FetchResponse is the status, flat headers, and body from Fetch or XHR.
type FetchResponse = fetch.Response

// ErrScheme means a fetch URL is empty or not http or https.
var ErrScheme = fetch.ErrScheme

// ErrPrivateAddress means a fetch URL resolved to a blocked address:
// loopback, private, link-local, multicast, or unspecified.
var ErrPrivateAddress = fetch.ErrPrivateAddress

// ErrDowngrade means a redirect stepped from https down to http.
var ErrDowngrade = fetch.ErrDowngrade

// AllowPrivate returns a context that skips the private-address block, for
// tests and trusted local tooling. Scheme and userinfo checks still apply.
func AllowPrivate(ctx context.Context) context.Context {
	return fetch.AllowPrivate(ctx)
}

// Fetch GETs rawURL. It sends no body.
func Fetch(ctx context.Context, rawURL string) (FetchResponse, error) {
	return fetch.Do(ctx, "GET", rawURL, nil, nil)
}

// XHR sends method, header, and body to rawURL.
// An empty method is GET.
func XHR(ctx context.Context, method, rawURL string, header map[string]string, body []byte) (FetchResponse, error) {
	return fetch.Do(ctx, method, rawURL, header, body)
}
