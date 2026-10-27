package main

import (
	"strings"
	"unicode"
)

const cssVarPrefix = 2

var (
	cssValueWords = keywordSet(
		"auto", "inherit", "initial", "none", "normal", "revert", "unset",
		"absolute", "relative", "fixed", "sticky", "block", "inline", "flex",
		"grid", "hidden", "visible", "center", "bold", "italic", "transparent",
	)
	cssRuleAtRules = keywordSet("@media", "@supports", "@keyframes", "@layer", "@container", "@scope", "@document")
)

// scanLineCSS highlights one line of CSS given the current scanner mode and comment state.
func scanLineCSS(state lineState, line string) ([]segment, lineState) {
	inComment := state.inComment
	mode := state.mode
	if mode == "" {
		mode = "selector"
	}
	container := state.container

	r := []rune(line)
	var segs []segment

	i := 0
lineLoop:
	for i < len(r) {
		if inComment {
			end := indexRunes(r, i, "*/", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classCom)

				break lineLoop
			}

			segs = addSeg(segs, string(r[i:end+len("*/")]), classCom)
			i = end + len("*/")
			inComment = false

			continue
		}

		c := r[i]

		switch {
		case unicode.IsSpace(c):
			j := i + 1
			for j < len(r) && unicode.IsSpace(r[j]) {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classPlain)
			i = j
		case hasRunes(r, i, "/*"):
			end := indexRunes(r, i+len("/*"), "*/", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classCom)
				inComment = true

				break lineLoop
			}

			segs = addSeg(segs, string(r[i:end+len("*/")]), classCom)
			i = end + len("*/")
		case c == '"' || c == '\'':
			end := indexRunes(r, i+1, string(c), true)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classStr)

				break lineLoop
			}

			segs = addSeg(segs, string(r[i:end+1]), classStr)
			i = end + 1
		case mode == "value":
			switch {
			case c == ';':
				segs = addSeg(segs, ";", classPunc)
				mode = "prop"
				i++
			case c == '}':
				segs = addSeg(segs, "}", classPunc)
				mode = "selector"
				i++
			case c == '#':
				j := i + 1
				for j < len(r) && isHexRune(r[j]) {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classNum)
				i = j
			case c == '!' && hasRunes(r, i, "!important"):
				segs = addSeg(segs, "!important", classKw)
				i += len("!important")
			case c == '-' && i+1 < len(r) && r[i+1] == '-':
				j := scanCSSVar(r, i)
				segs = addSeg(segs, string(r[i:j]), classVar)
				i = j
			case unicode.IsDigit(c) || (c == '.' && i+1 < len(r) && unicode.IsDigit(r[i+1])):
				j := scanCSSNumber(r, i)
				segs = addSeg(segs, string(r[i:j]), classNum)
				i = j
			case isIdentStart(c):
				j := i + 1
				for j < len(r) && (isIdentRune(r[j]) || r[j] == '-') {
					j++
				}

				word := string(r[i:j])
				class := classPlain

				if j < len(r) && r[j] == '(' {
					class = classFn
				} else if cssValueWords[word] {
					class = classKw
				}

				segs = addSeg(segs, word, class)
				i = j
			case strings.ContainsRune("(),/+*", c):
				segs = addSeg(segs, string(c), classPunc)
				i++
			default:
				segs = addSeg(segs, string(c), classPlain)
				i++
			}
		case mode == "prop":
			switch {
			case c == ':':
				segs = addSeg(segs, ":", classPunc)
				mode = "value"
				i++
			case c == ';':
				segs = addSeg(segs, ";", classPunc)
				i++
			case c == '}':
				segs = addSeg(segs, "}", classPunc)
				mode = "selector"
				i++
			case c == '{':
				segs = addSeg(segs, "{", classPunc)
				i++
			case isIdentStart(c) || (c == '-' && i+1 < len(r) && r[i+1] == '-') || c == '-':
				j := i + 1
				for j < len(r) && (isIdentRune(r[j]) || r[j] == '-') {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classAttr)
				i = j
			case unicode.IsSpace(c):
				segs = addSeg(segs, string(c), classPlain)
				i++
			default:
				segs = addSeg(segs, string(c), classPunc)
				i++
			}
		default: // selector
			switch {
			case c == '@':
				j := i + 1
				for j < len(r) && isIdentRune(r[j]) {
					j++
				}

				word := string(r[i:j])
				segs = addSeg(segs, word, classKw)
				container = cssRuleAtRules[word]
				i = j
			case c == '{':
				segs = addSeg(segs, "{", classPunc)
				if container {
					mode = "selector"
				} else {
					mode = "prop"
				}

				container = false
				i++
			case c == '}':
				segs = addSeg(segs, "}", classPunc)
				mode = "selector"
				i++
			case c == ';':
				segs = addSeg(segs, ";", classPunc)
				i++
			case c == '.' || c == '#' || c == ':' || c == '[':
				j := i + 1
				for j < len(r) && (isIdentRune(r[j]) || r[j] == '-' || r[j] == ']' || r[j] == '"' || r[j] == '\'') {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classTag)
				i = j
			case c == '>' || c == '+' || c == '~' || c == '*':
				segs = addSeg(segs, string(c), classOp)
				i++
			case c == ',' || c == '(' || c == ')':
				segs = addSeg(segs, string(c), classPunc)
				i++
			case isIdentStart(c):
				j := i + 1
				for j < len(r) && (isIdentRune(r[j]) || r[j] == '-') {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classTag)
				i = j
			default:
				segs = addSeg(segs, string(c), classPlain)
				i++
			}
		}
	}

	return finishLine(segs), lineState{inComment: inComment, mode: mode, container: container}
}

