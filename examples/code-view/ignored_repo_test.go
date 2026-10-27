package main

import (
	"testing"
)

func TestRepoIgnoredFileVisible(t *testing.T) {
	ws, err := loadDiskWorkspace("../..")
	if err != nil {
		t.Fatal(err)
	}

	if ws.ByPath["motion.test"] == nil {
		t.Fatal("motion.test missing from scan")
	}

	t.Logf("meta ignored=%v", ws.ByPath["motion.test"].Ignored)

	var found bool
	for _, node := range ws.Tree {
		if node.Path == "motion.test" {
			found = true
			t.Logf("node ignored=%v", node.Ignored)
		}
	}

	if !found {
		t.Fatal("motion.test missing from tree")
	}
}
