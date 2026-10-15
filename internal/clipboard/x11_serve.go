//go:build linux && !android

package clipboard

func (x *xClient) onRequest(ev []byte) {
	defer recoverX(x, "x11 request")

	when := get32(ev[4:])
	who := get32(ev[12:])
	sel := get32(ev[16:])
	target := get32(ev[20:])
	prop := get32(ev[24:])

	if who == 0 {
		return
	}

	if prop == 0 {
		prop = target
	}

	x.mu.Lock()
	text := x.text
	clip, utf8, targets := x.clip, x.utf8, x.targets
	limit := x.max
	x.mu.Unlock()

	out := prop
	var err error

	switch {
	case sel != clip:
		out = 0
	case target == targets:
		vals := []uint32{targets, utf8, 31}
		err = x.send(prop32(who, prop, 4, vals))
	case target == utf8 || target == 31:
		if 24+len(text) > limit {
			out = 0
		} else {
			raw := []byte(text)
			err = x.send(prop8(who, prop, target, raw))
		}
	default:
		out = 0
	}

	if err != nil {
		out = 0
	}

	_ = x.send(sendEv(who, selEvent(when, who, sel, target, out)))
}
