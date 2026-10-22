package cat

import (
	"context"
	"net/http/httptest"
	"testing"
)

func stateReq(host, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequestWithContext(context.Background(), "GET", "http://"+host+"/state", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	NewInbox().Handler().ServeHTTP(rec, r)
	return rec
}

func TestStateOriginCheck(t *testing.T) {
	t.Parallel()
	if rec := stateReq("127.0.0.1:6969", "http://127.0.0.1:6969"); rec.Code != 200 {
		t.Fatalf("valid origin: %d", rec.Code)
	}
	if rec := stateReq("127.0.0.1:6969", ""); rec.Code != 403 {
		t.Fatalf("missing origin: %d", rec.Code)
	}
	if rec := stateReq("127.0.0.1:6969", "http://evil.com"); rec.Code != 403 {
		t.Fatalf("evil origin: %d", rec.Code)
	}
}

func TestCatSecurityHeaders(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequestWithContext(t.Context(), "GET", "http://127.0.0.1:6969/expressions", nil)
	rec := httptest.NewRecorder()
	NewInbox().Handler().ServeHTTP(rec, r)
	for k, want := range map[string]string{"X-Content-Type-Options": "nosniff", "X-Frame-Options": "DENY", "Referrer-Policy": "no-referrer"} {
		if rec.Header().Get(k) != want {
			t.Fatalf("%s = %q", k, rec.Header().Get(k))
		}
	}
	if got := rec.Header().Get("Permissions-Policy"); got == "" {
		t.Fatal("missing Permissions-Policy")
	}
}
