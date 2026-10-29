package main

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPastePendingHelpers(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	if _, ok := GetPendingCopy(); ok {
		t.Fatal("pending copy is set on a clean store")
	}

	SetPendingCopy("/tmp/outside/note.txt", false)

	got, ok := GetPendingCopy()
	if !ok || got.SrcPath != "/tmp/outside/note.txt" || got.IsCut {
		t.Fatalf("pending = %+v, %v; want the staged path, copy mode", got, ok)
	}

	SetPendingCopy("/tmp/outside/cut.txt", true)

	if got, ok := GetPendingCopy(); !ok || !got.IsCut || got.SrcPath != "/tmp/outside/cut.txt" {
		t.Fatalf("cut pending = %+v, %v", got, ok)
	}

	ClearPendingCopy()

	if _, ok := GetPendingCopy(); ok {
		t.Fatal("pending copy survives ClearPendingCopy")
	}
}

func TestPastePendingCopyDisk(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "sub/b.go", "package sub\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	srcDir := t.TempDir()
	writeFile(t, srcDir, "note.txt", "pasted\n")
	src := filepath.Join(srcDir, "note.txt")

	SetPendingCopy(src, false)

	rel, err := app.PastePendingCopy("sub")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	if rel != "sub/note.txt" {
		t.Fatalf("rel = %q, want sub/note.txt", rel)
	}

	got, err := app.ws.readFile(rel)
	if err != nil || string(got) != "pasted\n" {
		t.Fatalf("pasted bytes = %q (%v)", got, err)
	}

	if app.ws.ByPath[rel] == nil {
		t.Fatal("sub/note.txt missing from the workspace")
	}

	if findTreeNode(app.ws.Tree, rel) == nil {
		t.Fatal("sub/note.txt has no tree node")
	}

	rel2, err := app.PastePendingCopy("sub")
	if err != nil {
		t.Fatalf("second paste: %v", err)
	}

	if rel2 != "sub/note (copy).txt" {
		t.Fatalf("second rel = %q, want sub/note (copy).txt", rel2)
	}

	if _, ok := GetPendingCopy(); !ok {
		t.Fatal("copy mode stopped being staged after a paste")
	}

	rootRel, err := app.PastePendingCopy("")
	if err != nil {
		t.Fatalf("root paste: %v", err)
	}

	if rootRel != "note.txt" {
		t.Fatalf("root rel = %q, want note.txt", rootRel)
	}
}

func TestPastePendingCopyCutClears(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	srcDir := t.TempDir()
	writeFile(t, srcDir, "cut.txt", "cut\n")

	SetPendingCopy(filepath.Join(srcDir, "cut.txt"), true)

	if _, err := app.PastePendingCopy(""); err != nil {
		t.Fatalf("cut paste: %v", err)
	}

	if _, ok := GetPendingCopy(); ok {
		t.Fatal("cut staging survives its paste")
	}
}

func TestPastePendingCopyRelativeStaged(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	stageCopy("a.go", false)

	rel, err := app.PastePendingCopy("")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	if rel != "a (copy).go" {
		t.Fatalf("rel = %q, want a (copy).go", rel)
	}

	got, err := app.ws.readFile(rel)
	if err != nil || string(got) != "package main\n" {
		t.Fatalf("pasted bytes = %q (%v)", got, err)
	}
}

func TestPastePendingCopySampleGuard(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	app := newTestApp(t)
	before := len(app.ws.Files)

	srcDir := t.TempDir()
	writeFile(t, srcDir, "note.txt", "nope\n")
	SetPendingCopy(filepath.Join(srcDir, "note.txt"), false)

	_, err := app.PastePendingCopy("")
	if err == nil || err.Error() != sampleReadOnlyMessage {
		t.Fatalf("paste error = %v, want %q", err, sampleReadOnlyMessage)
	}

	if len(app.ws.Files) != before {
		t.Fatal("sample workspace changed after a guarded paste")
	}

	if _, ok := GetPendingCopy(); !ok {
		t.Fatal("guarded paste consumed the staging")
	}
}

