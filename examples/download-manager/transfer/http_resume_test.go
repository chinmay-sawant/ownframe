package transfer

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
)

// TestHTTPResumesRange seeds a partial and validates a 206 resume.
func TestHTTPResumesRange(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	tr := newHTTP(t)
	body, out := fetchFull(t, tr, srv.URL+fixture.PathRange)

	dest, partial := httpPaths(t)
	if err := os.WriteFile(partial, body[:10000], 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := tr.Download(context.Background(), Request{
		URL: srv.URL + fixture.PathRange, Dest: dest, Partial: partial,
		Expected: int64(len(body)), Validators: out.Validators,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !res.Resumed {
		t.Error("206 resume not reported")
	}

	if got := readAll(t, dest); string(got) != string(body) {
		t.Error("resumed body mismatch")
	}
}

// TestHTTPRestartsWhenValidatorChanged sends a stale If-Range and gets a
// full 200 body.
func TestHTTPRestartsWhenValidatorChanged(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	tr := newHTTP(t)
	body, _ := fetchFull(t, tr, srv.URL+fixture.PathRange)

	dest, partial := httpPaths(t)
	if err := os.WriteFile(partial, body[:10000], 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := tr.Download(context.Background(), Request{
		URL: srv.URL + fixture.PathRange, Dest: dest, Partial: partial,
		Expected:   int64(len(body)),
		Validators: Validators{ETag: `"stale"`},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.Resumed {
		t.Error("stale validator reported resumed")
	}

	if got := readAll(t, dest); string(got) != string(body) {
		t.Error("restarted body mismatch")
	}
}
