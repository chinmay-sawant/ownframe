package fetch_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestFileScheme(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "file:///etc/passwd", "data:text/plain,hi", "javascript:alert(1)"} {
		_, err := fetch.Do(context.Background(), http.MethodGet, raw, nil, nil)
		if err != fetch.ErrScheme {
			t.Fatalf("%s: %v", raw, err)
		}
	}
}

func TestRedirectFile(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "file:///tmp/x")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	_, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if err != fetch.ErrScheme {
		t.Fatalf("err = %v", err)
	}
}

func TestRedirectHTTP(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/next", http.StatusFound)
			return
		}
		_, _ = io.WriteString(w, "next")
	}))
	t.Cleanup(srv.Close)

	got, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != http.StatusOK || string(got.Body) != "next" {
		t.Fatalf("status %d body %q", got.Status, got.Body)
	}
}
