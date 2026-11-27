package ui

import (
	"bytes"
	"context"
	"image/color"
	"image/png"
	"regexp"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
)

func TestThemeTokensPresent(t *testing.T) {
	tokens := []string{
		"brand", "brand-hover", "brand-tint", "titlebar-bg", "titlebar-fg", "rail-bg", "sidebar-bg", "canvas-bg",
		"surface-bg", "surface-elevated", "border", "border-strong", "text", "text-muted", "row-alt", "row-hover",
		"cell-selected", "null-bg", "null-fg", "diff-edit-bg", "diff-insert-bg", "diff-delete-bg",
		"status-ok", "status-warn", "status-err",
		"syn-kw", "syn-fn", "syn-str", "syn-num", "syn-type", "syn-comment", "syn-op",
		"env-local", "env-staging", "env-prod",
	}
	hex := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	for name, css := range map[string]string{"dark": darkCSS, "light": lightCSS} {
		for _, tok := range tokens {
			m := regexp.MustCompile(`--qr-` + regexp.QuoteMeta(tok) + `:\s*([^;]+);`).FindStringSubmatch(css)
			if m == nil {
				t.Errorf("%s.css lacks --qr-%s", name, tok)
				continue
			}
			if !hex.MatchString(strings.TrimSpace(m[1])) {
				t.Errorf("%s.css --qr-%s = %q is not a hex colour", name, tok, m[1])
			}
		}
	}
	if baseCSS == "" || strings.Contains(baseCSS, "#") && regexp.MustCompile(`(?m)^\s*--qr-`).MatchString(baseCSS) {
		t.Error("base.css must hold geometry only, no token definitions")
	}
	newRig(t, 1440, 960, "")
}

func TestShellFitsViewport(t *testing.T) {
	for _, sz := range [][2]int{{1440, 960}, {1280, 800}} {
		r := newRig(t, sz[0], sz[1], "")
		w, h := sz[0], sz[1]
		want := map[string][4]int{ // id: x, y, w, h
			"titlebar":  {0, 0, w, titleBarHeight},
			"rail":      {0, titleBarHeight, railWidth, h - titleBarHeight - statusHeight},
			"sidebar":   {railWidth, titleBarHeight, 248, h - titleBarHeight - statusHeight},
			"statusbar": {0, h - statusHeight, w, statusHeight},
		}
		for id, g := range want {
			b := r.box(id)
			if int(b.X) != g[0] || int(b.Y) != g[1] || int(b.W) != g[2] || int(b.H) != g[3] {
				t.Errorf("%dx%d %s = %v,%v %vx%v, want %v", w, h, id, b.X, b.Y, b.W, b.H, g)
			}
		}
		ws := r.box("workspace")
		if int(ws.X+ws.W) > w || int(ws.Y+ws.H) > h-statusHeight {
			t.Errorf("workspace overflows: %+v", ws)
		}
		tabs := r.box("tab-strip")
		if int(tabs.H) != 34 {
			t.Errorf("tab bar height %v, want 34", tabs.H)
		}
		for _, b := range r.a.Page().Boxes() {
			if int(b.X+b.W) > w+1 || int(b.Y+b.H) > h+1 && !strings.HasPrefix(b.ID, "grid") && !strings.HasPrefix(b.ID, "ed-") {
				t.Errorf("%dx%d box %q overflows the viewport: %+v", w, h, b.ID, b)
			}
		}
	}
}

func pixelAt(r *rig, x, y int) color.RGBA {
	img, err := png.Decode(bytes.NewReader(r.a.Page().Probe().PNG()))
	if err != nil {
		r.t.Fatal(err)
	}
	c, _ := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	return c
}

func TestDefaultDarkAndThemeToggle(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	if !r.a.Dark() {
		t.Fatal("must start dark")
	}
	dark := pixelAt(r, 700, 250) // editor canvas
	if want := (color.RGBA{0x1c, 0x1c, 0x1c, 0xff}); dark != want {
		t.Fatalf("dark canvas = %v, want %v", dark, want)
	}
	tab := r.a.newQueryTab("SELECT 42")
	r.a.activeTab = len(r.a.tabs) - 1
	_ = tab
	r.must(r.a.redraw(r.ctx))
	nTabs, text := len(r.a.tabs), r.a.activeTabPtr().ed.text

	r.tap("tb-theme")
	if r.a.Dark() {
		t.Fatal("toggle did not switch to light")
	}
	light := pixelAt(r, 700, 250)
	if want := (color.RGBA{0xff, 0xff, 0xff, 0xff}); light != want {
		t.Fatalf("light canvas = %v, want %v", light, want)
	}
	if light == dark {
		t.Fatal("light and dark canvas identical")
	}
	r.tap("tb-theme")
	if !r.a.Dark() || pixelAt(r, 700, 250) != dark {
		t.Fatal("second toggle did not restore dark")
	}
	if len(r.a.tabs) != nTabs || r.a.activeTabPtr().ed.text != text {
		t.Fatal("toggle lost tabs or query text")
	}
	if v, _ := r.a.loadPref("theme"); v != "dark" {
		t.Fatalf("theme pref = %q", v)
	}
}

