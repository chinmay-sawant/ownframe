package main

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestBuilderGalleryRendersAllTabs(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	for _, word := range []string{
		"Widgets gallery", "Buttons", "data-action=\"tab:widgets\"",
		"data-action=\"tab:settings\"", "data-action=\"tab:layout\"",
		"data-action=\"tab:todo\"", "data-action=\"tab:list\"",
		"data-action=\"tab:theme\"", "data-action=\"tab:compare\"",
		"data-action=\"tab:dialog\"",
		"data-action=\"win-dev\"",
	} {
		if !strings.Contains(html, word) {
			t.Errorf("gallery HTML missing %q", word)
		}
	}
	if strings.Contains(html, "<script") {
		t.Fatal("gallery should not depend on scripts")
	}
	if _, err := ui.Render(a.screen()); err != nil {
		t.Fatalf("builder cannot serialize gallery: %v", err)
	}
}

func TestBuilderGalleryTabRouting(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, tab := range []string{
		"tab:settings", "tab:layout", "tab:todo", "tab:list", "tab:theme", "tab:compare", "tab:dialog", "tab:widgets",
	} {
		if err := pagetest.ClickAction(ctx, a.page, tab); err != nil {
			t.Fatalf("click %s: %v", actionName(tab), err)
		}
	}
	if a.tab != "widgets" {
		t.Fatalf("expected to land back on widgets, got %q", a.tab)
	}
	// Spot-check one action per tab through the click path.
	for _, action := range []string{
		"widgets:badges:inc",
		"settings:notify",
		"layout:more",
		"todo:inc",
		"list:filter:go",
		"theme:radius:12px",
		"compare:mode:html",
		"dialog:open",
	} {
		tab := strings.SplitN(action, ":", 2)[0]
		if err := pagetest.ClickAction(ctx, a.page, "tab:"+tab); err != nil {
			t.Fatalf("click tab:%s: %v", tab, err)
		}
		if err := pagetest.ClickAction(ctx, a.page, action); err != nil {
			t.Fatalf("click %s: %v", action, err)
		}
	}
	if a.widgets.badges != defaultWidgetsState().badges+1 {
		t.Fatalf("widgets badge not incremented: %+v", a.widgets)
	}
	if a.settings.Notify == defaultSettingsState().Notify {
		t.Fatal("settings notify should have toggled")
	}
	if a.layout.Boxes != defaultLayoutState().Boxes+1 {
		t.Fatalf("layout boxes not raised: %+v", a.layout)
	}
	if a.todos.Counter != 1 {
		t.Fatalf("todo counter not incremented: %+v", a.todos)
	}
	if a.files.Filter != "go" {
		t.Fatalf("list filter not applied: %+v", a.files)
	}
	if a.theme.Radius != "12px" {
		t.Fatalf("theme radius not applied: %+v", a.theme)
	}
	if a.compare.Mode != "html" {
		t.Fatalf("compare mode not applied: %+v", a.compare)
	}
	if !a.dialog.open {
		t.Fatalf("dialog did not open: %+v", a.dialog)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "dialog:confirm"}); err != nil {
		t.Fatal(err)
	}
	if a.dialog.open || !a.dialog.confirmed {
		t.Fatalf("dialog confirm did not close and record: %+v", a.dialog)
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "win-dev"}); err != nil {
		t.Fatal(err)
	}
	if a.page.Window().DevTools() {
		t.Fatal("win-dev should toggle the dock off")
	}
	if err := a.handleClick(ctx, ownframe.Box{Action: "win-dev"}); err != nil {
		t.Fatal(err)
	}
	if !a.page.Window().DevTools() {
		t.Fatal("win-dev should toggle the dock back on")
	}
}

func TestBuilderDialogBackdropClick(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:dialog"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "dialog:open"); err != nil {
		t.Fatal(err)
	}
	if !a.dialog.open || !strings.Contains(a.page.HTML(), "dialog-backdrop") {
		t.Fatal("dialog should be open over a backdrop")
	}
	if err := a.page.Click(ctx, 30, 400); err != nil {
		t.Fatal(err)
	}
	if a.dialog.open || strings.Contains(a.page.HTML(), "dialog-backdrop") {
		t.Fatal("backdrop click did not close the dialog")
	}
}

