//go:build !js && !android && !ios

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/desktop-cat/cat"
)

// mediaTimeout bounds post I/O; mediaHeaderBytes caps response headers.
const (
	mediaTimeout     = 5 * time.Second
	mediaHeaderBytes = 1 << 16
)

// mediaClient never uses the shared global client: it has its own timeouts
// and a TLS 1.2 minimum, so global client state cannot change posts.
var mediaClient = &http.Client{
	Timeout: mediaTimeout,
	Transport: &http.Transport{
		Proxy:                  nil,
		TLSHandshakeTimeout:    mediaTimeout,
		ResponseHeaderTimeout:  mediaTimeout,
		MaxResponseHeaderBytes: mediaHeaderBytes,
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12},
	},
}

func postMedia(ctx context.Context, endpoint string, n cat.Notification) error {
	ctx, cancel := context.WithTimeout(ctx, postTimeout)
	defer cancel()
	data, err := json.Marshal(n)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if u, err := url.Parse(endpoint); err == nil && u.Scheme != "" && u.Host != "" {
		request.Header.Set("Origin", u.Scheme+"://"+u.Host)
	}
	response, err := mediaClient.Do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("media notification returned %s", response.Status)
	}
	return nil
}
