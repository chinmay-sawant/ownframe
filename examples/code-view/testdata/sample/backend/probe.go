package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
	"time"
)

// Result is one probe outcome. Err is empty when the request completed.
type Result struct {
	Service  string
	At       time.Time
	State    string
	Status   int
	Total    time.Duration
	TTFB     time.Duration
	Err      string
	Attempts int
}

// States a probe can report.
const (
	StateUp   = "up"
	StateSlow = "slow"
	StateDown = "down"
)

const (
	defaultAttempts = 3
	retryBackoff    = 150 * time.Millisecond
	bodyLimit       = 64 << 10
	probeUserAgent  = "beacon/1.4 (+https://example.com/beacon)"
)

// Prober runs HTTP checks with a bounded client and a latency budget.
type Prober struct {
	client  *http.Client
	timeout time.Duration
	budget  time.Duration
	log     *slog.Logger
}

// NewProber builds a Prober. A zero timeout becomes four seconds.
func NewProber(timeout, budget time.Duration, log *slog.Logger) *Prober {
	if timeout <= 0 {
		timeout = 4 * time.Second
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 3 * time.Second}).DialContext,
		MaxIdleConns:          32,
		MaxIdleConnsPerHost:   4,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ResponseHeaderTimeout: timeout,
		ForceAttemptHTTP2:     true,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
	}

	return &Prober{
		client:  &http.Client{Transport: transport, Timeout: timeout},
		timeout: timeout,
		budget:  budget,
		log:     log,
	}
}

// Probe checks one service, retrying transport errors and 5xx answers.
func (p *Prober) Probe(ctx context.Context, svc Service) Result {
	result := Result{Service: svc.Name, At: time.Now().UTC(), State: StateDown}

	var last error

	for attempt := 1; attempt <= defaultAttempts; attempt++ {
		result.Attempts = attempt

		attemptCtx, cancel := context.WithTimeout(ctx, p.timeout)
		start := time.Now()
		status, ttfb, err := p.once(attemptCtx, svc.URL)
		cancel()

		result.Total = time.Since(start)
		result.TTFB = ttfb
		result.Status = status

		if err == nil && status < http.StatusInternalServerError {
			result.State = StateUp
			if result.Total > p.budget {
				result.State = StateSlow
			}

			return result
		}

		if err != nil {
			last = err
		} else {
			last = fmt.Errorf("status %d", status)
		}

		if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
			break
		}

		if attempt < defaultAttempts {
			select {
			case <-ctx.Done():
				result.Err = ctx.Err().Error()

				return result
			case <-time.After(retryBackoff * time.Duration(attempt)):
			}
		}
	}

	if last != nil {
		result.Err = last.Error()
	}

	p.log.Debug("probe", "service", svc.Name, "state", result.State, "err", result.Err, "attempts", result.Attempts)

	return result
}

// once sends one request and returns the status, the time to first byte, and
// a transport error.
func (p *Prober) once(ctx context.Context, url string) (int, time.Duration, error) {
	var (
		firstByte time.Time
		start     = time.Now()
	)

	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			if firstByte.IsZero() {
				firstByte = time.Now()
			}
		},
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("User-Agent", probeUserAgent)
	req.Header.Set("Accept", "text/html,application/json;q=0.9,*/*;q=0.8")

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	// Drain a bounded prefix so the connection can be reused. A broken body
	// is a failed check, not an ignored one.
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, bodyLimit)); err != nil && !errors.Is(err, io.EOF) {
		return resp.StatusCode, firstByte.Sub(start), fmt.Errorf("read body: %w", err)
	}

	return resp.StatusCode, firstByte.Sub(start), nil
}
