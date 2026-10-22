package fixture

import (
	"io"
	"testing"
)

// TestStableRoutes covers ok, chunked, and redirect.
func TestStableRoutes(t *testing.T) {
	srv := NewServer()
	defer srv.Close()

	resp := testGet(t, srv.URL+PathOK)

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != 200 || len(body) != 64<<10 {
		t.Fatalf("ok: status %d len %d", resp.StatusCode, len(body))
	}

	resp = testGet(t, srv.URL+PathChunked)

	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.ContentLength != -1 || len(body) != 64<<10 {
		t.Fatalf("chunked: len %d body %d", resp.ContentLength, len(body))
	}

	resp = testGet(t, srv.URL+PathRedirect)

	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != 200 || len(body) != 64<<10 {
		t.Fatalf("redirect: status %d len %d", resp.StatusCode, len(body))
	}
}

// fillFor rebuilds the stable body for comparison.
func fillFor(t *testing.T, s *Server) []byte {
	t.Helper()

	return s.fill(seedFrom(PathRange))
}
