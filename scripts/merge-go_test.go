package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func writeMergeFixture(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestMergedProgramKeepsBehaviorAndComments(t *testing.T) {
	dir := t.TempDir()
	a, b, dest := filepath.Join(dir, "a.go"), filepath.Join(dir, "b.go"), filepath.Join(dir, "joined.go")
	writeMergeFixture(t, a, "// Package main is a merge fixture.\npackage main\nimport \"fmt\"\n// main calls the helper.\nfunc main() { fmt.Println(message()) } // trailing comment\n")
	writeMergeFixture(t, b, "package main\nimport \"strings\"\n// message returns the result.\nfunc message() string { return strings.ToUpper(\"kept\") }\n")
	if err := merge([]string{dest, a, b}, true); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	for _, comment := range []string{"Package main is a merge fixture.", "main calls the helper.", "trailing comment", "message returns the result."} {
		if !strings.Contains(string(body), comment) {
			t.Fatalf("lost comment %q", comment)
		}
	}
	output, err := exec.Command("go", "run", dest).CombinedOutput()
	if err != nil || string(output) != "KEPT\n" {
		t.Fatalf("merged program: %s, %v", output, err)
	}
	for _, path := range []string{a, b} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("source still exists or could not be checked: %s: %v", path, err)
		}
	}
}

func TestUnsafeMergeLeavesSourcesUntouched(t *testing.T) {
	cases := []struct{ name, file, body string }{
		{"constraint", "a.go", "//go:build linux\n\npackage fixture\n"},
		{"platform", "a_linux.go", "package fixture\n"},
		{"platform test", "a_linux_test.go", "package fixture\n"},
		{"cgo", "a.go", "package fixture\nimport \"C\"\n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			a := filepath.Join(dir, tc.file)
			b, dest := filepath.Join(dir, "b.go"), filepath.Join(dir, "joined.go")
			if strings.HasSuffix(a, "_test.go") {
				b, dest = filepath.Join(dir, "b_test.go"), filepath.Join(dir, "joined_test.go")
			}
			writeMergeFixture(t, a, tc.body)
			writeMergeFixture(t, b, "package fixture\n")
			if err := merge([]string{dest, a, b}, true); err == nil {
				t.Fatal("unsafe merge was accepted")
			}
			body, err := os.ReadFile(a)
			if err != nil || string(body) != tc.body {
				t.Fatal("rejected merge changed its source")
			}
			if _, err := os.Stat(dest); !os.IsNotExist(err) {
				t.Fatal("rejected merge wrote a destination")
			}
		})
	}
}
