package page

import (
	"context"
	"strings"
	"testing"
)

// A Windows-style session: preedit grows, the candidate anchor follows it,
// cancel leaves the value alone, and commit replaces the preedit.
func TestIMECompositionLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := imeField(t, "ab")

	idle, ok := p.IMEAnchor()
	if !ok {
		t.Fatal("missing idle anchor")
	}

	// The clause cursor sits at the end of the preedit, as ebiten reports
	// before the user selects a clause.
	if err := p.SetIMEComposition(ctx, "かな", 6, 6); err != nil {
		t.Fatal(err)
	}

	moved, ok := p.IMEAnchor()
	if !ok || moved.X <= idle.X {
		t.Fatalf("candidate anchor did not follow preedit: idle=%v composing=%v", idle.X, moved.X)
	}

	if got := p.FormValue("e"); got != "ab" {
		t.Fatalf("value during composition = %q", got)
	}

	// Cancel.
	if err := p.SetIMEComposition(ctx, "", 0, 0); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("e"); got != "ab" {
		t.Fatalf("value after cancel = %q", got)
	}

	back, _ := p.IMEAnchor()
	if back.X != idle.X {
		t.Fatalf("anchor after cancel = %v, want %v", back.X, idle.X)
	}

	// Commit: the window clears the preedit, then types the result.
	if err := p.SetIMEComposition(ctx, "かな", 0, 6); err != nil {
		t.Fatal(err)
	}

	if err := p.SetIMEComposition(ctx, "", 0, 0); err != nil {
		t.Fatal(err)
	}

	if err := p.Type(ctx, "仮名"); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("e"); got != "ab仮名" {
		t.Fatalf("value after commit = %q", got)
	}

	if strings.Contains(p.source, `data-ownframe-composition="1"`) {
		t.Fatal("preedit remained after commit")
	}
}

func TestIMECompositionNeedsFocusedField(t *testing.T) {
	t.Parallel()

	p := bedPage(t, `<input id="e" type="text"`+bedWide+`>`, nil)
	if err := p.SetIMEComposition(context.Background(), "か", 0, 3); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(p.source, `data-ownframe-composition="1"`) {
		t.Fatal("unfocused page painted a preedit")
	}
}
