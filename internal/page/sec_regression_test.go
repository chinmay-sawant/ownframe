package page

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestSec_PageSourceSizeCapped pins the template size cap.
func TestSec_PageSourceSizeCapped(t *testing.T) {
	t.Parallel()
	if _, err := New(Config{HTML: strings.Repeat("x", MaxSourceSize+1), Width: 10, Height: 10}); !errors.Is(err, ErrSourceTooLarge) {
		t.Fatalf("err = %v", err)
	}
}

// TestSec_PageDOMLimits pins the nesting and node caps at every parse
// entry: New, Load, and Route registration.
func TestSec_PageDOMLimits(t *testing.T) {
	t.Parallel()
	deep := strings.Repeat("<div>", MaxDOMDepth+1) + "x" + strings.Repeat("</div>", MaxDOMDepth+1)
	if _, err := New(Config{HTML: deep, Width: 10, Height: 10}); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatalf("deep: %v", err)
	}
	many := strings.Repeat("<i>x</i>", MaxDOMNodes+1)
	if _, err := New(Config{HTML: many, Width: 10, Height: 10}); !errors.Is(err, ErrTooManyNodes) {
		t.Fatalf("nodes: %v", err)
	}
	p, err := New(Config{HTML: "<p>x</p>", Width: 10, Height: 10})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Load(context.Background(), deep); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatalf("load: %v", err)
	}
	if err := p.Route("a", many); !errors.Is(err, ErrTooManyNodes) {
		t.Fatalf("route: %v", err)
	}
}
