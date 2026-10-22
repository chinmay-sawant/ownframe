package fetch_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestDoRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(make([]byte, fetch.MaxBodyBytes+1))
	}))
	t.Cleanup(srv.Close)

	_, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if !errors.Is(err, fetch.ErrBodyTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

func TestDoAcceptsBodyAtLimit(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(make([]byte, fetch.MaxBodyBytes))
	}))
	t.Cleanup(srv.Close)

	got, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Body) != fetch.MaxBodyBytes {
		t.Fatalf("body len = %d", len(got.Body))
	}
}

func TestDoRespectsContextDeadline(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = io.WriteString(w, "late")
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := fetch.Do(fetch.AllowPrivate(ctx), http.MethodGet, srv.URL, nil, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
}
