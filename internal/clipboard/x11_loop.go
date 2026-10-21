//go:build linux && !android

package clipboard

import "time"

func (x *xClient) fail() {
	x.once.Do(func() {
		x.mu.Lock()
		close(x.dead)

		for _, ch := range x.wait {
			ch <- xReply{err: ErrX11}
		}

		x.wait = nil
		x.mu.Unlock()
		_ = x.c.Close()
	})
}

func (x *xClient) send(req []byte) error {
	x.mu.Lock()
	x.seq++
	_, err := x.c.Write(req)
	x.mu.Unlock()

	if err != nil {
		x.fail()
	}

	return err
}

func (x *xClient) round(req []byte) ([]byte, error) {
	ch := make(chan xReply, 1)

	x.mu.Lock()
	x.seq++
	seq := x.seq
	x.wait[seq] = ch
	_, err := x.c.Write(req)
	x.mu.Unlock()

	if err != nil {
		x.fail()

		return nil, err
	}

	select {
	case rep := <-ch:
		if rep.err != nil {
			return nil, rep.err
		}

		return rep.buf, nil
	case <-time.After(xWait):
		x.mu.Lock()
		delete(x.wait, seq)
		x.mu.Unlock()

		return nil, ErrX11
	case <-x.dead:
		return nil, ErrX11
	}
}

func (x *xClient) deliver(seq uint16, rep xReply) {
	x.mu.Lock()
	ch := x.wait[seq]
	delete(x.wait, seq)
	x.mu.Unlock()

	if ch != nil {
		ch <- rep
	}
}
