package main

import (
	"os/exec"
	"testing"
)

// TestRepoIgnoredFileVisible builds a small git repository with an ignored
// *.test fixture and checks the scan keeps it visible in both the file index
// and the tree, flagged as ignored. The fixture lives in a temp repo so the
// test never depends on whatever untracked files happen to sit in the
// enclosing checkout.
func TestRepoIgnoredFileVisible(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q")

	writeFile(t, dir, ".gitignore", "*.test\n")
	writeFile(t, dir, "motion.test", "ignored\n")
	writeFile(t, dir, "main.go", "package main\n\nfunc main() {}\n")

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatal(err)
	}

	if ws.ByPath["motion.test"] == nil {
		t.Fatal("motion.test missing from scan")
	}

	if !ws.ByPath["motion.test"].Ignored {
		t.Error("motion.test has Ignored=false, want true")
	}

	t.Logf("meta ignored=%v", ws.ByPath["motion.test"].Ignored)

	var found bool
	for _, node := range ws.Tree {
		if node.Path == "motion.test" {
			found = true
			t.Logf("node ignored=%v", node.Ignored)

			if !node.Ignored {
				t.Error("tree node motion.test has Ignored=false, want true")
			}
		}
	}

	if !found {
		t.Fatal("motion.test missing from tree")
	}

	if ws.ByPath["main.go"] == nil {
		t.Fatal("main.go missing from scan")
	}

	if ws.ByPath["main.go"].Ignored {
		t.Error("main.go has Ignored=true, want false")
	}
}
