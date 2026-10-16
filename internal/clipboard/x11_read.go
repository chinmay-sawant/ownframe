//go:build linux && !android

package clipboard

import "io"

func (x *xClient) loop() {
	defer recoverX(x, "x11 loop")

	for {
		var hdr [32]byte
		if _, err := io.ReadFull(x.c, hdr[:]); err != nil {
			x.fail()

			return
		}

		if hdr[0] == 0 {
			x.deliver(get16(hdr[2:]), xReply{err: ErrX11})

			continue
		}

		if hdr[0] == 1 {
			x.reply(hdr)

			continue
		}

		switch hdr[0] & 0x7f {
		case 30:
			ev := make([]byte, 32)
			copy(ev, hdr[:])
			go x.onRequest(ev)
		case 31:
			x.gotSel(get32(hdr[20:]))
		case 29:
			x.mu.Lock()
			x.owned = false
			x.mu.Unlock()
		}
	}
}

func (x *xClient) reply(hdr [32]byte) {
	n := get32(hdr[4:])
	if n > 1<<20 {
		x.fail()

		return
	}

	buf := make([]byte, 32+int(n)*4)
	copy(buf, hdr[:])

	if n > 0 {
		if _, err := io.ReadFull(x.c, buf[32:]); err != nil {
			x.fail()

			return
		}
	}

	x.deliver(get16(hdr[2:]), xReply{buf: buf})
}

func (x *xClient) gotSel(prop uint32) {
	x.mu.Lock()
	ch := x.note
	x.note = nil
	x.mu.Unlock()

	if ch != nil {
		ch <- prop
	}
}
