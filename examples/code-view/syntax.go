package main

import (
	"strings"
	"unicode"
)

// segment is one run of source text with one highlight class. Templates print
// it as <span class="t-{{.Class}}">{{.Text}}</span>; every value is a plain
// string, so html/template escapes it.
type segment struct {
	Text  string
	Class string
}

const (
	classPlain = "plain"
	classKw    = "kw"
	classType  = "type"
	classFn    = "fn"
	classStr   = "str"
	classNum   = "num"
	classCom   = "com"
	classOp    = "op"
	classPunc  = "punc"
	classVar   = "var"
	classLit   = "lit"
	classDec   = "dec"
	classTag   = "tag"
	classAttr  = "attr"
	classVal   = "val"
	classMeta  = "meta"
	classH1    = "h1"
	classH2    = "h2"
	classH3    = "h3"
	classH4    = "h4"
	classH5    = "h5"
	classH6    = "h6"
	classBold  = "bold"
	classItal  = "ital"
	classCode  = "code"
	classLink  = "link"
	classQuote = "quote"
	classHR    = "hr"
)

const (
	langGo       = "go"
	langPython   = "python"
	langJS       = "javascript"
	langMarkdown = "markdown"
	langCSS      = "css"
	langJSON     = "json"
	langHTML     = "html"
	langYAML     = "yaml"
	langShell    = "shell"
	langPlain    = "plain"
)

// langInfo is the label and the two-letter badge for one language.
type langInfo struct {
	Label string
	Badge string
}

// languageOf maps a path to the language the viewer highlights. Unknown text
// files use the plain tokenizer; binary files never reach this function.
func languageOf(path string) string {
	name := strings.ToLower(path)

	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}

	switch {
	case strings.HasSuffix(name, ".go") || strings.HasSuffix(name, ".mod") || strings.HasSuffix(name, ".sum"):
		if strings.HasSuffix(name, ".mod") || strings.HasSuffix(name, ".sum") {
			return langPlain
		}

		return langGo
	case strings.HasSuffix(name, ".py") || strings.HasSuffix(name, ".pyi"):
		return langPython
	case strings.HasSuffix(name, ".js") || strings.HasSuffix(name, ".mjs") ||
		strings.HasSuffix(name, ".cjs") || strings.HasSuffix(name, ".jsx") ||
		strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".tsx"):
		return langJS
	case strings.HasSuffix(name, ".md") || strings.HasSuffix(name, ".markdown"):
		return langMarkdown
	case strings.HasSuffix(name, ".css") || strings.HasSuffix(name, ".scss"):
		return langCSS
	case strings.HasSuffix(name, ".json"):
		return langJSON
	case strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm"):
		return langHTML
	case strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml"):
		return langYAML
	case strings.HasSuffix(name, ".sh") || strings.HasSuffix(name, ".bash") || name == "makefile":
		return langShell
	default:
		return langPlain
	}
}

// languageInfo is the display label and badge for a language key.
func languageInfo(lang string) langInfo {
	switch lang {
	case langGo:
		return langInfo{Label: "Go", Badge: "Go"}
	case langPython:
		return langInfo{Label: "Python", Badge: "Py"}
	case langJS:
		return langInfo{Label: "JavaScript", Badge: "JS"}
	case langMarkdown:
		return langInfo{Label: "Markdown", Badge: "Md"}
	case langCSS:
		return langInfo{Label: "CSS", Badge: "Cs"}
	case langJSON:
		return langInfo{Label: "JSON", Badge: "{}"}
	case langHTML:
		return langInfo{Label: "HTML", Badge: "<>"}
	case langYAML:
		return langInfo{Label: "YAML", Badge: "Ya"}
	case langShell:
		return langInfo{Label: "Shell", Badge: "Sh"}
	default:
		return langInfo{Label: "Plain text", Badge: "Tx"}
	}
}

// lineState captures the cross-line scanner continuation state.
type lineState struct {
	// C-family (Go, Python, JS)
	inBlock   bool
	inString  string
	pending   string
	funcAwait bool
	paren     int

	// Markdown
	inFence bool
	inFront bool

	// CSS
	mode      string
	container bool

	// HTML & CSS
	inComment bool
	inTag     bool
}

// initialLineState returns the starting scanner state for a language.
func initialLineState(lang string) lineState {
	if lang == langCSS {
		return lineState{mode: "selector"}
	}

	return lineState{}
}

