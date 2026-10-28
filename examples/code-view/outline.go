package main

import (
	"strings"
)

const (
	maxOutlineRows  = 60
	maxHeadingLevel = 6
	maxSymbolRunes  = 64
	pythonIndent    = 4
)

// outline returns the symbols the details panel shows. It is a lightweight
// scan, not a parser: function, type, class, heading, and selector lines are
// enough to navigate a file.
func outline(lang string, lines []string) []symbolRow {
	rows := make([]symbolRow, 0, maxOutlineRows)

	switch lang {
	case langGo:
		rows = outlineGo(lines, rows)
	case langPython:
		rows = outlinePython(lines, rows)
	case langJS:
		rows = outlineJS(lines, rows)
	case langMarkdown:
		rows = outlineMarkdown(lines, rows)
	case langCSS:
		rows = outlineCSS(lines, rows)
	}

	if len(rows) > maxOutlineRows {
		rows = rows[:maxOutlineRows]
	}

	return rows
}

// addSymbol appends one symbol unless the row cap is reached.
func addSymbol(rows []symbolRow, name, kind string, line, depth int) []symbolRow {
	name = strings.TrimSpace(name)
	if name == "" || len(rows) >= maxOutlineRows {
		return rows
	}

	if runes := []rune(name); len(runes) > maxSymbolRunes {
		name = string(runes[:63]) + "…"
	}

	return append(rows, symbolRow{Name: name, Kind: kind, Line: line, No: line + 1, Depth: depth})
}

// outlineGo collects func, type, const, and var declarations.
func outlineGo(lines []string, rows []symbolRow) []symbolRow {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "func "):
			rest := strings.TrimPrefix(trimmed, "func ")
			if strings.HasPrefix(rest, "(") {
				if end := strings.Index(rest, ") "); end >= 0 {
					rest = rest[end+2:]
				}
			}

			rows = addSymbol(rows, cutName(rest), "fn", i, 0)
		case strings.HasPrefix(trimmed, "type "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "type ")), "type", i, 0)
		case strings.HasPrefix(trimmed, "const "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "const ")), "const", i, 0)
		case strings.HasPrefix(trimmed, "var "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "var ")), "var", i, 0)
		}
	}

	return rows
}

// outlinePython collects def and class lines with their indentation.
func outlinePython(lines []string, rows []symbolRow) []symbolRow {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		depth := (len(line) - len(strings.TrimLeft(line, " "))) / pythonIndent

		switch {
		case strings.HasPrefix(trimmed, "def "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "def ")), "fn", i, depth)
		case strings.HasPrefix(trimmed, "async def "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "async def ")), "fn", i, depth)
		case strings.HasPrefix(trimmed, "class "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "class ")), "class", i, depth)
		}
	}

	return rows
}

// outlineJS collects function, class, and arrow-assignment lines.
func outlineJS(lines []string, rows []symbolRow) []symbolRow {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(trimmed, "export default function "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "export default function ")), "fn", i, 0)
		case strings.HasPrefix(trimmed, "export function "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "export function ")), "fn", i, 0)
		case strings.HasPrefix(trimmed, "function "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "function ")), "fn", i, 0)
		case strings.HasPrefix(trimmed, "export class "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "export class ")), "class", i, 0)
		case strings.HasPrefix(trimmed, "class "):
			rows = addSymbol(rows, cutName(strings.TrimPrefix(trimmed, "class ")), "class", i, 0)
		case strings.HasPrefix(trimmed, "const ") || strings.HasPrefix(trimmed, "let ") || strings.HasPrefix(trimmed, "var "):
			if strings.Contains(trimmed, "= (") || strings.Contains(trimmed, "= async") || strings.Contains(trimmed, "= function") {
				rows = addSymbol(rows, cutName(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(trimmed, "const "), "let "), "var ")), "fn", i, 0)
			}
		}
	}

	return rows
}

// outlineMarkdown collects ATX headings.
func outlineMarkdown(lines []string, rows []symbolRow) []symbolRow {
	inFence := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence

			continue
		}

		if inFence {
			continue
		}

		r := []rune(strings.TrimLeft(line, " "))
		level, marker := headingLevel(r)
		if level == 0 || marker > len(r) {
			continue
		}

		rows = addSymbol(rows, string(r[marker:]), headingKind(level), i, level-1)
	}

	return rows
}

// headingKind maps a level to the outline label.
func headingKind(level int) string {
	kinds := [...]string{"h1", "h2", "h3", "h4", "h5", "h6"}
	if level < 1 || level > len(kinds) {
		return "h1"
	}

	return kinds[level-1]
}

// outlineCSS collects selectors and at-rules.
func outlineCSS(lines []string, rows []symbolRow) []symbolRow {
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasSuffix(trimmed, "{") {
			continue
		}

		name := strings.TrimSpace(strings.TrimSuffix(trimmed, "{"))
		if name == "" || strings.HasPrefix(name, "/*") {
			continue
		}

		kind := "rule"
		if strings.HasPrefix(name, "@") {
			kind = "at"
		}

		rows = addSymbol(rows, name, kind, i, 0)
	}

	return rows
}

// cutName keeps the leading identifier of a declaration.
func cutName(rest string) string {
	rest = strings.TrimSpace(rest)

	for i, r := range rest {
		switch r {
		case '(', ' ', '[', '{', '=', ':', '<':
			return strings.TrimSpace(rest[:i])
		}
	}

	return rest
}
