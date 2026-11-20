package inputlab

import (
	"bytes"
	"context"
	"errors"
	"image/png"
	"os"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestMain keeps every clipboard test off the desktop clipboard.
func TestMain(m *testing.M) {
	ownframe.UseMemory(true)
	os.Exit(m.Run())
}

func boxTextByID(t *testing.T, app *App, id string) string {
	t.Helper()
	var text string
	found := false
	for _, b := range app.Boxes() {
		if b.ID != id {
			continue
		}
		text = b.Text
		found = true
	}
	if !found {
		t.Fatalf("no box id=%q", id)
	}
	return text
}

func assertReplayed(t *testing.T, app *App) {
	t.Helper()
	if app.Page().Probe().Display() == nil {
		t.Fatal("page fell back to the bitmap")
	}
}

func TestSmokeFramePng(t *testing.T) {
	t.Parallel()
	app, _ := newTest(t)
	img, err := png.Decode(bytes.NewReader(app.PNG()))
	if err != nil {
		t.Fatal(err)
	}
	if img.Bounds().Dx() != DefaultWidth || img.Bounds().Dy() < DefaultHeight {
		t.Fatalf("png = %v, want width %d and height >= %d", img.Bounds(), DefaultWidth, DefaultHeight)
	}
}

func TestBindingControlsWriteView(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "b-name")
	if err := app.Type(ctx, "ada"); err != nil {
		t.Fatal(err)
	}
	if got := boxTextByID(t, app, "b-state"); !strings.Contains(got, "ada") {
		t.Fatalf("b-state = %q, want the typed name", got)
	}

	clickID(t, app, ctx, "b-agree")
	if !app.View().BAgree {
		t.Fatal("BAgree = false, want true")
	}

	clickID(t, app, ctx, "b-color")
	if got := app.View().BColor; got != "Green" {
		t.Fatalf("BColor = %q, want Green", got)
	}
	clickID(t, app, ctx, "b-color")
	if got := app.View().BColor; got != "Blue" {
		t.Fatalf("BColor = %q, want Blue", got)
	}

	if got := app.View().BStatus; !strings.Contains(got, "changed") {
		t.Fatalf("BStatus = %q, want a change note", got)
	}
}

func TestBindHooksEmailAgreeAndPlanWriteView(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "l-email")
	if err := app.Type(ctx, ".net"); err != nil {
		t.Fatal(err)
	}
	if got := app.View().LEmail; got != "you@example.com.net" {
		t.Fatalf("LEmail = %q, want you@example.com.net", got)
	}
	if got := app.View().LStatus; !strings.Contains(got, "l-email") {
		t.Fatalf("LStatus = %q, want a note about l-email", got)
	}

	clickID(t, app, ctx, "l-pro")
	if got := app.View().LPlan; got != "pro" {
		t.Fatalf("LPlan = %q, want pro", got)
	}
	if got := app.View().LStatus; !strings.Contains(got, "l-pro") {
		t.Fatalf("LStatus = %q, want a note about l-pro", got)
	}

	clickID(t, app, ctx, "l-free")
	if got := app.View().LPlan; got != "free" {
		t.Fatalf("LPlan = %q, want free", got)
	}

	clickID(t, app, ctx, "l-agree")
	if !app.View().LAgree {
		t.Fatal("LAgree = false, want true")
	}
	if got := app.View().LStatus; !strings.Contains(got, "l-agree") {
		t.Fatalf("LStatus = %q, want a note about l-agree", got)
	}
}

func TestBeforeEditVetoesLockedName(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)
	errVeto := errors.New("inputlab: locked field veto")
	app.Page().AddHandlers(ownframe.Handlers{
		BeforeEdit: func(ctx context.Context, box ownframe.Box) error {
			if box.ID == "l-name" && app.FormValue("l-name") == "locked" {
				return errVeto
			}
			return app.onBeforeEdit(ctx, box)
		},
	})

	clickID(t, app, ctx, "l-name")
	if err := app.Type(ctx, "new"); err == nil {
		t.Fatal("Type on the locked name returned nil, want a veto error")
	}
	if got := app.FormValue("l-name"); got != "locked" {
		t.Fatalf("FormValue = %q, want locked", got)
	}
	if got := app.View().LName; got != "locked" {
		t.Fatalf("LName = %q, want locked", got)
	}
	if got := app.View().LStatus; got != "" {
		t.Fatalf("LStatus = %q, want unchanged and empty", got)
	}
}

