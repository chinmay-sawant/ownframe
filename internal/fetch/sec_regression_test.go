package fetch_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// TestSec_FetchUserinfoRejected pins userinfo rejection, including a
// redirect that smuggles credentials.
func TestSec_FetchUserinfoRejected(t *testing.T) {
	t.Parallel()
	if _, err := fetch.Do(context.Background(), http.MethodGet, "http://user:pass@example.test/", nil, nil); !errors.Is(err, fetch.ErrCredentialsInURL) {
		t.Fatalf("direct: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://user:pass@"+r.Host+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	if _, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil); !errors.Is(err, fetch.ErrCredentialsInURL) {
		t.Fatalf("redirect: %v", err)
	}
}

// TestSec_FetchSchemeRejected pins the http(s) allowlist, including a
// redirect to a blocked scheme.
func TestSec_FetchSchemeRejected(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"", "file:///etc/passwd", "data:text/plain,hi", "javascript:alert(1)"} {
		if _, err := fetch.Do(context.Background(), http.MethodGet, raw, nil, nil); !errors.Is(err, fetch.ErrScheme) {
			t.Fatalf("%q: %v", raw, err)
		}
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "file:///tmp/x")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)
	if _, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil); !errors.Is(err, fetch.ErrScheme) {
		t.Fatalf("redirect: %v", err)
	}
}