// syntaxCache stores line-level cached tokens and continuation states
// to allow fast incremental re-tokenization during editing.
type syntaxCache struct {
	lang       string
	lines      []string
	tokenCache [][]segment
	states     []lineState
	dirty      map[int]bool
}

// newSyntaxCache constructs an empty syntaxCache for a language.
func newSyntaxCache(lang string) *syntaxCache {
	return &syntaxCache{
		lang:  lang,
		dirty: make(map[int]bool),
	}
}

// markDirty records that a source line has been edited.
func (c *syntaxCache) markDirty(line int) {
	if c.dirty == nil {
		c.dirty = make(map[int]bool)
	}

	c.dirty[line] = true
}

// tokenize tokenizes lines incrementally, reusing cached tokens for lines that
// did not change and whose starting state was not altered by earlier edits.
func (c *syntaxCache) tokenize(lines []string) [][]segment {
	n := len(lines)
	if n == 0 {
		c.lines = nil
		c.tokenCache = nil
		c.states = nil
		c.dirty = make(map[int]bool)

		return nil
	}

	// Line count changed: reuse common clean prefix and tokenize suffix.
	if len(c.lines) != n {
		firstDiff := 0
		for firstDiff < len(c.lines) && firstDiff < n && c.lines[firstDiff] == lines[firstDiff] {
			firstDiff++
		}

		newLines := make([]string, n)
		copy(newLines, lines)
		newTokenCache := make([][]segment, n)
		newStates := make([]lineState, n)

		if firstDiff > 0 {
			copy(newTokenCache[:firstDiff], c.tokenCache[:firstDiff])
			copy(newStates[:firstDiff], c.states[:firstDiff])
		}

		st := initialLineState(c.lang)
		if firstDiff > 0 {
			prev := firstDiff - 1
			// firstDiff is the shared clean prefix length, so prev is a
			// valid index into both slices; the guard makes it explicit.
			if prev < len(lines) && prev < len(c.states) {
				st = c.states[prev]
				_, st = scanLineWithState(c.lang, st, prev, lines[prev])
			}
		}

		for i := firstDiff; i < n; i++ {
			newStates[i] = st
			segs, next := scanLineWithState(c.lang, st, i, lines[i])
			newTokenCache[i] = segs
			st = next
		}

		c.lines = newLines
		c.tokenCache = newTokenCache
		c.states = newStates
		c.dirty = make(map[int]bool)

		out := make([][]segment, n)
		copy(out, c.tokenCache)

		return out
	}

	// Same line count: find first dirty or modified line.
	firstDirty := -1
	for i, line := range lines {
		if (c.dirty != nil && c.dirty[i]) || line != c.lines[i] {
			firstDirty = i

			break
		}
	}

	if firstDirty < 0 {
		// Fast path: entire file matches cache.
		out := make([][]segment, n)
		copy(out, c.tokenCache)

		return out
	}

	// Incremental re-tokenization starting at firstDirty.
	st := c.states[firstDirty]
	i := firstDirty
	for i < n {
		isDirty := (c.dirty != nil && c.dirty[i]) || lines[i] != c.lines[i] || c.states[i] != st
		if !isDirty {
			// Clean line whose starting state matches cached state.
			break
		}

		c.states[i] = st
		c.lines[i] = lines[i]
		segs, next := scanLineWithState(c.lang, st, i, lines[i])
		c.tokenCache[i] = segs
		st = next
		i++
	}

	c.dirty = make(map[int]bool)

	out := make([][]segment, n)
	copy(out, c.tokenCache)

	return out
}

// scanLineWithState highlights one line in any supported language given its continuation state.
func scanLineWithState(lang string, state lineState, li int, line string) ([]segment, lineState) {
	switch lang {
	case langGo, langPython, langJS:
		return scanLineC(cSpecFor(lang), state, line)
	case langMarkdown:
		return scanLineMarkdown(li, state, line)
	case langCSS:
		return scanLineCSS(state, line)
	case langHTML:
		return scanLineHTML(state, line)
	case langJSON:
		return scanLineJSON(line), lineState{}
	case langYAML:
		return scanLineYAML(line), lineState{}
	case langShell:
		return scanLineShell(line), lineState{}
	default:
		return finishLine([]segment{{Text: line, Class: classPlain}}), lineState{}
	}
}

