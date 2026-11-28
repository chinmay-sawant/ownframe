package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleTemplate = `<!DOCTYPE html>
<html>
<head><style>.ok { color: green; }</style></head>
<body>
  {{/* header comment */}}
  <div id="status-badge" class="{{if .Online}}ok{{else}}off{{end}}">{{.Title}}</div>
  <span id="counter">{{.Count:int}}</span>
  <span id="price">{{.Price:float64}}</span>
  <button data-action="save">Save</button>
  <button data-action="add-task">Add</button>
  <div id="slot">{{raw .BodyHTML}}</div>
  <div id="task-list">
    {{- range .Tasks -}}
    <div id="task-{{.ID:int}}" class="row{{if not .Done}} pending{{end}}">{{.Name}}</div>
    {{- end -}}
  </div>
</body>
</html>`

func TestParseAndEmitSample(t *testing.T) {
	t.Parallel()

	view, err := ParseTemplate("dashboard.html", "DashboardView", sampleTemplate)
	if err != nil {
		t.Fatalf("ParseTemplate failed: %v", err)
	}

	// Verify extracted static IDs (dynamic "task-{{.ID:int}}" must be excluded).
	wantIDs := []string{"status-badge", "counter", "price", "slot", "task-list"}
	if len(view.IDs) != len(wantIDs) {
		t.Fatalf("got IDs %v, want %v", view.IDs, wantIDs)
	}
	for i, want := range wantIDs {
		if view.IDs[i] != want {
			t.Fatalf("ID[%d] = %q, want %q", i, view.IDs[i], want)
		}
	}

	// Verify extracted data-action names.
	wantActions := []string{"save", "add-task"}
	if len(view.Actions) != len(wantActions) {
		t.Fatalf("got Actions %v, want %v", view.Actions, wantActions)
	}
	for i, want := range wantActions {
		if view.Actions[i] != want {
			t.Fatalf("Action[%d] = %q, want %q", i, view.Actions[i], want)
		}
	}

	src, err := EmitFile("main", []*ParsedView{view})
	if err != nil {
		t.Fatalf("EmitFile failed: %v", err)
	}

	// Must be valid Go syntax and already gofmt-formatted.
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, "dashboard_gen.go", src, parser.AllErrors); err != nil {
		t.Fatalf("generated code failed go/parser: %v\n%s", err, src)
	}
	reformatted, err := format.Source(src)
	if err != nil {
		t.Fatalf("format.Source failed: %v", err)
	}
	if !bytes.Equal(src, reformatted) {
		t.Fatal("generated code is not gofmt-canonical")
	}

	code := string(src)
	normalized := strings.Join(strings.Fields(code), " ")
	checks := []string{
		`IDStatusBadge = "status-badge"`,
		`IDTaskList = "task-list"`,
		`ActionSave = "save"`,
		`ActionAddTask = "add-task"`,
		`type DashboardView struct`,
		`Count int`,
		`Price float64`,
		`Online bool`,
		`BodyHTML string`,
		`Tasks []DashboardViewTask`,
		`func NewDashboardView() *DashboardView`,
		`func (v *DashboardView) WithTitle(val string) *DashboardView`,
		`func (v *DashboardView) WithCount(val int) *DashboardView`,
		`func (v *DashboardView) WithTasks(items ...DashboardViewTask) *DashboardView`,
		`func (v *DashboardView) AddTask(item DashboardViewTask) *DashboardView`,
		`func (v *DashboardView) RenderTo(b *strings.Builder)`,
		`func (v *DashboardView) Render(w io.Writer, data any) error`,
		`html.EscapeString(v.Title)`,
		`strconv.AppendInt(numBuf[:0], int64(v.Count), 10)`,
		`strconv.AppendFloat(numBuf[:0], v.Price, 'f', -1, 64)`,
		`b.WriteString(v.BodyHTML)`,
		`func (p *DashboardViewPatch) StatusBadgeText(text string) *DashboardViewPatch`,
		`func (p *DashboardViewPatch) StatusBadgeAddClass(className string) *DashboardViewPatch`,
		`func (p *DashboardViewPatch) TaskListAppendHTML(fragment string) *DashboardViewPatch`,
		`func (p *DashboardViewPatch) Apply(ctx context.Context, page *ownframe.Page) error`,
	}
	for _, want := range checks {
		if !strings.Contains(normalized, want) {
			t.Errorf("generated code missing %q", want)
		}
	}
	if strings.Contains(code, "header comment") {
		t.Error("template comment leaked into generated code")
	}
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		tmpl string
	}{
		{"unclosed action", `<div>{{.Title</div>`},
		{"unclosed if", `{{if .Ok}}<span>hi</span>`},
		{"unclosed range", `{{range .Items}}<span>{{.Name}}</span>`},
		{"unexpected else", `<div>{{else}}</div>`},
		{"unexpected end", `<div>{{end}}</div>`},
		{"conflicting types", `{{.Count:int}}{{.Count:float64}}`},
		{"raw non-string", `{{raw .Count:int}}`},
		{"empty action", `{{}}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if _, err := ParseTemplate("bad.html", "BadView", tc.tmpl); err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestCLIRunSingleFileAndDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inFile := filepath.Join(dir, "panel.html")
	outFile := filepath.Join(dir, "panel_gen.go")

	const htmlSrc = `<div id="hdr">{{.Heading}}</div>`
	if err := os.WriteFile(inFile, []byte(htmlSrc), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	err := run([]string{"-in", inFile, "-out", outFile, "-pkg", "demo", "-type", "PanelView"}, &stdout)
	if err != nil {
		t.Fatalf("run single file failed: %v", err)
	}
	//nolint:gosec // Test reads temp output file created inside t.TempDir().
	got, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), "package demo") || !strings.Contains(string(got), "type PanelView struct") {
		t.Fatalf("unexpected generated output:\n%s", got)
	}

	// Test stdout mode with directory input.
	stdout.Reset()
	if err := run([]string{"-in", dir, "-out", "-", "-pkg", "demo"}, &stdout); err != nil {
		t.Fatalf("run directory mode failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "type PanelView struct") {
		t.Fatalf("directory mode missing PanelView:\n%s", stdout.String())
	}
}
