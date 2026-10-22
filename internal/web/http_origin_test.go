package web

import (
	"net/http/httptest"
	"testing"
)

func TestSameOriginRules(t *testing.T) {
	t.Parallel()
	mk := func(origin string) bool {
		r := httptest.NewRequestWithContext(t.Context(), "GET", "http://example.com/", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		return sameOrigin(r)
	}
	if !mk("") || !mk("http://example.com") {
		t.Fatal("missing or matching origin must pass for GETs")
	}
	if mk("https://example.com") || mk("http://evil.com") {
		t.Fatal("scheme or host mismatch must fail")
	}
}

func TestRequireOriginRules(t *testing.T) {
	t.Parallel()
	mk := func(origin string) bool {
		r := httptest.NewRequestWithContext(t.Context(), "POST", "http://example.com/click", nil)
		if origin != "" {
			r.Header.Set("Origin", origin)
		}
		return requireOrigin(r)
	}
	if mk("") || !mk("http://example.com") {
		t.Fatal("mutating routes need a present matching origin")
	}
	if mk("https://example.com") || mk("http://evil.com") {
		t.Fatal("scheme or host mismatch must fail")
	}
}

func TestClickOriginGate(t *testing.T) {
	t.Parallel()
	if rec := clickRec(&auditClick{}, "http://example.com/click?x=1&y=2", ""); rec.Code != 403 {
		t.Fatalf("missing origin: %d", rec.Code)
	}
	if rec := clickRec(&auditClick{}, "http://example.com/click?x=1&y=2", "http://evil.com"); rec.Code != 403 {
		t.Fatalf("evil origin: %d", rec.Code)
	}
	if rec := clickRec(&auditClick{}, "http://example.com/click?x=1&y=2", "http://example.com"); rec.Code != 303 {
		t.Fatalf("valid origin: %d", rec.Code)
	}
}
