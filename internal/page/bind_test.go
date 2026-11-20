package page

import (
	"context"
	"reflect"
	"strings"
	"testing"
)

type bindData struct {
	Text  string
	Flag  bool
	Pick  string
	Count int
	inner string
}

func bindCtl(tag, kind string) Control {
	return Control{ID: "x", Tag: tag, Type: kind}
}

func TestBindText(t *testing.T) {
	t.Parallel()

	for _, c := range []Control{
		bindCtl("input", "text"),
		bindCtl("textarea", ""),
		bindCtl("select", ""),
	} {
		c.Bind, c.Value = "Text", "old"
		d := &bindData{Text: "Ann"}
		got, ok := bindRead(&Page{data: d}, c)
		if !ok || got.Value != "Ann" {
			t.Fatalf("%s read %q %v", c.Tag, got.Value, ok)
		}
		c.Value = "Bob"
		if !bindWrite(&Page{data: d}, c) || d.Text != "Bob" {
			t.Fatalf("%s write %q", c.Tag, d.Text)
		}
	}
}

func TestBindCheckbox(t *testing.T) {
	t.Parallel()

	c := bindCtl("input", "checkbox")
	c.Bind = "Flag"
	d := &bindData{Flag: true}
	got, ok := bindRead(&Page{data: d}, c)
	if !ok || !got.Checked {
		t.Fatalf("read %v %v", got.Checked, ok)
	}
	if !bindWrite(&Page{data: d}, got) || !d.Flag {
		t.Fatalf("write %v", d.Flag)
	}
	c.Checked = false
	if !bindWrite(&Page{data: d}, c) || d.Flag {
		t.Fatalf("write %v", d.Flag)
	}
}

func TestBindRadio(t *testing.T) {
	t.Parallel()

	yes := bindCtl("input", "radio")
	yes.Bind, yes.Value = "Pick", "b"
	no := yes
	no.Value = "a"

	d := &bindData{Pick: "b"}
	got, ok := bindRead(&Page{data: d}, yes)
	if !ok || !got.Checked {
		t.Fatalf("b read %v %v", got.Checked, ok)
	}
	got, ok = bindRead(&Page{data: d}, no)
	if !ok || got.Checked {
		t.Fatalf("a read %v %v", got.Checked, ok)
	}

	no.Checked = true
	if !bindWrite(&Page{data: d}, no) || d.Pick != "a" {
		t.Fatalf("write %q", d.Pick)
	}
	if !bindWrite(&Page{data: d}, yes) || d.Pick != "a" {
		t.Fatalf("unchecked %q", d.Pick)
	}
}

