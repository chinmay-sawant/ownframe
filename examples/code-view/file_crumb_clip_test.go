package main

import (
	"context"
	"math"
	"strings"
	"testing"
)

func TestClipFileCrumbPath(t *testing.T) {
	name := "a_very_long_component_filename_for_wrap_check.tsx"
	dir := "some/quite/long/nested/directory/structure/with_many_levels"

	if got, want := len([]rune(name)), 49; got != want {
		t.Fatalf("fixture name is %d runes, want %d", got, want)
	}

	short, dshort := clipFileCrumbPath(886, 1280, "probe.go", "backend")
	if short != "probe.go" || dshort != "backend" {
		t.Fatalf("short path clipped to %q %q", short, dshort)
	}

	short, dshort = clipFileCrumbPath(886, 1280, name, dir)
	if short != name {
		t.Fatalf("name clipped to %q, want full name", short)
	}

	if !strings.HasSuffix(dshort, "…") {
		t.Fatalf("dir %q has no ellipsis", dshort)
	}

	// budget = (886 - 23 - 400) / 6 = 77; the 49-rune name fits, so the
	// directory keeps 77 - 49 - 2 runes plus the ellipsis.
	if got, want := len([]rune(dshort)), 77-49-2+1; got != want {
		t.Fatalf("dir is %d runes, want %d", got, want)
	}

	if got, dshort := clipFileCrumbPath(406, 720, name, dir); dshort != "" || !strings.HasSuffix(got, "…") {
		t.Fatalf("narrow clipped to %q %q, want clipped name and empty dir", got, dshort)
	}

	if got, dshort := clipFileCrumbPath(886, 1280, "", ""); got != "" || dshort != "" {
		t.Fatalf("empty path = %q %q", got, dshort)
	}
}

func TestClipSecondCrumbPath(t *testing.T) {
	short, dshort := clipSecondCrumbPath(886, 1280, "probe.go", "backend")
	if short != "probe.go" || dshort != "backend" {
		t.Fatalf("short path clipped to %q %q", short, dshort)
	}

	// budget = (442 - 23 - 190) / 6 = 38; the 49-rune name alone exceeds
	// it, so the name clips and the directory drops.
	name := "a_very_long_component_filename_for_wrap_check.tsx"
	short, dshort = clipSecondCrumbPath(442, 1280, name, "some/dir")
	if dshort != "" || !strings.HasSuffix(short, "…") {
		t.Fatalf("narrow second pane clipped to %q %q", short, dshort)
	}

	if got, want := len([]rune(short)), 38; got != want {
		t.Fatalf("short is %d runes, want %d", got, want)
	}
}

func TestClipTreeName(t *testing.T) {
	if got := clipTreeName(drawerMaxW, treePad, false, false, "probe.go"); got != "probe.go" {
		t.Fatalf("short name clipped to %q", got)
	}

	dir := "some very long directory name with spaces and extra words to force clipping"
	// overhead = 10 + 2 + 18 + 7, avail = 348 - 8 - 37, budget = 303 / 6.
	short := clipTreeName(drawerMaxW, treePad, true, false, dir)
	if !strings.HasSuffix(short, "…") {
		t.Fatalf("dir %q has no ellipsis", short)
	}

	if got, want := len([]rune(short)), 50; got != want {
		t.Fatalf("dir is %d runes, want %d", got, want)
	}

	deeper := clipTreeName(drawerMaxW, treePad+treeIndent, true, false, dir)
	if len([]rune(deeper)) >= len([]rune(short)) {
		t.Fatalf("deeper indent %q is not shorter than %q", deeper, short)
	}

	asFile := clipTreeName(drawerMaxW, treePad, false, false, dir)
	if len([]rune(asFile)) >= len([]rune(short)) {
		t.Fatalf("file row %q is not shorter than dir row %q", asFile, short)
	}

	withNote := clipTreeName(drawerMaxW, treePad, true, true, dir)
	if len([]rune(withNote)) >= len([]rune(short)) {
		t.Fatalf("noted row %q is not shorter than %q", withNote, short)
	}

	if got := clipTreeName(drawerMaxW, 1000, false, false, dir); got != "…" {
		t.Fatalf("hopeless indent clipped to %q, want ellipsis", got)
	}
}

