package page

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

func TestRewriteButtonsEscapedQuotesRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="submit" id="go" value="a&quot;b" data-label="<b>not a tag</b>">`
	got := rewriteButtons(src)

	if !strings.Contains(got, "<button") {
		t.Fatalf("rewriteButtons dropped button: %q", got)
	}

	// Attribute text is part of the serialized attribute value, not executable
	// markup in the rewritten output.
	if got == src {
		t.Fatalf("button was not rewritten: %q", got)
	}
}

func TestRewriteButtonsNestedInputAttrsRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="submit" id="go" value="run" data-x="<input type='text' value='nested'>">`
	got := rewriteButtons(src)

	if strings.Count(got, "<button") != 1 {
		t.Fatalf("expected one button, got %q", got)
	}

	if strings.Count(got, "<button") != 1 {
		t.Fatalf("expected one button, got %q", got)
	}
}

func TestScanControlsNestedInputAttrsRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="text" id="name" value="one" data-x="<input type='text' id='evil'>">`
	spans := form.Scan(src)

	if len(spans) != 1 || spans[0].Control.ID != "name" {
		t.Fatalf("spans = %#v", spans)
	}
}

func TestScanControlsDoubleEscapedQuotesRegression(t *testing.T) {
	t.Parallel()

	src := `<input type="text" id="q" value="&amp;quot;quoted&amp;quot;">`
	spans := form.Scan(src)

	if len(spans) != 1 {
		t.Fatalf("spans = %#v", spans)
	}

	// The value is unescaped exactly once by the scanner, so a double-
	// escaped quote stays escaped rather than becoming a live quote.
	if strings.Contains(spans[0].Control.Value, `"`) {
		t.Fatalf("quote unexpectedly unescaped twice: %q", spans[0].Control.Value)
	}
}

func TestAttrValueDoubleEscapedQuotesRegression(t *testing.T) {
	t.Parallel()

	raw := `<input value="&amp;quot;x&amp;quot;">`
	got := form.AttrValue(raw, "value")

	if got == `"x"` {
		t.Fatalf("attrValue collapsed double escaping: %q", got)
	}
}
