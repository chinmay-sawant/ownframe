package main

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	delimOpen  = "{{"
	delimClose = "}}"
	delimLen   = 2
)

// ParsedView is the compile-time representation of one HTML template.
type ParsedView struct {
	SourceName string
	TypeName   string
	Root       *StructSchema
	Structs    []*StructSchema
	Nodes      []Node
	StaticSize int
	IDs        []string
	Actions    []string
}

// StructSchema describes one generated Go struct (root view or range item).
type StructSchema struct {
	Name   string
	Fields []*FieldSpec
	byName map[string]*FieldSpec
}

// FieldSpec is one typed field on a generated struct.
type FieldSpec struct {
	Name     string
	GoType   string
	IsSlice  bool
	IsRaw    bool
	Explicit bool
	ItemType *StructSchema
}

// NodeKind identifies a template AST node.
type NodeKind int

const (
	nodeText NodeKind = iota
	nodeExpr
	nodeIf
	nodeRange
)

// Node is one element in the parsed template AST.
type Node struct {
	Kind      NodeKind
	Text      string
	Field     string
	GoType    string
	Raw       bool
	Negate    bool
	Then      []Node
	Else      []Node
	Body      []Node
	ItemType  string
	FieldSpec *FieldSpec
}

// ParseTemplate parses an HTML template string into a typed AST and extracts
// static element IDs and data-action names.
func ParseTemplate(sourceName, typeName, input string) (*ParsedView, error) {
	if !isValidIdent(typeName) {
		return nil, fmt.Errorf("invalid Go type name %q", typeName)
	}

	root := newStructSchema(typeName)
	p := &templateParser{
		input:   input,
		root:    root,
		structs: []*StructSchema{root},
	}

	nodes, term, err := p.parseBlock(root, false)
	if err != nil {
		return nil, err
	}
	if term != "" {
		return nil, fmt.Errorf("unexpected {{%s}} at top level", term)
	}

	ids, actions := extractHTMLMetadata(input)

	return &ParsedView{
		SourceName: sourceName,
		TypeName:   typeName,
		Root:       root,
		Structs:    p.structs,
		Nodes:      nodes,
		StaticSize: staticByteSize(nodes),
		IDs:        ids,
		Actions:    actions,
	}, nil
}

func newStructSchema(name string) *StructSchema {
	return &StructSchema{
		Name:   name,
		byName: make(map[string]*FieldSpec),
	}
}

func (s *StructSchema) ensureField(name, goType string, explicit, isSlice, isRaw bool, itemType *StructSchema) (*FieldSpec, error) {
	if !isValidIdent(name) {
		return nil, fmt.Errorf("invalid field name %q", name)
	}
	if existing, ok := s.byName[name]; ok {
		if existing.IsSlice != isSlice {
			return nil, fmt.Errorf("field %s.%s used as both slice and scalar", s.Name, name)
		}
		if explicit {
			if existing.Explicit && existing.GoType != goType {
				return nil, fmt.Errorf("conflicting types for %s.%s: %s vs %s", s.Name, name, existing.GoType, goType)
			}
			existing.GoType = goType
			existing.Explicit = true
		}
		if isRaw {
			existing.IsRaw = true
		}

		return existing, nil
	}

	spec := &FieldSpec{
		Name:     name,
		GoType:   goType,
		IsSlice:  isSlice,
		IsRaw:    isRaw,
		Explicit: explicit,
		ItemType: itemType,
	}
	s.byName[name] = spec
	s.Fields = append(s.Fields, spec)

	return spec, nil
}

type templateParser struct {
	input   string
	pos     int
	root    *StructSchema
	structs []*StructSchema
}