func TestTitleBarActions(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	for _, act := range []string{"win-min", "win-max", "win-close"} {
		b := r.box(act)
		if b.Action != act {
			t.Errorf("%s data-action = %q", act, b.Action)
		}
	}
	for _, id := range []string{"tb-conn", "tb-palette", "tb-theme"} {
		r.box(id)
	}
	before := r.a.dark
	r.tap("tb-theme")
	if r.a.dark == before {
		t.Fatal("theme toggle did not repaint")
	}
	tree, err := r.a.Page().Access().AccessibilityTree()
	r.must(err)
	named := false
	for _, n := range tree.Nodes {
		if n.ElementID == "tb-theme" && strings.Contains(n.Name, "Dark") {
			named = true
		}
	}
	if !named {
		t.Fatal("theme button should now offer the Dark theme in its accessible name")
	}
}

func TestDragGuardYieldsToOverlays(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	a := r.a
	if !a.Draggable(400, 10) {
		t.Fatal("title strip should drag")
	}
	if a.Draggable(400, 100) {
		t.Fatal("body must not drag")
	}
	if a.Draggable(1440-20, 10) {
		t.Fatal("window controls must not drag")
	}
	for _, open := range []func(){
		func() { a.modal = "conn" },
		func() { a.menuOpen = "conn" },
		func() { a.ac.open = true },
	} {
		open()
		if a.Draggable(400, 10) {
			t.Fatal("drag must yield to an open overlay")
		}
		a.modal, a.menuOpen, a.ac.open = "", "", false
	}
}

func TestStatusBarReadout(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	for _, id := range []string{"st-state", "st-version", "st-db", "st-theme"} {
		r.box(id)
	}
	txt := r.texts()
	for _, want := range []string{"Connected", "SQLite", "QuarryDark"} {
		if !strings.Contains(txt, want) {
			t.Errorf("status bar missing %q in %q", want, txt)
		}
	}
	r.a.activeTabPtr().ed = newEditor("SELECT * FROM users")
	r.chord("enter", false)
	r.settle()
	if !r.has("st-last") || !strings.Contains(r.texts(), "500 rows") {
		t.Fatalf("last query readout missing: %q", r.texts())
	}
	r.tap("tb-theme")
	if !strings.Contains(r.texts(), "QuarryLight") {
		t.Fatal("theme indicator did not follow the toggle")
	}
}

func TestConnectionModalFlow(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.tap("rail-connections")
	if !r.has("conn-form") {
		t.Fatal("connection manager did not open")
	}
	r.tap("cd-engine-postgresql")
	if !r.has("cd-host") || r.has("cd-path") {
		t.Fatal("PostgreSQL form fields expected")
	}
	r.tap("cd-engine-sqlite")
	if !r.has("cd-path") || r.has("cd-host") {
		t.Fatal("SQLite form fields expected")
	}
	// Reject empty required fields.
	r.a.page.Form().SetFormValue("cd-name", "Edge")
	r.a.page.Form().SetFormValue("cd-path", "")
	r.tap("conn-save")
	if r.a.connDraft.ResultOK || r.a.connDraft.Result == "" {
		t.Fatalf("empty path should be rejected, got %q", r.a.connDraft.Result)
	}
	// Valid save.
	r.a.page.Form().SetFormValue("cd-path", r.a.cfg.FilePath)
	r.tap("conn-save")
	if !r.a.connDraft.ResultOK {
		t.Fatalf("save failed: %s", r.a.connDraft.Result)
	}
	if got := r.a.savedConnections(); len(got) != 1 || got[0].Name != "Edge" {
		t.Fatalf("saved = %+v", got)
	}
	r.tap("conn-test")
	if !r.a.connDraft.ResultOK || !strings.Contains(r.a.connDraft.Result, "Connected") {
		t.Fatalf("test connection: %q", r.a.connDraft.Result)
	}
	// Prod defaults to read-only.
	r.tap("cd-env-prod")
	if !r.a.connDraft.ReadOnly {
		t.Fatal("Prod must default to read-only")
	}
}

func TestDropSQLiteFileOpensConnection(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	other, err := driver.EnsureDemoDB(r.ctx, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	r.must(r.a.onDrop(r.ctx, []ownframe.Drop{{Name: "other.sqlite", Path: other}}))
	if r.a.cfg.FilePath != other {
		t.Fatalf("active connection = %q, want %q", r.a.cfg.FilePath, other)
	}
	found := false
	for _, sc := range r.a.savedConnections() {
		found = found || sc.FilePath == other
	}
	if !found {
		t.Fatal("dropped file not added to recent connections")
	}
	r.must(r.a.onDrop(r.ctx, []ownframe.Drop{{Name: "notes.txt", Path: "/tmp/notes.txt"}}))
	if r.a.cfg.FilePath != other {
		t.Fatal("non-sqlite drop must be ignored")
	}
}

var _ = context.Background
