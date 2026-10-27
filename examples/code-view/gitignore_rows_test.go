package main

import (
	"bytes"
	"context"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestGitignoreRows builds a small repository with ignored files and folders
// and checks the explorer keeps a row for each one, flags it as ignored, and
// styles it without fading the whole row.
func TestGitignoreRows(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q")

	writeFile(t, dir, ".gitignore", "build/\nnode_modules/\n*.test\nsecret.txt\nempty/\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")
	writeFile(t, dir, "secret.txt", "secret\n")
	writeFile(t, dir, "motion.test", "ignored\n")
	writeFile(t, dir, "build/out.txt", "build output\n")
	writeFile(t, dir, "node_modules/pkg/index.js", "module.exports = 1\n")

	if err := os.MkdirAll(filepath.Join(dir, "empty"), 0o700); err != nil {
		t.Fatalf("create empty dir: %v", err)
	}

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New(%q): %v", dir, err)
	}

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	rows := make(map[string]treeRow, len(app.View().Tree))
	for _, row := range app.View().Tree {
		rows[row.Path] = row
	}

	ignoredPaths := []string{"build", "node_modules", "motion.test", "secret.txt", "empty"}
	for _, path := range ignoredPaths {
		row, ok := rows[path]
		if !ok {
			t.Errorf("missing tree row %q", path)
			continue
		}

		if !row.Ignored {
			t.Errorf("tree row %q has Ignored=false, want true", path)
		}
	}

	mainRow, ok := rows["main.go"]
	if !ok {
		t.Error(`missing tree row "main.go"`)
	} else if mainRow.Ignored {
		t.Error(`tree row "main.go" has Ignored=true, want false`)
	}

	shell := buildHTML()
	classHook := "{{if .Ignored}} ignored{{end}}"
	if !strings.Contains(shell, classHook) {
		t.Fatalf("sidebar template no longer emits the ignored class: missing %q", classHook)
	}

	tmpl, err := template.New("shell").Parse(shell)
	if err != nil {
		t.Fatalf("parse shell template: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, &app.view); err != nil {
		t.Fatalf("execute shell template: %v", err)
	}

	for _, path := range ignoredPaths {
		button := regexp.MustCompile(`<button class="([^"]*)" id="tree-` + regexp.QuoteMeta(path) + `"`)
		match := button.FindStringSubmatch(rendered.String())
		if match == nil {
			t.Errorf("rendered shell has no button for tree row %q", path)
			continue
		}

		if !hasClass(match[1], "ignored") {
			t.Errorf("rendered button for %q has class %q, want the ignored class", path, match[1])
		}
	}

	cssBytes, err := os.ReadFile(filepath.Join("components", "sidebar.css"))
	if err != nil {
		t.Fatalf("read sidebar.css: %v", err)
	}

	rule := regexp.MustCompile(`(?s)([^{}]*\.trow\.ignored[^{}]*)\{([^{}]*)\}`)
	matches := rule.FindAllStringSubmatch(string(cssBytes), -1)
	if len(matches) == 0 {
		t.Fatal("sidebar.css has no .trow.ignored rule")
	}

	var hasBase bool

	for _, match := range matches {
		for _, selector := range strings.Split(match[1], ",") {
			if strings.Join(strings.Fields(selector), " ") != ".trow.ignored" {
				continue
			}

			hasBase = true
			if !strings.Contains(match[2], "background-color") {
				t.Errorf(".trow.ignored rule has no solid background: %q", strings.TrimSpace(match[2]))
			}
		}

		if strings.Contains(strings.ToLower(match[2]), "opacity") {
			t.Errorf("ignored CSS uses opacity: selector %q declarations %q",
				strings.TrimSpace(match[1]), strings.TrimSpace(match[2]))
		}
	}

	if !hasBase {
		t.Error("sidebar.css has no base .trow.ignored rule")
	}
}

func hasClass(list, name string) bool {
	for _, field := range strings.Fields(list) {
		if field == name {
			return true
		}
	}

	return false
}
