package fetch

import (
	"errors"
	"net/http"
	"net/url"
)

// client is shared. Jar stays nil, so no cookie is stored or sent.
// Timeout is the default client timeout; a request context with a shorter
// deadline still wins.
var client = &http.Client{
	Timeout:       fetchTimeout,
	CheckRedirect: redirect,
}

// ErrBodyTooLarge means a response body exceeded the 10 MiB cap.
var ErrBodyTooLarge = errors.New("ownframe: fetch response body exceeds 10 MiB")

// ErrTooManyRedirects means the client stopped after 10 redirects.
var ErrTooManyRedirects = errors.New("ownframe: fetch stopped after 10 redirects")

// ErrHeadersTooLarge means response headers exceeded the 1 MiB cap.
var ErrHeadersTooLarge = errors.New("ownframe: fetch response headers exceed 1 MiB")

// ErrCredentialsInURL means a URL contained userinfo, which is not allowed.
var ErrCredentialsInURL = errors.New("ownframe: fetch URL contains userinfo")

// Aliases for the same limit error.
var (
	ErrTooLarge          = ErrBodyTooLarge
	ErrResponseTooLarge  = ErrBodyTooLarge
	ErrBodyLimitExceeded = ErrBodyTooLarge
)

// MaxBodyBytes is the response body cap (10 MiB).
const (
	MaxBodyBytes         = 10 * 1024 * 1024
	MaxResponseBodyBytes = MaxBodyBytes
	MaxFetchBodyBytes    = MaxBodyBytes
	MaxHeaderBytes       = 1 * 1024 * 1024
)

func checkURL(raw string) error {
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

	return nil
}

func web(u *url.URL) bool {
	return u != nil && (u.Scheme == "http" || u.Scheme == "https")
}