func TestPlainFormSendReadsControls(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "f-email")
	if err := app.Type(ctx, "ada@ex.com"); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, ctx, "f-remember")
	clickID(t, app, ctx, "f-pro")
	clickID(t, app, ctx, "f-color")

	clickID(t, app, ctx, "f-doc")
	if err := app.Type(ctx, "notes.txt"); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, ctx, "f-send")

	const want = "email=ada@ex.com remember=true plan=pro color=Green file=notes.txt"
	if got := boxTextByID(t, app, "status"); got != want {
		t.Fatalf("status = %q, want %q", got, want)
	}
}

func TestControlsTypingFocusDeleteToggleAndCycle(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "f-email")
	if got := app.FocusedField(); got != "f-email" {
		t.Fatalf("FocusedField = %q, want f-email", got)
	}
	if err := app.Type(ctx, "one two"); err != nil {
		t.Fatal(err)
	}
	if err := app.DeleteWord(ctx); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("f-email"); got != "one " {
		t.Fatalf("after DeleteWord = %q, want %q", got, "one ")
	}
	if err := app.Backspace(ctx); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("f-email"); got != "one" {
		t.Fatalf("after Backspace = %q, want one", got)
	}
	if got := boxTextByID(t, app, "status"); !strings.Contains(got, "form=one|") {
		t.Fatalf("status = %q, want form=one|", got)
	}

	clickID(t, app, ctx, "f-remember")
	if !app.FormChecked("f-remember") {
		t.Fatal("f-remember = false, want true")
	}

	clickID(t, app, ctx, "f-pro")
	if !app.FormChecked("f-pro") {
		t.Fatal("f-pro = false, want true")
	}
	if app.FormChecked("f-free") {
		t.Fatal("f-free = true, want false after checking f-pro")
	}
	if got := app.FocusedField(); got != "f-pro" {
		t.Fatalf("FocusedField = %q, want f-pro", got)
	}

	clickID(t, app, ctx, "f-color")
	if got := app.FormValue("f-color"); got != "Green" {
		t.Fatalf("f-color = %q, want Green", got)
	}
	clickID(t, app, ctx, "f-color")
	clickID(t, app, ctx, "f-color")
	if got := app.FormValue("f-color"); got != "Red" {
		t.Fatalf("f-color = %q, want Red after the third click", got)
	}

	status := boxTextByID(t, app, "status")
	for _, want := range []string{"plan=pro", "color=Red"} {
		if !strings.Contains(status, want) {
			t.Fatalf("status = %q, want %q", status, want)
		}
	}

	clickID(t, app, ctx, "f-note")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("f-note"); got != "hello" {
		t.Fatalf("f-note = %q, want hello", got)
	}

	clickID(t, app, ctx, "f-doc")
	if err := app.Type(ctx, "notes.txt"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("f-doc"); got != "notes.txt" {
		t.Fatalf("f-doc = %q, want notes.txt", got)
	}
}

func TestDocPickerUpdatesStatus(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	app.Page().Testing().SetPicker(func(context.Context, string) (string, error) {
		return "/tmp/notes.txt", nil
	})

	clickID(t, app, ctx, "f-doc")

	if got := app.FormValue("f-doc"); got != "/tmp/notes.txt" {
		t.Fatalf("f-doc = %q", got)
	}
	if got := boxTextByID(t, app, "f-doc"); got != "/tmp/notes.txt" {
		t.Fatalf("f-doc text = %q", got)
	}
	if got := boxTextByID(t, app, "status"); !strings.Contains(got, "doc=/tmp/notes.txt") {
		t.Fatalf("status = %q", got)
	}
}

func TestLongDocGrowsField(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)
	before := boxByID(t, app, "f-doc").H

	clickID(t, app, ctx, "f-doc")
	if err := app.Type(ctx, "some/very/long/path/to/a/file/that/keeps/going/document.pdf"); err != nil {
		t.Fatal(err)
	}

	if got := boxByID(t, app, "f-doc").H; got <= before {
		t.Fatalf("f-doc height = %v, want more than %v", got, before)
	}
}

