package transfer

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
)

// TestHTTP416FinalizesCompletePartial adopts a full partial.
func TestHTTP416FinalizesCompletePartial(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	tr := newHTTP(t)
	body, out := fetchFull(t, tr, srv.URL+fixture.PathRange)

	dest, partial := httpPaths(t)
	if err := os.WriteFile(partial, body, 0o600); err != nil {
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
		t.Error("416 completion not reported as resumed")
	}

	if got := readAll(t, dest); string(got) != string(body) {
		t.Error("416 body mismatch")
	}
}

// TestHTTPChangingContent restarts when the entity changed.
func TestHTTPChangingContent(t *testing.T) {
	srv := fixture.NewServer()
	defer srv.Close()

	tr := newHTTP(t)
	body, out := fetchFull(t, tr, srv.URL+fixture.PathChanging)

	dest, partial := httpPaths(t)
	if err := os.WriteFile(partial, body[:10000], 0o600); err != nil {
		t.Fatal(err)
	}

	res, err := tr.Download(context.Background(), Request{
		URL: srv.URL + fixture.PathChanging, Dest: dest, Partial: partial,
		Validators: out.Validators,
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	if res.Resumed {
		t.Error("changed content reported resumed")
	}

	if res.Validators.ETag == out.Validators.ETag {
		t.Error("changed route kept the old ETag")
	}
}
