package fetch_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestBlockedHeadersRejected(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)
	ctx := fetch.AllowPrivate(context.Background())
	for _, h := range []string{"Cookie", "Authorization", "Host", "Content-Length", "Transfer-Encoding", "Connection", "cookie"} {
		_, err := fetch.Do(ctx, http.MethodGet, srv.URL, map[string]string{h: "x"}, nil)
		if !errors.Is(err, fetch.ErrBlockedHeader) {
			t.Fatalf("%s: err = %v", h, err)
		}
	}
}

func TestMethodAllowlist(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	ctx := fetch.AllowPrivate(context.Background())
	if _, err := fetch.Do(ctx, http.MethodTrace, srv.URL, nil, nil); !errors.Is(err, fetch.ErrMethod) {
		t.Fatalf("TRACE: err = %v", err)
	}
	if _, err := fetch.Do(ctx, "CONNECT", srv.URL, nil, nil); !errors.Is(err, fetch.ErrMethod) {
		t.Fatalf("CONNECT: err = %v", err)
	}
	if _, err := fetch.Do(ctx, http.MethodPut, srv.URL, nil, nil); err != nil {
		t.Fatalf("PUT: err = %v", err)
	}
}
