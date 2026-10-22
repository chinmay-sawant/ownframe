package fetch

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// Transport timeouts bound slow servers. Proxy stays nil so a hostile
// environment cannot route fetches elsewhere; call UseSystemProxy to
// opt into HTTP(S)_PROXY handling.
const (
	dialTimeout           = 10 * time.Second
	tlsTimeout            = 10 * time.Second
	headerTimeout         = 10 * time.Second
	expectContinueTimeout = 2 * time.Second
	idleConnTimeout       = 30 * time.Second
	dialKeepAlive         = 30 * time.Second
)

// pinnedTransport builds a fresh transport with a validating dialer.
// It never clones the shared global transport, so global tunables cannot
// change fetch behavior.
func pinnedTransport() *http.Transport {
	return &http.Transport{
		Proxy:                  nil,
		DialContext:            dialValidated,
		TLSHandshakeTimeout:    tlsTimeout,
		ResponseHeaderTimeout:  headerTimeout,
		MaxResponseHeaderBytes: MaxHeaderBytes,
		TLSClientConfig:        &tls.Config{MinVersion: tls.VersionTLS12},
		ExpectContinueTimeout:  expectContinueTimeout,
		IdleConnTimeout:        idleConnTimeout,
	}
}

// UseSystemProxy opts into HTTP(S)_PROXY for callers behind a proxy.
func UseSystemProxy() {
	if t, ok := client.Transport.(*http.Transport); ok {
		t.Proxy = http.ProxyFromEnvironment
	}
}

// dialer bounds connection setup; the request context still wins first.
func dialer() *net.Dialer {
	return &net.Dialer{Timeout: dialTimeout, KeepAlive: dialKeepAlive}
}
