package web

import (
	"bufio"
	"context"
	"errors"
	"net"
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

func TestServeShutdownWithOpenSSEStream(t *testing.T) {
	t.Parallel()

	ln, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	app := &domScreen{generation: 1, width: 200, height: 100, source: "<html><body>hi</body></html>"}

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- ServeWithOptionsContext(ctx, app, addr, Options{Listener: ln})
	}()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+addr+"/events", nil)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("read initial SSE line: %v", err)
	}
	if !strings.HasPrefix(line, "retry:") {
		t.Fatalf("unexpected SSE preamble: %q", line)
	}

	cancel()
	select {
	case err := <-serveErr:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("ServeWithOptionsContext err = %v, want context.Canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("ServeWithOptionsContext hung on shutdown with open SSE stream")
	}
}
