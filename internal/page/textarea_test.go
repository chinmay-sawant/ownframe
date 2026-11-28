package page

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func paintedCaretOp(t *testing.T, p *Page) layout.DisplayOp {
	t.Helper()

	d := p.draw.Display
	if d == nil || len(d.Ops) == 0 {
		t.Fatal("no display")
	}

	op := d.Ops[len(d.Ops)-1]
	if op.Kind != DisplayOpFillRect || op.W > d.PointsPerPixel+0.01 {
		t.Fatalf("last op is not a caret: kind=%v w=%.2f", op.Kind, op.W)
	}

	return op
}

func TestTextareaEnterNewlineAndInputSubmit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := editPage(t, `
		<input id="in" type="text" value="abc" style="display:block;width:240px;height:24px">
		<textarea id="ta" style="display:block;width:240px;min-height:24px;font-size:14px">hello</textarea>
	`)

	submits := 0
	p.Handle(Handlers{Submit: func(context.Context) error {
		submits++
		return nil
	}})

	// Single-line input: Enter does not insert \n and Submit fires handler.
	if err := p.Focus(ctx, "in"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if p.FormValue("in") != "abc" {
		t.Fatalf("input value = %q, want abc", p.FormValue("in"))
	}
	if submits != 1 {
		t.Fatalf("input submits = %d, want 1", submits)
	}

	// Textarea: Enter inserts \n, does not submit, grows box height, and moves caret onto the new empty line.
	if err := p.Focus(ctx, "ta"); err != nil {
		t.Fatal(err)
	}
	beforeBox := boxByID(p, "ta")
	beforeCaret := paintedCaretOp(t, p)

	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	// Simulating the window's submitIfEnter() right after KeyDown("enter"):
	if err := p.Submit(ctx); err != nil {
		t.Fatal(err)
	}
	if submits != 1 {
		t.Fatalf("textarea plain enter submitted: submits=%d", submits)
	}
	if got := p.FormValue("ta"); got != "hello\n" {
		t.Fatalf("textarea value = %q, want %q", got, "hello\n")
	}

	afterBox := boxByID(p, "ta")
	afterCaret := paintedCaretOp(t, p)
	if afterBox.H <= beforeBox.H {
		t.Fatalf("textarea height did not grow on trailing newline: before=%.2f after=%.2f", beforeBox.H, afterBox.H)
	}
	if afterCaret.Y <= beforeCaret.Y+1 {
		t.Fatalf("trailing newline caret Y=%.2f did not move below line 1 Y=%.2f", afterCaret.Y, beforeCaret.Y)
	}
	if afterCaret.X >= beforeCaret.X-1 {
		t.Fatalf("trailing newline caret X=%.2f did not move left of end-of-line X=%.2f", afterCaret.X, beforeCaret.X)
	}

	// Typing on the new line appends after the newline.
	if err := p.Type(ctx, "world"); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("ta"); got != "hello\nworld" {
		t.Fatalf("textarea value = %q, want %q", got, "hello\nworld")
	}

	// Ctrl+Enter submits from a focused textarea.
	if err := p.KeyDown(ctx, "ctrl+enter"); err != nil {
		t.Fatal(err)
	}
	if submits != 2 {
		t.Fatalf("ctrl+enter submits = %d, want 2", submits)
	}
}

