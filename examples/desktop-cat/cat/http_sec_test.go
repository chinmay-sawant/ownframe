package cat

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func notifyReq(method, host, origin, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequestWithContext(context.Background(), method, "http://"+host+"/notify", strings.NewReader(body))
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	NewInbox().Handler().ServeHTTP(rec, r)
	return rec
}

func TestNotifyOriginCheck(t *testing.T) {
	t.Parallel()
	ok := `{"expression":"happy","message":"done","source":"t"}`
	if rec := notifyReq("POST", "127.0.0.1:6969", "http://127.0.0.1:6969", ok); rec.Code != 202 {
		t.Fatalf("valid origin: %d %s", rec.Code, rec.Body.String())
	}
	if rec := notifyReq("POST", "127.0.0.1:6969", "", ok); rec.Code != 403 {
		t.Fatalf("missing origin fails closed: %d %s", rec.Code, rec.Body.String())
	}
	if rec := notifyReq("POST", "127.0.0.1:6969", "http://evil.com", ok); rec.Code != 403 {
		t.Fatalf("evil origin: %d", rec.Code)
	}
	if rec := notifyReq("POST", "127.0.0.1:6969", "https://127.0.0.1:6969", ok); rec.Code != 403 {
		t.Fatalf("scheme mismatch: %d", rec.Code)
	}
}

func TestNotifyGenericErrors(t *testing.T) {
	t.Parallel()
	good := "http://127.0.0.1:6969"
	rec := notifyReq("POST", "127.0.0.1:6969", good, `{"expression":"nope","message":"x"}`)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "invalid notification") {
		t.Fatalf("submit error: %d %s", rec.Code, rec.Body.String())
	}
	rec = notifyReq("POST", "127.0.0.1:6969", good, `{"expression":`)
	if rec.Code != 400 || strings.Contains(rec.Body.String(), "unexpected") {
		t.Fatalf("decode error: %d %s", rec.Code, rec.Body.String())
	}
}
