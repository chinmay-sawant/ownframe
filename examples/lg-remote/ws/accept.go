package ws

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1" //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket handshake, not a security hash.
	"encoding/base64"
	"strings"
)

func nonce() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(b[:])
}

func accept(r *bufio.Reader, key string) error {
	status, err := r.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.Contains(status, "101") {
		return ErrHandshake
	}

	sum := sha1.Sum([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")) //nolint:gosec // reason: SHA1 is required by RFC 6455 for the WebSocket handshake, not a security hash.
	want := base64.StdEncoding.EncodeToString(sum[:])
	ok := false

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}

		if line == "\r\n" {
			break
		}

		if !strings.HasPrefix(strings.ToLower(line), "sec-websocket-accept:") {
			continue
		}

		got := strings.TrimSpace(line[len("sec-websocket-accept:"):])
		ok = got == want
	}

	if !ok {
		return ErrHandshake
	}

	return nil
}
