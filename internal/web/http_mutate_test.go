package web

import (
	"errors"
	"strings"
	"testing"
)

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
