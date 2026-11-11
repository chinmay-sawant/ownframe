package chadcn_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

type managedControls struct {
	input                        *chadcn.Interactions
	view                         *ui.View
	active                       string
	value                        float64
	open, disabled, removed, bad bool
	hideTrigger, hideFirstTab    bool
	changes, keys, clicks        int
}

func newManagedControls(t *testing.T) *managedControls {
	t.Helper()
	f := &managedControls{input: chadcn.NewInteractions(), active: "a", value: 25}
	view, err := f.input.Mount(t.Context(), ownframe.Config{
		Width: 640, Height: 480, Theme: chadcn.Light + chadcn.Stylesheet,
	}, f.screen)
	if err != nil {
		t.Fatal(err)
	}
	f.view = view
	f.input.Handle(ownframe.Handlers{KeyDown: func(context.Context, string) error { f.keys++; return nil }})
	if err := view.Page.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	return f
}

func (f *managedControls) refresh(ctx context.Context) error { return f.input.Update(ctx, f.screen) }

func (f *managedControls) screen() *ui.Element {
	root := ui.Div()
	if !f.hideTrigger {
		root.Children(ui.Button("Open").ID("open").OnClick(func(ctx context.Context) error { f.open = true; return f.refresh(ctx) }))
	}
	items := []chadcn.TabItem{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}, {ID: "c", Label: "C"}}
	if f.hideFirstTab {
		items = items[1:]
	}
	root.Children(
		ui.Button("Outside").ID("outside").OnClick(func(context.Context) error { f.clicks++; return nil }),
		chadcn.Tabs(chadcn.TabsOptions{Interactions: f.input, Active: f.active,
			Items:    items,
			OnSelect: func(ctx context.Context, id string) error { f.active = id; return f.refresh(ctx) },
		}),
	)
	if !f.removed {
		id := "volume"
		if f.bad {
			id = ""
		}
		root.Children(chadcn.Slider(chadcn.SliderOptions{Interactions: f.input, ID: id,
			Value: f.value, Step: 5, Disabled: f.disabled,
			OnValueChange: func(ctx context.Context, value float64) error {
				f.value = value
				f.changes++
				return f.refresh(ctx)
			},
		}).Width(ui.Px(300)))
	}
	root.Children(chadcn.Dialog(chadcn.DialogOptions{Interactions: f.input,
		ID: "modal", Open: f.open, Title: "Managed dialog",
		OnOpenChange: func(ctx context.Context, open bool) error { f.open = open; return f.refresh(ctx) },
	}, ui.Input().ID("editor"), ui.Button("Confirm").ID("confirm")))
	return root
}

func managedBox(t *testing.T, p *ownframe.Page, id string) ownframe.Box {
	t.Helper()
	b, ok := pagetest.BoxByID(p, id)
	if !ok {
		t.Fatalf("missing control %s", id)
	}
	return b
}

func TestManagedTabsKeyboardAndHandlerComposition(t *testing.T) {
	f := newManagedControls(t)
	p := f.view.Page
	if err := p.Form().Focus(t.Context(), "a-tab"); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ key, want string }{
		{key: "arrowright", want: "b"}, {key: "end", want: "c"},
		{key: "arrowright", want: "a"}, {key: "arrowleft", want: "c"}, {key: "home", want: "a"},
	} {
		if err := p.Edit().KeyDown(t.Context(), test.key); err != nil {
			t.Fatal(err)
		}
		if f.active != test.want || p.Form().FocusID() != test.want+"-tab" {
			t.Fatalf("%s: active %s focus %s", test.key, f.active, p.Form().FocusID())
		}
	}
	if err := p.Edit().KeyDown(t.Context(), "arrowup"); err != nil {
		t.Fatal(err)
	}
	if f.keys != 1 {
		t.Fatal("horizontal tabs swallowed an unrelated key or ran handlers twice")
	}
	b := managedBox(t, p, "outside")
	if err := p.Click(t.Context(), b.X+b.W/2, b.Y+b.H/2); err != nil {
		t.Fatal(err)
	}
	if f.clicks != 1 {
		t.Fatal("interaction installation replaced builder OnClick dispatch")
	}
}

