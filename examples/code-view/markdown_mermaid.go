package main

import "strings"

const (
	previewDiagramPad  = 2
	previewBoxOverhead = 4
	diagramMinCols     = 2
	diagramBoxBorder   = 2
	diagramHalf        = 2
	diagramArrowBars   = 5
	diagramArrowPlain  = 6
	seqSlotMin         = 8
	seqSlotMax         = 18
)

// diagramRow marks one rendered diagram line. The marker and the row kind let
// the preview stylesheet tell a diagram from source code.
func diagramRow(text string, cols int) previewRow {
	return previewRow{
		Kind:     "diagram",
		Marker:   "◇",
		Segments: []segment{{Text: lineOrSpace(clipRunes(text, cols)), Class: classPlain}},
	}
}

// clipRunes caps one rendered line at the preview width.
func clipRunes(text string, cols int) string {
	r := []rune(text)
	if len(r) <= cols {
		return text
	}

	if cols < diagramMinCols {
		return string(r[:cols])
	}

	return string(r[:cols-1]) + "…"
}

// isMermaidInfo reports whether a fence info string asks for a diagram.
func isMermaidInfo(info string) bool {
	fields := strings.Fields(info)

	return len(fields) > 0 && strings.EqualFold(fields[0], "mermaid")
}

// mermaidKind returns the supported diagram type of a fence.
func mermaidKind(src []string) string {
	for _, line := range src {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		fields := strings.Fields(trimmed)
		head := strings.ToLower(fields[0])

		switch head {
		case "flowchart", "graph":
			return "flowchart"
		case "sequencediagram":
			return "sequence"
		default:
			return ""
		}
	}

	return ""
}

// mermaidTypeName returns the first statement's word for the note row.
func mermaidTypeName(src []string) string {
	for _, line := range src {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		fields := strings.Fields(trimmed)

		return strings.TrimSuffix(fields[0], ";")
	}

	return ""
}

// appendMermaid renders a supported diagram as text rows, or a note and code
// rows for an unsupported one. It returns one row offset per source line.
func appendMermaid(rows []previewRow, src []string, cols int) ([]previewRow, []int) {
	switch mermaidKind(src) {
	case "flowchart":
		return appendFlowchart(rows, src, cols)
	case "sequence":
		return appendSequence(rows, src, cols)
	}

	base := len(rows)
	name := mermaidTypeName(src)
	note := "Mermaid diagram type is not rendered"
	if name != "" {
		note = "Mermaid diagram type " + name + " is not rendered"
	}

	rows = append(rows, diagramRow(note, cols))

	lineRows := make([]int, len(src))

	for i, line := range src {
		rows = append(rows, previewRow{Kind: "code", Segments: []segment{{Text: lineOrSpace(line), Class: classCode}}})
		lineRows[i] = len(rows) - 1 - base
	}

	return rows, lineRows
}

// flowEdge is one parsed flowchart connector.
type flowEdge struct {
	from       string
	to         string
	fromLabel  string
	toLabel    string
	label      string
	style      string
	undirected bool
}

// flowStatement keeps a source line in rendering order.
type flowStatement struct {
	line    int
	node    string
	edge    int
	text    string
	hasEdge bool
}

