package ws

import (
	"context"
	"crypto/tls"
	"net"
)

func dial(ctx context.Context, scheme, name, host string) (net.Conn, error) {
	d := &net.Dialer{Timeout: dialTimeout}
	if scheme != "wss" {
		return d.DialContext(ctx, "tcp", host)
	}

	td := tls.Dialer{
		NetDialer: d,
		Config: &tls.Config{
			ServerName: name,
			MinVersion: tls.VersionTLS12,
		},
	}

	return td.DialContext(ctx, "tcp", host)
}
