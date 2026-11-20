package page

import (
	"bytes"
	"context"
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"image/png"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBeforeEditErrorAbortsType(t *testing.T) {
	ctx := context.Background()
	p := editPage(t, `<input id="e" type="text" style="display:block;width:120px;height:20px">`)
	clickBox(t, p, "e")

	gen := p.Generation()
	boom := errors.New("stop")
	p.Handle(Handlers{BeforeEdit: func(context.Context, Box) error { return boom }})

	if err := p.Type(ctx, "x"); !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}

	if p.FormValue("e") != "" {
		t.Fatalf("value = %q", p.FormValue("e"))
	}

	if p.Generation() != gen {
		t.Fatal("page redrew after abort")
	}
}

func TestBeforeEditFiresOnCheckbox(t *testing.T) {
	p := editPage(t, `<input id="c" type="checkbox">`)

	var got Box
	p.Handle(Handlers{BeforeEdit: func(_ context.Context, b Box) error {
		got = b
		return nil
	}})

	clickBox(t, p, "c")

	if got.ID != "c" || !p.FormChecked("c") {
		t.Fatalf("box = %+v checked = %v", got, p.FormChecked("c"))
	}
}

func editPage(t *testing.T, html string) *Page {
	t.Helper()

	p, err := New(Config{HTML: html, Width: 320, Height: 240})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func clickBox(t *testing.T, p *Page, id string) {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID != id {
			continue
		}

		if err := p.Click(context.Background(), b.X+b.W/2, b.Y+b.H/2); err != nil {
			t.Fatal(err)
		}

		return
	}

	t.Fatalf("no box %q", id)
}

func TestBeforeEditFiresBeforeCut(t *testing.T) {
	ctx := context.Background()
	view := &struct{ Email string }{Email: "ab"}
	p := editPage(t, `<input id="e" type="text" data-bind="Email" style="display:block;width:120px;height:20px">`)
	p.SetData(view)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	clickBox(t, p, "e")

	var got Box
	p.Handle(Handlers{BeforeEdit: func(_ context.Context, b Box) error {
		got = b
		return nil
	}})

	text, ok, err := p.Cut(ctx)
	if err != nil || !ok || text != "ab" {
		t.Fatalf("cut = %q %v %v", text, ok, err)
	}

	if got.ID != "e" {
		t.Fatalf("before edit box = %+v", got)
	}

	if view.Email != "" {
		t.Fatalf("bound = %q", view.Email)
	}
}

