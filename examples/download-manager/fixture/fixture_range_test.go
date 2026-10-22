package fixture

import (
	"io"
	"net/http"
	"testing"
)

// TestRangeRoute covers 206, If-Range, and 416.
func TestRangeRoute(t *testing.T) {
	srv := NewServer()
	defer srv.Close()

	req, _ := http.NewRequestWithContext(t.Context(), "GET", srv.URL+PathRange, nil)
	req.Header.Set("Range", "bytes=10-19")

	resp, err := testClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()

	if resp.StatusCode != 206 || string(body) != string(fillFor(t, srv)[10:20]) {
		t.Fatalf("206: status %d len %d", resp.StatusCode, len(body))
	}

	if got := resp.Header.Get("Content-Range"); got != "bytes 10-19/65536" {
		t.Errorf("Content-Range %q", got)
	}

	req, _ = http.NewRequestWithContext(t.Context(), "GET", srv.URL+PathRange, nil)
	req.Header.Set("Range", "bytes=10-19")
	req.Header.Set("If-Range", `"stale"`)

	resp, err = testClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("stale If-Range status %d", resp.StatusCode)
	}

	req, _ = http.NewRequestWithContext(t.Context(), "GET", srv.URL+PathRange, nil)
	req.Header.Set("Range", "bytes=999999-")

	resp, err = testClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	if resp.StatusCode != 416 {
		t.Errorf("416 status %d", resp.StatusCode)
	}
}
