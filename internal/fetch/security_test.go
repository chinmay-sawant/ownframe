package fetch_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestUserinfoRejected(t *testing.T) {
	t.Parallel()

	_, err := fetch.Do(context.Background(), http.MethodGet, "http://user:pass@example.test/", nil, nil)
	if !errors.Is(err, fetch.ErrCredentialsInURL) {
		t.Fatalf("err = %v", err)
	}
}

func TestRedirectUserinfoRejected(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "http://user:pass@"+r.Host+"/next")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(srv.Close)

	_, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if !errors.Is(err, fetch.ErrCredentialsInURL) {
		t.Fatalf("err = %v", err)
	}
}

func TestHeaderCapRejected(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Big", strings.Repeat("a", fetch.MaxHeaderBytes+1))
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)

	_, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if !errors.Is(err, fetch.ErrHeadersTooLarge) {
		t.Fatalf("err = %v", err)
	}
}