func TestPastePendingCopyMissingSource(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	SetPendingCopy(filepath.Join(dir, "gone.txt"), false)

	rel, err := app.PastePendingCopy("")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	if rel != "gone.txt" {
		t.Fatalf("rel = %q, want gone.txt", rel)
	}

	got, err := app.ws.readFile(rel)
	if err != nil || len(got) != 0 {
		t.Fatalf("missing source pasted %d bytes (%v), want an empty file", len(got), err)
	}
}

func TestPastePendingCopyBinary(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	dir := t.TempDir()
	writeFile(t, dir, "sub/b.go", "package sub\n")

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	raw := []byte{0x89, 'P', 'N', 'G', 0x00, 0x01, 0x02, 0xff}
	srcDir := t.TempDir()
	writeFile(t, srcDir, "pic.png", string(raw))

	SetPendingCopy(filepath.Join(srcDir, "pic.png"), false)

	rel, err := app.PastePendingCopy("sub")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	if rel != "sub/pic.png" {
		t.Fatalf("rel = %q, want sub/pic.png", rel)
	}

	got, err := app.ws.readFile(rel)
	if err != nil || !bytes.Equal(got, raw) {
		t.Fatalf("pasted bytes = %v (%v), want the raw PNG bytes", got, err)
	}

	if app.ws.ByPath[rel] == nil {
		t.Fatal("sub/pic.png missing from the workspace")
	}

	if findTreeNode(app.ws.Tree, rel) == nil {
		t.Fatal("sub/pic.png has no tree node")
	}
}

func TestPastePendingCopyFullScan(t *testing.T) {
	ClearPendingCopy()
	defer ClearPendingCopy()

	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "sub/b.go", "package sub\n")

	ws, err := loadDiskWorkspace(dir)
	if err != nil {
		t.Fatalf("loadDiskWorkspace: %v", err)
	}

	app := &App{ws: ws}

	srcDir := t.TempDir()
	writeFile(t, srcDir, "note.txt", "full\n")
	SetPendingCopy(filepath.Join(srcDir, "note.txt"), false)

	rel, err := app.PastePendingCopy("sub")
	if err != nil {
		t.Fatalf("paste: %v", err)
	}

	if rel != "sub/note.txt" {
		t.Fatalf("rel = %q, want sub/note.txt", rel)
	}

	if findTreeNode(ws.Tree, rel) == nil {
		t.Fatal("sub/note.txt has no tree node after the full-scan finish")
	}
}

func TestResolvePasteDirFocus(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	if err := app.Page().Form().Focus(ctx, ""); err != nil {
		t.Fatalf("clear focus: %v", err)
	}

	if got := app.ResolvePasteDir(""); got != "" {
		t.Fatalf("root fallback = %q, want the workspace root", got)
	}

	if err := app.Page().Form().Focus(ctx, "tree-backend"); err != nil {
		t.Fatalf("focus dir: %v", err)
	}

	if got := app.ResolvePasteDir(""); got != "backend" {
		t.Fatalf("focused dir = %q, want backend", got)
	}

	if err := app.Page().Form().Focus(ctx, "tree-backend/probe.go"); err != nil {
		t.Fatalf("focus file: %v", err)
	}

	if got := app.ResolvePasteDir(""); got != "backend" {
		t.Fatalf("focused file parent = %q, want backend", got)
	}

	if got := app.ResolvePasteDir("web/index.html"); got != "web" {
		t.Fatalf("explicit file = %q, want web", got)
	}

	if got := app.ResolvePasteDir("web"); got != "web" {
		t.Fatalf("explicit dir = %q, want web", got)
	}

	if err := app.Page().Form().Focus(ctx, "tree-backend"); err != nil {
		t.Fatalf("focus dir: %v", err)
	}

	if got := app.ResolvePasteDir("../evil"); got != "backend" {
		t.Fatalf("escape fallback = %q, want the focused backend dir", got)
	}

	if err := app.Page().Form().Focus(ctx, ""); err != nil {
		t.Fatalf("clear focus: %v", err)
	}

	if got := app.ResolvePasteDir("../evil"); got != "" {
		t.Fatalf("escape with no focus = %q, want the workspace root", got)
	}
}

