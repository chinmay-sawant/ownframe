package ui

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

const (
	testWidth  = 480
	testHeight = 320
)

func TestNewPageConnectsNativeClickHandler(t *testing.T) {
	t.Parallel()
	clicked := 0
	root := Div().Padding(Px(10)).Children(
		Button("Save").ID("save").OnClick(func(context.Context) error {
			clicked++
			return nil
		}),
	)

	page, err := NewPage(ownframe.Config{
		Title: "UI Builder",
		Width: testWidth, Height: testHeight,
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
		Title: "{{.Admin}}", Width: testWidth, Height: testHeight,
	}, Div().Children(Text("{{.Password}}")))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(page.HTML(), "{{") {
		t.Fatalf("raw template actions in builder page: %s", page.HTML())
	}
}

func TestCustomPropertiesReachEngineLayout(t *testing.T) {
	t.Parallel()
	const variableWidth = 137
	const fallbackWidth = 83
	root := Div().CustomProperty("--CardWidth", "137px").Children(
		Div().ID("inherited").Width(Var("--CardWidth")).Height(Px(20)),
		Div().ID("fallback").Width(Var("--missing", "83px")).Height(Px(20)),
	)
	page, err := NewPage(ownframe.Config{Width: testWidth, Height: testHeight}, root)
	if err != nil {
		t.Fatal(err)
	}
	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, box := range page.Boxes() {
		want := 0
		switch box.ID {
		case "inherited":
			want = variableWidth
		case "fallback":
			want = fallbackWidth
		default:
			continue
		}
		seen[box.ID] = true
		if box.W != float64(want) {
			t.Errorf("%s width = %v, want %d", box.ID, box.W, want)
		}
	}
	for _, id := range []string{"inherited", "fallback"} {
		if !seen[id] {
			t.Errorf("missing rendered box %s", id)
		}
	}
}
