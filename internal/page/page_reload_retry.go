package page

// retryPending retries the bytes a parse error kept. A later stat change
// drops the pending bytes and takes the new read.
func (p *Page) retryPending() (bool, error) {
	pend := p.watch.pending
	wp := p.watch.at(pend.kind)

	info, err := srcFiles.Stat(wp.path)
	if err != nil {
		return false, p.statNote(wp, err)
	}

	wp.statErr = ""
	if !sameStat(info, pend.info) {
		p.watch.pending = nil

		return p.pollSource(pend.kind)
	}

	return p.applySource(pend.kind, pend.data, info)
}
