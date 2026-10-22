package fetch

import (
	"crypto/tls"
	"net/http"
	"testing"
)

func TestTransportLimits(t *testing.T) {
	t.Parallel()
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("custom transport")
	}
	if tr.Proxy != nil {
		t.Fatal("proxy must be nil by default")
	}
	if tr.TLSHandshakeTimeout == 0 || tr.ResponseHeaderTimeout == 0 {
		t.Fatal("timeouts missing")
	}
	if tr.MaxResponseHeaderBytes == 0 {
		t.Fatal("header cap missing")
	}
	if tr.TLSClientConfig == nil || tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Fatal("TLS 1.2 minimum missing")
	}
}
