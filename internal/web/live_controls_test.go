package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

const liveControlsHTML = `<!DOCTYPE html><html><head><title>Controls</title></head><body>
<textarea id="note" data-bind="Note" style="display:block;width:200px;height:40px"></textarea>
<select id="city" data-bind="City" style="display:block;width:120px;height:24px">
<option value="paris">Paris</option>
<option value="rome" selected>Rome</option>
</select>
<input id="ok" type="checkbox" data-bind="Ok" style="display:block;width:16px;height:16px">
<input id="a" type="radio" name="g" value="a" data-bind="Group" style="display:block;width:16px;height:16px">
<input id="b" type="radio" name="g" value="b" data-bind="Group" checked style="display:block;width:16px;height:16px">
</body></html>`

type liveControlsView struct {
	Note  string
	City  string
	Ok    bool
	Group string
}

func TestLiveControlsRoundTrip(t *testing.T) {
	t.Parallel()

	view := &liveControlsView{}
	p, err := page.New(page.Config{HTML: liveControlsHTML, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(view)
	p.Handle(page.Handlers{})
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}

	srv := &server{app: p}
	origin := "http://example.com"

	rec := httptest.NewRecorder()
	srv.page(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, origin+"/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /: %d %q", rec.Code, rec.Body.String())
	}
	assertLiveControls(t, rec.Body.String())

	note := url.Values{"id": {"note"}, "value": {"hello <b>"}}.Encode()
	if rec = formRec(t, srv.change, "/change", note, origin); rec.Code != http.StatusNoContent {
		t.Fatalf("note: %d %q", rec.Code, rec.Body.String())
	}
	if view.Note != "hello <b>" {
		t.Fatalf("Note = %q", view.Note)
	}

	if rec = formRec(t, srv.change, "/change", "id=city&value=paris", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("city: %d %q", rec.Code, rec.Body.String())
	}
	if view.City != "paris" {
		t.Fatalf("City = %q", view.City)
	}

	if rec = formRec(t, srv.change, "/change", "id=ok&checked=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("ok: %d %q", rec.Code, rec.Body.String())
	}
	if !view.Ok {
		t.Fatal("Ok = false")
	}

	if rec = formRec(t, srv.change, "/change", "id=a&checked=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("radio: %d %q", rec.Code, rec.Body.String())
	}
	if view.Group != "a" {
		t.Fatalf("Group = %q", view.Group)
	}

	rec = httptest.NewRecorder()
	srv.dom(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, origin+"/dom", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /dom: %d %q", rec.Code, rec.Body.String())
	}
	var state domState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if !optionSelected(state.HTML, "paris") {
		t.Fatalf("paris option is not selected:\n%s", state.HTML)
	}
	if strings.Contains(state.HTML, "<b>") || !strings.Contains(state.HTML, "hello &lt;b&gt;") {
		t.Fatalf("textarea text is not escaped:\n%s", state.HTML)
	}
	if strings.Contains(state.HTML, "<iframe") {
		t.Fatalf("dom html has an iframe:\n%s", state.HTML)
	}
}

func assertLiveControls(t *testing.T, body string) {
	t.Helper()

	first := strings.Index(body, "<option")
	second := -1
	if first >= 0 {
		second = strings.Index(body[first+1:], "<option")
	}
	if !strings.Contains(body, "<textarea") || !strings.Contains(body, "<select") || first < 0 || second < 0 ||
		!strings.Contains(body, "paris") || !strings.Contains(body, "rome") {
		t.Fatalf("GET / is missing the live controls:\n%s", body)
	}
	if strings.Contains(body, "<iframe") {
		t.Fatalf("GET / has an iframe:\n%s", body)
	}
	if strings.Contains(body, `data-ownframe-field="select"`) {
		t.Fatalf("select element has data-ownframe-field:\n%s", body)
	}
}

func optionSelected(html, value string) bool {
	needle := `value="` + value + `"`
	rest := html
	for {
		i := strings.Index(rest, "<option")
		if i < 0 {
			return false
		}
		rest = rest[i:]
		end := strings.IndexByte(rest, '>')
		if end < 0 {
			return false
		}
		tag := rest[:end]
		rest = rest[end+1:]
		if strings.Contains(tag, needle) && strings.Contains(tag, "selected") {
			return true
		}
	}
}
