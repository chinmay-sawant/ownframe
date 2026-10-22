package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLimitMutating(t *testing.T) {
	t.Parallel()
	srv := &server{}
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	serve := srv.limit(ok).ServeHTTP
	mk := func(remote string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), "POST", "http://x/click", nil)
		req.RemoteAddr = remote
		rec := httptest.NewRecorder()
		serve(rec, req)
		return rec
	}
	if rec := mk("10.9.8.7:1"); rec.Code != 200 {
		t.Fatalf("first: %d", rec.Code)
	}
	denied := false
	for i := 0; i < 40; i++ {
		if rec := mk("10.9.8.7:1"); rec.Code == http.StatusTooManyRequests {
			denied = true
			break
		}
	}
	if !denied {
		t.Fatal("burst never limited")
	}
	rec := httptest.NewRecorder()
	serve(rec, httptest.NewRequestWithContext(t.Context(), "GET", "http://x/click", nil))
	if rec.Code != 200 {
		t.Fatal("GET limited")
	}
}
