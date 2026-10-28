package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSearchFile(t *testing.T, dir, name, text string) {
	t.Helper()

	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func searchWorkspaceDir(t *testing.T, files map[string]string) *workspace {
	t.Helper()

	dir := t.TempDir()
	for name, text := range files {
		writeSearchFile(t, dir, name, text)
	}

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	return ws
}

func assertWorkspaceHitsSorted(t *testing.T, hits []workspaceHit) {
	t.Helper()

	for i := 1; i < len(hits); i++ {
		prev, cur := hits[i-1], hits[i]
		if prev.Path > cur.Path ||
			(prev.Path == cur.Path && prev.Line > cur.Line) ||
			(prev.Path == cur.Path && prev.Line == cur.Line && prev.Col > cur.Col) {
			t.Fatalf("hits not sorted at %d: %+v before %+v", i, prev, cur)
		}
	}
}

// TestWorkspaceSearchFindsAcrossFiles covers both the embedded sample and a
// real disk workspace. The sample does not contain "secret", so its half
// searches the project name and the temp workspace covers the requested query.
func TestWorkspaceSearchFindsAcrossFiles(t *testing.T) {
	t.Parallel()

	t.Run("sample", func(t *testing.T) {
		t.Parallel()

		ws, err := loadSampleWorkspace()
		if err != nil {
			t.Fatal(err)
		}

		result := searchWorkspace(context.Background(), ws, "Beacon", true, 0)
		if result.Query != "Beacon" {
			t.Fatalf("query = %q", result.Query)
		}

		if result.Truncated || result.Note != "" {
			t.Fatalf("unexpected truncation: %+v", result)
		}

		if len(result.Hits) == 0 || result.Files < 2 {
			t.Fatalf("hits = %d files = %d", len(result.Hits), result.Files)
		}

		seen := make(map[string]bool)
		for _, hit := range result.Hits {
			seen[hit.Path] = true

			if ws.ByPath[hit.Path] == nil {
				t.Errorf("hit in unlisted file %s", hit.Path)
			}

			if hit.Line < 1 || hit.Col < 1 {
				t.Errorf("bad position %s:%d:%d", hit.Path, hit.Line, hit.Col)
			}

			if hit.Before+hit.Match+hit.After != hit.Text {
				t.Errorf("split does not rebuild %q: %q + %q + %q", hit.Text, hit.Before, hit.Match, hit.After)
			}

			if !strings.EqualFold(hit.Match, "Beacon") {
				t.Errorf("match = %q", hit.Match)
			}

			if len([]rune(hit.Text)) > workspaceSearchTextRunes {
				t.Errorf("text is %d runes", len([]rune(hit.Text)))
			}
		}

		for _, path := range []string{"README.md", "docs/architecture.md", "docs/deployment.md", "web/index.html"} {
			if !seen[path] {
				t.Errorf("no hit in %s", path)
			}
		}

		if result.Files != len(seen) {
			t.Errorf("files = %d, want %d", result.Files, len(seen))
		}

		assertWorkspaceHitsSorted(t, result.Hits)
	})

	t.Run("secret", func(t *testing.T) {
		t.Parallel()

		ws := searchWorkspaceDir(t, map[string]string{
			"alpha.txt":       "\tsecret\r\n",
			"nested/beta.txt": "secret first\nmiddle\nsecret secret\n",
			"gamma.txt":       "nothing here\n",
		})

		result := searchWorkspace(context.Background(), ws, "secret", true, 0)
		if result.Query != "secret" || result.Truncated || result.Note != "" {
			t.Fatalf("result = %+v", result)
		}

		if len(result.Hits) != 4 || result.Files != 2 {
			t.Fatalf("hits = %d files = %d", len(result.Hits), result.Files)
		}

		wantPaths := []string{"alpha.txt", "nested/beta.txt", "nested/beta.txt", "nested/beta.txt"}
		wantCols := []int{5, 1, 1, 8}

		for i, hit := range result.Hits {
			if hit.Path != wantPaths[i] || hit.Col != wantCols[i] {
				t.Errorf("hit %d = %s:%d:%d, want %s:%d", i, hit.Path, hit.Line, hit.Col, wantPaths[i], wantCols[i])
			}

			if hit.Before+hit.Match+hit.After != hit.Text {
				t.Errorf("split does not rebuild %q", hit.Text)
			}
		}

		if result.Hits[0].Text != "    secret" {
			t.Errorf("tab expansion = %q", result.Hits[0].Text)
		}

		if result.Hits[3].Before != "secret " || result.Hits[3].After != "" {
			t.Errorf("last split = %q + %q + %q", result.Hits[3].Before, result.Hits[3].Match, result.Hits[3].After)
		}

		assertWorkspaceHitsSorted(t, result.Hits)
	})
}

func TestWorkspaceSearchCaseSensitive(t *testing.T) {
	t.Parallel()

	ws := searchWorkspaceDir(t, map[string]string{
		"case.txt": "Secret\nsecret\nSECRET\nsecret\n",
	})

	exact := searchWorkspace(context.Background(), ws, "Secret", true, 0)
	if len(exact.Hits) != 1 || exact.Hits[0].Line != 1 || exact.Hits[0].Match != "Secret" {
		t.Fatalf("exact = %+v", exact.Hits)
	}

	folded := searchWorkspace(context.Background(), ws, "Secret", false, 0)
	if len(folded.Hits) != 4 {
		t.Fatalf("folded hits = %d", len(folded.Hits))
	}

	want := []string{"Secret", "secret", "SECRET", "secret"}
	for i, hit := range folded.Hits {
		if hit.Line != i+1 || hit.Match != want[i] {
			t.Errorf("folded %d = line %d %q", i, hit.Line, hit.Match)
		}
	}

	lower := searchWorkspace(context.Background(), ws, "secret", true, 0)
	if len(lower.Hits) != 2 || lower.Hits[0].Line != 2 || lower.Hits[1].Line != 4 {
		t.Fatalf("lower = %+v", lower.Hits)
	}
}

func TestWorkspaceSearchLongLines(t *testing.T) {
	t.Parallel()

	ws := searchWorkspaceDir(t, map[string]string{
		"long.txt": "\t" + strings.Repeat("界", 235) + "Needle " + strings.Repeat("x", 100) + "needle\n",
	})
	result := searchWorkspace(context.Background(), ws, "needle", false, 0)
	if len(result.Hits) != 2 || result.Truncated {
		t.Fatalf("result = %+v", result)
	}

	for i, want := range []struct {
		col   int
		match string
	}{{240, "Needle"}, {347, "needle"}} {
		hit := result.Hits[i]
		if hit.Line != 1 || hit.Col != want.col || hit.Match != want.match {
			t.Errorf("hit %d = %+v, want column %d match %q", i, hit, want.col, want.match)
		}
		if hit.Before+hit.Match+hit.After != hit.Text || len([]rune(hit.Text)) > workspaceSearchTextRunes {
			t.Errorf("invalid preview for hit %d: %+v", i, hit)
		}
	}
}

func TestWorkspaceSearchLongQueryPreview(t *testing.T) {
	t.Parallel()

	query := strings.Repeat("a", workspaceSearchTextRunes+10)
	ws := searchWorkspaceDir(t, map[string]string{"long.txt": "prefix" + query + "suffix"})
	result := searchWorkspace(context.Background(), ws, query, true, 0)
	if len(result.Hits) != 1 {
		t.Fatalf("hits = %d, want 1", len(result.Hits))
	}

	hit := result.Hits[0]
	if hit.Col != 7 || hit.Before+hit.Match+hit.After != hit.Text || len([]rune(hit.Text)) != workspaceSearchTextRunes {
		t.Fatalf("unexpected preview: %+v", hit)
	}
}

func TestWorkspaceSearchLimit(t *testing.T) {
	t.Parallel()

	t.Run("stops", func(t *testing.T) {
		t.Parallel()

		ws := searchWorkspaceDir(t, map[string]string{
			"hits.txt": strings.Repeat("needle\n", 5),
		})

		result := searchWorkspace(context.Background(), ws, "needle", true, 3)
		if len(result.Hits) != 3 || !result.Truncated {
			t.Fatalf("hits = %d truncated = %v", len(result.Hits), result.Truncated)
		}

		if result.Files != 1 || !strings.Contains(result.Note, "3") {
			t.Fatalf("files = %d note = %q", result.Files, result.Note)
		}
	})

	t.Run("exact", func(t *testing.T) {
		t.Parallel()

		ws := searchWorkspaceDir(t, map[string]string{
			"hits.txt": strings.Repeat("needle\n", 5),
		})

		result := searchWorkspace(context.Background(), ws, "needle", true, 5)
		if len(result.Hits) != 5 || result.Truncated || result.Note != "" {
			t.Fatalf("hits = %d truncated = %v note = %q", len(result.Hits), result.Truncated, result.Note)
		}
	})

	t.Run("default", func(t *testing.T) {
		t.Parallel()

		files := make(map[string]string, 26)
		for i := 0; i < 26; i++ {
			files[fmt.Sprintf("f%02d.txt", i)] = strings.Repeat("needle\n", workspaceSearchFileHits)
		}

		ws := searchWorkspaceDir(t, files)

		result := searchWorkspace(context.Background(), ws, "needle", true, 0)
		if len(result.Hits) != workspaceSearchLimit {
			t.Fatalf("hits = %d", len(result.Hits))
		}

		if result.Files != 25 {
			t.Fatalf("files = %d", result.Files)
		}

		if !result.Truncated || !strings.Contains(result.Note, "200") {
			t.Fatalf("truncated = %v note = %q", result.Truncated, result.Note)
		}
	})
}

func TestWorkspaceSearchPerFileCap(t *testing.T) {
	t.Parallel()

	ws := searchWorkspaceDir(t, map[string]string{
		"a.txt": strings.Repeat("needle\n", 10),
		"b.txt": "needle\nneedle\n",
	})

	result := searchWorkspace(context.Background(), ws, "needle", true, 0)

	if len(result.Hits) != workspaceSearchFileHits+2 {
		t.Fatalf("hits = %d", len(result.Hits))
	}

	if result.Files != 2 || !result.Truncated {
		t.Fatalf("files = %d truncated = %v", result.Files, result.Truncated)
	}

	if !strings.Contains(result.Note, "8") {
		t.Fatalf("note = %q", result.Note)
	}

	for i := 0; i < workspaceSearchFileHits; i++ {
		if hit := result.Hits[i]; hit.Path != "a.txt" || hit.Line != i+1 {
			t.Errorf("hit %d = %s:%d", i, hit.Path, hit.Line)
		}
	}

	if hit := result.Hits[workspaceSearchFileHits]; hit.Path != "b.txt" || hit.Line != 1 {
		t.Errorf("first b hit = %s:%d", hit.Path, hit.Line)
	}
}

func TestWorkspaceSearchSkipsBinaryAndBig(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeSearchFile(t, dir, "text.txt", "secret one\n")
	writeSearchFile(t, dir, "binary.bin", "secret\x00binary\n")
	writeSearchFile(t, dir, "big.txt", "secret\n"+strings.Repeat("x", maxFileBytes))

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	if ws.ByPath["binary.bin"] == nil {
		t.Fatal("binary fixture was not scanned")
	}

	if meta := ws.ByPath["big.txt"]; meta == nil || !meta.Big {
		t.Fatalf("big fixture = %+v", meta)
	}

	result := searchWorkspace(context.Background(), ws, "secret", true, 0)
	if len(result.Hits) != 1 || result.Hits[0].Path != "text.txt" {
		t.Fatalf("hits = %+v", result.Hits)
	}

	if result.Files != 1 || result.Truncated || result.Note != "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestWorkspaceSearchCancel(t *testing.T) {
	t.Parallel()

	ws := searchWorkspaceDir(t, map[string]string{
		"a.txt": "secret one\nsecret two\n",
		"b.txt": "secret three\nsecret four\n",
	})

	full := searchWorkspace(context.Background(), ws, "secret", true, 0)
	if len(full.Hits) != 4 {
		t.Fatalf("full hits = %d", len(full.Hits))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	stopped := searchWorkspace(ctx, ws, "secret", true, 0)
	if stopped.Note != "Search cancelled." {
		t.Fatalf("note = %q", stopped.Note)
	}

	if len(stopped.Hits) != 0 || stopped.Files != 0 || stopped.Query != "secret" {
		t.Fatalf("stopped = %+v", stopped)
	}

	partial := searchWorkspace(&countdownContext{Context: context.Background(), remaining: 3}, ws, "secret", true, 0)
	if partial.Note != "Search cancelled." {
		t.Fatalf("partial note = %q", partial.Note)
	}

	if len(partial.Hits) == 0 || len(partial.Hits) >= len(full.Hits) {
		t.Fatalf("partial hits = %d", len(partial.Hits))
	}

	if partial.Files != 1 {
		t.Fatalf("partial files = %d", partial.Files)
	}

	for _, hit := range partial.Hits {
		if hit.Path != "a.txt" {
			t.Errorf("partial hit = %s:%d", hit.Path, hit.Line)
		}
	}
}

// countdownContext cancels after a fixed number of Err checks so the test can
// inspect a pass that stopped between lines.
type countdownContext struct {
	context.Context
	remaining int
}

func (c *countdownContext) Err() error {
	if c.remaining <= 0 {
		return context.Canceled
	}

	c.remaining--

	return nil
}
