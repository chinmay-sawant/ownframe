//go:build linux && !android

package filepick

import (
	"strings"
	"testing"
)

func TestWindowScriptEscapesSingleQuote(t *testing.T) {
	t.Parallel()

	title := `it's; $(Remove-Item C:\x)`
	got := windowScript(title)
	quoted := strings.ReplaceAll(title, "'", "''")
	want := `$f.Title = '` + quoted + `';`
	if !strings.Contains(got, want) {
		t.Fatalf("got %q, want fragment %q", got, want)
	}
}
