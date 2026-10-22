package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeadersSet(t *testing.T) {
	t.Parallel()
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	rec := httptest.NewRecorder()
	securityHeaders(next).ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), "GET", "/", nil))
	for k, want := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Referrer-Policy": "no-referrer"} {
		if rec.Header().Get(k) != want {
			t.Fatalf("%s = %q", k, rec.Header().Get(k))
		}
	}
	if got := rec.Header().Get("Permissions-Policy"); got == "" {
		t.Fatal("missing Permissions-Policy")
	}
}

func TestPanicRecoveryReturns500(t *testing.T) {
	t.Parallel()
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), "GET", "/", nil)
	panicRecovery(next).ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("code = %d", rec.Code)
	}
}
