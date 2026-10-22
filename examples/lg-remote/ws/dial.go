package ws

import (
	"bufio"
	"context"
	"net"
	"net/url"
	"strings"
	"time"
)

// Conn is one upgraded WebSocket.
type Conn struct {
	c   net.Conn
	r   *bufio.Reader
	wmu chan struct{}
}

// Dial upgrades raw to a WebSocket. wss verifies the TV certificate
// with the system roots and TLS 1.2 or newer.
func Dial(ctx context.Context, raw string) (*Conn, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	host := u.Host
	if !strings.Contains(host, ":") {
		if u.Scheme == "wss" {
			host += ":443"
		} else {
			host += ":80"
		}
	}

	nc, err := dial(ctx, u.Scheme, u.Hostname(), host)
	if err != nil {
		return nil, err
	}

	key := nonce()
	path := u.RequestURI()
	if path == "" {
		path = "/"
	}

	req := "GET " + path + " HTTP/1.1\r\n" +
		"Host: " + u.Host + "\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Key: " + key + "\r\n" +
		"Sec-WebSocket-Version: 13\r\n\r\n"
	if _, err = nc.Write([]byte(req)); err != nil {
		_ = nc.Close()
		return nil, err
	}

	br := bufio.NewReader(nc)
	if err = accept(br, key); err != nil {
		_ = nc.Close()
		return nil, err
	}

	_ = nc.SetDeadline(time.Time{})

	return &Conn{c: nc, r: br, wmu: make(chan struct{}, 1)}, nil
}
