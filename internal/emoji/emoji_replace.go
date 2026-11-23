package emoji

import "strings"

// imgOpen and imgClose wrap one supported sequence. The data attribute
// carries the 1em sizing rule, and the alt keeps the raw runes for boxes
// and readers that do not paint the image.
const (
	imgOpen  = `<img data-ownframe-emoji="1" src="emoji/`
	imgMid   = `" alt="`
	imgClose = `">`
)

// Replace swaps supported emoji sequences in the text nodes of html for
// image tags and reports whether anything changed. Tags, comments, and
// script, style, textarea, and title content pass through untouched, so attribute values
// such as value="😂" are never rewritten.
func Replace(html string) (string, bool) {
	if !maybeHasEmoji(html) {
		return html, false
	}

	var b strings.Builder

	b.Grow(len(html))

	changed := false
	i := 0

	for i < len(html) {
		c := html[i]
		switch {
		case c == '<' && strings.HasPrefix(html[i:], "<!--"):
			end := strings.Index(html[i:], "-->")
			if end < 0 {
				// An unclosed comment runs to the end of the input.
				b.WriteString(html[i:])
				i = len(html)

				continue
			}

			b.WriteString(html[i : i+end+len("-->")])
			i += end + len("-->")
		case c == '<' && tagStart(html, i):
			open := i+1 >= len(html) || html[i+1] != '/'
			name, end := tagName(html, i)
			b.WriteString(html[i:end])
			i = end

			if open && rawTextTag(name) {
				rest, ok := rawText(html, i, name)
				b.WriteString(rest)

				if ok {
					i += len(rest)
				} else {
					i = len(html)
				}
			}
		case c < asciiLimit:
			b.WriteByte(c)
			i++
		default:
			seq, file := match(html, i)
			if file == "" {
				_, size := decodeRune(html, i)
				b.WriteString(html[i : i+size])
				i += size

				continue
			}

			b.WriteString(imgOpen)
			b.WriteString(file)
			b.WriteString(imgMid)
			b.WriteString(seq)
			b.WriteString(imgClose)
			i += len(seq)
			changed = true
		}
	}

	return b.String(), changed
}