func TestResolvePasteDirDiskFallback(t *testing.T) {
	dir := t.TempDir()

	app, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	// Created after the scan, so the tree never listed it.
	if err := os.Mkdir(filepath.Join(dir, "fresh"), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if got := app.ResolvePasteDir("fresh"); got != "fresh" {
		t.Fatalf("unscanned dir = %q, want fresh", got)
	}

	if got := app.ResolvePasteDir("fresh/nope.txt"); got != "" {
		t.Fatalf("missing file fallback = %q, want the workspace root", got)
	}
}

func TestParseClipboardFileList(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.png", "PNGDATA")
	writeFile(t, dir, "sp ace.txt", "hello\n")

	aURI := "file://" + filepath.ToSlash(filepath.Join(dir, "a.png"))
	spaceURI := "file://" + strings.ReplaceAll(filepath.ToSlash(filepath.Join(dir, "sp ace.txt")), " ", "%20")

	text := "# text/uri-list comment\n" + aURI + "\n" + spaceURI + "\n\n" +
		"file:///nonexistent-missing.png\nrelative/path.txt\n../traversal.txt\n"

	got := ParseClipboardFileList(text)
	want := []string{filepath.Join(dir, "a.png"), filepath.Join(dir, "sp ace.txt")}

	if len(got) != len(want) {
		t.Fatalf("files = %q, want %q", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("files = %q, want %q", got, want)
		}
	}
}

func TestParseClipboardWSL(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "TestDistro")

	dir := t.TempDir()
	writeFile(t, dir, "w.png", "data")
	abs := filepath.Join(dir, "w.png")

	unc := `\\wsl$\TestDistro` + strings.ReplaceAll(filepath.ToSlash(abs), "/", "\\")

	if got := ParseClipboardFileList(unc + "\n"); len(got) != 1 || got[0] != abs {
		t.Fatalf("wsl unc = %q, want [%q]", got, abs)
	}

	foreign := `\\wsl$\OtherDistro` + strings.ReplaceAll(filepath.ToSlash(abs), "/", "\\")

	if got := ParseClipboardFileList(foreign + "\n"); len(got) != 0 {
		t.Fatalf("foreign distro = %q, want it dropped", got)
	}

	if got := ParseClipboardFileList("C:\\Users\\nope-missing-12345.png\n"); len(got) != 0 {
		t.Fatalf("missing drive path = %q, want it dropped", got)
	}
}

func TestParseClipboardTranslation(t *testing.T) {
	t.Setenv("WSL_DISTRO_NAME", "TestDistro")

	cases := []struct {
		line string
		want string
		ok   bool
	}{
		{`C:\Users\pic.png`, "/mnt/c/Users/pic.png", true},
		{`c:/a/b.jpg`, "/mnt/c/a/b.jpg", true},
		{`\\wsl$\TestDistro\home\u\f.png`, "/home/u/f.png", true},
		{"", "", false},
		{"# comment", "", false},
		{"relative/x.png", "", false},
		{"/a/../../b", "", false},
		{"file://remotehost/s/f.png", "", false},
		{"file:garbage", "", false},
		{`\\server\share\x`, "", false},
	}

	for _, tc := range cases {
		got, ok := translateClipboardPath(tc.line)
		if ok != tc.ok || got != tc.want {
			t.Errorf("translateClipboardPath(%q) = %q, %v; want %q, %v", tc.line, got, ok, tc.want, tc.ok)
		}
	}
}

func TestParseClipboardRejects(t *testing.T) {
	inputs := []string{
		"",
		"   \n  \n",
		"# comment\n",
		"relative/path.png\n",
		"../traversal.txt\n",
		"a/../../b.png\n",
		"file://remotehost/share/f.png\n",
		"file:garbage\n",
		"file:///nonexistent-missing-12345.png\n",
		`C:\definitely\missing-12345.png` + "\n",
	}

	for _, in := range inputs {
		if got := ParseClipboardFileList(in); len(got) != 0 {
			t.Errorf("ParseClipboardFileList(%q) = %q, want empty", in, got)
		}
	}
}
