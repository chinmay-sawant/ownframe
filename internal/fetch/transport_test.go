package fetch

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"testing"
)

func TestTransportLimits(t *testing.T) {
	t.Parallel()
	tr, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("custom transport")
	}
	if tr.Proxy != nil {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
		if err != nil {
			t.Fatal(err)
		}
		u, err := tr.Proxy(req)
		if err != nil || u != nil {
			t.Fatalf("proxy must be disabled by default: u=%v err=%v", u, err)
		}
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

func TestUseSystemProxyRaceFree(t *testing.T) {
	prev := useSystemProxy.Load()
	t.Cleanup(func() { useSystemProxy.Store(prev) })

	tr, ok := client.Transport.(*http.Transport)
	if !ok || tr.Proxy == nil {
		t.Fatal("expected transport with proxyFunc")
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				UseSystemProxy()
			}
		}()
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_, _ = tr.Proxy(req)
			}
		}()
	}
	wg.Wait()
}
