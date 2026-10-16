package eyes

import (
	"strings"
	"testing"
)

func TestControlsStayRightAlignedAndBelowTopEdge(t *testing.T) {
	html := buildHTML()
	topbar, eyes := strings.Index(html, `id="topbar"`), strings.Index(html, `id="eyes"`)
	if topbar < 0 || eyes < 0 || topbar >= eyes ||
		!strings.Contains(html, `id="dev-toggle"`) ||
		!strings.Contains(html, `id="mode-toggle"`) ||
		strings.Contains(html, `id="brow-left"`) ||
		strings.Contains(html, `id="brow-right"`) ||
		strings.Contains(html, "info-") || strings.Contains(stylesCSS, ".info-panel") ||
		!strings.Contains(stylesCSS, "align-items: flex-end") ||
		!strings.Contains(stylesCSS, "padding: 28px 14px 0 0") {
		t.Fatal("controls lost their right alignment, lower offset, or two-button layout")
	}
}