func TestCaretClickMovesPaintedBar(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const value = "1234567890"
	p := bftPage(t, `<input id="t" type="text" value="`+value+`"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, value)
	if op == nil {
		t.Fatal("missing text run")
	}

	left, width := runSpan(p, op)
	base := runBase(p, op)
	box := boxByID(p, "t")
	padX := box.X + box.W - 4
	if padX <= left+width {
		t.Fatalf("field has no empty padding: text ends at %.1f, box ends at %.1f", left+width, box.X+box.W)
	}

	if err := p.SelectAt(ctx, padX, base-4); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	n := runeLen(value)
	if p.FocusID() != "t" || st.Caret != n || st.Anchor != n {
		t.Fatalf("padding focus %q caret %d anchor %d", p.FocusID(), st.Caret, st.Anchor)
	}

	endX := paintedCaretX(t, p)
	focused := runOp(p, value)
	if focused == nil {
		t.Fatal("missing focused text run")
	}

	runEnd := focused.X + focused.W
	tol := 2 * p.draw.Display.PointsPerPixel
	delta := endX - runEnd
	if delta < 0 {
		delta = -delta
	}
	if delta > tol {
		t.Fatalf("padding caret x %.2f, run ends at %.2f", endX, runEnd)
	}

	if err := p.SelectAt(ctx, left+width*0.2, base-4); err != nil {
		t.Fatal(err)
	}

	st = formState(p)
	if p.FocusID() != "t" || st.Caret != 2 || st.Anchor != 2 {
		t.Fatalf("glyph focus %q caret %d anchor %d", p.FocusID(), st.Caret, st.Anchor)
	}

	glyphX := paintedCaretX(t, p)
	if glyphX >= endX {
		t.Fatalf("painted caret x %.2f did not move left of %.2f", glyphX, endX)
	}
}

func paintedCaretX(t *testing.T, p *Page) float64 {
	t.Helper()

	d := p.draw.Display
	if d == nil || len(d.Ops) == 0 {
		t.Fatal("no display")
	}

	op := d.Ops[len(d.Ops)-1]
	if op.Kind != DisplayOpFillRect || op.W > d.PointsPerPixel {
		t.Fatalf("caret kind %v width %.2f", op.Kind, op.W)
	}

	return op.X
}

func TestCutChangeGetsBox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="a" type="text"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "a")
	if err := p.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	var want Box
	for _, b := range p.Boxes() {
		if b.ID == "a" {
			want = b
		}
	}

	var got []Box
	p.Handle(Handlers{Change: func(_ context.Context, box Box) error {
		got = append(got, box)
		return nil
	}})

	text, ok, err := p.Cut(ctx)
	if err != nil || !ok || text != "ab" {
		t.Fatalf("cut %q %v %v", text, ok, err)
	}
	if len(got) != 1 || !reflect.DeepEqual(got[0], want) {
		t.Fatalf("change %v want %v", got, want)
	}
	if p.FormValue("a") != "" {
		t.Fatalf("value %q", p.FormValue("a"))
	}
}

func TestCutWritesField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := &bftView{Email: "ada"}
	p := bftPage(t, bftField)
	bftDraw(t, p, view)
	bftClick(t, p, "email")

	text, ok, err := p.Cut(ctx)
	if err != nil || !ok || text != "ada" {
		t.Fatalf("cut %q %v %v", text, ok, err)
	}
	if view.Email != "" || p.FormValue("email") != "" {
		t.Fatalf("view %q form %q", view.Email, p.FormValue("email"))
	}
}

func TestCutErrorStopsRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="a" type="text"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "a")
	if err := p.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	boom := errors.New("boom")
	p.Handle(Handlers{Change: func(context.Context, Box) error { return boom }})
	gen := p.Generation()

	text, ok, err := p.Cut(ctx)
	if !errors.Is(err, boom) || !ok || text != "ab" {
		t.Fatalf("cut %q %v %v", text, ok, err)
	}
	if p.FormValue("a") != "" {
		t.Fatalf("value %q", p.FormValue("a"))
	}
	if p.Generation() != gen {
		t.Fatalf("generation %d -> %d", gen, p.Generation())
	}
}

func TestWholeSelectionPaintsASelectionSpan(t *testing.T) {
	src := `<input id="e" type="text" value="x">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "x"}
	got := rewriteControls(src, []controlSpan{{Control: ctrl, Start: 0, End: len(src)}}, nil, "e", true)
	part := got
	if i := strings.Index(got, "<span"); i >= 0 {
		part = got[i:]
	}
	if !strings.Contains(part, `<span data-ownframe-selection="1">x</span>`) {
		t.Fatalf("whole selection has no selection span: %s", got)
	}
}

func TestCaretUsesTheTextColor(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML:   `<input id="e" type="text" value="ab" style="color:#ff0000">`,
		Width:  300,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	d := p.Display()
	if d == nil {
		t.Fatal("no display")
	}

	found := false
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.W >= 2 || op.H <= 5 || op.H > 20 {
			continue
		}

		found = true

		if op.R < 0.99 || op.G > 0.01 || op.B > 0.01 {
			t.Fatalf("caret color = %.3f,%.3f,%.3f, want the text red", op.R, op.G, op.B)
		}
	}

	if !found {
		t.Fatal("no caret op")
	}
}

func TestCaretBlinkDoesNotMoveFieldText(t *testing.T) {
	p, err := New(Config{HTML: `<style>body{margin:0;background:white}</style><input id="e" value="Focused text stays visible">`, Width: 360, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err = p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}
	for _, caret := range []int{0, 13, 27} {
		st := formState(p)
		st.Caret, st.Anchor = caret, caret
		resetCaretBlink(p)
		if err = p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		shown, err := png.Decode(bytes.NewReader(p.PNG()))
		if err != nil {
			t.Fatal(err)
		}
		st = formState(p)
		st.BlinkAt = time.Now().Add(-caretBlinkInterval)
		if err = p.Tick(ctx); err != nil {
			t.Fatal(err)
		}
		hidden, err := png.Decode(bytes.NewReader(p.PNG()))
		if err != nil {
			t.Fatal(err)
		}
		minX, maxX := shown.Bounds().Dx(), -1
		for y := 0; y < shown.Bounds().Dy(); y++ {
			for x := 0; x < shown.Bounds().Dx(); x++ {
				if shown.At(x, y) == hidden.At(x, y) {
					continue
				}
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
			}
		}
		if maxX < minX {
			t.Fatalf("caret at %d did not change any pixels", caret)
		}
		if maxX-minX > 4 {
			t.Fatalf("caret at %d changed pixels across x=%d..%d; expected only the caret column to change", caret, minX, maxX)
		}
	}
}

func TestCaretBlinkHidesAndShows(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="e" type="text" value="ab">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpFillRect {
		t.Fatalf("visible caret kind = %v, want fill rect", got)
	}

	st := formState(p)
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpNoop {
		t.Fatalf("hidden caret kind = %v, want noop", got)
	}

	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpNoop {
		t.Fatalf("caret returned early with kind %v", got)
	}

	st = formState(p)
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.draw.Display.Ops[len(p.draw.Display.Ops)-1].Kind; got != layout.DisplayOpFillRect {
		t.Fatalf("caret did not return, kind = %v", got)
	}
}

func TestEditResetsBlink(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="e" type="text" value="ab">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if formState(p).BlinkOn {
		t.Fatal("blink did not turn off")
	}

	if err := p.Type(ctx, "c"); err != nil {
		t.Fatal(err)
	}

	if !formState(p).BlinkOn {
		t.Fatal("typing did not reset the blink")
	}
}

func TestCaretReplacesSelection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	st := formState(p)
	st.Anchor, st.Caret = 1, 3

	if err := p.Type(ctx, "Z"); err != nil {
		t.Fatal(err)
	}

	got := formState(p)
	if p.FormValue("t") != "aZd" || got.Caret != 2 || got.Anchor != 2 {
		t.Fatalf("value %q caret %d anchor %d", p.FormValue("t"), got.Caret, got.Anchor)
	}
}

func TestSelectAllSetsFullRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")

	if err := p.SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	start, end := st.Bounds(4)
	if start != 0 || end != 4 || st.Caret != 4 || !st.All {
		t.Fatalf("range [%d,%d) caret %d all %v", start, end, st.Caret, st.All)
	}

	if !p.FormSelected("t") {
		t.Fatal("not selected")
	}
}

func TestCaretSurvivesRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	setCaret(p, 4, false)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	st := formState(p)
	if st.Caret != 4 || st.Anchor != 4 || p.FocusID() != "t" {
		t.Fatalf("caret %d anchor %d focus %q", st.Caret, st.Anchor, p.FocusID())
	}
}

func TestCaretAndSelectionSurviveResize(t *testing.T) {
	t.Parallel()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	setCaret(p, 2, false)
	setCaret(p, 4, true)
	p.SetSize(600, 400)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	st := formState(p)
	if st.Caret != 4 || st.Anchor != 2 || p.FocusID() != "t" || p.FormValue("t") != "abcdef" {
		t.Fatal("resize changed the live field, caret, selection, or focus")
	}
}

func TestCaretEditsAtOffset(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	setCaret(p, 2, false)

	if err := p.Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("t") != "abXcd" || formState(p).Caret != 3 {
		t.Fatalf("insert value %q caret %d", p.FormValue("t"), formState(p).Caret)
	}

	if err := p.Backspace(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("t") != "abcd" || formState(p).Caret != 2 {
		t.Fatalf("backspace value %q caret %d", p.FormValue("t"), formState(p).Caret)
	}

	if err := p.DeleteWord(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("t") != "cd" || formState(p).Caret != 0 {
		t.Fatalf("delete word value %q caret %d", p.FormValue("t"), formState(p).Caret)
	}
}

func TestCaretMoveDoesNotEdit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)

	before, after := 0, 0
	p.Handle(Handlers{
		BeforeEdit: func(context.Context, Box) error { before++; return nil },
		Change:     func(context.Context, Box) error { after++; return nil },
	})

	bftClick(t, p, "t")
	setCaret(p, 2, false)

	if err := p.KeyDown(ctx, "arrowleft"); err != nil {
		t.Fatal(err)
	}

	if before != 0 || after != 0 {
		t.Fatalf("caret move fired before %d change %d", before, after)
	}

	if err := p.Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}

	if before != 1 || after != 1 {
		t.Fatalf("edit fired before %d change %d", before, after)
	}
}

func TestFileShowsNoCaret(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="f" type="file" value="a.txt">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "f"); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(p.source, `data-ownframe-focus="1"`) {
		t.Fatalf("focus missing: %s", p.source)
	}

	if strings.Contains(p.source, `data-ownframe-caret="1"`) {
		t.Fatalf("file paints a caret: %s", p.source)
	}

	st := formState(p)
	st.BlinkAt = time.Now().Add(-caretBlinkInterval)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(p.source, `data-ownframe-caret="1"`) {
		t.Fatalf("file blinked a caret: %s", p.source)
	}

	if err := p.Type(ctx, "b"); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("f"); got != "a.txtb" {
		t.Fatalf("typed file value %q", got)
	}

	if strings.Contains(p.source, `data-ownframe-caret="1"`) {
		t.Fatalf("file paints a caret after type: %s", p.source)
	}
}

