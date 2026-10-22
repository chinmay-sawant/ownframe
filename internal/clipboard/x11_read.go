//go:build linux && !android

package clipboard

import "io"

// reqGate bounds concurrent SelectionRequest handlers. The X loop stays
// responsive: excess requests run inline instead of spawning goroutines.
var reqGate = make(chan struct{}, 16)

func (x *xClient) loop() {
	defer recoverX(x, "x11 loop")

	for {
		var hdr [xEvSize]byte
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
		case xSelReq:
			ev := make([]byte, xEvSize)
			copy(ev, hdr[:])
			select {
			case reqGate <- struct{}{}:
				go func() {
					defer func() { <-reqGate }()
					x.onRequest(ev)
				}()
			default:
				x.onRequest(ev)
			}
		case xSelNotify:
			x.gotSel(get32(hdr[20:]))
		case xSelClear:
			x.mu.Lock()
			x.owned = false
			x.mu.Unlock()
		}
	}
}

func (x *xClient) reply(hdr [xEvSize]byte) {
	n := get32(hdr[4:])
	if n > xMaxWords {
		x.fail()

		return
	}

	buf := make([]byte, xEvSize+int(n)*4)
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
