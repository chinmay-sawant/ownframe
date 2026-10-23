package draw

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/fetch"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

var errRedraw = errors.New("redraw")

// TestRedrawSequenceOrder checks decision order on a fake surface.
// The page redraw tests still call Page.Redraw.
func TestRedrawSequenceOrder(t *testing.T) {
	t.Parallel()

	const (
		stopped = "AfterBusy QueueBatch BeginSample CountRedraw ExecTemplate"
		drew    = " DropPNG StyleStaged DrawStaged AfterRedrawDraw"
		full    = stopped + " SyncForm RenderState" + drew + " FinishRedraw"
	)

	cases := []struct {
		name     string
		f        fake
		skipSame bool
		want     string
		wantErr  error
	}{
		{
			name: "busy asks and does not draw",
			f:    fake{busy: true, queued: true},
			want: "AfterBusy AfterAsk",
		},
		{
			name: "open batch returns before the clock",
			f:    fake{queued: true},
			want: "AfterBusy QueueBatch",
		},
		{
			name:     "unchanged skips the draw",
			f:        fake{unchanged: true},
			skipSame: true,
			want:     stopped + " SyncForm RenderState Unchanged FinishRedraw",
		},
		{
			name:     "perf off clears counters before the template",
			f:        fake{clear: true, unchanged: true},
			skipSame: true,
			want: "AfterBusy QueueBatch BeginSample ClearPerf CountRedraw " +
				"ExecTemplate SyncForm RenderState Unchanged FinishRedraw",
		},
		{
			name: "changed frame styles and draws",
			want: full,
		},
		{
			name:     "skipSame still draws when the frame changed",
			skipSame: true,
			want:     stopped + " SyncForm RenderState Unchanged" + drew + " FinishRedraw",
		},
		{
			name:    "template error stops before the form sync",
			f:       fake{execErr: errRedraw},
			want:    stopped + " FinishRedraw",
			wantErr: errRedraw,
		},
		{
			name:    "style error stops before the draw",
			f:       fake{styleErr: errRedraw},
			want:    stopped + " SyncForm RenderState DropPNG StyleStaged FinishRedraw",
			wantErr: errRedraw,
		},
		{
			name:    "draw error stops before the follow-up",
			f:       fake{drawErr: errRedraw},
			want:    stopped + " SyncForm RenderState DropPNG StyleStaged DrawStaged FinishRedraw",
			wantErr: errRedraw,
		},
		{
			name:    "follow-up error is returned",
			f:       fake{afterErr: errRedraw},
			want:    full,
			wantErr: errRedraw,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.f
			err := Run(context.Background(), &f, tc.skipSame)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}

			got := strings.Join(f.calls, " ")
			if got != tc.want {
				t.Fatalf("calls = %s", got)
			}
		})
	}
}

func TestRedrawRejectsNilContext(t *testing.T) {
	t.Parallel()

	f := &fake{}
	//nolint:staticcheck // reason: intentionally passes nil to verify ErrNilContext.
	err := Run(nil, f, false)
	if !errors.Is(err, fetch.ErrNilContext) {
		t.Fatalf("err = %v", err)
	}

	if len(f.calls) != 0 {
		t.Fatalf("calls = %v", f.calls)
	}
}

func TestRedrawStopsOnCanceledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	f := &fake{}
	err := Run(ctx, f, false)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}

	if len(f.calls) != 0 {
		t.Fatalf("calls = %v", f.calls)
	}
}

