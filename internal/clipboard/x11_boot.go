//go:build linux && !android

package clipboard

func (x *xClient) bootstrap() error {
	x.win = x.newID()

	if err := x.send(winReq(x.win, x.root)); err != nil {
		return err
	}

	var err error
	if x.clip, err = x.atom("CLIPBOARD"); err != nil {
		return err
	}

	if x.utf8, err = x.atom("UTF8_STRING"); err != nil {
		return err
	}

	if x.targets, err = x.atom("TARGETS"); err != nil {
		return err
	}

	if x.incr, err = x.atom("INCR"); err != nil {
		return err
	}

	x.prop, err = x.atom("OWNFRAME_CLIP")

	return err
}

func (x *xClient) atom(name string) (uint32, error) {
	rep, err := x.round(internReq(name))
	if err != nil || len(rep) < 12 || rep[0] != 1 {
		return 0, ErrX11
	}

	id := get32(rep[8:])
	if id == 0 {
		return 0, ErrX11
	}

	return id, nil
}

func (x *xClient) newID() uint32 {
	x.mu.Lock()
	defer x.mu.Unlock()

	inc := x.mask & -x.mask
	if inc == 0 {
		inc = 1
	}

	x.xid += inc

	return x.xid | x.base
}
