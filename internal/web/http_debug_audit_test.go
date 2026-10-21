package web

import (
	"errors"
	"image"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

func errSecret() error { return errors.New("secret /tmp/x") }

type auditDbg struct {
	host.Screen
	reloadErr string
}

func (f *auditDbg) Boxes() []layout.Box      { return nil }
func (f *auditDbg) Generation() uint64       { return 1 }
func (f *auditDbg) Display() *layout.Display { return nil }
func (f *auditDbg) Image() image.Image       { return nil }
func (f *auditDbg) Size() (int, int)         { return 10, 10 }
func (f *auditDbg) DevTools() bool           { return true }
func (f *auditDbg) SetDevTools(bool)         {}
func (f *auditDbg) Stats() host.Stats        { return host.Stats{LastReloadError: f.reloadErr} }

func TestDebugSanitizesReloadError(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	(&server{app: &auditDbg{reloadErr: "bad\nline\r/tmp/x"}}).debug(rec, httptest.NewRequest("GET", "/debug/state", nil))
	body := rec.Body.String()
	if !strings.Contains(body, "bad_line_/tmp/x") || strings.Contains(body, "bad\nline") {
		t.Fatalf("unsanitized: %q", body)
	}
	if got := sanitizeLine("a\nb\rc"); got != "a_b_c" {
		t.Fatalf("sanitizeLine = %q", got)
	}
}

func TestClickHidesError(t *testing.T) {
	t.Parallel()
	rec := clickRec(&auditClick{err: errSecret()}, "http://example.com/click?x=1&y=2", "http://example.com")
	if rec.Code != 500 || strings.TrimSpace(rec.Body.String()) != "internal server error" {
		t.Fatalf("click: %d %q", rec.Code, rec.Body.String())
	}
}