func TestBindE2ECheckboxAndRadioReflectSetData(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	body := `<input type="checkbox" id="ok" data-bind="OK">` +
		`<input type="radio" id="free" name="plan" value="free" data-bind="Pick">` +
		`<input type="radio" id="pro" name="plan" value="pro" data-bind="Pick">`
	view := bdtView{}
	p := bftPage(t, body)
	bftDraw(t, p, &view)

	view.OK = true
	view.Pick = "pro"
	p.SetData(&view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if !p.FormChecked("ok") || !p.FormChecked("pro") || p.FormChecked("free") {
		t.Fatalf("ok=%v free=%v pro=%v", p.FormChecked("ok"), p.FormChecked("free"), p.FormChecked("pro"))
	}

	view.OK = false
	view.Pick = "free"
	p.SetData(&view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FormChecked("ok") || !p.FormChecked("free") || p.FormChecked("pro") {
		t.Fatalf("ok=%v free=%v pro=%v", p.FormChecked("ok"), p.FormChecked("free"), p.FormChecked("pro"))
	}
}

type bdtView struct {
	Email string
	OK    bool
	Pick  string
}

const bdtField = `<input id="email" type="text" data-bind="Email" style="display:block;width:220px;height:28px"><p id="out">{{.Email}}</p>`

func TestBindE2ESetDataPushesChangedField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := bdtView{Email: "ada"}
	p := bftPage(t, bdtField)
	bftDraw(t, p, &view)

	if p.FormValue("email") != "ada" || bftBox(t, p, "email") != "ada" {
		t.Fatalf("form %q box %q", p.FormValue("email"), bftBox(t, p, "email"))
	}

	view.Email = "grace"
	p.SetData(&view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("email") != "grace" {
		t.Fatalf("FormValue %q", p.FormValue("email"))
	}
	if got := bftBox(t, p, "out"); got != "grace" {
		t.Fatalf("paragraph %q", got)
	}
}

func TestBindE2EValueDataStaysInert(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := bdtView{Email: "grace"}
	p := bftPage(t, bdtField)
	bftDraw(t, p, view)

	if p.FormValue("email") != "" {
		t.Fatalf("FormValue %q", p.FormValue("email"))
	}

	bftClick(t, p, "email")
	if err := p.Type(ctx, "ada"); err != nil {
		t.Fatal(err)
	}

	if p.FormValue("email") != "ada" || view.Email != "grace" {
		t.Fatalf("FormValue %q view %q", p.FormValue("email"), view.Email)
	}
	if got := bftBox(t, p, "out"); got != "grace" {
		t.Fatalf("paragraph %q", got)
	}
}

type bedView struct {
	Email string
	Agree bool
	Plan  string
}

const bedWide = ` style="display:block;width:220px;height:28px"`

func bedPage(t *testing.T, body string, data any) *Page {
	t.Helper()

	p, err := New(Config{HTML: `<body style="margin:8px">` + body + `</body>`, Width: 640, Height: 480})
	if err != nil {
		t.Fatal(err)
	}

	p.SetData(data)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

func bedClick(t *testing.T, p *Page, id string) {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			if err := p.Click(context.Background(), b.X+b.W/2, b.Y+b.H/2); err != nil {
				t.Fatal(err)
			}

			return
		}
	}

	t.Fatalf("no box %q", id)
}

func TestBindEditTypeWritesStruct(t *testing.T) {
	t.Parallel()

	view := &bedView{}
	p := bedPage(t, `<input id="e" type="text" data-bind="Email"`+bedWide+`>`, view)
	bedClick(t, p, "e")
	if err := p.Type(context.Background(), "ab"); err != nil {
		t.Fatal(err)
	}

	if view.Email != "ab" || p.FormValue("e") != "ab" {
		t.Fatalf("view %q form %q", view.Email, p.FormValue("e"))
	}
}

func TestBindEditUnboundWritesNothing(t *testing.T) {
	t.Parallel()

	view := &bedView{Email: "keep"}
	p := bedPage(t, `<input id="e" type="text"`+bedWide+`>`, view)
	bedClick(t, p, "e")
	if err := p.Type(context.Background(), "ab"); err != nil {
		t.Fatal(err)
	}

	if view.Email != "keep" {
		t.Fatalf("view %q", view.Email)
	}
}

func TestBindE2EBackspaceWritesThroughToTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := bftView{Email: "abc"}
	p := bftPage(t, bftField)
	bftDraw(t, p, &view)
	bftClick(t, p, "email")
	if err := p.Backspace(ctx); err != nil {
		t.Fatal(err)
	}

	if view.Email != "ab" {
		t.Fatalf("view.Email %q", view.Email)
	}
	if p.FormValue("email") != "ab" {
		t.Fatalf("FormValue %q", p.FormValue("email"))
	}
	if got := bftBox(t, p, "out"); got != "ab" {
		t.Fatalf("paragraph %q", got)
	}
}

func TestBindE2EPasteWritesThroughToTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := bftView{}
	p := bftPage(t, bftField)
	bftDraw(t, p, &view)
	bftClick(t, p, "email")
	if err := p.Paste(ctx, "ada"); err != nil {
		t.Fatal(err)
	}

	if view.Email != "ada" {
		t.Fatalf("view.Email %q", view.Email)
	}
	if p.FormValue("email") != "ada" {
		t.Fatalf("FormValue %q", p.FormValue("email"))
	}
	if got := bftBox(t, p, "out"); got != "ada" {
		t.Fatalf("paragraph %q", got)
	}
}

type bftView struct {
	Email string
}

func bftPage(t *testing.T, body string) *Page {
	t.Helper()

	p, err := New(Config{
		HTML:   `<body style="margin:8px">` + body + `</body>`,
		Width:  640,
		Height: 480,
	})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

func bftDraw(t *testing.T, p *Page, data any) {
	t.Helper()

	p.SetData(data)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func bftBox(t *testing.T, p *Page, id string) string {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			return b.Text
		}
	}

	t.Fatalf("no box %q", id)

	return ""
}

