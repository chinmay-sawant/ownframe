package fetch

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
)

func parseURL(t *testing.T, raw string) *url.URL {
	t.Helper()

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}

	return u
}

func TestRedirectRejectsPrivateHost(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{
		"http://127.0.0.1:8080/",
		"http://10.0.0.1/",
		"http://169.254.169.254/latest/meta-data/",
	} {
		req := &http.Request{URL: parseURL(t, raw)}
		via := []*http.Request{{URL: parseURL(t, "http://93.184.216.34/")}}
		if err := redirect(req, via); !errors.Is(err, ErrPrivateAddress) {
			t.Fatalf("%s: err = %v", raw, err)
		}
	}
}

func TestRedirectRejectsDowngrade(t *testing.T) {
	t.Parallel()

	req := &http.Request{URL: parseURL(t, "http://93.184.216.34/next")}
	via := []*http.Request{{URL: parseURL(t, "https://93.184.216.34/")}}
	if err := redirect(req, via); !errors.Is(err, ErrDowngrade) {
		t.Fatalf("err = %v", err)
	}
}

func TestRedirectAllowsPublicHop(t *testing.T) {
	t.Parallel()

	req := &http.Request{URL: parseURL(t, "http://93.184.216.34/next")}
	if err := redirect(req, nil); err != nil {
		t.Fatalf("err = %v", err)
	}
}
