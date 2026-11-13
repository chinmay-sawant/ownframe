package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestEventsHeadersAndReplay(t *testing.T) {
	t.Parallel()

	app := &stateScreen{generation: 9, width: 320, height: 200}
	srv := &server{app: app, subs: map[chan []byte]struct{}{}}
	srv.publishIfChangedLocked()

	ctx, cancel := context.WithCancel(t.Context())
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/events", nil)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		defer close(done)
		srv.events(rec, req)
	}()
	time.Sleep(200 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("events handler did not exit on cancel")
	}

	res := rec.Result()
	if got := res.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
	body := rec.Body.String()
	if !strings.HasPrefix(body, "retry: 3000\n\n") {
		t.Fatalf("missing retry preamble: %q", body)
	}
	if !strings.Contains(body, "data: ") {
		t.Fatalf("missing snapshot replay: %q", body)
	}
	if !strings.Contains(body, `"domGeneration":9`) {
		t.Fatalf("replay missing domGeneration: %q", body)
	}
}

func TestEventsPublishesOnGenerationBump(t *testing.T) {
	t.Parallel()

	app := &domScreen{generation: 4, width: 100, height: 50, source: "<html><body></body></html>"}
	srv := &server{app: app, subs: map[chan []byte]struct{}{}}
	srv.publishIfChangedLocked()

	ch := srv.subscribe(1)
	defer srv.unsubscribe(ch)
	select {
	case <-ch:
	default:
	}

	app.generation = 5
	srv.publishIfChangedLocked()

	select {
	case msg := <-ch:
		if !strings.Contains(string(msg), `"domGeneration":5`) {
			t.Fatalf("event missing bumped domGeneration: %q", msg)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("no event after generation bump")
	}
}

func TestEventsSilentWhenIdle(t *testing.T) {
	t.Parallel()

	app := &domScreen{generation: 7, width: 100, height: 50, source: "<html><body></body></html>"}
	srv := &server{app: app, subs: map[chan []byte]struct{}{}}
	srv.publishIfChangedLocked()

	ch := srv.subscribe(1)
	defer srv.unsubscribe(ch)
	select {
	case <-ch:
	default:
	}

	srv.publishIfChangedLocked()

	select {
	case msg := <-ch:
		t.Fatalf("idle publish emitted: %q", msg)
	case <-time.After(200 * time.Millisecond):
	}
}