// tokenize splits every line into highlighted segments. Each call tokenizes
// from scratch with a fresh cache, so concurrent callers share no state.
func tokenize(lang string, lines []string) [][]segment {
	return newSyntaxCache(lang).tokenize(lines)
}

// highlightLine tokenizes one line on its own. Diff rows use it because a
// unified diff interleaves lines from different sides of the file.
func highlightLine(lang, text string) []segment {
	if text == "" {
		return []segment{{Text: " ", Class: classPlain}}
	}

	segs, _ := scanLineWithState(lang, initialLineState(lang), 0, text)

	return finishLine(segs)
}

// addSeg appends text with a class, merging with the previous run when the
// class matches. A trailing space moves to the next run: the shaper trims a
// space at the end of a run, so "Service string" would print as
// "Servicestring" when the type starts a new run.
func addSeg(dst []segment, text, class string) []segment {
	if text == "" {
		return dst
	}

	if n := len(dst); n > 0 && dst[n-1].Class == class {
		dst[n-1].Text += text

		return dst
	}

	if n := len(dst); n > 0 {
		text = shiftTrailing(dst, n-1, text)
	}

	return append(dst, segment{Text: text, Class: class})
}

// shiftTrailing moves the spaces at the end of segment i to the front of the
// next text. It returns the next text with the spaces prepended.
func shiftTrailing(segs []segment, i int, next string) string {
	text := segs[i].Text
	trimmed := strings.TrimRight(text, " ")
	if trimmed == text {
		return next
	}

	segs[i].Text = trimmed

	return text[len(trimmed):] + next
}

// shiftAll applies the same rule between every pair of runs.
func shiftAll(segs []segment) []segment {
	for i := 0; i+1 < len(segs); i++ {
		segs[i+1].Text = shiftTrailing(segs, i, segs[i+1].Text)
	}

	return segs
}

// finishLine keeps an empty line visible with one space.
func finishLine(segs []segment) []segment {
	if len(segs) == 0 {
		return []segment{{Text: " ", Class: classPlain}}
	}

	return segs
}

// keywordSet builds a lookup from a list, so no package-level map is needed.
func keywordSet(words ...string) map[string]bool {
	set := make(map[string]bool, len(words))

	for _, word := range words {
		set[word] = true
	}

	return set
}

// cSpec is the shared lexical profile for Go, Python, and JavaScript.
type cSpec struct {
	lang         string
	keywords     map[string]bool
	types        map[string]bool
	constants    map[string]bool
	lineComment  string
	blockComment bool
	quotes       []string
	backtick     bool
	triple       bool
	hashComment  bool
	decorators   bool
}

// cSpecFor returns the keyword sets and comment rules for one C-family
// language.
func cSpecFor(lang string) cSpec {
	spec := cSpec{lang: lang}

	switch lang {
	case langGo:
		spec.keywords = keywordSet(
			"break", "case", "chan", "const", "continue", "default", "defer",
			"else", "fallthrough", "for", "func", "go", "goto", "if", "import",
			"interface", "map", "package", "range", "return", "select", "struct",
			"switch", "type", "var",
		)
		spec.types = keywordSet(
			"any", "bool", "byte", "complex64", "complex128", "error", "float32",
			"float64", "int", "int8", "int16", "int32", "int64", "rune", "string",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
		)
		spec.constants = keywordSet("true", "false", "nil", "iota")
		spec.lineComment = "//"
		spec.blockComment = true
		spec.quotes = []string{"\"", "'"}
		spec.backtick = true
	case langPython:
		spec.keywords = keywordSet(
			"and", "as", "assert", "async", "await", "break", "class", "continue",
			"def", "del", "elif", "else", "except", "finally", "for", "from",
			"global", "if", "import", "in", "is", "lambda", "nonlocal", "not",
			"or", "pass", "raise", "return", "try", "while", "with", "yield",
		)
		spec.types = keywordSet("bool", "bytes", "dict", "float", "frozenset", "int", "list", "object", "set", "str", "tuple")
		spec.constants = keywordSet("None", "True", "False")
		spec.hashComment = true
		spec.quotes = []string{"\"", "'"}
		spec.triple = true
		spec.decorators = true
	case langJS:
		spec.keywords = keywordSet(
			"async", "await", "break", "case", "catch", "class", "const",
			"continue", "debugger", "default", "delete", "do", "else", "export",
			"extends", "finally", "for", "from", "function", "get", "if",
			"import", "in", "instanceof", "let", "new", "of", "return", "set",
			"static", "switch", "throw", "try", "typeof", "var", "void", "while",
			"with", "yield",
		)
		spec.types = keywordSet("Array", "Boolean", "Date", "Error", "Map", "Number", "Object", "Promise", "RegExp", "Set", "String")
		spec.constants = keywordSet("true", "false", "null", "undefined", "NaN", "Infinity")
		spec.lineComment = "//"
		spec.blockComment = true
		spec.quotes = []string{"\"", "'"}
		spec.backtick = true
	}

	return spec
}

