package transfer

import (
	"context"
	"net/http"
	"time"
)

// HTTPOptions tunes the real transport. Zero values pick the defaults
// named on each field.
type HTTPOptions struct {
	// Client replaces the whole HTTP client. Nil builds one from the
	// fields below.
	Client *http.Client
	// ConnectTimeout bounds dial and TLS handshake. Zero means 10s.
	ConnectTimeout time.Duration
	// ResponseTimeout bounds the wait for response headers. Zero means 20s.
	ResponseTimeout time.Duration
	// StallTimeout fails a transfer that reads no byte for this long.
	// Zero means 30s.
	StallTimeout time.Duration
	// UserAgent is sent with every request. Empty sends none.
	UserAgent string
}

// HTTP is the real network transport.
type HTTP struct {
	client *http.Client
	ua     string
	stall  time.Duration
}

// NewHTTP builds the HTTP transport.
func NewHTTP(opts HTTPOptions) *HTTP {
	if opts.Client == nil {
		opts.Client = newClient(opts)
	}

	stall := opts.StallTimeout
	if stall <= 0 {
		stall = stallTimeout
	}

	return &HTTP{client: opts.Client, ua: opts.UserAgent, stall: stall}
}

// Download streams one URL to req.Partial, checking a resume with 206,
// Content-Range, and If-Range, then finalizes to req.Dest. A cancelled
// context leaves the partial file in place for a later resume.
func (h *HTTP) Download(ctx context.Context, req Request, report Reporter) (Outcome, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	prep, err := prepare(ctx, req, h.ua)
	if err != nil {
		return Outcome{}, err
	}

	resp, err := h.client.Do(prep.request)
	if err != nil {
		return Outcome{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	p, err := negotiate(resp, req, prep)
	if err != nil {
		return Outcome{}, err
	}

	if p.complete {
		return h.complete(req, p, Outcome{Resumed: prep.resumed})
	}

	return h.receive(ctx, cancel, req, resp, p, report)
}
