package main

import (
	"strings"
	"unicode"
)

const (
	minHRRunes = 3
	markerLen  = 2
)

// tokenizeMarkdown highlights headings, fences, quotes, lists, tables, and
// inline emphasis. The whole file goes through one pass so a code fence and a
// front-matter block carry across lines.
func tokenizeMarkdown(lines []string) [][]segment {
	out := make([][]segment, len(lines))
	inFence := false
	inFront := len(lines) > 0 && strings.TrimSpace(lines[0]) == "---"

	for li, line := range lines {
		trimmed := strings.TrimSpace(line)

		if inFront && li > 0 && trimmed == "---" {
			inFront = false
			out[li] = finishLine([]segment{{Text: line, Class: classMeta}})

			continue
		}

		if inFront {
			out[li] = finishLine([]segment{{Text: line, Class: classMeta}})

			continue
		}

		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			out[li] = finishLine([]segment{{Text: line, Class: classMeta}})

			continue
		}

		if inFence {
			out[li] = finishLine([]segment{{Text: line, Class: classCode}})

			continue
		}

		out[li] = finishLine(markdownLine(line))
	}

	return out
}

// markdownLine highlights one line outside a fence.
func markdownLine(line string) []segment {
	var segs []segment

	r := []rune(line)
	trimmed := strings.TrimSpace(line)

	switch {
	case trimmed == "":
		return segs
	case isMarkdownHR(trimmed):
		return addSeg(segs, line, classHR)
	}

	if level, marker := headingLevel(r); level > 0 {
		segs = addSeg(segs, string(r[:marker]), classPunc)
		segs = addSeg(segs, string(r[marker:]), headingClass(level))

		return segs
	}

	i := 0
	for i < len(r) && r[i] == ' ' {
		i++
	}
	segs = addSeg(segs, string(r[:i]), classPlain)

	if i < len(r) && r[i] == '>' {
		j := i
		for j < len(r) && (r[j] == '>' || r[j] == ' ') {
			j++
		}

		segs = addSeg(segs, string(r[i:j]), classPunc)
		segs = append(segs, markdownInline(string(r[j:]))...)

		return segs
	}

	if i < len(r) && strings.ContainsRune("-*+", r[i]) && i+1 < len(r) && r[i+1] == ' ' {
		segs = addSeg(segs, string(r[i]), classPunc)
		segs = append(segs, markdownInline(string(r[i+1:]))...)

		return segs
	}

	if i < len(r) && unicode.IsDigit(r[i]) {
		j := i
		for j < len(r) && unicode.IsDigit(r[j]) {
			j++
		}

		if j < len(r) && r[j] == '.' && j+1 < len(r) && r[j+1] == ' ' {
			segs = addSeg(segs, string(r[i:j+1]), classPunc)
			segs = append(segs, markdownInline(string(r[j+1:]))...)

			return segs
		}
	}

	if i < len(r) && r[i] == '|' {
		segs = append(segs, markdownTable(r[i:])...)

		return segs
	}

	segs = append(segs, markdownInline(string(r[i:]))...)

	return segs
}

// headingLevel returns the heading level and the offset after the markers.
func headingLevel(r []rune) (int, int) {
	n := 0

	for n < len(r) && n < maxHeadingLevel && r[n] == '#' {
		n++
	}

	if n == 0 || n >= len(r) || r[n] != ' ' {
		return 0, 0
	}

	return n, n
}

// headingClass maps a level to its token class.
func headingClass(level int) string {
	classes := [...]string{classH1, classH2, classH3, classH4, classH5, classH6}
	if level < 1 || level > len(classes) {
		return classH1
	}

	return classes[level-1]
}

// isMarkdownHR reports whether a trimmed line is a thematic break.
func isMarkdownHR(trimmed string) bool {
	if len(trimmed) < minHRRunes {
		return false
	}

	c := trimmed[0]
	if c != '-' && c != '*' && c != '_' {
		return false
	}

	for i := range len(trimmed) {
		if trimmed[i] != c {
			return false
		}
	}

	return true
}

