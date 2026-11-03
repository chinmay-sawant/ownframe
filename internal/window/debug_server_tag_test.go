//go:build ownframe_debug

package window

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestCheckDebugBindAddr(t *testing.T) {
	for _, addr := range []string{"127.0.0.1:8128", "localhost:8128", "[::1]:8128", "127.0.0.1:0"} {
		if err := checkDebugBindAddr(addr); err != nil {
			t.Fatalf("%s: %v, want nil", addr, err)
		}
	}
	for _, addr := range []string{"", "127.0.0.1", ":8128", "0.0.0.0:8128", "[::]:8128", "192.168.1.10:8128", "example.com:8128"} {
		if err := checkDebugBindAddr(addr); err == nil {
			t.Fatalf("%s: nil error, want a rejection", addr)
		}
	}
}

func TestDebugAllowedHost(t *testing.T) {
	if !debugAllowedHost("127.0.0.1:8128", "") {
		t.Fatal("loopback host rejected")
	}
	if !debugAllowedHost("localhost:8128", "") {
		t.Fatal("localhost rejected")
	}
	if debugAllowedHost("example.com:8128", "127.0.0.1") {
		t.Fatal("foreign host passed")
	}
	if !debugAllowedHost("example.com:8128", "example.com") {
		t.Fatal("configured host rejected")
	}
}

// TestDebugServerServesState starts the real listener on an ephemeral
// loopback port, publishes one snapshot, and reads it back with the field
// names the sampler expects.
func TestDebugServerServesState(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	game := newDevShell(newDevScreen())
	game.debugEnabled = true
	game.perf = true
	bound, stop, err := startDebugServer(ctx, game, "127.0.0.1:0")
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer stop()
	game.maybePublishDebug()
	client := &http.Client{Timeout: 5 * time.Second}
	state, err := fetchDebugJSON(client, "http://"+bound+"/debug/state")
	if err != nil {
		t.Fatalf("state: %v", err)
	}
	stats, ok := state["stats"].(map[string]any)
	if !ok {
		t.Fatalf("state = %v, want a stats object", state)
	}
	for _, key := range []string{"Redraws", "Parses", "Cascades", "Layouts", "Repaints"} {
		if _, ok := stats[key]; !ok {
			t.Fatalf("stats = %v, want %s", stats, key)
		}
	}
	runtime, ok := state["runtime"].(map[string]any)
	if !ok {
		t.Fatalf("state = %v, want a runtime object", state)
	}
	for _, key := range []string{"heap_alloc", "goroutines"} {
		if _, ok := runtime[key]; !ok {
			t.Fatalf("runtime = %v, want %s", runtime, key)
		}
	}
	if _, ok := state["frame"].(map[string]any); !ok {
		t.Fatalf("state = %v, want a frame object", state)
	}
	if status := fetchDebugStatus(client, "http://"+bound+"/debug/pprof/"); status != http.StatusOK {
		t.Fatalf("pprof index = %d, want 200", status)
	}
	if status := fetchDebugStatus(client, "http://"+bound+"/debug/state"); status != http.StatusOK {
		t.Fatalf("state status = %d, want 200", status)
	}
}

func fetchDebugJSON(client *http.Client, url string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var state map[string]any
	if err := json.Unmarshal(body, &state); err != nil {
		return nil, err
	}
	return state, nil
}

func fetchDebugStatus(client *http.Client, url string) int {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return 0
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode
}
