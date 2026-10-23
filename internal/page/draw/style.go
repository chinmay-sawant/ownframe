package draw

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// StyleDocument calls Style and, when track is true, stores the elapsed time in styleTime.
// A nil styleTime skips the store. The Styled result is Style's result, including a Style error.
func StyleDocument(
	ctx context.Context,
	track bool,
	styleTime *time.Duration,
	cache *render.Cache,
	source string,
	view Viewport,
	state render.State,
) (Styled, error) {
	start := Clock(track)
	styled, err := Style(ctx, cache, source, view, state)
	if styleTime != nil {
		Note(track, styleTime, start)
	}

	return styled, err
}