// cState carries the tokens that a line break does not end.
type cState struct {
	spec      cSpec
	inBlock   bool
	inString  string
	pending   string
	funcAwait bool
	paren     int
}

// scanLineC tokenizes one line using C-family lexical rules (Go, Python, JS).
func scanLineC(spec cSpec, state lineState, line string) ([]segment, lineState) {
	st := cState{
		spec:      spec,
		inBlock:   state.inBlock,
		inString:  state.inString,
		pending:   state.pending,
		funcAwait: state.funcAwait,
		paren:     state.paren,
	}
	segs := finishLine(st.scanLine([]rune(line)))
	next := lineState{
		inBlock:   st.inBlock,
		inString:  st.inString,
		pending:   st.pending,
		funcAwait: st.funcAwait,
		paren:     st.paren,
	}

	return segs, next
}

// scanLine tokenizes one line and keeps block comments, multiline strings,
// and the name-after-keyword state across calls.
func (st *cState) scanLine(r []rune) []segment {
	var segs []segment

	i := 0
	for i < len(r) {
		if st.inString != "" {
			end := indexRunes(r, i, st.inString, st.spec.lang != langGo)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classStr)

				return segs
			}

			segs = addSeg(segs, string(r[i:end+len([]rune(st.inString))]), classStr)
			i = end + len([]rune(st.inString))
			st.inString = ""

			continue
		}

		if st.inBlock {
			end := indexRunes(r, i, "*/", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classCom)

				return segs
			}

			segs = addSeg(segs, string(r[i:end+len("*/")]), classCom)
			i = end + len("*/")
			st.inBlock = false

			continue
		}

		if delim := st.openString(r, i); delim != "" {
			segs, i = st.scanString(segs, r, i, delim)

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
		case st.spec.lineComment != "" && hasRunes(r, i, st.spec.lineComment):
			segs = addSeg(segs, string(r[i:]), classCom)

			return segs
		case st.spec.hashComment && c == '#':
			segs = addSeg(segs, string(r[i:]), classCom)

			return segs
		case st.spec.blockComment && hasRunes(r, i, "/*"):
			end := indexRunes(r, i+len("/*"), "*/", false)
			if end < 0 {
				segs = addSeg(segs, string(r[i:]), classCom)
				st.inBlock = true

				return segs
			}

			segs = addSeg(segs, string(r[i:end+len("*/")]), classCom)
			i = end + len("*/")
		case st.spec.decorators && c == '@':
			j := i + 1
			for j < len(r) && (isIdentRune(r[j]) || r[j] == '.') {
				j++
			}

			segs = addSeg(segs, string(r[i:j]), classDec)
			i = j
		case unicode.IsDigit(c) || (c == '.' && i+1 < len(r) && unicode.IsDigit(r[i+1])):
			j := scanNumber(r, i)
			segs = addSeg(segs, string(r[i:j]), classNum)
			i = j
		case isIdentStart(c):
			j := i + 1
			for j < len(r) && isIdentRune(r[j]) {
				j++
			}

			word := string(r[i:j])
			segs = addSeg(segs, word, st.classify(word))
			i = j
		case strings.ContainsRune("+-*/%=<>!&|^~?:", c):
			segs = addSeg(segs, string(c), classOp)
			i++
		case strings.ContainsRune("()[]{}.,;", c):
			if st.funcAwait {
				switch c {
				case '(':
					st.paren++
				case ')':
					if st.paren > 0 {
						st.paren--
					}
				}
			}

			if strings.ContainsRune("(={[,;", c) && st.pending != "" {
				st.pending = ""
			}

			segs = addSeg(segs, string(c), classPunc)
			i++
		default:
			segs = addSeg(segs, string(c), classPlain)
			i++
		}
	}

	return segs
}