func bftClick(t *testing.T, p *Page, id string) {
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

const bftField = `<input id="email" type="text" data-bind="Email" style="display:block;width:220px;height:28px"><p id="out">{{.Email}}</p>`

func TestBindE2ETypeWritesThroughToTemplate(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := bftView{}
	p := bftPage(t, bftField)
	bftDraw(t, p, &view)
	bftClick(t, p, "email")
	if err := p.Type(ctx, "ada"); err != nil {
		t.Fatal(err)
	}

	if view.Email != "ada" {
		t.Fatalf("view.Email %q", view.Email)
	}
	if p.FormValue("email") != "ada" {
		t.Fatalf("FormValue %q", p.FormValue("email"))
	}
	if got := bftBox(t, p, "out"); !strings.Contains(got, "ada") {
		t.Fatalf("paragraph %q", got)
	}
}

func TestBindInert(t *testing.T) {
	t.Parallel()

	c := Control{Type: "text", Bind: "Text", Value: "x"}
	d := &bindData{}
	for _, data := range []any{nil, bindData{}, (*bindData)(nil), new(int)} {
		if _, ok := bindRead(&Page{data: data}, c); ok {
			t.Fatalf("read %T", data)
		}
		if bindWrite(&Page{data: data}, c) || d.Text != "" {
			t.Fatalf("write %T", data)
		}
	}
	if _, ok := bindRead(nil, c); ok {
		t.Fatal("nil page")
	}
	if bindWrite(nil, c) {
		t.Fatal("nil page write")
	}
}

func TestBindIgnored(t *testing.T) {
	t.Parallel()

	d := &bindData{Text: "keep", Count: 3, inner: "hidden"}
	for _, c := range []Control{
		{Type: "text", Bind: "Nope"},
		{Type: "text", Bind: "inner"},
		{Type: "text", Bind: "Count"},
		{Type: "text"},
		{Type: "checkbox", Bind: "Text"},
		{Type: "checkbox", Bind: "Count"},
		{Tag: "textarea", Bind: "Count"},
		{Type: "radio", Bind: "Count"},
		{Type: "submit", Bind: "Text"},
	} {
		if _, ok := bindRead(&Page{data: d}, c); ok {
			t.Fatalf("read %+v", c)
		}
		if bindWrite(&Page{data: d}, c) {
			t.Fatalf("write %+v", c)
		}
	}
	if d.Text != "keep" || d.Count != 3 || d.inner != "hidden" {
		t.Fatalf("mutated %+v", d)
	}
}

func TestBindTarget(t *testing.T) {
	t.Parallel()

	f, ok := bindTarget(&Page{data: &bindData{Text: "hi"}}, "Text")
	if !ok || f.Kind() != reflect.String || f.String() != "hi" {
		t.Fatalf("target %v %v", f, ok)
	}
	for _, name := range []string{"", "Nope", "inner"} {
		if _, ok := bindTarget(&Page{data: &bindData{}}, name); ok {
			t.Fatalf("target %q", name)
		}
	}
}

func TestBindEditTogglesWriteStruct(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, body, id string
		check          func(bedView, *Page) bool
	}{
		{"checkbox", `<input id="c" type="checkbox" data-bind="Agree">`, "c",
			func(v bedView, p *Page) bool { return v.Agree && p.FormChecked("c") }},
		{"radio", `<input id="a" type="radio" name="g" value="free" data-bind="Plan">` +
			`<input id="b" type="radio" name="g" value="pro" data-bind="Plan"` + bedWide + `>`, "b",
			func(v bedView, p *Page) bool { return v.Plan == "pro" && p.FormChecked("b") }},
		{"select", `<select id="s" data-bind="Plan"` + bedWide + `><option value="free">Free</option>` +
			`<option value="pro" selected>Pro</option></select>`, "s",
			func(v bedView, p *Page) bool { return v.Plan == "free" && p.FormValue("s") == "free" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			view := &bedView{}
			p := bedPage(t, tc.body, view)
			bedClick(t, p, tc.id)
			if !tc.check(*view, p) {
				t.Fatalf("view %+v", *view)
			}
		})
	}
}
