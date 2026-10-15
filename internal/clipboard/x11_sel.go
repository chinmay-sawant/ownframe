//go:build linux && !android

package clipboard

import "time"

func xWrite(text string) {
	x, err := xOpen()
	if err != nil {
		return
	}

	if err = x.claim(text); err != nil {
		x.fail()
	}
}

func xRead() (string, bool) {
	x, err := xOpen()
	if err != nil {
		return "", false
	}

	x.mu.Lock()
	owned, text := x.owned, x.text
	x.mu.Unlock()

	if owned {
		return text, true
	}

	return x.pull()
}

func (x *xClient) claim(text string) error {
	x.mu.Lock()
	x.text = text
	x.mu.Unlock()

	if err := x.send(ownReq(x.win, x.clip)); err != nil {
		return err
	}

	rep, err := x.round(ownerReq(x.clip))
	if err != nil || len(rep) < 12 || rep[0] != 1 {
		return ErrX11
	}

	x.mu.Lock()
	x.owned = get32(rep[8:]) == x.win
	x.mu.Unlock()

	return nil
}

func (x *xClient) pull() (string, bool) {
	ch := make(chan uint32, 1)

	x.mu.Lock()
	x.note = ch
	x.mu.Unlock()

	err := x.send(convReq(x.win, x.clip, x.utf8, x.prop))
	if err != nil {
		return "", false
	}

	var prop uint32

	select {
	case prop = <-ch:
	case <-time.After(xWait):
		x.mu.Lock()

		if x.note == ch {
			x.note = nil
		}

		x.mu.Unlock()

		return "", false
	case <-x.dead:
		return "", false
	}

	if prop == 0 {
		return "", false
	}

	rep, err := x.round(getReq(x.win, prop))
	if err != nil || len(rep) < 32 || rep[0] != 1 || rep[1] != 8 {
		return "", false
	}

	if get32(rep[8:]) == x.incr {
		return "", false
	}

	n := int(get32(rep[16:]))
	if n < 0 || 32+n > len(rep) {
		return "", false
	}

	return string(rep[32 : 32+n]), true
}
