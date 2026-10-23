package draw

import (
	"context"
	"runtime"
	"time"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// RedrawSurface is the page state one redraw reads and the actions it takes.
// The page adapter delegates each method. This package does not import page.
type RedrawSurface interface {
	// AfterBusy reports that the after-redraw callback is on the stack.
	AfterBusy() bool
	// AfterAsk records a redraw requested from inside that callback.
	AfterAsk()
	// QueueBatch records this draw while a batch is open.
	// It reports that the caller should return without drawing.
	QueueBatch(skipSame bool) bool
	// BeginSample samples the allocator when perf is on and a memory sample is due.
	// The bool tells the caller to clear opt-in counters before starting the clock.
	BeginSample() (Sample, bool)
	// ClearPerf zeroes opt-in counters before the clock starts.
	ClearPerf()
	// CountRedraw counts this redraw.
	CountRedraw()
	// FinishRedraw records the redraw duration and, when memOK is true, the
	// bytes allocated since that sample. A redraw that did not sample leaves
	// the previous allocation count.
	FinishRedraw(track bool, memOK bool, mem runtime.MemStats, start time.Time)
	// ExecTemplate runs the template and returns the executed source.
	ExecTemplate(track bool) (string, error)
	// SyncForm stores the form-synced source and returns that same string.
	SyncForm(executed string) string
	// RenderState is the pointer, theme, and image input for this draw.
	RenderState() render.State
	// Unchanged reports that state matches the last committed frame.
	Unchanged(state render.State) bool
	// DropPNG forgets a cached picture before a new commit.
	DropPNG()
	// StyleStaged styles the synced source.
	StyleStaged(ctx context.Context, track bool, source string, state render.State) (*css.Document, error)
	// DrawStaged commits the frame.
	// executed is the template output from before SyncForm, not the synced source.
	DrawStaged(
		ctx context.Context,
		track bool,
		styled *css.Document,
		state render.State,
		executed string,
	) error
	// AfterRedrawDraw serves one redraw the callback asked for.
	AfterRedrawDraw(ctx context.Context) error
}

// Run draws one frame.
// A nil or done context returns before any surface call.
// A callback that is already drawing records the request and returns,
// so the commit does not start another draw inside itself.
// An open batch records the draw and returns.
// The clock starts after a perf clear. The template runs, then the form
// is synced into the source. When skipSame is set, an unchanged frame
// returns after that sync. Otherwise the page drops its PNG, styles,
// draws, and serves one follow-up the callback asked for.
// FinishRedraw runs on the way out once the clock has started, including
// when a later step returns an error.
func Run(ctx context.Context, s RedrawSurface, skipSame bool) error {
	if err := useContext(ctx); err != nil {
		return err
	}
	if s.AfterBusy() {
		// The after-redraw callback asked for a draw. The commit that runs
		// the callback must not start another one inside it, so the request
		// waits for the commit to finish.
		s.AfterAsk()

		return nil
	}
	if s.QueueBatch(skipSame) {
		return nil
	}

	sample, clear := s.BeginSample()
	if clear {
		s.ClearPerf()
	}
	sample.StartClock()
	s.CountRedraw()

	defer s.FinishRedraw(sample.Track, sample.MemOK, sample.Mem, sample.Start)

	executed, err := s.ExecTemplate(sample.Track)
	if err != nil {
		return err
	}

	source := s.SyncForm(executed)
	state := s.RenderState()
	if skipSame && s.Unchanged(state) {
		return nil
	}

	s.DropPNG()
	styled, err := s.StyleStaged(ctx, sample.Track, source, state)
	if err != nil {
		return err
	}
	if err := s.DrawStaged(ctx, sample.Track, styled, state, executed); err != nil {
		return err
	}

	return s.AfterRedrawDraw(ctx)
}

func useContext(ctx context.Context) error {
	if ctx == nil {
		return fetch.ErrNilContext
	}

	return ctx.Err()
}
