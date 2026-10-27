package cat

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNotificationHTTP(t *testing.T) {
	inbox := NewInbox()
	handler := inbox.Handler()
	for _, tc := range []struct {
		method, path, body string
		status             int
	}{
		{"GET", "/expressions", "", 200},
		{"POST", "/notify", `{"expression":"sleepy","message":"It is midnight. You should sleep now.","source":"Open Code"}`, 202},
		{"GET", "/state", "", 200},
		{"POST", "/notify", `{"expression":"missing","message":"done"}`, 400},
		{"POST", "/notify", `{"expression":"happy","message":""}`, 400},
		{"POST", "/notify", `{"expression":"happy","message":"done","extra":true}`, 400},
		{"POST", "/notify", `{"expression":"happy","message":"done"}{}`, 400},
		{"GET", "/notify", "", 405},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequestWithContext(t.Context(), tc.method, "http://127.0.0.1:6969"+tc.path, strings.NewReader(tc.body))
		if tc.path == "/notify" || tc.path == "/state" {
			req.Header.Set("Origin", "http://127.0.0.1:6969")
		}
		handler.ServeHTTP(w, req)
		if w.Code != tc.status {
			t.Fatalf("%s %s: %d %s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}
	n, v := inbox.Latest()
	if n.Expression != "sleepy" || v != 2 {
		t.Fatal("invalid requests replaced notification")
	}
	if len(Expressions()) != 45 {
		t.Fatalf("expression names: %d", len(Expressions()))
	}
}