func TestEditingShortcutsAndButtons(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "e-note")
	if err := app.Type(ctx, "abc"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("e-note"); got != "abc" {
		t.Fatalf("typed = %q, want abc", got)
	}

	clickID(t, app, ctx, "e-undo")
	if got := app.FormValue("e-note"); got != "" {
		t.Fatalf("undo = %q, want empty", got)
	}

	clickID(t, app, ctx, "e-redo")
	if got := app.FormValue("e-note"); got != "abc" {
		t.Fatalf("redo = %q, want abc", got)
	}

	clickID(t, app, ctx, "e-selectall")
	if !app.FormSelected("e-note") {
		t.Fatal("e-selectall did not select e-note")
	}

	if err := app.Type(ctx, "z"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("e-note"); got != "z" {
		t.Fatalf("selected type = %q, want z", got)
	}

	before := app.View().Stamp
	clickID(t, app, ctx, "e-redraw")
	if got := app.View().Stamp; got != before+1 {
		t.Fatalf("Stamp = %d, want %d", got, before+1)
	}
	if got := app.FormValue("e-note"); got != "z" {
		t.Fatalf("after redraw = %q, want z", got)
	}
}

func TestEditingCopyCutPasteAndUndo(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "e-note")
	if err := app.Type(ctx, "abc"); err != nil {
		t.Fatal(err)
	}
	if err := app.Paste(ctx, "def"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("e-note"); got != "abcdef" {
		t.Fatalf("paste = %q, want abcdef", got)
	}

	text, ok, err := app.Copy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || text != "abcdef" {
		t.Fatalf("copy = %q ok=%v, want abcdef", text, ok)
	}

	text, ok, err = app.Cut(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || text != "abcdef" {
		t.Fatalf("cut = %q ok=%v, want abcdef", text, ok)
	}
	if got := app.FormValue("e-note"); got != "" {
		t.Fatalf("after cut = %q, want empty", got)
	}

	if err := app.onUndo(ctx); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("e-note"); got != "abcdef" {
		t.Fatalf("undo after cut = %q, want abcdef", got)
	}
}

func TestClipboardButtonsWriteStatusAndText(t *testing.T) {
	app, ctx := newTest(t)

	clickID(t, app, ctx, "c-copy")
	if got := app.View().Status; !strings.Contains(got, "nothing selected") {
		t.Fatalf("empty copy status = %q", got)
	}

	clickID(t, app, ctx, "c-left")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, ctx, "c-copy")
	if got := app.View().Status; !strings.Contains(got, "copied: hello") {
		t.Fatalf("copy status = %q", got)
	}
	gotClip, err := ownframe.Read()
	if err != nil || gotClip != "hello" {
		t.Fatalf("clipboard after copy = %q err=%v, want hello", gotClip, err)
	}

	clickID(t, app, ctx, "c-cut")
	if got := app.FormValue("c-left"); got != "" {
		t.Fatalf("after cut = %q, want empty", got)
	}
	if got := app.View().Status; !strings.Contains(got, "cut: hello") {
		t.Fatalf("cut status = %q", got)
	}
	gotClip, err = ownframe.Read()
	if err != nil || gotClip != "hello" {
		t.Fatalf("clipboard after cut = %q err=%v, want hello", gotClip, err)
	}

	clickID(t, app, ctx, "c-pastego")
	if got := app.FormValue("c-left"); !strings.Contains(got, "from Go") {
		t.Fatalf("after pastego = %q, want from Go", got)
	}

	clickID(t, app, ctx, "c-selectall")
	if !app.Page().Form().FormSelected("c-left") {
		t.Fatal("c-selectall did not select c-left")
	}
}

func TestClipboardSelectAllSurvivesCaretPlacement(t *testing.T) {
	app, ctx := newTest(t)

	clickID(t, app, ctx, "c-left")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}

	box := boxByID(t, app, "c-selectall")
	if err := app.Page().Testing().SelectAt(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, ctx, "c-selectall")
	if !app.Page().Form().FormSelected("c-left") {
		t.Fatal("c-selectall lost the selection after the caret placement")
	}
}