func TestTextareaVerticalArrowsHardLinesAndDesiredX(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const initial = "first line long\n\nthird line long"
	p := editPage(t, `<textarea id="ta" style="display:block;width:280px;font-size:14px">`+initial+`</textarea>`)

	if err := p.Focus(ctx, "ta"); err != nil {
		t.Fatal(err)
	}
	// Move to line 1, col 10 ("first line| long").
	if err := p.KeyDown(ctx, "ctrl+home"); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		if err := p.KeyDown(ctx, "arrowright"); err != nil {
			t.Fatal(err)
		}
	}
	if got := formState(p).Caret; got != 10 {
		t.Fatalf("start caret = %d, want 10", got)
	}
	origX := paintedCaretOp(t, p).X

	// Down to the empty middle line (offset 16).
	if err := p.KeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != 16 {
		t.Fatalf("empty line caret = %d, want 16", got)
	}

	// Down again to line 3: desiredX should restore column near col 10 (offset 17 + ~10 = 27).
	if err := p.KeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}
	line3Caret := formState(p).Caret
	line3X := paintedCaretOp(t, p).X
	if line3Caret < 24 || line3Caret > 30 {
		t.Fatalf("line 3 caret = %d, want near 27 (desiredX preserved across empty line)", line3Caret)
	}
	if math.Abs(line3X-origX) > 10 {
		t.Fatalf("line 3 caret X=%.2f drifted from origX=%.2f", line3X, origX)
	}

	// Up back to empty line, then up back to line 1 near col 10.
	if err := p.KeyDown(ctx, "arrowup"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != 16 {
		t.Fatalf("back to empty line caret = %d, want 16", got)
	}
	if err := p.KeyDown(ctx, "arrowup"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != 10 {
		t.Fatalf("back to line 1 caret = %d, want 10", got)
	}

	// Horizontal move resets desiredX.
	if err := p.KeyDown(ctx, "home"); err != nil {
		t.Fatal(err)
	}
	if formState(p).HasDesiredX {
		t.Fatal("home did not clear desiredX")
	}
	if err := p.KeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != 17 {
		t.Fatalf("after horizontal reset, line 3 caret = %d, want 17", got)
	}
}

func TestTextareaVerticalArrowsSoftWrapAndShiftSelect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const wrapped = "alpha bravo charlie delta echo foxtrot golf hotel india juliet kilo lima"
	p := editPage(t, `<textarea id="ta" style="display:block;width:140px;font-size:14px">`+wrapped+`</textarea>`)

	if err := p.Focus(ctx, "ta"); err != nil {
		t.Fatal(err)
	}
	// Move to offset 6 ("alpha |bravo...") on visual line 0.
	if err := p.KeyDown(ctx, "ctrl+home"); err != nil {
		t.Fatal(err)
	}
	for range 6 {
		if err := p.KeyDown(ctx, "arrowright"); err != nil {
			t.Fatal(err)
		}
	}
	line0Caret := formState(p).Caret
	line0Op := paintedCaretOp(t, p)

	// ArrowDown crosses into visual line 1.
	if err := p.KeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}
	line1Caret := formState(p).Caret
	line1Op := paintedCaretOp(t, p)
	if line1Caret <= line0Caret {
		t.Fatalf("soft-wrap arrowdown did not advance caret: line0=%d line1=%d", line0Caret, line1Caret)
	}
	if line1Op.Y <= line0Op.Y+1 {
		t.Fatalf("soft-wrap arrowdown did not move caret Y down: line0=%.2f line1=%.2f", line0Op.Y, line1Op.Y)
	}

	// Shift+ArrowDown extends selection from line 1 to line 2.
	if err := p.KeyDown(ctx, "shift+arrowdown"); err != nil {
		t.Fatal(err)
	}
	st := formState(p)
	if st.Anchor != line1Caret || st.Caret <= line1Caret {
		t.Fatalf("shift+arrowdown anchor=%d caret=%d want anchor=%d caret>%d", st.Anchor, st.Caret, line1Caret, line1Caret)
	}
	sel, ok, err := p.Copy(ctx)
	if err != nil || !ok || len(sel) == 0 {
		t.Fatalf("copy selection = %q ok=%v err=%v", sel, ok, err)
	}

	// Shift+ArrowUp brings caret back to line 1.
	if err := p.KeyDown(ctx, "shift+arrowup"); err != nil {
		t.Fatal(err)
	}
	if formState(p).Caret != line1Caret {
		t.Fatalf("shift+arrowup caret = %d, want %d", formState(p).Caret, line1Caret)
	}
}

