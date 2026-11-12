package page

import (
	"context"
	"strings"
	"testing"
)

func TestApplyWebTextWritesBind(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := &struct{ Email string }{}
	p := editPage(t, `<input id="e" type="text" data-bind="Email" style="display:block;width:120px;height:20px">`)
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var changes int
	p.Handle(Handlers{Change: func(_ context.Context, b Box) error {
		if b.ID != "e" {
			t.Errorf("change id %q", b.ID)
		}
		changes++

		return nil
	}})

	if err := p.ApplyWebText(ctx, "e", "ada@x"); err != nil {
		t.Fatal(err)
	}
	if view.Email != "ada@x" || p.FormValue("e") != "ada@x" {
		t.Fatalf("bound %q form %q", view.Email, p.FormValue("e"))
	}
	if changes != 1 {
		t.Fatalf("changes = %d, want 1", changes)
	}

	if err := p.ApplyWebText(ctx, "e", "ada@x"); err != nil {
		t.Fatal(err)
	}
	if changes != 1 {
		t.Fatalf("changes after same value = %d, want 1", changes)
	}
}

func TestApplyWebTextPasswordKeepsText(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := editPage(t, `<input id="pw" type="password" style="display:block;width:120px;height:20px">`)
	if err := p.ApplyWebText(ctx, "pw", "secret"); err != nil {
		t.Fatal(err)
	}

	c, ok := p.control("pw")
	if !ok || c.Value != "secret" || p.FormValue("pw") != "secret" {
		t.Fatalf("stored %+v form %q", c, p.FormValue("pw"))
	}
	if strings.Contains(c.Value, "•") {
		t.Fatalf("stored bullets %q", c.Value)
	}
}

func TestApplyWebCheckedFiresChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	view := &struct{ Agree bool }{}
	p := editPage(t, `<input id="c" type="checkbox" data-bind="Agree" style="display:block;width:20px;height:20px">`)
	p.SetData(view)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	var changes int
	p.Handle(Handlers{Change: func(_ context.Context, b Box) error {
		if b.ID != "c" {
			t.Errorf("change id %q", b.ID)
		}
		changes++

		return nil
	}})

	if err := p.ApplyWebChecked(ctx, "c", true); err != nil {
		t.Fatal(err)
	}
	if !view.Agree || !p.FormChecked("c") {
		t.Fatalf("bound %v form %v", view.Agree, p.FormChecked("c"))
	}
	if changes != 1 {
		t.Fatalf("changes = %d, want 1", changes)
	}

	if err := p.ApplyWebChecked(ctx, "c", true); err != nil {
		t.Fatal(err)
	}
	if changes != 1 {
		t.Fatalf("changes after same checked = %d, want 1", changes)
	}
}

func TestWebControlsListsLiveFields(t *testing.T) {
	t.Parallel()

	body := `<input id="pw" type="password" name="pass" value="secret" style="display:block;width:120px;height:20px">` +
		`<input id="e" type="text" value="hi" style="display:block;width:120px;height:20px">` +
		`<input id="ok" type="checkbox" checked style="display:block;width:20px;height:20px">` +
		`<select id="city" name="city" style="display:block;width:120px;height:20px">` +
		`<option value="a">A</option><option value="b" selected>B</option></select>`
	p := editPage(t, body)

	got := p.WebControls()
	ids := make([]string, len(got))
	for i, c := range got {
		ids[i] = c.ID
	}
	if strings.Join(ids, ",") != "city,e,ok,pw" {
		t.Fatalf("ids %v", ids)
	}

	byID := map[string]WebControl{}
	for _, c := range got {
		byID[c.ID] = c
	}

	pw := byID["pw"]
	if pw.Tag != "input" || pw.Type != "password" || pw.Name != "pass" || pw.Value != "secret" {
		t.Fatalf("password %+v", pw)
	}
	if strings.Contains(pw.Value, "•") {
		t.Fatalf("password bullets %q", pw.Value)
	}
	if pw.Options != nil || byID["e"].Options != nil || byID["ok"].Options != nil {
		t.Fatal("options set on a non-select")
	}
	if !byID["ok"].Checked || byID["ok"].Tag != "input" || byID["ok"].Type != "checkbox" {
		t.Fatalf("checkbox %+v", byID["ok"])
	}
	if byID["e"].Tag != "input" || byID["e"].Type != "text" || byID["e"].Value != "hi" {
		t.Fatalf("text %+v", byID["e"])
	}

	city := byID["city"]
	if city.Tag != "select" || city.Name != "city" || city.Value != "b" || len(city.Options) != 2 {
		t.Fatalf("select %+v", city)
	}
	if city.Options[0].Value != "a" || city.Options[0].Label != "A" || city.Options[0].Selected {
		t.Fatalf("option a %+v", city.Options[0])
	}
	if city.Options[1].Value != "b" || city.Options[1].Label != "B" || !city.Options[1].Selected {
		t.Fatalf("option b %+v", city.Options[1])
	}
}

