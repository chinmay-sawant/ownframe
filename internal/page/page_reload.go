package page

import (
	"context"
	"crypto/sha256"
)

// PollReload stats and rereads the watched files. It reports whether the
// page changed, and returns the first error seen. A read that does not parse
// keeps the last good template or theme and is retried on the next poll.
func (p *Page) PollReload(ctx context.Context) (bool, error) {
	if err := useContext(ctx); err != nil {
		return false, err
	}

	if !p.watch.on() {
		return false, nil
	}

	applied := false
	var first error

	if p.watch.pending != nil {
		applied, first = p.retryPending()
	} else {
		for _, kind := range []watchKind{watchHTML, watchTheme} {
			ok, err := p.pollSource(kind)
			if ok {
				applied = true
			}

			if err != nil && first == nil {
				first = err
			}
		}
	}

	if !applied {
		p.noteReloadError(first)

		return false, first
	}

	if err := p.Redraw(ctx); err != nil {
		p.noteReloadError(err)

		return false, err
	}

	p.noteReload(first)
	p.resolveStates()

	return true, first
}

// pollSource rereads one path when its stat changed and its hash differs.
func (p *Page) pollSource(kind watchKind) (bool, error) {
	wp := p.watch.at(kind)
	if wp.path == "" {
		return false, nil
	}

	info, err := srcFiles.Stat(wp.path)
	if err != nil {
		return false, p.statNote(wp, err)
	}

	wp.statErr = ""
	if sameStat(info, wp.info) {
		return false, nil
	}

	if info.Size() > maxSourceSize {
		err := sourceTooLarge(wp.path, info.Size())
		return false, p.statNote(wp, err)
	}

	data, err := srcFiles.ReadFile(wp.path)
	if err != nil {
		return false, p.statNote(wp, err)
	}

	if int64(len(data)) > maxSourceSize {
		err := sourceTooLarge(wp.path, int64(len(data)))
		return false, p.statNote(wp, err)
	}

	if wp.info != nil && wp.sum == sha256.Sum256(data) {
		wp.info = info

		return false, nil
	}

	return p.applySource(kind, data, info)
}