func TestClipboardCopyCutPasteUndoRedo(t *testing.T) {
	app, ctx := newTest(t)

	clickID(t, app, ctx, "c-left")
	if err := app.Type(ctx, "hello"); err != nil {
		t.Fatal(err)
	}
	if err := app.SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	text, ok, err := app.Copy(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || text != "hello" {
		t.Fatalf("copy = %q ok=%v, want hello", text, ok)
	}

	text, ok, err = app.Cut(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || text != "hello" {
		t.Fatalf("cut = %q ok=%v, want hello", text, ok)
	}
	if got := app.FormValue("c-left"); got != "" {
		t.Fatalf("after cut = %q, want empty", got)
	}

	prev := app.FormValue("c-left")
	if err := app.Paste(ctx, "clip"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("c-left"); !strings.Contains(got, "clip") {
		t.Fatalf("pasted = %q, want clip", got)
	}

	clickID(t, app, ctx, "c-undo")
	if got := app.FormValue("c-left"); got != prev {
		t.Fatalf("undo = %q, want %q", got, prev)
	}

	clickID(t, app, ctx, "c-redo")
	if got := app.FormValue("c-left"); !strings.Contains(got, "clip") {
		t.Fatalf("redo = %q, want clip", got)
	}
}

func TestClipboardButtonsUseLastFocusedField(t *testing.T) {
	app, ctx := newTest(t)

	clickID(t, app, ctx, "c-right")
	if err := app.Type(ctx, "rr"); err != nil {
		t.Fatal(err)
	}

	clickID(t, app, ctx, "c-copy")
	if got := app.View().Status; !strings.Contains(got, "copied: rr") {
		t.Fatalf("copy status = %q, want copied: rr", got)
	}

	clickID(t, app, ctx, "c-selectall")
	if !app.Page().Form().FormSelected("c-right") {
		t.Fatal("c-selectall did not select the right field")
	}

	clickID(t, app, ctx, "c-pastego")
	if got := app.FormValue("c-right"); !strings.Contains(got, "from Go") {
		t.Fatalf("c-right = %q, want from Go", got)
	}

	clickID(t, app, ctx, "c-cut")
	if got := app.FormValue("c-right"); got != "" {
		t.Fatalf("c-right after cut = %q, want empty", got)
	}
	if got := app.View().Status; !strings.Contains(got, "cut: ") {
		t.Fatalf("cut status = %q, want cut: ...", got)
	}
	if got := app.FormValue("c-left"); got != "" {
		t.Fatalf("c-left = %q, want untouched", got)
	}
}

func TestClipboardPastePullsClipboardTextAndReplacesSelected(t *testing.T) {
	app, ctx := newTest(t)

	ownframe.Write("first\nsecond")
	clickID(t, app, ctx, "c-left")
	clickID(t, app, ctx, "c-paste")

	if got := app.FormValue("c-left"); got != "firstsecond" {
		t.Fatalf("pasted = %q, want firstsecond", got)
	}
	if got := app.View().Status; !strings.Contains(got, "pasted: firstsecond") {
		t.Fatalf("status = %q", got)
	}

	clickID(t, app, ctx, "c-selectall")
	clickID(t, app, ctx, "c-pastego")
	if got := app.FormValue("c-left"); got != "from Go" {
		t.Fatalf("after paste over selection = %q, want from Go", got)
	}
}

func TestInputFieldsAndLongPage(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	clickID(t, app, ctx, "i-email")
	if got := app.FocusedField(); got != "i-email" {
		t.Fatalf("focused = %q, want i-email", got)
	}
	if err := app.Type(ctx, "a@b.c"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("i-email"); got != "a@b.c" {
		t.Fatalf("i-email = %q, want a@b.c", got)
	}

	clickID(t, app, ctx, "i-remember")
	if !app.FormChecked("i-remember") {
		t.Fatal("i-remember checkbox did not toggle")
	}

	last := boxByID(t, app, "row-30")
	if last.Y <= float64(DefaultHeight) {
		t.Fatalf("row-30 y = %v, want below the fold (%d)", last.Y, DefaultHeight)
	}
}

func TestStatesAndPointerInput(t *testing.T) {
	t.Parallel()
	app, ctx := newTest(t)

	plain := app.PNG()
	if len(plain) == 0 {
		t.Fatal("png is empty after the first redraw")
	}
	assertReplayed(t, app)

	clickID(t, app, ctx, "s-focus-input")
	if got := app.FocusedField(); got != "s-focus-input" {
		t.Fatalf("focused = %q, want s-focus-input", got)
	}

	focused := app.PNG()
	if bytes.Equal(focused, plain) {
		t.Fatal("focus did not change the png")
	}
	assertReplayed(t, app)

	hoverBtn := boxByID(t, app, "s-hover-btn")
	hx, hy := hoverBtn.X+hoverBtn.W/2, hoverBtn.Y+hoverBtn.H/2
	before := app.Generation()
	if err := app.Hover(ctx, hx, hy); err != nil {
		t.Fatal(err)
	}
	if app.Generation() <= before {
		t.Fatalf("hover generation = %d, want past %d", app.Generation(), before)
	}
	if hovered := app.PNG(); bytes.Equal(hovered, focused) {
		t.Fatal("hover did not change the png")
	}
	assertReplayed(t, app)

	pressBtn := boxByID(t, app, "s-press-btn")
	px, py := pressBtn.X+pressBtn.W/2, pressBtn.Y+pressBtn.H/2
	if err := app.Press(ctx, px, py); err != nil {
		t.Fatal(err)
	}
	assertReplayed(t, app)

	if err := app.Release(ctx); err != nil {
		t.Fatal(err)
	}

	beforeCheck := app.PNG()
	clickID(t, app, ctx, "s-agree")
	if !app.FormChecked("s-agree") {
		t.Fatal("s-agree is not checked after the click")
	}
	if checked := app.PNG(); bytes.Equal(checked, beforeCheck) {
		t.Fatal("checked did not change the png")
	}
	assertReplayed(t, app)
}