func TestBuilderDialogButtonsCloseByID(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:dialog"); err != nil {
		t.Fatal(err)
	}
	clickCenter := func(id string) {
		t.Helper()
		var found bool
		for _, b := range a.page.Boxes() {
			if b.ID == id {
				if err := a.page.Click(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
					t.Fatal(err)
				}
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("box %q not found", id)
		}
	}
	for _, tc := range []struct {
		open string
		want bool
	}{
		{"dialog-x", false},
		{"dialog-cancel", false},
		{"dialog-confirm", true},
	} {
		if err := pagetest.ClickAction(ctx, a.page, "dialog:open"); err != nil {
			t.Fatal(err)
		}
		if !a.dialog.open {
			t.Fatal("dialog should be open")
		}
		clickCenter(tc.open)
		if a.dialog.open {
			t.Fatalf("button %s did not close the dialog", tc.open)
		}
		if a.dialog.confirmed != tc.want {
			t.Fatalf("button %s: confirmed=%v, want %v", tc.open, a.dialog.confirmed, tc.want)
		}
	}
}

func TestBuilderComparePreview(t *testing.T) {
	t.Parallel()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:compare"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "compare:open"); err != nil {
		t.Fatal(err)
	}
	if !a.compare.Preview || !strings.Contains(a.page.HTML(), "compare-preview-backdrop") {
		t.Fatal("compare preview should open over a backdrop")
	}
	if err := a.page.Click(ctx, 30, 400); err != nil {
		t.Fatal(err)
	}
	if a.compare.Preview || strings.Contains(a.page.HTML(), "compare-preview-backdrop") {
		t.Fatal("backdrop click did not close the compare preview")
	}
	// Reopen and close via the × button.
	if err := pagetest.ClickAction(ctx, a.page, "compare:open"); err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, b := range a.page.Boxes() {
		if b.ID == "compare-preview-x" {
			if err := a.page.Click(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
				t.Fatal(err)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("compare preview × box not found")
	}
	if a.compare.Preview {
		t.Fatal("compare preview × did not close the preview")
	}
}

func TestBuilderLayoutSnippetCopyAndCount(t *testing.T) {
	t.Parallel()
	ownframe.UseMemory(true)
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:layout"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "layout:dir:column"); err != nil {
		t.Fatal(err)
	}
	html := a.page.HTML()
	if !strings.Contains(html, "FlexDirection") || !strings.Contains(html, "column") {
		t.Fatal("snippet did not switch to column")
	}
	if err := pagetest.ClickAction(ctx, a.page, "layout:more"); err != nil {
		t.Fatal(err)
	}
	if a.layout.Boxes != defaultLayoutState().Boxes+1 {
		t.Fatalf("more did not add a box: %+v", a.layout)
	}
	if html := a.page.HTML(); !strings.Contains(html, "Box 5") {
		t.Fatal("snippet did not gain Box 5 after +More")
	}
	if err := pagetest.ClickAction(ctx, a.page, "layout:copy"); err != nil {
		t.Fatal(err)
	}
	got, err := ownframe.Read()
	if err != nil {
		t.Fatal(err)
	}
	if got != a.layoutSnippet() {
		t.Fatal("clipboard does not hold the layout snippet")
	}
}

func actionName(action string) string {
	if i := strings.Index(action, ":"); i >= 0 {
		return action[i+1:]
	}
	return action
}

func TestGalleryDialogKeyboardAndFormReset(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	a, err := newBuilderApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := a.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	a.page.Form().SetFormValue("widgets-input", "user edit")
	if err := pagetest.ClickAction(ctx, a.page, "widgets:input:clear"); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FormValue("widgets-input"); got != "" {
		t.Fatalf("clear retained %q", got)
	}
	if err := pagetest.ClickAction(ctx, a.page, "tab:dialog"); err != nil {
		t.Fatal(err)
	}
	if err := a.page.Form().Focus(ctx, "dialog-open"); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.ClickAction(ctx, a.page, "dialog:open"); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FocusID(); got != "dialog-x" {
		t.Fatalf("modal initial focus=%q", got)
	}
	if err := a.page.Form().FocusPrev(ctx); err != nil {
		t.Fatal(err)
	}
	if got := a.page.Form().FocusID(); got != "dialog-cancel" {
		t.Fatalf("modal reverse Tab=%q", got)
	}
	if err := a.page.Edit().KeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}
	if a.dialog.open {
		t.Fatal("Escape did not close dialog")
	}
	if got := a.page.Form().FocusID(); got != "dialog-open" {
		t.Fatalf("restored focus=%q", got)
	}
}