func TestApplyWebTextSelectStoresValue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	body := `<select id="city" style="display:block;width:120px;height:20px">` +
		`<option value="a">A</option><option value="b" selected>B</option></select>`
	p := editPage(t, body)

	var changes int
	p.Handle(Handlers{Change: func(context.Context, Box) error {
		changes++
		return nil
	}})

	if err := p.ApplyWebText(ctx, "city", "a"); err != nil {
		t.Fatal(err)
	}
	if p.FormValue("city") != "a" {
		t.Fatalf("value %q", p.FormValue("city"))
	}
	if changes != 1 {
		t.Fatalf("changes = %d, want 1", changes)
	}

	var selected string
	for _, c := range p.WebControls() {
		if c.ID != "city" {
			continue
		}
		for _, opt := range c.Options {
			if opt.Selected {
				selected = opt.Value
			}
		}
	}
	if selected != "a" {
		t.Fatalf("selected %q", selected)
	}

	if err := p.ApplyWebText(ctx, "city", "a"); err != nil {
		t.Fatal(err)
	}
	if changes != 1 {
		t.Fatalf("changes after same value = %d, want 1", changes)
	}
}

func TestApplyWebNilAndIgnored(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	var none *Page
	if none.WebControls() != nil {
		t.Fatal("nil page controls")
	}
	if err := none.ApplyWebText(ctx, "e", "x"); err != nil {
		t.Fatal(err)
	}
	if err := none.ApplyWebChecked(ctx, "c", true); err != nil {
		t.Fatal(err)
	}

	body := `<input id="e" type="text" value="keep" disabled style="display:block;width:120px;height:20px">` +
		`<input id="c" type="checkbox" style="display:block;width:20px;height:20px">` +
		`<input id="d" type="checkbox" disabled style="display:block;width:20px;height:20px">`
	p := editPage(t, body)

	var changes int
	p.Handle(Handlers{Change: func(context.Context, Box) error {
		changes++
		return nil
	}})

	if err := p.ApplyWebText(ctx, "missing", "x"); err != nil {
		t.Fatal(err)
	}
	if err := p.ApplyWebText(ctx, "e", "next"); err != nil {
		t.Fatal(err)
	}
	if err := p.ApplyWebText(ctx, "c", "on"); err != nil {
		t.Fatal(err)
	}
	if err := p.ApplyWebChecked(ctx, "missing", true); err != nil {
		t.Fatal(err)
	}
	if err := p.ApplyWebChecked(ctx, "d", true); err != nil {
		t.Fatal(err)
	}
	if changes != 0 || p.FormValue("e") != "keep" || p.FormChecked("c") || p.FormChecked("d") {
		t.Fatalf("changes %d value %q c %v d %v", changes, p.FormValue("e"), p.FormChecked("c"), p.FormChecked("d"))
	}
}