// appendFlowchart renders a text flowchart for TD, TB, or LR graphs.
func appendFlowchart(rows []previewRow, src []string, cols int) ([]previewRow, []int) {
	base := len(rows)
	lineRows := make([]int, len(src))

	nodes := map[string]string{}
	order := make([]string, 0, len(src))
	edges := make([]flowEdge, 0, len(src))
	statements := make([]flowStatement, 0, len(src))

	ensure := func(id, label string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}

		if _, ok := nodes[id]; !ok {
			order = append(order, id)
			nodes[id] = id
		}

		if label != "" {
			nodes[id] = label
		}
	}

	direction := "TD"

	for i, raw := range src {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		fields := strings.Fields(trimmed)
		head := strings.ToLower(fields[0])

		switch head {
		case "flowchart", "graph":
			if len(fields) > 1 {
				direction = strings.ToUpper(fields[1])
			}

			continue
		case "subgraph", "end", "direction", "classdef", "class", "style", "linkstyle", "click":
			continue
		}

		if edge, ok := parseFlowEdge(trimmed); ok {
			ensure(edge.from, edge.fromLabel)
			ensure(edge.to, edge.toLabel)
			edges = append(edges, edge)
			statements = append(statements, flowStatement{line: i, edge: len(edges) - 1, hasEdge: true})

			continue
		}

		if id, label, ok := parseFlowNode(trimmed); ok {
			ensure(id, label)
			statements = append(statements, flowStatement{line: i, node: id})

			continue
		}

		statements = append(statements, flowStatement{line: i, text: trimmed})
	}

	rendered := map[string]bool{}
	horizontal := direction == "LR"

	for _, statement := range statements {
		lineRows[statement.line] = len(rows) - base

		if statement.text != "" {
			rows = append(rows, diagramRow(statement.text, cols))

			continue
		}

		if statement.hasEdge {
			edge := edges[statement.edge]
			if horizontal {
				rows = append(rows, flowHorizontalRows(nodes[edge.from], nodes[edge.to], edge, cols)...)

				continue
			}

			if !rendered[edge.from] {
				rows = append(rows, flowNodeRows(nodes[edge.from], cols)...)
				rendered[edge.from] = true
			}

			rows = append(rows, flowConnectorRows(nodes[edge.from], edge, cols)...)

			if !rendered[edge.to] {
				rows = append(rows, flowNodeRows(nodes[edge.to], cols)...)
				rendered[edge.to] = true
			}

			continue
		}

		if statement.node == "" || rendered[statement.node] {
			continue
		}

		rows = append(rows, flowNodeRows(nodes[statement.node], cols)...)

		rendered[statement.node] = true
	}

	for _, id := range order {
		if rendered[id] {
			continue
		}

		rows = append(rows, flowNodeRows(nodes[id], cols)...)

		rendered[id] = true
	}

	if len(rows) == base {
		rows = append(rows, diagramRow("empty flowchart", cols))
	}

	clampLineRows(lineRows, len(rows)-base)

	return rows, lineRows
}

// clampLineRows keeps every source anchor inside the rendered diagram.
func clampLineRows(lineRows []int, total int) {
	for i, rel := range lineRows {
		if rel < 0 || rel >= total {
			lineRows[i] = 0
		}
	}
}

// parseFlowEdge parses one flowchart edge.
func parseFlowEdge(line string) (flowEdge, bool) {
	operators := [...]string{"-.->", "==>", "-->", "---"}

	for _, operator := range operators {
		index := strings.Index(line, operator)
		if index <= 0 {
			continue
		}

		left := strings.TrimSpace(strings.TrimSuffix(line[:index], ";"))
		right := strings.TrimSpace(line[index+len(operator):])
		if right == "" {
			continue
		}

		label := ""
		if strings.HasPrefix(right, "|") {
			end := strings.Index(right[1:], "|")
			if end >= 0 {
				label = strings.TrimSpace(right[1 : 1+end])
				right = strings.TrimSpace(right[1+end+1:])
			}
		}

		from, fromLabel, ok := parseFlowNode(left)
		to, toLabel, okTo := parseFlowNode(right)
		if !ok || !okTo || from == "" || to == "" {
			continue
		}

		edge := flowEdge{from: from, to: to, fromLabel: fromLabel, toLabel: toLabel, label: label, style: "solid"}

		switch operator {
		case "---":
			edge.undirected = true
		case "-.->":
			edge.style = "dotted"
		case "==>":
			edge.style = "thick"
		}

		return edge, true
	}

	return flowEdge{}, false
}

