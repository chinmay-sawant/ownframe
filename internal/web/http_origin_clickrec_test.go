package web

import (
	"context"
	"net/http/httptest"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type auditClick struct {
	host.Screen
	err error
}

func (f *auditClick) Click(context.Context, float64, float64) error { return f.err }

func clickRec(app *auditClick, target, origin string) *httptest.ResponseRecorder {
	r := httptest.NewRequestWithContext(context.Background(), "POST", target, nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	(&server{app: app}).click(rec, r)
	return rec
}
