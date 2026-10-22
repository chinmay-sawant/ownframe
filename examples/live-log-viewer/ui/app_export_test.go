package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportSelectionWritesCSV(t *testing.T) {
	ff := &fakeFeed{}
	ff.seed(10)
	dir := t.TempDir()
	a := newApp(t, ff, Options{ExportDir: dir})

	pumpUntil(t, a, func() bool { return a.pager.Len() == 10 }, "page")

	a.selected = 3
	a.selAnchor = 6
	a.startExport()
	pumpUntil(t, a, func() bool { return !a.exporting && strings.Contains(a.note, "exported") }, "export")

	files, err := os.ReadDir(dir)
	if err != nil || len(files) != 1 {
		t.Fatalf("export dir = %v %v", files, err)
	}

	data, err := os.ReadFile(filepath.Join(dir, files[0].Name())) //nolint:gosec // reason: test reads files it just exported into TempDir
	if err != nil {
		t.Fatal(err)
	}

	if lines := strings.Count(string(data), "\n"); lines != 4 {
		t.Fatalf("exported lines = %d", lines)
	}
}
