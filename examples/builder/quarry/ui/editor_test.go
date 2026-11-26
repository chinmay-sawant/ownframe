package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/store"
)

// focusEditor presses inside the editor body so typing reaches it.
func (r *rig) focusEditor() {
	r.t.Helper()
	b := r.box("ed-line-0")
	x, y := b.X+b.W-20, b.Y+b.H/2
	r.must(r.a.Page().Edit().Press(r.ctx, x, y))
	r.must(r.a.Page().Click(r.ctx, x, y))
	r.must(r.a.Page().Edit().Release(r.ctx))
}

func TestEditorTypingCaretAndUndo(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.focusEditor()
	r.typ("SELECT 1;\nSELECT 2")
	ed := r.a.activeTabPtr().ed
	if ed.text != "SELECT 1;\nSELECT 2" || ed.lineCount() != 2 {
		t.Fatalf("text = %q", ed.text)
	}
	r.key("arrowup")
	if l, _ := ed.lineCol(ed.caret); l != 0 {
		t.Fatalf("caret line %d after ArrowUp", l)
	}
	r.key("end")
	r.key("shift+arrowleft")
	if ed.selectionText() != ";" {
		t.Fatalf("selection = %q", ed.selectionText())
	}
	r.must(r.a.onSelectAll(r.ctx))
	if got, ok, _ := r.a.onCopy(r.ctx); !ok || got != ed.text {
		t.Fatalf("copy = %q %v", got, ok)
	}
	r.must(r.a.onBackspace(r.ctx))
	if ed.text != "" {
		t.Fatalf("backspace over selection left %q", ed.text)
	}
	r.must(r.a.onUndo(r.ctx))
	if ed.text != "SELECT 1;\nSELECT 2" {
		t.Fatalf("undo = %q", ed.text)
	}
	r.must(r.a.onRedo(r.ctx))
	if ed.text != "" {
		t.Fatal("redo did not reapply")
	}
	r.must(r.a.onUndo(r.ctx))
	// Clicking a line positions the caret there.
	_, err := r.a.onPress(r.ctx, r.box("ed-line-1"), 1000, r.box("ed-line-1").Y+9)
	r.must(err)
	if l, c := ed.lineCol(ed.caret); l != 1 || c != len("SELECT 2") {
		t.Fatalf("click caret = %d,%d", l, c)
	}
	if !r.a.activeTabPtr().dirty {
		t.Fatal("editing should mark the tab dirty")
	}
	r.png("editor")
}

func TestSyntaxHighlightColoursPaint(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.a.activeTabPtr().ed = newEditor("SELECT count(*) FROM users WHERE id > 5 -- note")
	r.must(r.a.redraw(r.ctx))
	colours := map[string]uint32{}
	for _, op := range r.a.Page().Probe().Display().Ops {
		if op.Kind == ownframe.DisplayOpText {
			s := strings.TrimSpace(op.Text)
			switch s {
			case "SELECT", "count", "5", "-- note":
				colours[s] = uint32(op.R*255+0.5)<<16 | uint32(op.G*255+0.5)<<8 | uint32(op.B*255+0.5)
			}
		}
	}
	if len(colours) != 4 {
		t.Fatalf("expected four token runs, got %v", colours)
	}
	seen := map[uint32]bool{}
	for _, c := range colours {
		seen[c] = true
	}
	if len(seen) != 4 {
		t.Fatalf("keyword, function, number and comment need distinct colours: %v", colours)
	}
	if colours["SELECT"] != 0x8ab4f8 || colours["-- note"] != 0x6b6b6b {
		t.Fatalf("keyword/comment colours not the QuarryDark tokens: %v", colours)
	}
}