// tokenizeCSS highlights selectors, declarations, and values. It tracks one
// mode at a time rather than a full parser: selector, property, or value.
func tokenizeCSS(lines []string) [][]segment {
	out := make([][]segment, len(lines))
	st := lineState{mode: "selector"}

	for li, line := range lines {
		out[li], st = scanLineCSS(st, line)
	}

	return out
}

// scanCSSVar returns the end of one custom property name.
func scanCSSVar(r []rune, i int) int {
	j := i + cssVarPrefix

	for j < len(r) && (isIdentRune(r[j]) || r[j] == '-') {
		j++
	}

	return j
}

// scanCSSNumber returns the end of one number plus its unit or percent.
func scanCSSNumber(r []rune, i int) int {
	j := i

	for j < len(r) && (unicode.IsDigit(r[j]) || r[j] == '.' || r[j] == '_') {
		j++
	}

	for j < len(r) && (unicode.IsLetter(r[j]) || r[j] == '%') {
		j++
	}

	return j
}

// isHexRune reports whether r is a hex digit.
func isHexRune(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

// tokenizeJSON highlights keys, strings, numbers, and literals.
func tokenizeJSON(lines []string) [][]segment {
	out := make([][]segment, len(lines))

	for li, line := range lines {
		r := []rune(line)
		var segs []segment

		i := 0
		for i < len(r) {
			c := r[i]

			switch {
			case unicode.IsSpace(c):
				j := i + 1
				for j < len(r) && unicode.IsSpace(r[j]) {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classPlain)
				i = j
			case c == '"':
				end := indexRunes(r, i+1, "\"", true)
				if end < 0 {
					end = len(r) - 1
				}

				text := string(r[i : end+1])
				k := end + 1
				for k < len(r) && unicode.IsSpace(r[k]) {
					k++
				}

				class := classStr
				if k < len(r) && r[k] == ':' {
					class = classAttr
				}

				segs = addSeg(segs, text, class)
				i = end + 1
			case c == '-':
				segs = addSeg(segs, "-", classOp)
				i++
			case unicode.IsDigit(c):
				j := scanNumber(r, i)
				segs = addSeg(segs, string(r[i:j]), classNum)
				i = j
			case isIdentStart(c):
				j := i + 1
				for j < len(r) && isIdentRune(r[j]) {
					j++
				}

				word := string(r[i:j])
				class := classPlain
				if word == "true" || word == "false" || word == "null" {
					class = classLit
				}

				segs = addSeg(segs, word, class)
				i = j
			case strings.ContainsRune("{}[],:", c):
				segs = addSeg(segs, string(c), classPunc)
				i++
			default:
				segs = addSeg(segs, string(c), classPlain)
				i++
			}
		}

		out[li] = finishLine(segs)
	}

	return out
}

// scanLineHTML colours one line of HTML given the active tag and comment continuation state.
func scanLineHTML(state lineState, line string) ([]segment, lineState) {
	inComment := state.inComment
	inTag := state.inTag

	r := []rune(line)
	var segs []segment

	i := 0
htmlLoop:
	for i < len(r) {
		if inComment {
			end := indexRunes(r, i, "-->", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classCom)

				break htmlLoop
			}

			segs = addSeg(segs, string(r[i:end+len("-->")]), classCom)
			i = end + len("-->")
			inComment = false

			continue
		}

		c := r[i]

		switch {
		case hasRunes(r, i, "<!--") && !inTag:
			end := indexRunes(r, i+len("<!--"), "-->", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classCom)
				inComment = true

				break htmlLoop
			}

			segs = addSeg(segs, string(r[i:end+len("-->")]), classCom)
			i = end + len("-->")
		case !inTag && (c == '<' && i+1 < len(r) && (isIdentStart(r[i+1]) || r[i+1] == '/' || r[i+1] == '!')):
			inTag = true
			j := i + 1
			if j < len(r) && r[j] == '/' {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classPunc)
			i = j

			if i < len(r) && r[i] == '!' {
				j = i + 1
				for j < len(r) && unicode.IsLetter(r[j]) {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classMeta)
				i = j
			} else if i < len(r) && isIdentStart(r[i]) {
				j := i + 1
				for j < len(r) && (isIdentRune(r[j]) || r[j] == '-') {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classTag)
				i = j
			}
		case inTag:
			switch {
			case c == '>':
				segs = addSeg(segs, ">", classPunc)
				inTag = false
				i++
			case c == '/' && i+1 < len(r) && r[i+1] == '>':
				segs = addSeg(segs, "/>", classPunc)
				inTag = false
				i += 2
			case c == '=':
				segs = addSeg(segs, "=", classOp)
				i++
			case c == '"' || c == '\'':
				end := indexRunes(r, i+1, string(c), false)
				if end < 0 {
					segs = addSeg(segs, string(r[i:]), classStr)

					break htmlLoop
				}

				segs = addSeg(segs, string(r[i:end+1]), classStr)
				i = end + 1
			case unicode.IsSpace(c):
				j := i + 1
				for j < len(r) && unicode.IsSpace(r[j]) {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classPlain)
				i = j
			case isIdentStart(c):
				j := i + 1
				for j < len(r) && (isIdentRune(r[j]) || r[j] == '-' || r[j] == ':') {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classAttr)
				i = j
			default:
				segs = addSeg(segs, string(c), classVal)
				i++
			}
		default:
			j := i + 1
			for j < len(r) && r[j] != '<' {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classPlain)
			i = j
		}
	}

	return finishLine(segs), lineState{inComment: inComment, inTag: inTag}
}

