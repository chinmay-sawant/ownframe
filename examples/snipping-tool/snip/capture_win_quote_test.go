package snip

import "testing"

func TestPSQuoteEscapesSingleQuote(t *testing.T) {
	if got := psQuote(`C:\temp\plain.png`); got != `'C:\temp\plain.png'` {
		t.Fatalf("psQuote = %q", got)
	}
	if got := psQuote(`C:\temp\o'b.png`); got != `'C:\temp\o''b.png'` {
		t.Fatalf("psQuote = %q", got)
	}
	if got := psQuote(`a''b`); got != `'a''''b'` {
		t.Fatalf("psQuote = %q", got)
	}
}
