package fetch_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
)

// TestSec_FetchPrivateBlocked pins the SSRF private-address block and
// the method/header guard. Literal loopback needs no DNS.
func TestSec_FetchPrivateBlocked(t *testing.T) {
	t.Parallel()
	ctx := fetch.AllowPrivate(context.Background())
	if _, err := fetch.Do(context.Background(), http.MethodGet, "http://127.0.0.1:1/", nil, nil); !errors.Is(err, fetch.ErrPrivateAddress) {
		t.Fatalf("private: %v", err)
	}
	if _, err := fetch.Do(ctx, "BREW", "http://127.0.0.1:1/", nil, nil); !errors.Is(err, fetch.ErrMethod) {
		t.Fatalf("method: %v", err)
	}
	if _, err := fetch.Do(ctx, http.MethodGet, "http://127.0.0.1:1/", map[string]string{"Cookie": "x"}, nil); !errors.Is(err, fetch.ErrBlockedHeader) {
		t.Fatalf("header: %v", err)
	}
}
