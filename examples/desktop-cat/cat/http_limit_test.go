package cat

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotifyRateLimited(t *testing.T) {
	t.Parallel()
	handler := NewInbox().Handler()
	mk := func() *httptest.ResponseRecorder {
		body := `{"expression":"happy","message":"hi"}`
		r := httptest.NewRequestWithContext(t.Context(), "POST", "http://127.0.0.1:6969/notify", strings.NewReader(body))
		r.RemoteAddr = "10.7.7.7:9"
		r.Header.Set("Origin", "http://127.0.0.1:6969")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		return rec
	}
	if rec := mk(); rec.Code != 202 {
		t.Fatalf("first: %d", rec.Code)
	}
	limited := false
	for i := 0; i < 40; i++ {
		if rec := mk(); rec.Code == 429 {
			limited = true
			break
		}
	}
	if !limited {
		t.Fatal("burst never limited")
	}
}