func TestRedrawPassesStepInputs(t *testing.T) {
	t.Parallel()

	styled := &css.Document{}
	state := render.State{Hover: "h", Active: "a"}
	f := &fake{
		sample:    Sample{Track: true},
		executed:  "tmpl",
		synced:    "synced",
		state:     state,
		styled:    styled,
		unchanged: false,
	}
	ctx := context.Background()
	if err := Run(ctx, f, true); err != nil {
		t.Fatal(err)
	}

	if !f.queueSkip {
		t.Fatal("skipSame")
	}

	if strings.Contains(strings.Join(f.calls, " "), "ClearPerf") {
		t.Fatal("cleared")
	}

	if !f.execTrack {
		t.Fatal("template track")
	}

	if f.syncedIn != "tmpl" {
		t.Fatalf("synced from %q", f.syncedIn)
	}

	if !f.unchangedSeen {
		t.Fatal("unchanged skipped")
	}

	if f.unchangedState.Hover != "h" || f.unchangedState.Active != "a" {
		t.Fatalf("unchanged %+v", f.unchangedState)
	}

	if f.styleCtx != ctx || !f.styleTrack || f.styleSource != "synced" {
		t.Fatalf("style track %v source %q", f.styleTrack, f.styleSource)
	}

	if f.styleState.Hover != "h" {
		t.Fatalf("style state %+v", f.styleState)
	}

	if f.drawCtx != ctx || f.drawStyled != styled || f.drawExecuted != "tmpl" {
		t.Fatalf("draw executed %q", f.drawExecuted)
	}

	if !f.drawTrack || f.drawState.Active != "a" {
		t.Fatal("draw track or state")
	}

	if f.afterCtx != ctx {
		t.Fatal("after ctx")
	}

	if !f.finishTrack || f.finishStart.IsZero() {
		t.Fatal("finish")
	}
}

type fake struct {
	calls     []string
	busy      bool
	queued    bool
	queueSkip bool
	clear     bool
	sample    Sample
	executed  string
	execErr   error
	execTrack bool
	synced    string
	syncedIn  string
	state     render.State
	unchanged bool
	// unchangedSeen is set when Unchanged runs, so a false result is still visible.
	unchangedSeen  bool
	unchangedState render.State
	styled         *css.Document
	styleErr       error
	styleCtx       context.Context
	styleTrack     bool
	styleSource    string
	styleState     render.State
	drawErr        error
	drawCtx        context.Context
	drawTrack      bool
	drawStyled     *css.Document
	drawState      render.State
	drawExecuted   string
	afterErr       error
	afterCtx       context.Context
	finishTrack    bool
	finishStart    time.Time
}

func (f *fake) note(step string) {
	f.calls = append(f.calls, step)
}

func (f *fake) AfterBusy() bool {
	f.note("AfterBusy")

	return f.busy
}

func (f *fake) AfterAsk() { f.note("AfterAsk") }

func (f *fake) QueueBatch(skipSame bool) bool {
	f.note("QueueBatch")
	f.queueSkip = skipSame

	return f.queued
}

func (f *fake) BeginSample() (Sample, bool) {
	f.note("BeginSample")

	return f.sample, f.clear
}

func (f *fake) ClearPerf() { f.note("ClearPerf") }

func (f *fake) CountRedraw() { f.note("CountRedraw") }

func (f *fake) FinishRedraw(track bool, _ bool, _ runtime.MemStats, start time.Time) {
	f.note("FinishRedraw")
	f.finishTrack = track
	f.finishStart = start
}

func (f *fake) ExecTemplate(track bool) (string, error) {
	f.note("ExecTemplate")
	f.execTrack = track

	return f.executed, f.execErr
}

func (f *fake) SyncForm(executed string) string {
	f.note("SyncForm")
	f.syncedIn = executed

	return f.synced
}

func (f *fake) RenderState() render.State {
	f.note("RenderState")

	return f.state
}

func (f *fake) Unchanged(state render.State) bool {
	f.note("Unchanged")
	f.unchangedSeen = true
	f.unchangedState = state

	return f.unchanged
}

func (f *fake) DropPNG() { f.note("DropPNG") }

func (f *fake) StyleStaged(
	ctx context.Context,
	track bool,
	source string,
	state render.State,
) (*css.Document, error) {
	f.note("StyleStaged")
	f.styleCtx = ctx
	f.styleTrack = track
	f.styleSource = source
	f.styleState = state

	return f.styled, f.styleErr
}

func (f *fake) DrawStaged(
	ctx context.Context,
	track bool,
	styled *css.Document,
	state render.State,
	executed string,
) error {
	f.note("DrawStaged")
	f.drawCtx = ctx
	f.drawTrack = track
	f.drawStyled = styled
	f.drawState = state
	f.drawExecuted = executed

	return f.drawErr
}

func (f *fake) AfterRedrawDraw(ctx context.Context) error {
	f.note("AfterRedrawDraw")
	f.afterCtx = ctx

	return f.afterErr
}

var _ RedrawSurface = (*fake)(nil)
