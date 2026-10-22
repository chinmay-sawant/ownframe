package fixture

import (
	"net/http"
	"testing"
	"time"
)

// fixtureTimeout bounds fixture fetches.
const fixtureTimeout = 10 * time.Second

// testClient never uses the shared global client: timeouts stay
// explicit and global client state cannot change fixture fetches.
func testClient() *http.Client {
	return &http.Client{Timeout: fixtureTimeout}
}

// testGet fetches url with a request context through the test client.
func testGet(t *testing.T, url string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := testClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}