func TestManagedSliderKeysPointerAndCancellation(t *testing.T) {
	f := newManagedControls(t)
	p := f.view.Page
	if err := p.Form().Focus(t.Context(), "volume"); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		key  string
		want float64
	}{
		{key: "arrowright", want: 30}, {key: "home", want: 0}, {key: "end", want: 100}, {key: "arrowdown", want: 95},
	} {
		if err := p.Edit().KeyDown(t.Context(), test.key); err != nil {
			t.Fatal(err)
		}
		if f.value != test.want {
			t.Fatalf("%s value %g, want %g", test.key, f.value, test.want)
		}
	}
	rail := managedBox(t, p, "volume-rail")
	x, y := rail.X+rail.W/2, rail.Y+rail.H/2
	claimed, err := p.Testing().BeginDrag(t.Context(), x, y)
	if err != nil || !claimed || f.value != 50 {
		t.Fatalf("slider press/claim = %v, %v; value %g", claimed, err, f.value)
	}
	if err := p.Testing().MoveDrag(t.Context(), x+rail.W, y); err != nil {
		t.Fatal(err)
	}
	if f.value != 100 || !strings.Contains(p.HTML(), `aria-valuenow="100"`) {
		t.Fatal("slider drag did not clamp and publish its controlled value")
	}
	f.disabled = true
	if err := f.refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	changes := f.changes
	if err := p.Testing().MoveDrag(t.Context(), x, y); err != nil {
		t.Fatal(err)
	}
	if f.changes != changes {
		t.Fatal("disabled slider retained its drag callback")
	}
	if err := p.Testing().EndDrag(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func TestManagedDialogFocusEscapeAndReopen(t *testing.T) {
	f := newManagedControls(t)
	p := f.view.Page
	for range 2 {
		trigger := managedBox(t, p, "open")
		if err := p.Click(t.Context(), trigger.X+trigger.W/2, trigger.Y+trigger.H/2); err != nil {
			t.Fatal(err)
		}
		if !f.open || p.Form().FocusID() == "open" || p.Form().FocusID() == "" {
			t.Fatal("opening dialog did not focus an internal control")
		}
		for range 8 {
			if err := p.Form().FocusNext(t.Context()); err != nil {
				t.Fatal(err)
			}
			id := p.Form().FocusID()
			if id != "editor" && id != "confirm" && id != "modal-panel-close" {
				t.Fatalf("dialog focus escaped to %s", id)
			}
		}
		if err := p.Form().FocusPrev(t.Context()); err != nil {
			t.Fatal(err)
		}
		if err := p.Edit().KeyDown(t.Context(), "escape"); err != nil {
			t.Fatal(err)
		}
		if f.open || p.Form().FocusID() != "open" {
			t.Fatal("Escape failed to close and restore focus")
		}
	}
}

func TestManagedRegistrationFailureKeepsPreviousFrame(t *testing.T) {
	f := newManagedControls(t)
	before := f.view.Page.HTML()
	f.bad = true
	if err := f.refresh(t.Context()); err == nil {
		t.Fatal("invalid managed slider was accepted")
	}
	if f.view.Page.HTML() != before {
		t.Fatal("failed registration replaced the page source")
	}
	f.bad = false
	f.removed = true
	if err := f.refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := f.input.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err := f.view.Page.Edit().KeyDown(t.Context(), "arrowright"); err != nil {
		t.Fatal(err)
	}
	if f.keys != 1 || f.value != 25 {
		t.Fatal("teardown retained component input mechanics")
	}
}

func TestManagedRemovedFocusTargets(t *testing.T) {
	f := newManagedControls(t)
	p := f.view.Page
	if err := p.Form().Focus(t.Context(), "a-tab"); err != nil {
		t.Fatal(err)
	}
	f.hideFirstTab = true
	f.active = "b"
	if err := f.refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	if p.Form().FocusID() != "b-tab" {
		t.Fatalf("removed tab focus %q", p.Form().FocusID())
	}
	if err := p.Form().Focus(t.Context(), "open"); err != nil {
		t.Fatal(err)
	}
	f.open = true
	if err := f.refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	focused := p.Form().FocusID()
	f.hideTrigger = true
	if err := f.refresh(t.Context()); err != nil {
		t.Fatal(err)
	}
	if p.Form().FocusID() != focused {
		t.Fatalf("rebuild lost modal focus: %q -> %q", focused, p.Form().FocusID())
	}
	if err := p.Edit().KeyDown(t.Context(), "escape"); err != nil {
		t.Fatal(err)
	}
	if p.Form().FocusID() != "outside" {
		t.Fatalf("removed trigger fallback %q", p.Form().FocusID())
	}
}
