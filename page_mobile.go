package ownframe

import (
	"context"
	"time"
)

// PrepareMobile prepares this page for its native mobile viewport.
// Telegram tests drive the phone layout with it.
func (p *Page) PrepareMobile(ctx context.Context) error {
	return innerPage(p).PrepareMobile(ctx)
}

// NextFrameAt returns the deadline set by WantFrameAt and clears it.
// The departure-board example reads it from tests.
func (p *Page) NextFrameAt() time.Time {
	return innerPage(p).NextFrameAt()
}