// parseFlowNode returns the id and label of one flowchart node.
func parseFlowNode(text string) (string, string, bool) {
	text = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(text), ";"))
	if text == "" {
		return "", "", false
	}

	open := strings.IndexAny(text, "[{(")
	id := text
	label := ""

	if open >= 0 {
		id = strings.TrimSpace(text[:open])

		closeChar := byte(']')

		switch text[open] {
		case '{':
			closeChar = '}'
		case '(':
			closeChar = ')'
		}

		end := strings.LastIndexByte(text, closeChar)
		if end > open {
			label = strings.TrimSpace(text[open+1 : end])
			label = strings.Trim(label, `"`)
			label = strings.ReplaceAll(label, "<br/>", " ")
			label = strings.ReplaceAll(label, "<br>", " ")
		}
	}

	id = strings.Trim(id, `"`)
	if id == "" {
		return "", "", false
	}

	return id, label, true
}

// flowBoxLines draws one node as a box.
func flowBoxLines(label string, maxWidth int) ([]string, int) {
	if maxWidth < previewBoxOverhead+1 {
		maxWidth = previewBoxOverhead + 1
	}

	r := []rune(strings.TrimSpace(label))
	limit := maxWidth - previewBoxOverhead

	if len(r) > limit {
		if limit <= 1 {
			r = []rune("…")
		} else {
			r = append(r[:limit-1], '…')
		}
	}

	if len(r) == 0 {
		r = []rune(" ")
	}

	width := len(r) + previewBoxOverhead
	top := "┌" + strings.Repeat("─", len(r)+diagramBoxBorder) + "┐"
	middle := "│ " + string(r) + " │"
	bottom := "└" + strings.Repeat("─", len(r)+diagramBoxBorder) + "┘"

	return []string{top, middle, bottom}, width
}

// flowNodeRows renders one node box with a diagram marker on every line.
func flowNodeRows(label string, cols int) []previewRow {
	lines, _ := flowBoxLines(label, cols-previewDiagramPad)
	indent := strings.Repeat(" ", previewDiagramPad)
	rows := make([]previewRow, 0, len(lines))

	for _, line := range lines {
		rows = append(rows, diagramRow(indent+line, cols))
	}

	return rows
}

// flowConnectorRows draws the vertical connector after a source box.
func flowConnectorRows(fromLabel string, edge flowEdge, cols int) []previewRow {
	_, width := flowBoxLines(fromLabel, cols-previewDiagramPad)
	center := previewDiagramPad + width/diagramHalf
	indent := strings.Repeat(" ", center)
	rows := []previewRow{diagramRow(indent+"│", cols)}

	if edge.undirected {
		return append(rows, diagramRow(indent+"┴", cols))
	}

	arrow := "▼"
	if edge.label != "" {
		arrow += " " + edge.label
	}

	return append(rows, diagramRow(indent+arrow, cols))
}

// flowHorizontalRows draws one LR edge as a row of two boxes.
func flowHorizontalRows(fromLabel, toLabel string, edge flowEdge, cols int) []previewRow {
	arrow := flowHorizontalArrow(edge)
	maxBox := (cols - len([]rune(arrow)) - previewDiagramPad) / diagramHalf
	left, _ := flowBoxLines(fromLabel, maxBox)
	right, _ := flowBoxLines(toLabel, maxBox)
	gap := strings.Repeat(" ", len([]rune(arrow)))

	return []previewRow{
		diagramRow(left[0]+gap+right[0], cols),
		diagramRow(left[1]+arrow+right[1], cols),
		diagramRow(left[2]+gap+right[2], cols),
	}
}

// flowHorizontalArrow returns the connector between two LR boxes.
func flowHorizontalArrow(edge flowEdge) string {
	bar := "─"
	if edge.style == "dotted" {
		bar = "╌"
	}

	if edge.style == "thick" {
		bar = "━"
	}

	if edge.undirected {
		if edge.label != "" {
			return bar + bar + " " + edge.label + " " + bar + bar
		}

		return strings.Repeat(bar, diagramArrowPlain)
	}

	if edge.label != "" {
		return bar + bar + " " + edge.label + " " + bar + "▶"
	}

	return strings.Repeat(bar, diagramArrowBars) + "▶"
}

