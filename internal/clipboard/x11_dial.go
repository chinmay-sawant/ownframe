//go:build linux && !android

package clipboard

import (
	"io"
	"net"
	"os"
	"time"
)

func dialX() (*xClient, error) {
	host, num, screen, ok := displaySpec(os.Getenv("DISPLAY"))
	if !ok {
		return nil, ErrX11
	}

	conn, err := net.DialTimeout("unix", "/tmp/.X11-unix/X"+num, xWait)
	if err != nil {
		return nil, err
	}

	done := false
	defer func() {
		if !done {
			conn.Close()
		}
	}()

	_ = conn.SetDeadline(time.Now().Add(xWait))

	name, data := loadAuth(host, num)
	body, err := greet(conn, name, data)
	if err != nil {
		return nil, err
	}

	_ = conn.SetDeadline(time.Time{})

	base, mask, max, good := setupIDs(body)
	root, rootOK := rootAt(body, screen)
	if !good || !rootOK || mask == 0 {
		return nil, ErrX11
	}

	if max < xHdr24 {
		max = xHdr24
	}

	xc := &xClient{
		c:    conn,
		dead: make(chan struct{}),
		wait: map[uint16]chan xReply{},
		base: base,
		mask: mask,
		root: root,
		max:  max,
	}

	go xc.loop()
	done = true

	if err = xc.bootstrap(); err != nil {
		xc.fail()

		return nil, err
	}

	return xc, nil
}

func greet(conn net.Conn, name string, data []byte) ([]byte, error) {
	if _, err := conn.Write(setupReq([]byte(name), data)); err != nil {
		return nil, err
	}

	return readSetup(conn)
}

func readSetup(r io.Reader) ([]byte, error) {
	var hdr [8]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return nil, err
	}

	if hdr[0] != 1 {
		return nil, ErrX11
	}

	n := int(get16(hdr[6:]))
	if n < 8 || n > 1<<18 {
		return nil, ErrX11
	}

	buf := make([]byte, xHdr8+n*xWord)
	copy(buf, hdr[:])

	if _, err := io.ReadFull(r, buf[8:]); err != nil {
		return nil, err
	}

	return buf, nil
}
