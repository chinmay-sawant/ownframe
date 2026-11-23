package web

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func TestEventsSubscriberCap(t *testing.T) {
	t.Parallel()

	srv := &server{app: &domScreen{width: 10, height: 10}}
	chans := make([]chan []byte, 0, maxSubscribers)
	for range maxSubscribers {
		ch, ok := srv.trySubscribe(1)
		if !ok {
			t.Fatal("subscribe refused below the cap")
		}
		chans = append(chans, ch)
	}
	if _, ok := srv.trySubscribe(1); ok {
		t.Fatal("subscribe accepted past the cap")
	}

	rec := httptest.NewRecorder()
	srv.events(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/events", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("events past cap = %d, want 503", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "retry:") || rec.Header().Get("Content-Type") == "text/event-stream" {
		t.Fatalf("refused stream still looks like SSE: %q %v", rec.Body.String(), rec.Header())
	}

	srv.unsubscribe(chans[0])
	if _, ok := srv.trySubscribe(1); !ok {
		t.Fatal("slot not freed after unsubscribe")
	}
}

func TestEventsOnClosedServerWritesNoHeaders(t *testing.T) {
	t.Parallel()

	srv := &server{app: &domScreen{width: 10, height: 10}}
	srv.closeSubs()
	rec := httptest.NewRecorder()
	srv.events(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/events", nil))
	if rec.Code != http.StatusServiceUnavailable || strings.Contains(rec.Body.String(), "retry:") {
		t.Fatalf("closed server: %d %q", rec.Code, rec.Body.String())
	}
}

// fakeRemoteListener reports a non-loopback address.
type fakeRemoteListener struct{ net.Listener }

func (fakeRemoteListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("192.0.2.10"), Port: 8080}
}

func TestPreboundRemoteListenerNeedsOptIn(t *testing.T) {
	t.Parallel()

	ln := fakeRemoteListener{Listener: testListener(t)}
	err := ServeWithOptionsContext(t.Context(), &ctxFake{}, "127.0.0.1:8080", Options{Listener: ln})
	if !errors.Is(err, errRemoteBind) {
		t.Fatalf("err = %v, want errRemoteBind", err)
	}
	if err := checkBindAddr("", Options{Listener: ln, InsecureAllowRemoteBind: true}); err != nil {
		t.Fatalf("opt-in refused: %v", err)
	}
	if err := checkBindAddr("", Options{Listener: testListener(t)}); err != nil {
		t.Fatalf("loopback listener refused: %v", err)
	}
}

func TestDebugStateRequiresSameOrigin(t *testing.T) {
	t.Parallel()

	srv := &server{app: &domScreen{width: 10, height: 10}}
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/debug/state", nil)
	r.Header.Set("Origin", "http://evil.example")
	rec := httptest.NewRecorder()
	srv.debug(rec, r)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin /debug/state = %d, want 403", rec.Code)
	}
}

// pwScreen serves one password control.
type pwScreen struct{ domScreen }

func (*pwScreen) WebControls() []host.WebControl {
	return []host.WebControl{{ID: "pw", Tag: "input", Type: "password", Value: "hunter2"}, {ID: "u", Tag: "input", Type: "text", Value: "ann"}}
}

func (*pwScreen) SelectControls() map[string][]host.SelectOption { return nil }

func TestPasswordValuesRedactedOnRemoteBind(t *testing.T) {
	t.Parallel()

	src := `<body><span id="pw" data-ownframe-field="input"></span><span id="u" data-ownframe-field="input"></span></body>`
	for _, remote := range []bool{false, true} {
		srv := &server{app: &pwScreen{domScreen{width: 10, height: 10, source: src}}, redactPasswords: remote}
		rec := httptest.NewRecorder()
		srv.dom(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/dom", nil))
		page := httptest.NewRecorder()
		srv.page(page, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/", nil))
		for name, body := range map[string]string{"dom": rec.Body.String(), "page": page.Body.String()} {
			has := strings.Contains(body, "hunter2")
			if has == remote {
				t.Errorf("remote=%v %s: plaintext password present=%v", remote, name, has)
			}
			if !strings.Contains(body, "ann") {
				t.Errorf("remote=%v %s: text value lost", remote, name)
			}
		}
	}
}

func TestServeLogsBoundAddress(t *testing.T) {
	// Covered by reading ln.Addr(): an ephemeral ":0" must not print as ":0".
	ln := testListener(t)
	done := make(chan error, 1)
	ctx, cancel := context.WithCancel(t.Context())
	go func() { done <- ServeWithOptionsContext(ctx, &ctxFake{}, "127.0.0.1:0", Options{Listener: ln}) }()
	waitUp(t, ln.Addr().String(), done)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Logf("serve stopped with %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serve did not stop")
	}
}