func (p *templateParser) parseBlock(scope *StructSchema, allowElse bool) ([]Node, string, error) {
	var nodes []Node
	trimNextLeading := false

	for p.pos < len(p.input) {
		openIdx := strings.Index(p.input[p.pos:], delimOpen)
		if openIdx < 0 {
			chunk := p.input[p.pos:]
			if trimNextLeading {
				chunk = strings.TrimLeftFunc(chunk, unicode.IsSpace)
			}
			if chunk != "" {
				nodes = append(nodes, Node{Kind: nodeText, Text: chunk})
			}
			p.pos = len(p.input)

			return nodes, "", nil
		}

		chunk := p.input[p.pos : p.pos+openIdx]
		if trimNextLeading {
			chunk = strings.TrimLeftFunc(chunk, unicode.IsSpace)
			trimNextLeading = false
		}

		actionStart := p.pos + openIdx + delimLen
		closeRel := strings.Index(p.input[actionStart:], delimClose)
		if closeRel < 0 {
			return nil, "", errors.New("unclosed template action '{{'")
		}

		rawAction := p.input[actionStart : actionStart+closeRel]
		p.pos = actionStart + closeRel + delimLen

		trimPrev := strings.HasPrefix(rawAction, "-")
		if trimPrev {
			rawAction = rawAction[1:]
			chunk = strings.TrimRightFunc(chunk, unicode.IsSpace)
		}
		trimNext := strings.HasSuffix(rawAction, "-")
		if trimNext {
			rawAction = rawAction[:len(rawAction)-1]
			trimNextLeading = true
		}

		if chunk != "" {
			nodes = append(nodes, Node{Kind: nodeText, Text: chunk})
		}

		action := strings.TrimSpace(rawAction)
		if action == "" {
			return nil, "", errors.New("empty template action '{{}}'")
		}
		if strings.HasPrefix(action, "/*") && strings.HasSuffix(action, "*/") {
			continue
		}

		switch {
		case action == "end":
			return nodes, "end", nil
		case action == "else":
			if !allowElse {
				return nil, "", errors.New("unexpected {{else}}")
			}

			return nodes, "else", nil
		case strings.HasPrefix(action, "if "):
			node, err := p.parseIf(scope, strings.TrimSpace(action[len("if "):]))
			if err != nil {
				return nil, "", err
			}
			nodes = append(nodes, node)
		case strings.HasPrefix(action, "range "):
			node, err := p.parseRange(scope, strings.TrimSpace(action[len("range "):]))
			if err != nil {
				return nil, "", err
			}
			nodes = append(nodes, node)
		default:
			node, err := p.parseExpr(scope, action)
			if err != nil {
				return nil, "", err
			}
			nodes = append(nodes, node)
		}
	}

	return nodes, "", nil
}

func (p *templateParser) parseIf(scope *StructSchema, cond string) (Node, error) {
	negate := false
	if strings.HasPrefix(cond, "not ") {
		negate = true
		cond = strings.TrimSpace(cond[len("not "):])
	}

	fieldName, explicitType, hasExplicit, err := parseFieldSpec(cond)
	if err != nil {
		return Node{}, fmt.Errorf("invalid if condition %q: %w", cond, err)
	}
	goType := "bool"
	if hasExplicit {
		goType = explicitType
	}
	spec, err := scope.ensureField(fieldName, goType, hasExplicit, false, false, nil)
	if err != nil {
		return Node{}, err
	}

	thenNodes, term, err := p.parseBlock(scope, true)
	if err != nil {
		return Node{}, err
	}
	var elseNodes []Node
	switch term {
	case "end":
		// done
	case "else":
		elseNodes, term, err = p.parseBlock(scope, false)
		if err != nil {
			return Node{}, err
		}
		if term != "end" {
			return Node{}, fmt.Errorf("unclosed {{if %s}}: missing {{end}}", cond)
		}
	default:
		return Node{}, fmt.Errorf("unclosed {{if %s}}: missing {{end}}", cond)
	}

	return Node{
		Kind:      nodeIf,
		Field:     fieldName,
		Negate:    negate,
		Then:      thenNodes,
		Else:      elseNodes,
		FieldSpec: spec,
	}, nil
}

