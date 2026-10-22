package transfer

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
)

// TestHTTP416LengthMismatch refuses a 416 that disagrees with Expected.
func TestHTTP416LengthMismatch(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	tr := newHTTP(t)
	body, out := fetchFull(t, tr, srv.URL+fixture.PathRange)

	dest, partial := httpPaths(t)
	if err := os.WriteFile(partial, body, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := tr.Download(context.Background(), Request{
		URL: srv.URL + fixture.PathRange, Dest: dest, Partial: partial,
		Expected: int64(len(body)) + 1, Validators: out.Validators,
	}, nil)
	if !errors.Is(err, ErrBadRange) {
		t.Fatalf("want ErrBadRange, got %v", err)
	}
}

// TestHTTPNoValidatorsRestarts drops a partial that cannot be proven.
func TestHTTPNoValidatorsRestarts(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	tr := newHTTP(t)
	body, _ := fetchFull(t, tr, srv.URL+fixture.PathOK)

	dest, partial := httpPaths(t)
	if err := os.WriteFile(partial, body[:1000], 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := tr.Download(context.Background(), Request{
		URL: srv.URL + fixture.PathOK, Dest: dest, Partial: partial,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.Resumed {
		t.Error("unproven partial reported resumed")
	}

	if got := readAll(t, dest); string(got) != string(body) {
		t.Error("restarted body mismatch")
	}
}