// classify returns the class for one identifier and advances the
// name-after-keyword state.
func (st *cState) classify(word string) string {
	if st.pending != "" {
		class := st.pending
		st.pending = ""

		return class
	}

	if st.spec.keywords[word] {
		switch {
		case st.spec.lang == langGo && word == "func":
			st.funcAwait = true
			st.paren = 0
		case st.spec.lang == langGo && word == "type":
			st.pending = classType
		case st.spec.lang == langPython && word == "def":
			st.pending = classFn
		case st.spec.lang == langPython && word == "class":
			st.pending = classType
		case st.spec.lang == langJS && word == "function":
			st.pending = classFn
		case st.spec.lang == langJS && word == "class":
			st.pending = classType
		case st.spec.lang == langJS && (word == "const" || word == "let" || word == "var"):
			st.pending = classVar
		}

		return classKw
	}

	if st.funcAwait && st.paren == 0 {
		st.funcAwait = false

		return classFn
	}

	if st.spec.types[word] {
		return classType
	}

	if st.spec.constants[word] {
		return classLit
	}

	return classPlain
}

// scanString consumes one string that starts at i with delim. A Go raw string
// and a Python triple-quoted string can span lines; a single-quoted string
// ends at the line end.
func (st *cState) scanString(segs []segment, r []rune, i int, delim string) ([]segment, int) {
	multiline := delim == "`" || len(delim) > 1
	escapes := delim != "`"

	end := indexRunes(r, i+len([]rune(delim)), delim, escapes)
	if end < 0 {
		segs = addSeg(segs, string(r[i:]), classStr)
		if multiline {
			st.inString = delim
		}

		return segs, len(r)
	}

	segs = addSeg(segs, string(r[i:end+len([]rune(delim))]), classStr)

	return segs, end + len([]rune(delim))
}

// openString reports the opening delimiter at i, or "" when the rune does not
// start a string. Triple quotes win over a single quote.
func (st *cState) openString(r []rune, i int) string {
	if st.spec.backtick && r[i] == '`' {
		return "`"
	}

	if st.spec.triple {
		for _, q := range []string{"\"\"\"", "'''"} {
			if hasRunes(r, i, q) {
				return q
			}
		}
	}

	for _, q := range st.spec.quotes {
		if hasRunes(r, i, q) {
			return q
		}
	}

	return ""
}

// scanNumber returns the end of one number literal.
func scanNumber(r []rune, i int) int {
	j := i

	if r[j] == '0' && j+1 < len(r) {
		next := unicode.ToLower(r[j+1])
		if next == 'x' || next == 'b' || next == 'o' {
			j += 2
			for j < len(r) && (unicode.IsDigit(r[j]) || unicode.IsLetter(r[j]) || r[j] == '_') {
				j++
			}

			return j
		}
	}

	for j < len(r) && (unicode.IsDigit(r[j]) || r[j] == '_') {
		j++
	}

	if j < len(r) && r[j] == '.' {
		j++

		for j < len(r) && (unicode.IsDigit(r[j]) || r[j] == '_') {
			j++
		}
	}

	if j < len(r) && (r[j] == 'e' || r[j] == 'E') {
		k := j + 1
		if k < len(r) && (r[k] == '+' || r[k] == '-') {
			k++
		}

		if k < len(r) && unicode.IsDigit(r[k]) {
			j = k
			for j < len(r) && unicode.IsDigit(r[j]) {
				j++
			}
		}
	}

	for j < len(r) && (r[j] == 'i' || r[j] == 'j' || r[j] == 'n') {
		j++
	}

	return j
}

// isIdentStart reports whether r can begin an identifier.
func isIdentStart(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r)
}

// isIdentRune reports whether r can continue an identifier.
func isIdentRune(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r)
}

// hasRunes reports whether r[i:] starts with prefix.
func hasRunes(r []rune, i int, prefix string) bool {
	p := []rune(prefix)
	if i+len(p) > len(r) {
		return false
	}

	for k, c := range p {
		if r[i+k] != c {
			return false
		}
	}

	return true
}

// indexRunes returns the index of needle at or after start, or -1. When
// escapes is true, a backslash skips the next rune.
func indexRunes(r []rune, start int, needle string, escapes bool) int {
	n := []rune(needle)

	for i := start; i+len(n) <= len(r); i++ {
		if escapes && r[i] == '\\' {
			i++

			continue
		}

		match := true

		for k, c := range n {
			if r[i+k] != c {
				match = false

				break
			}
		}

		if match {
			return i
		}
	}

	return -1
}