func (p *templateParser) parseRange(scope *StructSchema, target string) (Node, error) {
	fieldName, explicitItemName, hasExplicit, err := parseFieldSpec(target)
	if err != nil {
		return Node{}, fmt.Errorf("invalid range target %q: %w", target, err)
	}

	itemStructName := scope.Name + singularize(fieldName)
	if hasExplicit {
		itemStructName = explicitItemName
	}

	var itemSchema *StructSchema
	if existing, ok := scope.byName[fieldName]; ok && existing.IsSlice && existing.ItemType != nil {
		itemSchema = existing.ItemType
	} else {
		itemSchema = newStructSchema(itemStructName)
		p.structs = append(p.structs, itemSchema)
	}

	spec, err := scope.ensureField(fieldName, "[]"+itemSchema.Name, true, true, false, itemSchema)
	if err != nil {
		return Node{}, err
	}

	bodyNodes, term, err := p.parseBlock(itemSchema, false)
	if err != nil {
		return Node{}, err
	}
	if term != "end" {
		return Node{}, fmt.Errorf("unclosed {{range %s}}: missing {{end}}", target)
	}

	return Node{
		Kind:      nodeRange,
		Field:     fieldName,
		Body:      bodyNodes,
		ItemType:  itemSchema.Name,
		FieldSpec: spec,
	}, nil
}

func (p *templateParser) parseExpr(scope *StructSchema, action string) (Node, error) {
	isRaw := false
	expr := action
	if strings.HasPrefix(expr, "raw ") {
		isRaw = true
		expr = strings.TrimSpace(expr[len("raw "):])
	}

	fieldName, explicitType, hasExplicit, err := parseFieldSpec(expr)
	if err != nil {
		return Node{}, fmt.Errorf("invalid expression %q: %w", action, err)
	}
	goType := "string"
	if hasExplicit {
		goType = explicitType
	}
	if isRaw && hasExplicit && goType != "string" {
		return Node{}, fmt.Errorf("raw expression %q must have string type, got %s", action, goType)
	}

	spec, err := scope.ensureField(fieldName, goType, hasExplicit, false, isRaw, nil)
	if err != nil {
		return Node{}, err
	}

	return Node{
		Kind:      nodeExpr,
		Field:     fieldName,
		GoType:    spec.GoType,
		Raw:       isRaw,
		FieldSpec: spec,
	}, nil
}

func parseFieldSpec(raw string) (fieldName, explicitType string, hasExplicit bool, err error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, ".") {
		return "", "", false, fmt.Errorf("expected field starting with '.', got %q", raw)
	}
	rest := raw[1:]
	if idx := strings.IndexByte(rest, ':'); idx >= 0 {
		fieldName = strings.TrimSpace(rest[:idx])
		explicitType = strings.TrimSpace(rest[idx+1:])
		if explicitType == "" {
			return "", "", false, fmt.Errorf("empty type annotation in %q", raw)
		}
		hasExplicit = true
	} else {
		fieldName = strings.TrimSpace(rest)
	}

	if !isValidIdent(fieldName) {
		return "", "", false, fmt.Errorf("invalid identifier %q", fieldName)
	}

	return fieldName, explicitType, hasExplicit, nil
}

func staticByteSize(nodes []Node) int {
	total := 0
	for _, n := range nodes {
		switch n.Kind {
		case nodeText:
			total += len(n.Text)
		case nodeIf:
			total += staticByteSize(n.Then) + staticByteSize(n.Else)
		case nodeRange:
			total += staticByteSize(n.Body)
		case nodeExpr:
			// Dynamic content; skip static estimate.
		}
	}

	return total
}

