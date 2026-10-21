package fetch_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

func TestPrivateAddressesRejected(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"http://127.0.0.1/",
		"http://127.0.0.1:8080/path",
		"http://10.0.0.1/",
		"http://172.16.0.1/",
		"http://192.168.1.1/",
		"http://169.254.169.254/latest/meta-data/",
		"http://0.0.0.0/",
		"http://[::1]/",
		"http://[ff02::1]/",
	} {
		_, err := fetch.Do(context.Background(), http.MethodGet, raw, nil, nil)
		if !errors.Is(err, fetch.ErrPrivateAddress) {
			t.Fatalf("%s: err = %v, want ErrPrivateAddress", raw, err)
		}
	}
}

func TestAllowPrivateReachesLoopback(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "local")
	}))
	t.Cleanup(srv.Close)

	got, err := fetch.Do(fetch.AllowPrivate(context.Background()), http.MethodGet, srv.URL, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if string(got.Body) != "local" {
		t.Fatalf("body = %q", got.Body)
	}
}
