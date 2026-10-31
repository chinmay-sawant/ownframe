package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

func TestNewPageConnectsNativeClickHandler(t *testing.T) {
	t.Parallel()
	clicked := 0
	root := Div().Padding("10px").Children(
		Button("Save").ID("save").OnClick(func(context.Context) error {
			clicked++
			return nil
		}),
	)

	page, err := NewPage(ownframe.Config{
		Title: "UI Builder",
		Width: 480, Height: 320,
	}, root)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(page.HTML(), "data-action=\"ownframe-ui-1\"") {
		t.Fatalf("builder action missing from page source: %s", page.HTML())
	}

	ctx := context.Background()
	if err := page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := pagetest.Click(ctx, page, "save"); err != nil {
		t.Fatal(err)
	}
	if clicked != 1 {
		t.Fatalf("click callback called %d times, want 1", clicked)
	}
}

func TestNewPageRejectsConflictingTemplate(t *testing.T) {
	t.Parallel()
	_, err := NewPage(ownframe.Config{HTML: "<p>existing</p>"}, Div())
	if err == nil {
		t.Fatal("expected error for Config.HTML conflict")
	}
	_, err = NewPage(ownframe.Config{File: "index.html"}, Div())
	if err == nil {
		t.Fatal("expected error for Config.File conflict")
	}
	_, err = NewPage(ownframe.Config{}, nil)
	if err == nil {
		t.Fatal("expected nil-tree error")
	}
}

func TestNewPageRejectsTemplateActionsInTitleAndText(t *testing.T) {
	t.Parallel()
	page, err := NewPage(ownframe.Config{
		Title: "{{.Admin}}", Width: 480, Height: 320,
	}, Div().Children(Text("{{.Password}}")))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(page.HTML(), "{{") {
		t.Fatalf("raw template actions in builder page: %s", page.HTML())
	}
}