// tokenizeHTML colours tags, attributes, and quoted values outside comments.
func tokenizeHTML(lines []string) [][]segment {
	out := make([][]segment, len(lines))
	var st lineState

	for li, line := range lines {
		out[li], st = scanLineHTML(st, line)
	}

	return out
}

// tokenizeYAML colours mapping keys, comments, anchors, and scalars.
func tokenizeYAML(lines []string) [][]segment {
	out := make([][]segment, len(lines))

	for li, line := range lines {
		r := []rune(line)
		var segs []segment

		i := 0
		for i < len(r) && r[i] == ' ' {
			i++
		}
		segs = addSeg(segs, string(r[:i]), classPlain)

		if i >= len(r) {
			out[li] = finishLine(segs)

			continue
		}

		switch {
		case r[i] == '#':
			segs = addSeg(segs, string(r[i:]), classCom)
		case hasRunes(r, i, "---") || hasRunes(r, i, "..."):
			segs = addSeg(segs, string(r[i:]), classMeta)
		case r[i] == '-':
			segs = addSeg(segs, "-", classPunc)
			i++
			segs = append(segs, tokenizeYAMLValue(r, i, true)...)
		case r[i] == '&' || r[i] == '*':
			j := i + 1
			for j < len(r) && (isIdentRune(r[j]) || r[j] == '-') {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classVar)
			segs = append(segs, tokenizeYAMLValue(r, j, false)...)
		default:
			key := scanYAMLKey(r, i)
			if key > i && key < len(r) && r[key] == ':' {
				segs = addSeg(segs, string(r[i:key]), classAttr)
				segs = addSeg(segs, ":", classPunc)
				segs = append(segs, tokenizeYAMLValue(r, key+1, false)...)
			} else {
				segs = append(segs, tokenizeYAMLValue(r, i, true)...)
			}
		}

		out[li] = finishLine(segs)
	}

	return out
}

