// Package fetch loads one http or https URL with net/http.
// It keeps no cookies and no cache.
package fetch

import (
	"errors"
	"time"
)

// ErrScheme means the URL is empty or not http or https.
// A redirect to any other scheme is the same error.
var ErrScheme = errors.New("ownframe: fetch scheme")

// ErrNilContext means Do got a nil context.
var ErrNilContext = errors.New("ownframe: nil context")

// fetchTimeout is the default client timeout.
const fetchTimeout = 30 * time.Second

// maxRedirects caps how many redirects the client follows.
const maxRedirects = 10

// Response is one completed request.
// Header is a flat map. Values for one key are joined with ", ".
type Response struct {
	Status int
	Header map[string]string
	Body   []byte
}
