package web

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type auditKbd struct {
	host.Screen
	typeErr error
	bsErr   error
}

func (f *auditKbd) Type(context.Context, string) error { return f.typeErr }
func (f *auditKbd) Backspace(context.Context) error    { return f.bsErr }

func mutateRec(t *testing.T, path, origin string, app host.Screen) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest("POST", "http://example.com"+path, strings.NewReader("text=x"))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	srv := &server{app: app}
	switch path {
	case "/type":
		srv.typeText(rec, r)
	default:
		srv.backspace(rec, r)
	}
	return rec
}

func TestClickBoundsParams(t *testing.T) {
	t.Parallel()
	many := "http://example.com/click?x=1&y=2"
	for i := 0; i < 20; i++ {
		many += "&k=1"
	}
	if rec := clickRec(&auditClick{}, many, "http://example.com"); rec.Code != 400 {
		t.Fatalf("param count: %d", rec.Code)
	}
	long := "http://example.com/click?x=" + strings.Repeat("1", 65) + "&y=2"
	if rec := clickRec(&auditClick{}, long, "http://example.com"); rec.Code != 400 {
		t.Fatalf("long value: %d", rec.Code)
	}
}

func TestMutatingNeedsOriginAndHidesErrors(t *testing.T) {
	t.Parallel()
	if rec := mutateRec(t, "/type", "", &auditKbd{}); rec.Code != 403 {
		t.Fatalf("missing origin: %d", rec.Code)
	}
	rec := mutateRec(t, "/type", "http://example.com", &auditKbd{typeErr: errors.New("secret")})
	if rec.Code != 500 || strings.TrimSpace(rec.Body.String()) != "internal server error" {
		t.Fatalf("type: %d %q", rec.Code, rec.Body.String())
	}
	rec = mutateRec(t, "/backspace", "http://example.com", &auditKbd{bsErr: errors.New("secret")})
	if rec.Code != 500 || strings.TrimSpace(rec.Body.String()) != "internal server error" {
		t.Fatalf("backspace: %d %q", rec.Code, rec.Body.String())
	}
}
