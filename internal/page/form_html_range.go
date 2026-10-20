package page

import "strings"

// rewriteControlsState copies source, swapping each kept control for
// paintable HTML and adding the default form styles.
func rewriteControlsState(source string, spans []controlSpan, live map[string]Control, st caretState) string {
	if !hasFieldSpans(spans) {
		return source
	}

	css := formCSS
	kept := keepSpans(source, spans)
	cssAt := headOpen(source)
	if cssAt < 0 {
		cssAt = findHead(source)
	}
	if cssAt >= 0 && spanCovers(kept, cssAt) {
		cssAt = -1
	}

	var b strings.Builder
	prev := 0
	placed := false
	write := func(end int) {
		if !placed && cssAt >= prev && cssAt <= end {
			b.WriteString(source[prev:cssAt])
			b.WriteString(css)
			b.WriteString(source[cssAt:end])
			placed = true

			return
		}

		b.WriteString(source[prev:end])
	}
	for _, sp := range kept {
		write(sp.Start)
		ctrl := pickControl(sp, live)
		b.WriteString(replaceControlState(source[sp.Start:sp.End], ctrl, st))
		prev = sp.End
	}
	write(len(source))
	if !placed {
		return css + b.String()
	}

	return b.String()
}
