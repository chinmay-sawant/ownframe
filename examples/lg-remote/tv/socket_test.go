package tv

import (
	"context"
	"crypto/sha1" //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket test handshake, not a security hash.
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/lg-remote/ws"
)

func socketURL(t *testing.T, reply func([]byte) []byte) string {
	t.Helper()
	var mu sync.Mutex
	clients := []net.Conn{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, buf, err := w.(http.Hijacker).Hijack()
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

func testClient(t *testing.T, payload func(map[string]any) map[string]any) *Client {
	t.Helper()
	url := socketURL(t, func(b []byte) []byte {
		var req map[string]any
		if json.Unmarshal(b, &req) != nil {
			return nil
		}
		out, _ := json.Marshal(map[string]any{"id": req["id"], "type": "response", "payload": payload(req)})
		return out
	})
	main, err := ws.Dial(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	c := &Client{main: main, wait: map[string]chan Message{}}
	go c.read(main)
	t.Cleanup(c.Close)
	return c
}
