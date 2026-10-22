package fixture

import (
	"io"
	"net/http"
	"testing"
)

// TestChangingRoute bumps the ETag on every hit.
func TestChangingRoute(t *testing.T) {
	srv := NewServer()
	defer srv.Close()

	first := etagOf(t, srv.URL+PathChanging)
	second := etagOf(t, srv.URL+PathChanging)

	if first == second {
		t.Fatalf("ETag did not change: %q", first)
	}
}

// TestStatusRoute answers the requested code.
func TestStatusRoute(t *testing.T) {
	srv := NewServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/status?code=404")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("status %d", resp.StatusCode)
	}
}

// TestInterruptRoute drops the connection mid-body.
func TestInterruptRoute(t *testing.T) {
	srv := NewServer()
	defer srv.Close()

	resp, err := http.Get(srv.URL + PathInterrupt)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, readErr := io.ReadAll(resp.Body)
	if readErr == nil {
		t.Fatal("interrupted body read cleanly")
	}

	if len(body) != 32<<10 {
		t.Errorf("read %d bytes before the drop", len(body))
	}
}

// etagOf fetches url and returns its ETag.
func etagOf(t *testing.T, url string) string {
	t.Helper()

	resp, err := http.Get(url) //nolint:gosec // reason: test fetches the local fixture URL.
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	_, _ = io.Copy(io.Discard, resp.Body)

	return resp.Header.Get("ETag")
}
