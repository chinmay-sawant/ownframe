package fetch_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestGETReturnsBodyAndStatus(t *testing.T) {
	t.Parallel()

	method := make(chan string, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method <- r.Method
		w.Header().Add("X-Tag", "one")
		w.Header().Add("X-Tag", "two")
		w.WriteHeader(http.StatusCreated)
		_, _ = io.WriteString(w, "hello")
	}))
	t.Cleanup(srv.Close)

	got, err := fetch.Do(fetch.AllowPrivate(context.Background()), "", srv.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if m := <-method; m != http.MethodGet || got.Status != http.StatusCreated || string(got.Body) != "hello" {
		t.Fatalf("method %s status %d body %q", m, got.Status, got.Body)
	}
	if got.Header["X-Tag"] != "one, two" {
		t.Fatalf("X-Tag = %q", got.Header["X-Tag"])
	}
}

func TestXHRPostSendsBody(t *testing.T) {
	t.Parallel()

	got := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		got <- b
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(srv.Close)

	res, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodPost, srv.URL, nil, []byte("ada"))
	if err != nil {
		t.Fatal(err)
	}
	if body := <-got; res.Status != http.StatusNoContent || !bytes.Equal(body, []byte("ada")) {
		t.Fatalf("status %d body %q", res.Status, body)
	}
}
