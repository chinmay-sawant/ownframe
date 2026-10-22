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
