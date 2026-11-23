package page

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCheckDOMIgnoresRawTextBodies(t *testing.T) {
	t.Parallel()

	many := strings.Repeat("<b>", MaxDOMNodes+10)
	for _, tag := range []string{"style", "SCRIPT"} {
		html := "<" + tag + ">" + many + "</" + tag + "><p>x</p>"
		if err := checkDOMLimits(html); err != nil {
			t.Fatalf("<%s> body counted as tags: %v", tag, err)
		}
	}
	// An unclosed script swallows the rest and still terminates.
	if err := checkDOMLimits("<script>" + many); err != nil {
		t.Fatalf("unclosed script: %v", err)
	}
	// Tags after the closing tag still count.
	if err := checkDOMLimits("<style>x</style>" + strings.Repeat("<i>", MaxDOMDepth+5)); !errors.Is(err, ErrDOMTooDeep) {
		t.Fatalf("depth after style: %v", err)
	}
}

// TestRedrawChecksExpandedTemplate keeps New's small template under the
// limits while {{range}} expands past them.
func TestRedrawChecksExpandedTemplate(t *testing.T) {
	t.Parallel()

	p, err := New(Config{
		HTML:   `<body>{{range .}}<p>x</p>{{end}}</body>`,
		Width:  200,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	p.SetData([]int{1})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.SetData(make([]int, MaxDOMNodes+10))
	if err := p.Redraw(ctx); !errors.Is(err, ErrTooManyNodes) {
		t.Fatalf("expanded template: %v, want ErrTooManyNodes", err)
	}
}
