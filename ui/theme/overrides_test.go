package theme_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

func TestOverridesFromFile(t *testing.T) {
	t.Parallel()
	files := fstest.MapFS{"themes/app.json": &fstest.MapFile{Data: []byte(`{"--of-primary":"#21694e","--mail-row-height":"104px"}`)}}
	properties, err := theme.LoadOverrides(files, "themes/app.json")
	if err != nil {
		t.Fatal(err)
	}
	sheet, err := properties.Stylesheet()
	if err != nil {
		t.Fatal(err)
	}
	if sheet != `:root{--mail-row-height:104px;--of-primary:#21694e;}` {
		t.Fatalf("stylesheet=%q", sheet)
	}
	rendered, err := ui.Render(properties.Apply(ui.Div().Color(theme.Primary()).Height(ui.LengthVar("--mail-row-height", ui.Px(80)))))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rendered, "--mail-row-height:104px;") {
		t.Fatal("app-owned property missing")
	}
	if _, err := theme.LoadOverrides(files, "missing.json"); err == nil {
		t.Fatal("missing file accepted")
	}
}

func TestOverrideDeclarationBoundaries(t *testing.T) {
	t.Parallel()
	for _, value := range []string{
		`null`, `[]`, `{"primary":"#123456"}`, `{"--of-primary":"red; color:blue"}`,
		`{"--of-primary":"red}body{color:blue"}`, `{"--of-primary":123}`,
	} {
		files := fstest.MapFS{"override.json": &fstest.MapFile{Data: []byte(value)}}
		if _, err := theme.LoadOverrides(files, "override.json"); err == nil {
			t.Errorf("accepted invalid overrides %s", value)
		}
	}
	invalid := theme.Overrides{"color": "red"}
	if _, err := ui.Render(invalid.Apply(ui.Div())); err == nil {
		t.Fatal("invalid component override accepted")
	}
}