// scanYAMLKey returns the end of a plain mapping key.
func scanYAMLKey(r []rune, i int) int {
	j := i

	for j < len(r) && r[j] != ':' && r[j] != '#' {
		j++
	}

	for j > i && unicode.IsSpace(r[j-1]) {
		j--
	}

	return j
}

// tokenizeYAMLValue colours a scalar, a quoted string, or an inline comment.
func tokenizeYAMLValue(r []rune, i int, trim bool) []segment {
	var segs []segment

	if trim {
		if i < len(r) && r[i] == '#' {
			return addSeg(segs, string(r[i:]), classCom)
		}
	}

	j := i
	for j < len(r) && r[j] != '#' {
		j++
	}

	value := r[i:j]
	if len(value) == 0 {
		return segs
	}

	segs = addSeg(segs, string(value), classVal)

	if j < len(r) && r[j] == '#' {
		segs = addSeg(segs, string(r[j:]), classCom)
	}

	return segs
}

// tokenizeShell colours keywords, variables, strings, and comments.
func tokenizeShell(lines []string) [][]segment {
	out := make([][]segment, len(lines))
	keywords := keywordSet(
		"case", "do", "done", "elif", "else", "esac", "fi", "for", "function",
		"if", "in", "local", "return", "select", "then", "until", "while",
		"export", "readonly", "shift", "source", "set", "unset",
	)

	for li, line := range lines {
		r := []rune(line)
		var segs []segment

		i := 0
	shellLoop:
		for i < len(r) {
			c := r[i]

			switch {
			case unicode.IsSpace(c):
				j := i + 1
				for j < len(r) && unicode.IsSpace(r[j]) {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classPlain)
				i = j
			case c == '#' && (i == 0 || unicode.IsSpace(r[i-1])):
				segs = addSeg(segs, string(r[i:]), classCom)

				break shellLoop
			case c == '"' || c == '\'':
				end := indexRunes(r, i+1, string(c), c == '"')
				if end < 0 {
					end = len(r) - 1
				}

				segs = addSeg(segs, string(r[i:end+1]), classStr)
				i = end + 1
			case c == '$':
				j := i + 1
				if j < len(r) && r[j] == '{' {
					for j < len(r) && r[j] != '}' {
						j++
					}

					if j < len(r) {
						j++
					}
				} else {
					for j < len(r) && (isIdentRune(r[j]) || r[j] == '?' || r[j] == '#' || unicode.IsDigit(r[j])) {
						j++
					}
				}

				segs = addSeg(segs, string(r[i:j]), classVar)
				i = j
			case unicode.IsDigit(c):
				j := i + 1
				for j < len(r) && (unicode.IsDigit(r[j]) || r[j] == '.') {
					j++
				}

				segs = addSeg(segs, string(r[i:j]), classNum)
				i = j
			case isIdentStart(c):
				j := i + 1
				for j < len(r) && isIdentRune(r[j]) {
					j++
				}

				word := string(r[i:j])
				class := classPlain
				if keywords[word] {
					class = classKw
				}

				segs = addSeg(segs, word, class)
				i = j
			case strings.ContainsRune("|&;><="+`+*[](){}!`, c):
				segs = addSeg(segs, string(c), classOp)
				i++
			default:
				segs = addSeg(segs, string(c), classPlain)
				i++
			}
		}

		out[li] = finishLine(segs)
	}

	return out
}

// scanLineJSON highlights one line of JSON.
func scanLineJSON(line string) []segment {
	out := tokenizeJSON([]string{line})
	if len(out) > 0 {
		return out[0]
	}
	return finishLine([]segment{{Text: line, Class: classPlain}})
}

// scanLineYAML highlights one line of YAML.
func scanLineYAML(line string) []segment {
	out := tokenizeYAML([]string{line})
	if len(out) > 0 {
		return out[0]
	}
	return finishLine([]segment{{Text: line, Class: classPlain}})
}

// scanLineShell highlights one line of shell.
func scanLineShell(line string) []segment {
	out := tokenizeShell([]string{line})
	if len(out) > 0 {
		return out[0]
	}
	return finishLine([]segment{{Text: line, Class: classPlain}})
}
