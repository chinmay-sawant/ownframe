package main

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	indentStepPx   = 10
	cellTextPadPx  = 4
	cellLineHeight = 21
	wrapLineHeight = 15
)

// fontFamilyCSS maps Style.FontFamily to a CSS font stack ("" = inherit).
func fontFamilyCSS(family string) string {
	switch family {
	case "mono":
		return "ui-monospace, 'DejaVu Sans Mono', Menlo, Consolas, monospace"
	case "serif":
		return "Georgia, 'DejaVu Serif', 'Times New Roman', serif"
	case "sans":
		return "'DejaVu Sans', Arial, Helvetica, sans-serif"
	}
	return ""
}

// edgeCSS renders one border edge declaration value ("" = none).
func edgeCSS(style string) string {
	switch style {
	case EdgeThin:
		return "1px solid var(--xl-cell-text)"
	case EdgeThick:
		return "2px solid var(--xl-cell-text)"
	case EdgeDouble:
		return "3px double var(--xl-cell-text)"
	}
	return ""
}

// vAlignHalves splits the free height evenly for middle alignment.
const vAlignHalves = 2

// wrapBorderPx is the border allowance subtracted when measuring wrap width.
const wrapBorderPx = 2

// avgGlyphEm is the average glyph width as a fraction of the font size.
const avgGlyphEm = 0.56

// wrapLineCount estimates how many lines text takes in a cell of the given
// pixel width at the font size (0 = default). It breaks on words and on "\n".
func wrapLineCount(text string, widthPx, fontSize int) int {
	if fontSize <= 0 {
		fontSize = defaultFontSize
	}
	avail := widthPx - 2*cellTextPadPx - wrapBorderPx
	charW := float64(fontSize) * avgGlyphEm
	perLine := max(int(float64(avail)/charW), 1)
	lines := 0
	for _, para := range strings.Split(text, "\n") {
		n, cur := 1, 0
		for _, w := range strings.Fields(para) {
			l := len([]rune(w))
			switch {
			case cur == 0:
				cur = l
			case cur+1+l <= perLine:
				cur += 1 + l
			default:
				n++
				cur = l
			}
			for cur > perLine {
				n++
				cur -= perLine
			}
		}
		lines += n
	}
	return max(lines, 1)
}

// applyStyleExtras sets the style properties that the base cell classes do
// not cover: font family, per-edge borders, vertical alignment and indent.
// cellH is the painted cell height and text the displayed text (used to place
// wrapped text). The grid calls it for every painted cell.
func (a *app) applyStyleExtras(el *ui.Element, st Style, cellW, cellH int, text string) {
	if f := fontFamilyCSS(st.FontFamily); f != "" {
		el.CSSRaw("font-family", f)
	}
	e := st.BorderEdges()
	if v := edgeCSS(e.Left); v != "" {
		el.CSSRaw("border-left", v)
	}
	if v := edgeCSS(e.Right); v != "" {
		el.CSSRaw("border-right", v)
	}
	if v := edgeCSS(e.Top); v != "" {
		el.CSSRaw("border-top", v)
	}
	if v := edgeCSS(e.Bottom); v != "" {
		el.CSSRaw("border-bottom", v)
	}
	if st.Indent > 0 {
		el.CSSRaw("padding-left", pxStr(a.zs(cellTextPadPx+st.Indent*indentStepPx)))
	}
	textH := a.zs(cellLineHeight)
	if st.Wrap {
		textH = a.zs(wrapLineCount(text, a.unz(cellW)-st.Indent*indentStepPx, st.FontSize) * wrapLineHeight)
	}
	pad := 0
	switch st.VAlign {
	case "middle":
		pad = (cellH - textH) / vAlignHalves
	case "bottom":
		pad = cellH - textH - 1
	}
	if pad > 0 {
		el.CSSRaw("padding-top", pxStr(pad))
	}
}
