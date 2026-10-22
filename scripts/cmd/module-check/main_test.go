package main

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchiveExcludesNestedModulesAndKeepsFixtures(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                               "module example.org/library\n\ngo 1.26.10\n",
		"api.go":                               "package library\n",
		"internal/replay/testdata/picture.txt": "fixture",
		"examples/go.mod":                      "module example.org/library/examples\n\ngo 1.26.10\n",
		"examples/main.go":                     "package main\n",
		"scripts/go.mod":                       "module example.org/library/scripts\n\ngo 1.26.10\n",
		"scripts/tool.go":                      "package main\n",
	}
	var names []string
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	data, version, err := archive(root, names, []byte(files["go.mod"]))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	fixture := false
	for _, file := range reader.File {
		name := strings.TrimPrefix(file.Name, version.Path+"@"+version.Version+"/")
		if strings.HasPrefix(name, "examples/") || strings.HasPrefix(name, "scripts/") {
			t.Fatal("nested module leaked:", name)
		}
		fixture = fixture || name == "internal/replay/testdata/picture.txt"
	}
	if !fixture {
		t.Fatal("library fixture missing from archive")
	}
	if err := os.WriteFile(filepath.Join(root, "api.go"), []byte("package library\n// changed\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, changed, err := archive(root, names, []byte(files["go.mod"]))
	if err != nil {
		t.Fatal(err)
	}
	if changed.Version == version.Version {
		t.Fatal("changed content reused synthetic version")
	}
}

func TestArchiveOmitsRootResidue(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.mod":                    "module example.org/library\n\ngo 1.26.10\n",
		"go.sum":                    "example.org/dep v1.0.0 h1:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=\n",
		"LICENSE":                   "license\n",
		"README.md":                 "readme\n",
		"api.go":                    "package library\n",
		"internal/engine/engine.go": "package engine\n",
		"internal/cmd/tool.go":      "package cmd\n",
		"pagetest/fixture.go":       "package pagetest\n",
		"go.work":                   "go 1.26.10\nuse .\n",
		"go.work.sum":               "example.org/dep v1.0.0 h1:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa=\n",
		".agents/skills/x.md":       "skill\n",
		".github/workflows/ci.yml":  "name: ci\n",
		".semgrep/rules.yml":        "rules: []\n",
		"cmd/preview/main.go":       "package main\n",
	}
	kept := map[string]bool{
		"go.mod":                    true,
		"go.sum":                    true,
		"LICENSE":                   true,
		"README.md":                 true,
		"api.go":                    true,
		"internal/engine/engine.go": true,
		"internal/cmd/tool.go":      true,
		"pagetest/fixture.go":       true,
	}
	var names []string
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	data, version, err := archive(root, names, []byte(files["go.mod"]))
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	prefix := version.Path + "@" + version.Version + "/"
	for _, file := range reader.File {
		name := strings.TrimPrefix(file.Name, prefix)
		if name == "go.work" || name == ".github/workflows/ci.yml" || moduleResidue(name) {
			t.Fatal("builder included", name)
		}
		delete(kept, name)
	}
	if len(kept) != 0 {
		t.Fatal("library files missing from archive:", kept)
	}
	if err := rejectArchive(reader, version); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{
		"go.work",
		"go.work.sum",
		".agents/skills/x.md",
		".github/workflows/ci.yml",
		".semgrep/rules.yml",
		"cmd/preview/main.go",
	} {
		fixture := zipFixture(t, prefix+name, prefix+"api.go")
		if err := rejectArchive(fixture, version); err == nil {
			t.Fatal("fixture zip included", name)
		}
	}
	clean := zipFixture(t,
		prefix+"api.go",
		prefix+"internal/engine/engine.go",
		prefix+"internal/cmd/tool.go",
		prefix+"pagetest/fixture.go",
		prefix+"go.mod",
		prefix+"go.sum",
		prefix+"LICENSE",
		prefix+"README.md",
	)
	if err := rejectArchive(clean, version); err != nil {
		t.Fatal(err)
	}
}

func zipFixture(t *testing.T, names ...string) *zip.Reader {
	t.Helper()
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, name := range names {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte("x")); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	data := buf.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	return reader
}

func TestExamplesReleaseRejectsFilesystemReplacements(t *testing.T) {
	base := "module github.com/chinmay-sawant/ownframe/examples\n\ngo 1.26.10\nrequire github.com/chinmay-sawant/ownframe v0.0.3\n"
	if err := validateExamples("examples/v0.0.3", []byte(base)); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct{ tag, manifest string }{
		{"v0.0.3", base},
		{"examples/not-a-version", base},
		{"examples/v0.0.3", base + "replace github.com/chinmay-sawant/ownframe => ../\n"},
		{"examples/v0.0.3", "module example.org/empty\n"},
	} {
		if err := validateExamples(test.tag, []byte(test.manifest)); err == nil {
			t.Fatalf("accepted invalid examples release: %+v", test)
		}
	}
}
