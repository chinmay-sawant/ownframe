package ui_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestViewUpdatesCallbacksWithoutNavigating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	oldClicks, newClicks, fallbackClicks := 0, 0, 0
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 320},
		ui.Button("Old").ID("button").OnClick(func(context.Context) error { oldClicks++; return nil }))
	if err != nil {
		t.Fatal(err)
	}
	view.Handle(ownframe.Handlers{Click: func(context.Context, ownframe.Box) error { fallbackClicks++; return nil }})
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "button"); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		err = view.Update(ctx, ui.Div().Children(
			ui.Button("New").ID("button").OnClick(func(context.Context) error { newClicks++; return nil }),
			ui.Button("Fallback").ID("fallback").Attr("data-action", "fallback")))
		if err != nil {
			t.Fatal(err)
		}
	}
	if !strings.Contains(view.Page.HTML(), "New") {
		t.Fatal("replacement source missing")
	}
	if err := view.Page.Back(ctx); !errors.Is(err, ownframe.ErrNoHistory) {
		t.Fatalf("updates added history: %v", err)
	}
	if err := pagetest.Click(ctx, view.Page, "button"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "fallback"); err != nil {
		t.Fatal(err)
	}
	if oldClicks != 1 || newClicks != 1 || fallbackClicks != 1 {
		t.Fatalf("dispatch old=%d new=%d fallback=%d", oldClicks, newClicks, fallbackClicks)
	}
	source := view.Page.HTML()
	if err := view.Update(ctx, ui.Tag("script")); err == nil {
		t.Fatal("invalid tree accepted")
	}
	if source != view.Page.HTML() {
		t.Fatal("invalid tree changed source")
	}
	if err := pagetest.Click(ctx, view.Page, "button"); err != nil {
		t.Fatal(err)
	}
	if newClicks != 2 {
		t.Fatal("failed update changed callbacks")
	}
	if err := view.Update(ctx, ui.Button("No callback").ID("button")); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, view.Page, "button"); err != nil {
		t.Fatal(err)
	}
	if newClicks != 2 {
		t.Fatal("removed callback survived")
	}
}

func TestViewPreservesLiveControlAndFocus(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	tree := func(value string) *ui.Element { return ui.Input().ID("field").Attr("value", value) }
	view, err := ui.Mount(ownframe.Config{Width: 480, Height: 320}, tree("initial"))
	if err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := view.Page.Form().Focus(ctx, "field"); err != nil {
		t.Fatal(err)
	}
	view.Page.Form().SetFormValue("field", "user edit")
	if err := view.Update(ctx, tree("new default")); err != nil {
		t.Fatal(err)
	}
	if got := view.Page.Form().FormValue("field"); got != "user edit" {
		t.Fatalf("value=%q", got)
	}
	if got := view.Page.Form().FocusID(); got != "field" {
		t.Fatalf("focus=%q", got)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	source := view.Page.HTML()
	if err := view.Update(cancelled, ui.Div()); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancel=%v", err)
	}
	if view.Page.HTML() != source {
		t.Fatal("cancel changed source")
	}
	if err := view.Update(ctx, ui.Div()); err != nil {
		t.Fatal(err)
	}
	if view.Page.Form().FocusID() != "" {
		t.Fatal("removed field kept focus")
	}
}

func TestReplaceCurrentPreservesForwardHistory(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := ui.NewPage(ownframe.Config{Width: 480, Height: 320}, ui.Text("first"))
	if err != nil {
		t.Fatal(err)
	}
	first := p.HTML()
	if err := p.Load(ctx, "<p>second</p>"); err != nil {
		t.Fatal(err)
	}
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Load(ctx, "<p>updated first</p>", ownframe.LoadOptions{ReplaceCurrent: true}); err != nil {
		t.Fatal(err)
	}
	if p.HTML() == first {
		t.Fatal("source not replaced")
	}
	if err := p.Forward(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != "<p>second</p>" {
		t.Fatal("forward history changed")
	}
	if err := p.Back(ctx); err != nil {
		t.Fatal(err)
	}
	if p.HTML() != "<p>updated first</p>" {
		t.Fatal("replacement not retained")
	}
}

func TestReplaceCurrentRejectsExecutionFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	p, err := ui.NewPage(ownframe.Config{Width: 480, Height: 320}, ui.Input().ID("field").Attr("value", "initial"))
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.Form().SetFormValue("field", "edited")
	if err := p.Form().Focus(ctx, "field"); err != nil {
		t.Fatal(err)
	}
	source := p.HTML()
	if err := p.Load(ctx, `{{call .Missing}}`, ownframe.LoadOptions{ReplaceCurrent: true}); err == nil {
		t.Fatal("invalid template execution accepted")
	}
	if p.HTML() != source {
		t.Fatal("failed replacement changed source")
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.Form().FormValue("field") != "edited" || p.Form().FocusID() != "field" {
		t.Fatal("failed replacement changed form")
	}
}
