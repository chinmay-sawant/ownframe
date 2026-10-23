package tv

import (
	"context"
	"errors"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/lg-remote/ws"
)

// Open dials the TV, pairs, and opens the pointer socket.
// Port 3000 is tried first. Port 3001 is the TLS fallback.
func Open(host, key string) (*Client, error) {
	conn, err := dialMain(host)
	if err != nil {
		return nil, err
	}

	c := &Client{
		main: conn,
		wait: map[string]chan Message{},
		Key:  ownframe.Secret(key),
	}
	goSafe("tv read", func() { c.read(conn) })

	ctx, cancel := context.WithTimeout(context.Background(), helloTimeout)
	_, _ = c.round(ctx, "hello", helloMsg())
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), sysTimeout)
	if m, err := c.round(ctx, "sys", sysMsg()); err == nil {
		c.Model = modelOf(m.Payload)
	}
	cancel()

	got, err := c.pair(key)
	if err != nil {
		c.Close()
		return nil, err
	}

	c.Key = ownframe.Secret(got)
	c.openPointer()
	c.readInfo()

	return c, nil
}

func dialMain(host string) (*ws.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mainTimeout)
	defer cancel()

	conn, err := ws.Dial(ctx, "ws://"+host+":3000/")
	if err == nil {
		return conn, nil
	}

	ctx, cancel = context.WithTimeout(context.Background(), tlsTimeout)
	defer cancel()

	conn, err = ws.Dial(ctx, "wss://"+host+":3001/")
	if err != nil {
		return nil, errors.New("the TV did not answer on the network")
	}

	return conn, nil
}