// seqItem is one message or note in a sequence diagram.
type seqItem struct {
	line    int
	from    string
	to      string
	text    string
	note    string
	message bool
	dashed  bool
}

// appendSequence renders a text sequence diagram.
func appendSequence(rows []previewRow, src []string, cols int) ([]previewRow, []int) {
	base := len(rows)
	lineRows := make([]int, len(src))

	participants := map[string]string{}
	order := make([]string, 0, len(src))
	items := make([]seqItem, 0, len(src))

	ensure := func(id, label string) {
		id = strings.TrimSpace(strings.Trim(id, `"`))
		if id == "" {
			return
		}

		if _, ok := participants[id]; !ok {
			order = append(order, id)
			participants[id] = id
		}

		if label != "" {
			participants[id] = label
		}
	}

	for i, raw := range src {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "%%") {
			continue
		}

		fields := strings.Fields(trimmed)
		head := strings.ToLower(fields[0])

		switch head {
		case "sequencediagram":
			continue
		case "participant", "actor":
			rest := strings.TrimSpace(trimmed[len(fields[0]):])
			id := rest
			label := ""
			if at := strings.Index(strings.ToLower(rest), " as "); at >= 0 {
				id = strings.TrimSpace(rest[:at])
				label = strings.TrimSpace(rest[at+len(" as "):])
			}

			ensure(id, label)

			continue
		case "note":
			items = append(items, seqItem{line: i, note: trimmed})

			continue
		}

		if from, to, text, dashed, ok := parseSeqMessage(trimmed); ok {
			ensure(from, "")
			ensure(to, "")
			items = append(items, seqItem{line: i, from: from, to: to, text: text, message: true, dashed: dashed})

			continue
		}

		if isSeqControl(head) {
			continue
		}

		items = append(items, seqItem{line: i, note: trimmed})
	}

	if len(order) == 0 {
		rows = append(rows, diagramRow("empty sequence diagram", cols))

		return rows, lineRows
	}

	labels := make([]string, len(order))
	slots := make([]int, len(order))

	for i, id := range order {
		labels[i] = participants[id]
		slots[i] = len([]rune(labels[i])) + previewBoxOverhead
		if slots[i] < seqSlotMin {
			slots[i] = seqSlotMin
		}

		if slots[i] > seqSlotMax {
			slots[i] = seqSlotMax
		}
	}

	fitSeqSlots(slots, cols-previewDiagramPad)

	centers := make([]int, len(order))
	offset := 1

	for i, slot := range slots {
		centers[i] = offset + slot/diagramHalf
		offset += slot
	}

	width := offset + 1

	rows = append(rows, diagramRow(seqHeaderRow(labels, slots, centers, width), cols))
	rows = append(rows, diagramRow(seqLifelineRow(centers, width), cols))

	for _, item := range items {
		lineRows[item.line] = len(rows) - base

		if !item.message {
			rows = append(rows, diagramRow("  "+item.note, cols))

			continue
		}

		from := seqParticipantIndex(order, labels, item.from)
		to := seqParticipantIndex(order, labels, item.to)
		if from < 0 || to < 0 {
			continue
		}

		rows = append(rows, diagramRow(seqMessageRow(item, centers, centers[from], width), cols))
		rows = append(rows, diagramRow(seqArrowRow(item, centers, centers[from], centers[to], width), cols))
	}

	clampLineRows(lineRows, len(rows)-base)

	return rows, lineRows
}

// fitSeqSlots shrinks participant columns until they fit the preview width.
func fitSeqSlots(slots []int, cols int) {
	total := 0
	for _, slot := range slots {
		total += slot
	}

	for total > cols {
		widest := -1

		for i, slot := range slots {
			if slot > 6 && (widest < 0 || slot > slots[widest]) {
				widest = i
			}
		}

		if widest < 0 {
			return
		}

		slots[widest]--
		total--
	}
}

