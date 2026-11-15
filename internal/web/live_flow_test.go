package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// liveDocHTML is a real document: inputs stay inputs, and data-bind names
// the view fields /change must write.
const liveDocHTML = `<!DOCTYPE html><html><head><title>Sign in</title></head><body>
<input id="email" type="text" data-bind="Email" style="display:block;width:200px;height:24px">
<input id="password" type="password" data-bind="Password" style="display:block;width:200px;height:24px">
<button id="login" data-action="login" style="display:block;width:120px;height:32px">Sign in</button>
</body></html>`

// liveDocView is the template data. SetData receives a pointer so data-bind
// can write Email and Password.
type liveDocView struct {
	Email    string
	Password string
}

func TestLiveDocumentTypesAndClicks(t *testing.T) {
	t.Parallel()

	p, err := page.New(page.Config{HTML: liveDocHTML, Width: 480, Height: 320})
	if err != nil {
		t.Fatal(err)
	}

	view := &liveDocView{}
	var changes int
	var action string
	p.Handle(page.Handlers{
		Change: func(context.Context, page.Box) error {
			changes++

			return nil
		},
		Click: func(_ context.Context, box page.Box) error {
			action = box.Action

			return nil
		},
	})
	p.SetData(view)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}

	srv := &server{app: p}
	origin := "http://example.com"

	body := liveDocBody(t, livePageRec(t, srv))
	for _, want := range []string{"<input", `type="password"`, `id="email"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET / missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "<iframe") || strings.Contains(body, "srcdoc") {
		t.Fatalf("GET / is a framed shell\n%s", body)
	}
	if !strings.Contains(body, `id="ownframe-boot"`) {
		t.Fatalf("GET / missing boot script\n%s", body)
	}

	rec := formRec(t, srv.change, "/change", "id=email&value=ada@ex.com", origin)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("email change: %d %q", rec.Code, rec.Body.String())
	}
	if view.Email != "ada@ex.com" {
		t.Fatalf("Email = %q", view.Email)
	}
	if changes == 0 {
		t.Fatal("Change was not called")
	}
	seen := changes

	rec = formRec(t, srv.change, "/change", "id=password&value=secret", origin)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("password change: %d %q", rec.Code, rec.Body.String())
	}
	if view.Password != "secret" {
		t.Fatalf("Password = %q", view.Password)
	}
	if changes == seen {
		t.Fatal("Change was not called for the password")
	}

	body = liveDocBody(t, livePageRec(t, srv))
	if strings.Contains(body, "<iframe") || strings.Contains(body, "srcdoc") {
		t.Fatalf("GET / after edit is a framed shell\n%s", body)
	}
	tag, ok := openTagByID(body, "password")
	if !ok || !strings.Contains(strings.ToLower(tag), "<input") {
		t.Fatalf("password input = %q\n%s", tag, body)
	}
	if !attrIs(tag, "value", "secret") {
		t.Fatalf("password value = %q", tag)
	}

	rec = formRec(t, srv.click, "/click", "id=login", origin)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("click: %d %q", rec.Code, rec.Body.String())
	}
	if action != "login" {
		t.Fatalf("click action = %q", action)
	}

	rec = formRec(t, srv.change, "/change", "id=email&value=ada@ex.com", "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("missing origin: %d %q", rec.Code, rec.Body.String())
	}
}

func livePageRec(t *testing.T, srv *server) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "http://example.com/", nil)
	srv.page(rec, req)

	return rec
}

func liveDocBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()

	if rec.Code != http.StatusOK {
		t.Fatalf("GET /: %d %q", rec.Code, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("Content-Type = %q", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("Cache-Control = %q", cc)
	}

	return rec.Body.String()
}

// openTagByID returns the open tag whose id attribute equals id.
func openTagByID(body, id string) (string, bool) {
	for _, needle := range []string{`id="` + id + `"`, `id='` + id + `'`} {
		from := 0
		for from < len(body) {
			rel := strings.Index(body[from:], needle)
			if rel < 0 {
				break
			}
			idx := from + rel
			start := strings.LastIndex(body[:idx], "<")
			if start < 0 {
				from = idx + len(needle)
				continue
			}
			endRel := strings.IndexByte(body[start:], '>')
			if endRel < 0 {
				return "", false
			}
			end := start + endRel
			if idx > end {
				from = idx + len(needle)
				continue
			}

			return body[start : end+1], true
		}
	}

	return "", false
}

// attrIs reports whether tag has name="value" or name='value'.
func attrIs(tag, name, value string) bool {
	return strings.Contains(tag, name+`="`+value+`"`) || strings.Contains(tag, name+`='`+value+`'`)
}
