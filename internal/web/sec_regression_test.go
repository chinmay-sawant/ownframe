package web

import (
	"net/http"
	"strings"
	"testing"
)

// TestSec_WebMissingOriginRejected pins the CSRF gate (mutating routes
// fail closed without a matching Origin).
func TestSec_WebMissingOriginRejected(t *testing.T) {
	t.Parallel()
	if rec := clickRec(&auditClick{}, "http://example.com/click?x=1&y=2", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("click: %d", rec.Code)
	}
	if rec := mutateRec(t, "/type", "", &auditKbd{}); rec.Code != http.StatusForbidden {
		t.Fatalf("type: %d", rec.Code)
	}
	if rec := mutateRec(t, "/backspace", "", &auditKbd{}); rec.Code != http.StatusForbidden {
		t.Fatalf("backspace: %d", rec.Code)
	}
}

// TestSec_WebClickBoundsRejected pins the click query caps.
func TestSec_WebClickBoundsRejected(t *testing.T) {
	t.Parallel()
	many := "http://example.com/click?x=1&y=2"
	for i := 0; i < 20; i++ {
		many += "&k=1"
	}
	if rec := clickRec(&auditClick{}, many, "http://example.com"); rec.Code != http.StatusBadRequest {
		t.Fatalf("flood: %d", rec.Code)
	}
	long := "http://example.com/click?x=" + strings.Repeat("1", 65) + "&y=2"
	if rec := clickRec(&auditClick{}, long, "http://example.com"); rec.Code != http.StatusBadRequest {
		t.Fatalf("long: %d", rec.Code)
	}
}
