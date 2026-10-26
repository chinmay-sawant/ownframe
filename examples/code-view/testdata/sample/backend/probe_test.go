package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeStates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		sleep  time.Duration
		want   string
	}{
		{name: "healthy", status: http.StatusOK, want: StateUp},
		{name: "redirect", status: http.StatusFound, want: StateUp},
		{name: "slow", status: http.StatusOK, sleep: 90 * time.Millisecond, want: StateSlow},
		{name: "server error", status: http.StatusBadGateway, want: StateDown},
		{name: "client error", status: http.StatusNotFound, want: StateDown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				if tc.sleep > 0 {
					time.Sleep(tc.sleep)
				}

				w.WriteHeader(tc.status)
			}))
			defer server.Close()

			prober := NewProber(2*time.Second, 50*time.Millisecond, slog.New(slog.DiscardHandler))
			result := prober.Probe(context.Background(), Service{Name: tc.name, URL: server.URL})

			if result.State != tc.want {
				t.Fatalf("state = %s, want %s (err %q)", result.State, tc.want, result.Err)
			}
			if result.Status != tc.status {
				t.Fatalf("status = %d, want %d", result.Status, tc.status)
			}
		})
	}
}

func TestProbeRetriesServerErrors(t *testing.T) {
	t.Parallel()

	var calls int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	prober := NewProber(time.Second, time.Second, slog.New(slog.DiscardHandler))
	result := prober.Probe(context.Background(), Service{Name: "flaky", URL: server.URL})

	if result.Attempts != defaultAttempts {
		t.Fatalf("attempts = %d, want %d", result.Attempts, defaultAttempts)
	}
	if calls != defaultAttempts {
		t.Fatalf("calls = %d, want %d", calls, defaultAttempts)
	}
	if result.State != StateDown {
		t.Fatalf("state = %s, want down", result.State)
	}
}

func TestProbeHonoursContext(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	prober := NewProber(time.Second, time.Second, slog.New(slog.DiscardHandler))
	result := prober.Probe(ctx, Service{Name: "slow", URL: server.URL})

	if result.State != StateDown {
		t.Fatalf("state = %s, want down", result.State)
	}
	if result.Err == "" {
		t.Fatal("expected an error")
	}
}