func TestEditorToolbarAndShortcuts(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT 1 AS a;\nSELECT 2 AS b;\nSELECT 3 AS c")
	// Caret on line 2: Ctrl+Enter runs only that statement.
	tab.ed.caret = tab.ed.offsetAt(1, 3)
	tab.ed.anchor = tab.ed.caret
	r.chord("enter", false)
	r.settle()
	if tab.grid == nil || tab.grid.cols[0].Name != "b" || len(tab.grid.rows) != 1 {
		t.Fatalf("statement at caret not run: %+v", tab.grid)
	}
	// Ctrl+Shift+Enter runs the whole script; messages list all three.
	r.chord("enter", true)
	r.settle()
	if len(tab.messages) != 3 || tab.grid.cols[0].Name != "c" {
		t.Fatalf("script run: msgs=%v", tab.messages)
	}
	// Toolbar Run button.
	tab.ed.caret, tab.ed.anchor = 0, 0
	r.tap("ed-run")
	r.settle()
	if tab.grid.cols[0].Name != "a" {
		t.Fatalf("Run button: %v", tab.grid.cols[0].Name)
	}
	// Run Selection runs only the selected text.
	tab.ed.anchor, tab.ed.caret = tab.ed.offsetAt(2, 0), len(tab.ed.text)
	r.tap("ed-run-sel")
	r.settle()
	if tab.grid.cols[0].Name != "c" {
		t.Fatalf("Run Selection: %v", tab.grid.cols[0].Name)
	}
	// Row limit selector.
	r.tap("ed-limit-100")
	if tab.rowLimit != 100 {
		t.Fatal("limit selector")
	}
	// Format uppercases and breaks clauses.
	tab.ed = newEditor("select id from users where id > 1 order by id")
	r.tap("ed-format")
	if !strings.Contains(tab.ed.text, "SELECT id\nFROM users\nWHERE id > 1\nORDER BY id") {
		t.Fatalf("format = %q", tab.ed.text)
	}
	// Ctrl+Space with F-key alternates.
	tab.ed = newEditor("SELECT 7 AS seven")
	r.key("f5")
	r.settle()
	if tab.grid.cols[0].Name != "seven" {
		t.Fatal("F5 should run")
	}
	r.chord("k", false)
	if r.a.modal != "palette" {
		t.Fatal("Ctrl+K opens the palette")
	}
}

func TestAutocompletePopoverSelection(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.focusEditor()
	r.typ("SELECT * FROM us")
	if !r.a.ac.open || !r.has("ac-popover") || !r.has("ac-0") {
		t.Fatal("popover should open while typing an identifier")
	}
	if r.a.ac.items[0].Label != "users" {
		t.Fatalf("top completion = %+v", r.a.ac.items[0])
	}
	if !strings.Contains(r.texts(), "TBL") {
		t.Fatal("kind badge TBL missing")
	}
	r.key("arrowdown")
	r.key("arrowup")
	r.key("enter")
	if got := r.a.activeTabPtr().ed.text; got != "SELECT * FROM users" {
		t.Fatalf("buffer = %q", got)
	}
	if r.a.ac.open {
		t.Fatal("popover should close after accept")
	}
	// Escape dismisses, Ctrl+Space reopens, click accepts.
	r.typ(" WHERE ")
	r.chord("space", false)
	if !r.a.ac.open {
		t.Fatal("Ctrl+Space should open the popover")
	}
	r.key("escape")
	if r.a.ac.open {
		t.Fatal("Escape should close the popover")
	}
	r.focusEditor()
	r.typ("na")
	r.tap("ac-0")
	if !strings.HasSuffix(r.a.activeTabPtr().ed.text, "WHERE ") && r.a.ac.open {
		t.Fatal("click accept")
	}
}

func TestAsyncQueryAndCancel(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("WITH RECURSIVE c(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM c WHERE x < 900000000) SELECT count(*) FROM c")
	r.chord("enter", false)
	if r.a.running == nil {
		t.Fatal("run should be in flight")
	}
	if !r.has("ed-cancel") || !r.has("ed-running") {
		t.Fatal("cancel button and spinner timer expected")
	}
	// The UI stays responsive while it runs: a theme toggle works.
	r.tap("tb-theme")
	r.tap("tb-theme")
	r.now = r.now.Add(2 * time.Second)
	if !r.a.onTick(r.ctx) {
		t.Fatal("elapsed timer should dirty the UI while running")
	}
	r.tap("ed-cancel")
	r.settle()
	if tab.err == nil || !strings.Contains(strings.ToLower(tab.err.Message), "cancel") {
		t.Fatalf("expected cancellation banner, got %+v", tab.err)
	}
	if r.a.running != nil {
		t.Fatal("run state not cleared")
	}
	// Escape also cancels.
	r.chord("enter", false)
	r.key("escape")
	r.settle()
	if tab.err == nil {
		t.Fatal("escape should cancel")
	}
}

