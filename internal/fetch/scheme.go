package fetch

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync/atomic"
)

var useSystemProxy atomic.Bool

// client is shared. Jar stays nil, so no cookie is stored or sent.
// Timeout is the default client timeout; a request context with a shorter
// deadline still wins. Transport pins DNS at dial time (FETCH-1).
var client = &http.Client{
	Transport:     pinnedTransport(),
	Timeout:       fetchTimeout,
	CheckRedirect: redirect,
}

// privateClient serves AllowPrivate contexts with its own transport, so a
// keep-alive connection opened to a local host is never reused by a normal
// call, and the reverse.
var privateClient = &http.Client{
	Transport:     pinnedTransport(),
	Timeout:       fetchTimeout,
	CheckRedirect: redirect,
}

// clientFor picks the pool for ctx.
func clientFor(ctx context.Context) *http.Client {
	if allowPrivate(ctx) {
		return privateClient
	}

	return client
}

// ErrBodyTooLarge means a response body exceeded the 10 MiB cap.
var ErrBodyTooLarge = errors.New("ownframe: fetch response body exceeds 10 MiB")

// ErrTooManyRedirects means the client stopped after 10 redirects.
var ErrTooManyRedirects = errors.New("ownframe: fetch stopped after 10 redirects")

// ErrHeadersTooLarge means response headers exceeded the 1 MiB cap.
var ErrHeadersTooLarge = errors.New("ownframe: fetch response headers exceed 1 MiB")

// ErrCredentialsInURL means a URL contained userinfo, which is not allowed.
var ErrCredentialsInURL = errors.New("ownframe: fetch URL contains userinfo")

// MaxBodyBytes is the response body cap (10 MiB).
const (
	MaxBodyBytes   = 10 * 1024 * 1024
	MaxHeaderBytes = 1 * 1024 * 1024
)

func web(u *url.URL) bool {
	return u != nil && (u.Scheme == "http" || u.Scheme == "https")
}