// markdownTable highlights pipes and a separator row.
func markdownTable(r []rune) []segment {
	var segs []segment

	separator := true

	for _, c := range r {
		if c != '|' && c != '-' && c != ':' && c != ' ' {
			separator = false

			break
		}
	}

	for i := 0; i < len(r); i++ {
		if r[i] == '|' {
			segs = addSeg(segs, "|", classPunc)
		} else {
			class := classPlain
			if separator && r[i] == '-' {
				class = classPunc
			}

			segs = addSeg(segs, string(r[i]), class)
		}
	}

	return segs
}

// markdownInline highlights code spans, emphasis, and links inside one line.
func markdownInline(text string) []segment {
	var segs []segment

	r := []rune(text)
	i := 0
	for i < len(r) {
		c := r[i]

		switch {
		case c == '\\' && i+1 < len(r):
			segs = addSeg(segs, string(r[i:i+2]), classPlain)
			i += 2
		case c == '`':
			end := indexRunes(r, i+1, "`", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classPlain)
				i = len(r)

				break
			}

			segs = addSeg(segs, string(r[i:end+1]), classCode)
			i = end + 1
		case hasRunes(r, i, "**") || hasRunes(r, i, "__"):
			marker := string(r[i : i+2])
			end := indexRunes(r, i+markerLen, marker, false)
			if end < 0 {
				segs = addSeg(segs, string(c), classPlain)
				i++

				break
			}

			segs = addSeg(segs, string(r[i:end+2]), classBold)
			i = end + markerLen
		case c == '[' || (c == '!' && i+1 < len(r) && r[i+1] == '['):
			open := i
			if c == '!' {
				open = i + 1
			}

			close := indexRunes(r, open+1, "](", false)
			if close >= 0 {
				if end := indexRunes(r, close+markerLen, ")", false); end >= 0 {
					segs = addSeg(segs, string(r[i:end+1]), classLink)
					i = end + 1

					break
				}
			}

			segs = addSeg(segs, string(c), classPlain)
			i++
		case c == '<':
			if end := indexRunes(r, i, ">", false); end >= 0 {
				segs = addSeg(segs, string(r[i:end+1]), classLink)
				i = end + 1
			} else {
				segs = addSeg(segs, "<", classPlain)
				i++
			}
		case c == 'h' && (hasRunes(r, i, "http://") || hasRunes(r, i, "https://")):
			j := i
			for j < len(r) && !unicode.IsSpace(r[j]) {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classLink)
			i = j
		case c == '*' || c == '_':
			end := indexRunes(r, i+1, string(c), false)
			if end <= i+1 || !emphasisOpen(r, i) || !emphasisClose(r, end) {
				segs = addSeg(segs, string(c), classPlain)
				i++

				break
			}

			segs = addSeg(segs, string(r[i:end+1]), classItal)
			i = end + 1
		default:
			j := i + 1
			for j < len(r) && !isMarkdownInlineStart(r, j) {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classPlain)
			i = j
		}
	}

	return segs
}

// emphasisOpen reports whether the marker at i starts at a word boundary.
func emphasisOpen(r []rune, i int) bool {
	if i == 0 {
		return true
	}

	return unicode.IsSpace(r[i-1]) || strings.ContainsRune("([{>", r[i-1])
}

// emphasisClose reports whether the marker at i ends at a word boundary.
func emphasisClose(r []rune, i int) bool {
	if i+1 >= len(r) {
		return true
	}

	return unicode.IsSpace(r[i+1]) || strings.ContainsRune(".,;:!?)]}", r[i+1])
}

// isMarkdownInlineStart reports whether r[i] may open an inline token.
func isMarkdownInlineStart(r []rune, i int) bool {
	c := r[i]

	if c == '`' || c == '[' || c == '!' || c == '*' || c == '_' || c == '<' || c == '\\' {
		return true
	}

	return c == 'h' && (hasRunes(r, i, "http://") || hasRunes(r, i, "https://"))
}