func TestTextareaHomeEndPerLineAndCtrlHomeEnd(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const text = "line one\nline two\nline three"
	p := editPage(t, `<textarea id="ta" style="display:block;width:240px;font-size:14px">`+text+`</textarea>`)

	if err := p.Focus(ctx, "ta"); err != nil {
		t.Fatal(err)
	}
	// Focus starts at end of "line three" (len=28). Move up to "line two".
	if err := p.KeyDown(ctx, "arrowup"); err != nil {
		t.Fatal(err)
	}
	// Home moves to start of "line two" (offset 9), NOT 0.
	if err := p.KeyDown(ctx, "home"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != 9 {
		t.Fatalf("home on line 2 = %d, want 9", got)
	}

	// Shift+End selects "line two" [9, 17).
	if err := p.KeyDown(ctx, "shift+end"); err != nil {
		t.Fatal(err)
	}
	sel, ok, err := p.Copy(ctx)
	if err != nil || !ok || sel != "line two" {
		t.Fatalf("shift+end copy = %q ok=%v err=%v, want %q", sel, ok, err, "line two")
	}

	// End without shift collapses to 17; Shift+Home selects "line two" backwards.
	if err := p.KeyDown(ctx, "end"); err != nil {
		t.Fatal(err)
	}
	if err := p.KeyDown(ctx, "shift+home"); err != nil {
		t.Fatal(err)
	}
	sel, ok, err = p.Copy(ctx)
	if err != nil || !ok || sel != "line two" {
		t.Fatalf("shift+home copy = %q ok=%v err=%v, want %q", sel, ok, err, "line two")
	}

	// Ctrl+Home moves to 0; Ctrl+End moves to end of whole value.
	if err := p.KeyDown(ctx, "ctrl+home"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != 0 {
		t.Fatalf("ctrl+home = %d, want 0", got)
	}
	if err := p.KeyDown(ctx, "ctrl+end"); err != nil {
		t.Fatal(err)
	}
	if got := formState(p).Caret; got != len(text) {
		t.Fatalf("ctrl+end = %d, want %d", got, len(text))
	}
}

func TestTextareaMouseTripleClickAndDragAcrossLines(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const text = "alpha one\nbravo two\ncharlie three"
	p := editPage(t, `<textarea id="ta" style="display:block;width:240px;font-size:14px">`+text+`</textarea>`)

	op1 := runOp(p, "alpha one")
	op2 := runOp(p, "bravo two")
	op3 := runOp(p, "charlie three")
	if op1 == nil || op2 == nil || op3 == nil {
		t.Fatalf("missing line ops: op1=%v op2=%v op3=%v", op1 != nil, op2 != nil, op3 != nil)
	}

	// Triple-click on line 2 selects only "bravo two".
	l2x, _ := runSpan(p, op2)
	l2y := runBase(p, op2) - 4
	if err := p.SelectLineAt(ctx, l2x+20, l2y); err != nil {
		t.Fatal(err)
	}
	sel, ok, err := p.Copy(ctx)
	if err != nil || !ok || sel != "bravo two" {
		t.Fatalf("triple-click line 2 = %q ok=%v err=%v, want %q", sel, ok, err, "bravo two")
	}

	// Double-click on "bravo" selects "bravo".
	if err := p.SelectWordAt(ctx, l2x+8, l2y); err != nil {
		t.Fatal(err)
	}
	sel, ok, err = p.Copy(ctx)
	if err != nil || !ok || sel != "bravo" {
		t.Fatalf("double-click word = %q ok=%v err=%v, want %q", sel, ok, err, "bravo")
	}

	// Click on line 1 and drag to line 3 selects across newlines.
	l1x, _ := runSpan(p, op1)
	l1y := runBase(p, op1) - 4
	l3x, l3w := runSpan(p, op3)
	l3y := runBase(p, op3) - 4
	if err := p.SelectAt(ctx, l1x+2, l1y); err != nil {
		t.Fatal(err)
	}
	if err := p.Drag(ctx, l3x+l3w, l3y); err != nil {
		t.Fatal(err)
	}
	sel, ok, err = p.Copy(ctx)
	if err != nil || !ok || !strings.Contains(sel, "\n") || !strings.HasSuffix(sel, "charlie three") {
		t.Fatalf("cross-line drag = %q ok=%v err=%v", sel, ok, err)
	}
}
