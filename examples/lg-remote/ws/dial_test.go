package ws

import (
	"bufio"
	"context"
	"crypto/sha1" //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket test handshake, not a security hash.
	"encoding/base64"
	"io"
	"net"
	"strings"
	"testing"
)

func TestDialSkipsOrigin(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = ln.Close() }()
	go serveOne(t, ln)

	conn, err := Dial(context.Background(), "ws://"+ln.Addr().String()+"/ws")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = conn.Close() }()
	got, err := conn.ReadText()
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "hi" {
		t.Fatalf("got %q", got)
	}
}

func serveOne(t *testing.T, ln net.Listener) {
	t.Helper()

	conn, err := ln.Accept()
	if err != nil {
		return
	}

	defer func() { _ = conn.Close() }()
	br := bufio.NewReader(conn)
	req := ""
	var key string

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return
		}

		req += line
		if strings.HasPrefix(line, "Sec-WebSocket-Key:") {
			key = strings.TrimSpace(line[len("Sec-WebSocket-Key:"):])
		}

		if line == "\r\n" {
			break
		}
	}

	if strings.Contains(strings.ToLower(req), "origin:") {
		t.Errorf("origin header sent")
	}

	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")) //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket test handshake, not a security hash.
	accept := base64.StdEncoding.EncodeToString(sum[:])
	_, _ = io.WriteString(conn, "HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\nConnection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: "+accept+"\r\n\r\n")
	_, _ = conn.Write([]byte{0x81, 2, 'h', 'i'})
}