func TestIMEAnchorTracksGlyphCaret(t *testing.T) {
	t.Parallel()
	p := imeField(t, "WWWW")
	setCaret(p, 1, false)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	first, ok := p.IMEAnchor()
	if !ok {
		t.Fatal("missing caret anchor")
	}
	setCaret(p, 3, false)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	second, ok := p.IMEAnchor()
	if !ok || second.X <= first.X {
		t.Fatalf("caret x positions = %v then %v", first.X, second.X)
	}
	field := boxByID(p, "e")
	if first.X <= field.X || second.X >= field.X+field.W {
		t.Fatalf("caret anchors outside field: first=%+v second=%+v field=%+v", first, second, field)
	}
}

func TestIMECompositionIsTemporary(t *testing.T) {
	p := imeField(t, "hello")
	beforeEdit, changes := 0, 0
	p.handlers.BeforeEdit = func(context.Context, Box) error {
		beforeEdit++
		return nil
	}
	p.handlers.Change = func(context.Context, Box) error {
		changes++
		return nil
	}
	before := p.Generation()
	if err := p.SetIMEComposition(context.Background(), "かな", 0, 6); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("e"); got != "hello" {
		t.Fatalf("committed value changed to %q", got)
	}
	if beforeEdit != 0 || changes != 0 {
		t.Fatalf("preedit called edit handlers: before=%d change=%d", beforeEdit, changes)
	}
	if p.Generation() <= before || !strings.Contains(p.source, `data-ownframe-composition="1"`) {
		t.Fatal("composition was not drawn")
	}
	if err := p.SetIMEComposition(context.Background(), "", 0, 0); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(p.source, `data-ownframe-composition="1"`) {
		t.Fatal("composition remained after clear")
	}
}

func TestIMEReplaceInsertsAndReplaces(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := imeField(t, "")

	if err := p.IMEReplace(ctx, 0, 0, "ab", 2); err != nil {
		t.Fatal(err)
	}

	_, caret, _, _, _ := p.IMEContext()
	if caret != 2 {
		t.Fatalf("caret after insert = %d", caret)
	}

	if err := p.IMEReplace(ctx, 0, 1, "X", 1); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("e"); got != "Xb" {
		t.Fatalf("value = %q", got)
	}
}

func TestIMEReplaceSnapsToRunes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := imeField(t, "héllo")

	// Byte 2 sits inside the two-byte é, so it snaps back to its start.
	if err := p.IMEReplace(ctx, 2, 3, "", 0); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("e"); got != "hllo" {
		t.Fatalf("value = %q", got)
	}

	// An empty replacement of an empty range deletes nothing.
	if err := p.IMEReplace(ctx, 3, 3, "", 0); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("e"); got != "hllo" {
		t.Fatalf("value after no-op = %q", got)
	}
}

// imeField focuses a text field and types value into it.
func imeField(t *testing.T, value string) *Page {
	t.Helper()

	p := bedPage(t, `<input id="e" type="text"`+bedWide+`>`, nil)
	bedClick(t, p, "e")

	if value != "" {
		if err := p.Type(context.Background(), value); err != nil {
			t.Fatal(err)
		}
	}

	return p
}

func TestIMEControlNeedsFocus(t *testing.T) {
	t.Parallel()

	p := bedPage(t, `<input id="e" type="text"`+bedWide+`>`, nil)

	if _, _, _, _, ok := p.IMEContext(); ok {
		t.Fatal("context without focus")
	}
}

func TestIMEContextAroundCaret(t *testing.T) {
	t.Parallel()

	p := imeField(t, "hello")
	setCaret(p, 2, false)

	box, caret, before, after, ok := p.IMEContext()
	if !ok || caret != 2 || before != "he" || after != "llo" {
		t.Fatalf("ok=%v caret=%d before=%q after=%q", ok, caret, before, after)
	}

	if box.W <= 0 || box.ID != "e" {
		t.Fatalf("box = %+v", box)
	}
}

func TestIMEContextCountsBytes(t *testing.T) {
	t.Parallel()

	p := imeField(t, "héllo")
	setCaret(p, 2, false)

	_, caret, before, _, ok := p.IMEContext()
	if !ok || caret != 3 || before != "hé" {
		t.Fatalf("ok=%v caret=%d before=%q", ok, caret, before)
	}
}

func TestIMEContextFollowsTheScroll(t *testing.T) {
	t.Parallel()

	p := imeField(t, "hello")
	_, doc := 0, 0
	for _, b := range p.Boxes() {
		if b.ID == "e" {
			doc = int(b.Y)
		}
	}

	p.SetScrollOffset(0, 120)

	box, _, _, _, ok := p.IMEContext()
	if !ok {
		t.Fatal("no context")
	}

	if got := int(box.Y); got != doc-120 {
		t.Fatalf("box y = %d, want %d", got, doc-120)
	}
}