func TestFileCrumbPathStaysOnOneLine(t *testing.T) {
	dir := t.TempDir()
	longDir := "some very long directory name with spaces and extra words to force clipping/nested dir part"
	longName := "animation card play full test.lock"
	rel := longDir + "/" + longName
	writeFile(t, dir, rel, "hello\nworld\n")
	writeFile(t, dir, "a.txt", "x\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.openFile(rel)

	for _, size := range [][2]int{{1200, 800}, {720, 540}} {
		app.Page().SetSize(size[0], size[1])

		if err := app.Page().Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		view := app.View()
		box, ok := boxByID(app.Page().Boxes(), "crumb-path")
		if !ok {
			t.Fatal("missing crumb-path")
		}

		if box.H > 20 {
			t.Fatalf("size %dx%d: crumb-path is %.0f px tall, want one 20 px line", size[0], size[1], box.H)
		}

		display := app.Page().Probe().Display()
		if display == nil {
			t.Fatal("page fell back to a bitmap")
		}

		scale := display.PixelPerPoint
		if scale == 0 {
			scale = 96.0 / 72.0
		}

		if size[0] == 1200 {
			if view.ActiveNameShort != longName {
				t.Fatalf("NameShort = %q, want full name", view.ActiveNameShort)
			}

			if !strings.HasSuffix(view.ActiveDirShort, "…") {
				t.Fatalf("DirShort = %q, want ellipsis", view.ActiveDirShort)
			}

			var nameY, dirY, nameX, dirX float64
			foundName, foundDir := false, false

			for _, op := range display.Ops {
				if op.Text == longName {
					nameX, nameY, foundName = op.X*scale, op.Y*scale, true
				}

				if op.Text == view.ActiveDirShort+"/" || op.Text == view.ActiveDirShort {
					if !foundDir || op.X*scale < dirX {
						dirX, dirY, foundDir = op.X*scale, op.Y*scale, true
					}
				}
			}

			if !foundName {
				t.Fatalf("full name %q not painted in the header", longName)
			}

			if !foundDir {
				t.Fatalf("directory %q not painted in the header", view.ActiveDirShort)
			}

			if math.Abs(dirY-nameY) > 0.5 {
				t.Fatalf("dir y=%.1f and name y=%.1f, want one line", dirY, nameY)
			}

			if dirX >= nameX {
				t.Fatalf("dir x=%.1f is not left of name x=%.1f", dirX, nameX)
			}
		} else {
			if view.ActiveDirShort != "" {
				t.Fatalf("narrow DirShort = %q, want empty", view.ActiveDirShort)
			}

			if !strings.HasSuffix(view.ActiveNameShort, "…") {
				t.Fatalf("narrow NameShort = %q, want ellipsis", view.ActiveNameShort)
			}

			var nameY float64
			foundName := false

			for _, op := range display.Ops {
				if op.Text == view.ActiveNameShort {
					if foundName {
						t.Fatalf("clipped name %q painted twice, want one line", view.ActiveNameShort)
					}

					nameY, foundName = op.Y*scale, true
				}
			}

			if !foundName {
				t.Fatalf("clipped name %q not painted in the header", view.ActiveNameShort)
			}

			_ = nameY
		}

		var treeShort string

		topDir := strings.Split(longDir, "/")[0]
		for _, row := range view.Tree {
			if row.Path == topDir {
				treeShort = row.Short
			}
		}

		if treeShort == "" {
			t.Fatalf("size %dx%d: long dir row missing from the tree", size[0], size[1])
		}

		if !strings.HasSuffix(treeShort, "…") {
			t.Fatalf("size %dx%d: tree Short = %q, want ellipsis", size[0], size[1], treeShort)
		}

		var wantY float64
		seen := false

		for _, op := range display.Ops {
			text := op.Text
			if text == "" || op.X*scale >= 314 {
				continue
			}

			if text != treeShort && (len(text) < 2 || !strings.Contains(topDir, text)) {
				continue
			}

			if !seen {
				wantY, seen = op.Y*scale, true
			} else if math.Abs(op.Y*scale-wantY) > 0.5 {
				t.Fatalf("size %dx%d: tree text %q at y=%.1f, want one line at y=%.1f", size[0], size[1], text, op.Y*scale, wantY)
			}
		}

		if !seen {
			t.Fatalf("size %dx%d: clipped tree name %q not painted", size[0], size[1], treeShort)
		}
	}
}