func TestQueryErrorBanner(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT 1;\nSELEKT * FROM users")
	r.chord("enter", true)
	r.settle()
	if tab.err == nil {
		t.Fatal("invalid SQL should error")
	}
	if !r.has("query-error") || !strings.Contains(r.texts(), "syntax error") {
		t.Fatalf("banner missing: %q", r.texts())
	}
	if tab.err.Line != 2 {
		t.Fatalf("error line = %d (%+v)", tab.err.Line, tab.err)
	}
	r.tap("err-jump")
	if l, _ := tab.ed.lineCol(tab.ed.caret); l != 1 {
		t.Fatalf("jump left caret on line %d", l)
	}
	b := r.box("query-error")
	rs := r.box("results-pane")
	if b.Y < rs.Y || b.Y+b.H > rs.Y+rs.H || b.W <= 0 {
		t.Fatalf("banner outside results pane: %+v vs %+v", b, rs)
	}
}

func TestMessagesSubTab(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT * FROM users;\nUPDATE users SET full_name = full_name WHERE id = 1")
	r.a.cfg.ReadOnly = false
	r.chord("enter", true)
	r.settle()
	r.tap("rs-messages")
	txt := r.texts()
	if !strings.Contains(txt, "Statement 1: 500 rows") || !strings.Contains(txt, "Statement 2: 1 rows affected") {
		t.Fatalf("messages = %q", txt)
	}
	if !r.has("msg-0") || !r.has("msg-1") {
		t.Fatal("both log entries expected")
	}
}

func TestExplainPlanView(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT * FROM orders WHERE user_id = 5")
	r.tap("ed-explain")
	r.settle()
	if tab.resultView != "explain" || tab.plan == nil {
		t.Fatalf("explain not shown: %+v", tab.err)
	}
	if !r.has("explain") || !strings.Contains(r.texts(), "Index Scan") {
		t.Fatalf("index search badge missing: %q", r.texts())
	}
	tab.ed = newEditor("SELECT * FROM audit_events WHERE payload LIKE '%x%'")
	r.tap("ed-explain")
	r.settle()
	if !strings.Contains(r.texts(), "Seq Scan") {
		t.Fatalf("full scan warning missing: %q", r.texts())
	}
	r.tap("explain-raw-toggle")
	if !r.has("explain-raw") {
		t.Fatal("raw view expected")
	}
	// Postgres JSON plan through the mock driver.
	p := newRig(t, 1440, 960, "mock:postgres")
	pt := p.a.activeTabPtr()
	pt.ed = newEditor("SELECT * FROM orders WHERE id = 1")
	p.tap("ed-analyze")
	p.settle()
	if pt.plan == nil || pt.plan.Root == nil || !strings.Contains(p.texts(), "plan") {
		t.Fatalf("postgres plan: %+v %v", pt.plan, pt.err)
	}
	p.png("explain-pg")
}

func TestQueryHistoryAndSnippets(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	tab := r.a.activeTabPtr()
	tab.ed = newEditor("SELECT count(*) AS n FROM orders")
	r.chord("enter", false)
	r.settle()
	hist, _ := r.a.prefs.ListHistory(context.Background(), r.a.cfg.ID, 10)
	if len(hist) != 1 || hist[0].SQL != "SELECT count(*) AS n FROM orders" || hist[0].Rows != 1 || hist[0].Error != "" {
		t.Fatalf("history = %+v", hist)
	}
	tab.ed = newEditor("SELEKT nope")
	r.chord("enter", false)
	r.settle()
	hist, _ = r.a.prefs.ListHistory(context.Background(), r.a.cfg.ID, 10)
	if len(hist) != 2 || hist[0].Error == "" {
		t.Fatalf("error not logged: %+v", hist)
	}
	tab.ed = newEditor("SELECT 'keep me'")
	r.tap("ed-snippet")
	if snips, _ := r.a.prefs.ListSnippets(context.Background()); len(snips) != 1 {
		t.Fatal("snippet not saved")
	}
	r.tap("rail-history")
	if !r.has("hist-list") || !r.has("hist-0") || !r.has("snippet-0") {
		t.Fatal("drawer rows missing")
	}
	r.a.histQuery = "orders"
	r.must(r.a.redraw(r.ctx))
	if r.has("hist-1") || !r.has("hist-0") {
		t.Fatal("search should narrow history to one row")
	}
	r.tap("snippet-0")
	if r.a.modal != "" || !strings.Contains(r.a.activeTabPtr().ed.text, "keep me") {
		t.Fatalf("restore failed: modal=%q text=%q", r.a.modal, r.a.activeTabPtr().ed.text)
	}
	_ = store.Memory
}