// seqHeaderRow places each participant label over its lifeline.
func seqHeaderRow(labels []string, slots, centers []int, width int) string {
	grid := make([]rune, width)

	for i := range grid {
		grid[i] = ' '
	}

	for i, label := range labels {
		text := []rune(label)
		limit := slots[i] - diagramBoxBorder
		if limit < 1 {
			limit = 1
		}

		if len(text) > limit {
			if limit <= 1 {
				text = []rune("…")
			} else {
				text = append(text[:limit-1], '…')
			}
		}

		start := centers[i] - len(text)/diagramHalf
		for j, r := range text {
			pos := start + j
			if pos >= 0 && pos < len(grid) {
				grid[pos] = r
			}
		}
	}

	return strings.TrimRight(string(grid), " ")
}

// seqLifelineRow draws one vertical lifeline per participant.
func seqLifelineRow(centers []int, width int) string {
	return strings.TrimRight(string(seqLifelineGrid(centers, width)), " ")
}

// seqLifelineGrid returns a row with a lifeline at every participant center.
func seqLifelineGrid(centers []int, width int) []rune {
	grid := make([]rune, width)

	for i := range grid {
		grid[i] = ' '
	}

	for _, center := range centers {
		if center >= 0 && center < len(grid) {
			grid[center] = '│'
		}
	}

	return grid
}

// seqMessageRow draws a message label beside its source lifeline.
func seqMessageRow(item seqItem, centers []int, from, width int) string {
	grid := seqLifelineGrid(centers, width)
	text := []rune(item.text)
	start := from + diagramHalf

	if start+len(text) >= width-1 {
		start = from - diagramHalf - len(text)
	}

	for i, r := range text {
		pos := start + i
		if pos >= 0 && pos < len(grid) {
			grid[pos] = r
		}
	}

	return strings.TrimRight(string(grid), " ")
}

// seqArrowRow draws the arrow between two lifelines.
func seqArrowRow(item seqItem, centers []int, from, to, width int) string {
	grid := seqLifelineGrid(centers, width)

	if from == to {
		return strings.TrimRight(string(grid), " ")
	}

	fill := '─'
	if item.dashed {
		fill = '╌'
	}

	if to > from {
		grid[from] = '├'

		for i := from + 1; i < to; i++ {
			if grid[i] == '│' {
				grid[i] = '┼'
			} else {
				grid[i] = fill
			}
		}

		grid[to] = '▶'
	} else {
		grid[from] = '┤'

		for i := to + 1; i < from; i++ {
			if grid[i] == '│' {
				grid[i] = '┼'
			} else {
				grid[i] = fill
			}
		}

		grid[to] = '◀'
	}

	return strings.TrimRight(string(grid), " ")
}

// parseSeqMessage parses one sequence message.
func parseSeqMessage(line string) (from, to, text string, dashed, ok bool) {
	operators := [...]string{"-->>", "->>", "-->", "->", "--x", "-x"}

	for _, operator := range operators {
		index := strings.Index(line, operator)
		if index <= 0 {
			continue
		}

		from = strings.TrimSpace(line[:index])
		rest := strings.TrimSpace(line[index+len(operator):])
		if from == "" || rest == "" {
			continue
		}

		to = rest

		if colon := strings.Index(rest, ":"); colon >= 0 {
			to = strings.TrimSpace(rest[:colon])
			text = strings.TrimSpace(rest[colon+1:])
		}

		if to == "" {
			continue
		}

		return from, to, text, strings.HasPrefix(operator, "--"), true
	}

	return "", "", "", false, false
}

// seqParticipantIndex resolves an id or a declared label to its column.
func seqParticipantIndex(ids, labels []string, name string) int {
	for i, id := range ids {
		if id == name {
			return i
		}
	}

	for i, label := range labels {
		if label == name {
			return i
		}
	}

	for i, id := range ids {
		if strings.EqualFold(id, name) {
			return i
		}
	}

	return -1
}

// isSeqControl reports whether a sequence line is a control keyword.
func isSeqControl(head string) bool {
	switch head {
	case "loop", "alt", "else", "opt", "par", "and", "end", "activate",
		"deactivate", "autonumber", "rect", "critical", "break", "box",
		"destroy", "create":
		return true
	}

	return false
}
