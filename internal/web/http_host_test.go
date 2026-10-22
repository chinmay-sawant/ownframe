package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHostAllowlist(t *testing.T) {
	t.Parallel()
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) })
	serve := requireLocalHost("app.test", next).ServeHTTP
	for _, host := range []string{"localhost:8128", "127.0.0.1:8128", "[::1]:8128", "app.test:8128"} {
		rec := httptest.NewRecorder()
		serve(rec, httptest.NewRequestWithContext(t.Context(), "GET", "http://"+host+"/", nil))
		if rec.Code != 200 {
			t.Fatalf("%s: %d", host, rec.Code)
		}
	}
	rec := httptest.NewRecorder()
	serve(rec, httptest.NewRequestWithContext(t.Context(), "GET", "http://evil.test/", nil))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("rebind: %d", rec.Code)
	}
}
