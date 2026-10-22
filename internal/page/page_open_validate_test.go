package page

import (
	"strings"
	"testing"
)

func TestValidateOpenTarget(t *testing.T) {
	t.Parallel()
	for _, ok := range []string{"https://example.test/p", "http://a.b/", "mailto:a@b.c"} {
		if _, err := ValidateOpenTarget(ok); err != nil {
			t.Fatalf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"javascript:alert(1)", "file:///etc/passwd", "data:text/html,hi", "  https://a.b/ ", "relative/path"} {
		if _, err := ValidateOpenTarget(bad); err == nil {
			t.Fatalf("%s allowed", strings.TrimSpace(bad))
		}
	}
}
