package web

import (
	"context"
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
	r := httptest.NewRequestWithContext(t.Context(), "POST", "http://example.com"+path, strings.NewReader("text=x"))
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
