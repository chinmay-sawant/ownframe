package tv

import (
	"crypto/sha1" //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket test handshake, not a security hash.
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func socketURL(t *testing.T, reply func([]byte) []byte) string {
	t.Helper()
	var mu sync.Mutex
	clients := []net.Conn{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			return
		}
		conn, buf, err := hj.Hijack()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		mu.Lock()
		clients = append(clients, conn)
		mu.Unlock()
		sum := sha1.Sum([]byte(r.Header.Get("Sec-WebSocket-Key") + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")) //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket test handshake, not a security hash.
		_, _ = buf.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + base64.StdEncoding.EncodeToString(sum[:]) + "\r\n\r\n")
		_ = buf.Flush()
		for {
			b, err := readSocketFrame(buf)
			if err != nil {
				return
			}
			if out := reply(b); out != nil {
				writeSocketFrame(conn, out)
			}
		}
	}))
	t.Cleanup(func() {
		s.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, c := range clients {
			_ = c.Close()
		}
	})
	return "ws" + strings.TrimPrefix(s.URL, "http")
}
