package ownframe

import (
	"context"
	"time"
)

// PrepareMobile prepares this page for its native mobile viewport.
func (t Testing) PrepareMobile(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PrepareMobile(ctx)
}

// NextFrameAt returns the deadline set by WantFrameAt and clears it.
func (t Testing) NextFrameAt() time.Time {
	p := innerPage(t.page)
	if p == nil {
		return time.Time{}
	}

	return p.NextFrameAt()
}