// extractHTMLMetadata scans static HTML start tags for id="..." and
// data-action="..." attributes. Dynamic values containing "{{" are skipped.
func extractHTMLMetadata(input string) (ids, actions []string) {
	seenID := make(map[string]bool)
	seenAction := make(map[string]bool)

	i := 0
	n := len(input)
	for i < n {
		if input[i] != '<' {
			i++

			continue
		}
		// Skip HTML comments.
		if strings.HasPrefix(input[i:], "<!--") {
			const commentOpenLen = len("<!--")
			const commentCloseLen = len("-->")
			end := strings.Index(input[i+commentOpenLen:], "-->")
			if end < 0 {
				break
			}
			i += commentOpenLen + end + commentCloseLen

			continue
		}
		// Skip closing tags, doctype, and processing instructions.
		if i+1 < n && (input[i+1] == '/' || input[i+1] == '!' || input[i+1] == '?') {
			end := strings.IndexByte(input[i:], '>')
			if end < 0 {
				break
			}
			i += end + 1

			continue
		}

		// Read tag name.
		tagStart := i + 1
		j := tagStart
		for j < n && isTagNameChar(input[j]) {
			j++
		}
		tagName := strings.ToLower(input[tagStart:j])
		if tagName == "" {
			i++

			continue
		}

		// Scan attributes until '>'.
		for j < n && input[j] != '>' {
			// Skip whitespace and self-closing '/'.
			for j < n && (isSpace(input[j]) || input[j] == '/') {
				j++
			}
			if j >= n || input[j] == '>' {
				break
			}
			// Skip template action inside tag attribute list (e.g. {{if .Checked}}checked{{end}}).
			if strings.HasPrefix(input[j:], delimOpen) {
				end := strings.Index(input[j+delimLen:], delimClose)
				if end < 0 {
					j = n

					break
				}
				j += delimLen + end + delimLen

				continue
			}

			attrStart := j
			for j < n && !isSpace(input[j]) && input[j] != '=' && input[j] != '>' && input[j] != '/' {
				j++
			}
			attrName := strings.ToLower(input[attrStart:j])

			for j < n && isSpace(input[j]) {
				j++
			}
			if j >= n || input[j] != '=' {
				continue
			}
			j++ // skip '='
			for j < n && isSpace(input[j]) {
				j++
			}
			if j >= n {
				break
			}

			var attrVal string
			if input[j] == '"' || input[j] == '\'' {
				quote := input[j]
				j++
				valStart := j
				for j < n && input[j] != quote {
					j++
				}
				attrVal = input[valStart:j]
				if j < n && input[j] == quote {
					j++
				}
			} else {
				valStart := j
				for j < n && !isSpace(input[j]) && input[j] != '>' {
					j++
				}
				attrVal = input[valStart:j]
			}

			if attrVal == "" || strings.Contains(attrVal, delimOpen) {
				continue
			}
			switch attrName {
			case "id":
				if !seenID[attrVal] {
					seenID[attrVal] = true
					ids = append(ids, attrVal)
				}
			case "data-action":
				if !seenAction[attrVal] {
					seenAction[attrVal] = true
					actions = append(actions, attrVal)
				}
			}
		}

		if j < n && input[j] == '>' {
			j++
		}

		// Skip raw text content of <style> and <script>.
		if tagName == "style" || tagName == "script" {
			closeTag := "</" + tagName
			end := indexCaseInsensitive(input[j:], closeTag)
			if end < 0 {
				break
			}
			i = j + end

			continue
		}

		i = j
	}

	return ids, actions
}

func indexCaseInsensitive(s, needle string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(needle))
}

func isTagNameChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '-' || b == ':'
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

func isValidIdent(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	return true
}

func toPascalCase(s string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			upperNext = true

			continue
		}
		if b.Len() == 0 && unicode.IsDigit(r) {
			b.WriteByte('N')
		}
		if upperNext {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

func singularize(name string) string {
	if len(name) > 1 && strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") {
		return name[:len(name)-1]
	}

	return name + "Item"
}
