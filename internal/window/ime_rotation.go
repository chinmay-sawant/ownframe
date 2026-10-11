package window

import "time"

const imeRotationWait = 250 * time.Millisecond

type imeRotationState struct {
	landscape bool
	known     bool
	rotatedAt time.Time
	pendingAt time.Time
	rotated   bool
}

func (r *imeRotationState) layout(w, h int, now time.Time) {
	landscape := w > h
	if r.known && r.landscape != landscape {
		r.rotatedAt = now
		r.rotated = !r.pendingAt.IsZero()
	}
	r.landscape, r.known = landscape, true
}

func (r *imeRotationState) endByUser(now time.Time) bool {
	if !r.rotatedAt.IsZero() && now.Sub(r.rotatedAt) < imeRotationWait {
		return true
	}
	r.pendingAt, r.rotated = now, false
	return false
}

func (r *imeRotationState) resolve(now time.Time) (wait, preserve bool) {
	if r.rotated {
		r.pendingAt, r.rotated = time.Time{}, false
		return false, true
	}
	if now.Sub(r.pendingAt) < imeRotationWait {
		return true, false
	}
	r.pendingAt = time.Time{}
	return false, false
}
