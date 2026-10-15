package fetch

import (
	"errors"
	"net/http"
	"net/url"
	"time"
)

// client is shared. Jar stays nil, so no cookie is stored or sent.
// Timeout is the default client timeout; a request context with a shorter
// deadline still wins.
var client = &http.Client{
	Timeout:       30 * time.Second,
	CheckRedirect: redirect,
}

// ErrBodyTooLarge means a response body exceeded the 10 MiB cap.
var ErrBodyTooLarge = errors.New("ownframe: fetch response body exceeds 10 MiB")

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
)

func checkURL(raw string) error {
	if raw == "" {
		return ErrScheme
	}

	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if !web(u) {
		return ErrScheme
	}

	return nil
}

func web(u *url.URL) bool {
	return u != nil && (u.Scheme == "http" || u.Scheme == "https")
}

func redirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	if !web(req.URL) {
		return ErrScheme
	}

	return nil
}
