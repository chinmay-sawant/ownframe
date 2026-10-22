package fetch

import (
	"errors"
	"net/http"
)

// ErrMethod means the fetch method is not on the allowlist.
var ErrMethod = errors.New("ownframe: fetch method not allowed")

// ErrBlockedHeader means a caller set a sensitive fetch header.
var ErrBlockedHeader = errors.New("ownframe: fetch header blocked")

// checkMethodHeader allowlists methods and blocks headers that would
// smuggle credentials or hop semantics into the shared client.
func checkMethodHeader(method string, header map[string]string) error {
	if method == "" {
		method = http.MethodGet
	}
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
	default:
		return ErrMethod
	}
	for key := range header {
		switch http.CanonicalHeaderKey(key) {
		case "Cookie", "Authorization", "Host", "Content-Length", "Transfer-Encoding", "Connection":
			return ErrBlockedHeader
		}
	}
	return nil
}
