package page

import "testing"

func TestInjectThemeEscapesStyleBreakout(t *testing.T) {
	t.Parallel()

	source := `<html><head></head></html>`
	theme := `p{content:"</style>"}`
	want := `<html><head><style>p{content:&#34;&lt;/style&gt;&#34;}</style></head></html>`

	if got := injectTheme(source, theme); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
