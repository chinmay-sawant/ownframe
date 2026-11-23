package ownframe_test

import (
	"context"
	"errors"
	"image"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestNilAndZeroPageMethodsDoNotPanic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	cases := []struct {
		name string
		p    *ownframe.Page
	}{
		{"nil", nil},
		{"zero", &ownframe.Page{}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			p := tc.p

			p.SetData("x")
			p.Handle(ownframe.Handlers{})
			p.AddHandlers(ownframe.Handlers{})
			if err := p.SetTheme("body{}"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("SetTheme err = %v, want ErrNilPage", err)
			}
			p.SetTick(nil)
			p.SetImage("logo", []byte{1})
			p.SetAudio([]byte{1}, "audio/wav")
			p.SetAudioPlayback(true, 0.5)
			if err := p.Redraw(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Redraw err = %v, want ErrNilPage", err)
			}
			if err := p.RedrawIfChanged(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("RedrawIfChanged err = %v, want ErrNilPage", err)
			}
			if err := p.Click(ctx, 1, 2); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Click err = %v, want ErrNilPage", err)
			}
			p.WantFrameAt(time.Now())
			p.WantFrameIn(time.Millisecond)
			if p.Title() != "" || p.Boxes() != nil || p.HTML() != "" {
				t.Fatal("expected zero Title/Boxes/HTML")
			}
			if w, h := p.Size(); w != 0 || h != 0 {
				t.Fatalf("Size = %d,%d", w, h)
			}
			if w, h := p.MinSize(); w != 0 || h != 0 {
				t.Fatalf("MinSize = %d,%d", w, h)
			}
			if w, h := p.MaxSize(); w != 0 || h != 0 {
				t.Fatalf("MaxSize = %d,%d", w, h)
			}
			if w, h := p.Clamp(10, 20); w != 0 || h != 0 {
				t.Fatalf("Clamp = %d,%d", w, h)
			}
			p.SetSize(100, 100)
			if err := p.PatchElement(ctx, "a"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("PatchElement err = %v, want ErrNilPage", err)
			}
			if err := p.PatchElements(ctx, nil); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("PatchElements err = %v, want ErrNilPage", err)
			}
			if err := p.Load(ctx, "<p>x</p>"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Load err = %v, want ErrNilPage", err)
			}
			if err := p.Load(ctx, "<p>x</p>", ownframe.LoadOptions{ReplaceCurrent: true}); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Load ReplaceCurrent err = %v, want ErrNilPage", err)
			}
			if err := p.Load(ctx, "<p>x</p>", ownframe.LoadOptions{AdoptTree: true}); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Load AdoptTree err = %v, want ErrNilPage", err)
			}
			if err := p.Back(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Back err = %v, want ErrNilPage", err)
			}
			if err := p.Forward(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Forward err = %v, want ErrNilPage", err)
			}
			if err := p.Route("go", "<p>x</p>"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Route err = %v, want ErrNilPage", err)
			}
			if p.ContextMenuFor("a") != nil {
				t.Fatal("expected nil ContextMenuFor")
			}
			if err := p.MenuAction(ctx, "a"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("MenuAction err = %v, want ErrNilPage", err)
			}

			// Form door
			f := p.Form()
			if f.FormValue("a") != "" || f.FormChecked("a") || f.FormSelected("a") || f.FocusID() != "" {
				t.Fatal("expected zero Form getters")
			}
			f.SetFormValue("a", "v")
			f.SetFormChecked("a", true)
			if err := f.Focus(ctx, "a"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Focus err = %v, want ErrNilPage", err)
			}
			if err := f.FocusNext(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("FocusNext err = %v, want ErrNilPage", err)
			}
			if err := f.FocusPrev(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("FocusPrev err = %v, want ErrNilPage", err)
			}
			f.SetFocusScope("a")
			f.RequestFocus("a")

			// Scroll door + root scroll methods
			s := p.Scroll()
			if x, y := s.ScrollOffset(); x != 0 || y != 0 {
				t.Fatalf("ScrollOffset = %d,%d", x, y)
			}
			s.SetScrollOffset(1, 2)
			s.SetWindowing(true)
			_ = s.Windowing()
			s.SetAllowScroll(true)
			_ = s.AllowScroll()
			s.SetViewportPinZ(10)
			p.SetScrollWindow(nil)
			p.ScrollTo(1, 2)
			p.ScrollBy(1, 2)

			// Spot-check Probe, Window, Access, Edit doors
			p.Probe().MarkRect(image.Rect(0, 0, 10, 10))
			_ = p.Probe().Generation()
			_ = p.Window().Animations()
			if _, err := p.Access().AccessibilityTree(); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("AccessibilityTree err = %v, want ErrNilPage", err)
			}
			if err := p.Access().AccessibilityPerform(ctx, ownframe.AccessibilityRequest{}); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("AccessibilityPerform err = %v, want ErrNilPage", err)
			}
			if _, err := p.Window().PollReload(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("PollReload err = %v, want ErrNilPage", err)
			}
			if err := p.Probe().EndRedrawBatch(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("EndRedrawBatch err = %v, want ErrNilPage", err)
			}
			if _, _, err := p.Edit().Copy(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Edit.Copy err = %v, want ErrNilPage", err)
			}
			if err := p.Edit().Type(ctx, "x"); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Edit.Type err = %v, want ErrNilPage", err)
			}
			if err := p.Testing().Tick(ctx); !errors.Is(err, ownframe.ErrNilPage) {
				t.Fatalf("Testing.Tick err = %v, want ErrNilPage", err)
			}
			_ = p.Edit().LastFocusID()
			_ = p.Edit().FocusedField()
		})
	}
}

// TestNewWithOptionsSkipsNilOption pins HYG-API-01: a nil Option no longer
// panics NewWithOptions.
func TestNewWithOptionsSkipsNilOption(t *testing.T) {
	t.Parallel()
	p, err := ownframe.NewWithOptions(ownframe.Config{HTML: "<p>x</p>", Width: 100, Height: 100}, nil, ownframe.WithPerf(false), nil)
	if err != nil || p == nil {
		t.Fatalf("NewWithOptions = %v, %v", p, err)
	}
}
