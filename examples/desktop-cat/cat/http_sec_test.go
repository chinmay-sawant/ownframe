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

func TestHostCheck(t *testing.T) {
	t.Parallel()
	inbox := NewInbox()
	handler := inbox.Handler()

	// Rejected non-loopback hosts
	for _, badHost := range []string{
		"evil.com",
		"evil.com:6969",
		"attacker.com",
		"attacker.com:1234",
		"example.com",
		"192.168.1.1:6969",
		"10.0.0.1",
		"",
	} {
		r := httptest.NewRequestWithContext(context.Background(), "GET", "http://placeholder/expressions", nil)
		r.Host = badHost
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		if rec.Code != 403 {
			t.Fatalf("bad host %q: expected 403, got %d", badHost, rec.Code)
		}
	}

	// Accepted loopback hosts
	for _, goodHost := range []string{
		"localhost",
		"localhost:6969",
		"127.0.0.1",
		"127.0.0.1:6969",
		"[::1]",
		"[::1]:6969",
		"127.0.0.2:8080",
	} {
		r := httptest.NewRequestWithContext(context.Background(), "GET", "http://placeholder/expressions", nil)
		r.Host = goodHost
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		if rec.Code != 200 {
			t.Fatalf("good host %q: expected 200, got %d", goodHost, rec.Code)
		}
	}
}

func TestDNSRebindingRejected(t *testing.T) {
	t.Parallel()
	ok := `{"expression":"happy","message":"done","source":"t"}`
	rec := notifyReq("POST", "evil.com:6969", "http://evil.com:6969", ok)
	if rec.Code != 403 {
		t.Fatalf("dns rebinding with matching evil host and origin should be 403, got %d", rec.Code)
	}
}
