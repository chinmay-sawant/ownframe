package sqlkit

import (
	"sort"
	"strings"
)

// Column is one introspected column.
type Column struct{ Name, Type string }

// Table is one introspected table or view. Schema is empty for SQLite "main".
type Table struct {
	Schema  string
	Name    string
	Columns []Column
}

// Schema is the introspected catalog autocomplete draws from.
type Schema struct{ Tables []Table }

// ItemKind is the badge shown next to a completion.
type ItemKind int

// Completion kinds.
const (
	ItemKeyword  ItemKind = iota // KW
	ItemFunction                 // FN
	ItemTable                    // TBL
	ItemColumn                   // COL
	ItemSchema                   // SCH
	ItemType                     // TYPE
)

// Badge returns the short label for the kind.
func (k ItemKind) Badge() string {
	return [...]string{"KW", "FN", "TBL", "COL", "SCH", "TYPE"}[k]
}

// Item is one completion candidate.
type Item struct {
	Label  string // shown in the list
	Insert string // text that replaces the typed prefix
	Kind   ItemKind
	Detail string // column type, owning table, etc.
	Score  int    // higher ranks first
}

// Completion is the result of Complete.
type Completion struct {
	// Replace is the byte range of src that Insert should overwrite.
	ReplaceStart, ReplaceEnd int
	Prefix                   string
	Items                    []Item
}

const maxItems = 50

// Base scores: a candidate kind that fits the context outranks the others.
const (
	scoreTableCtx        = 100
	scoreSchemaCtx       = 60
	scoreScopeColumn     = 120
	scoreLooseColumn     = 70
	scoreAlias           = 65
	scoreFunctionCol     = 50
	scoreKeywordCol      = 30
	scoreKeywordDefault  = 100
	scoreFunctionDefault = 40
	scoreTableDefault    = 30
	scoreQualified       = 100
	scoreCatalogTable    = 90

	quotePairLen    = 2
	minSubseqPrefix = 2
	matchExact      = 60
	matchPrefix     = 40
	matchWordStart  = 25
	matchContains   = 12
	matchSubseq     = 3
)

// Complete returns ranked completions for the caret at byte offset caret.
func Complete(src string, caret int, schema Schema, d Dialect) Completion {
	caret = max(0, min(caret, len(src)))
	start := caret
	for start > 0 && isIdentPart(src[start-1]) {
		start--
	}
	res := Completion{ReplaceStart: start, ReplaceEnd: caret, Prefix: src[start:caret]}

	// Statement bounds so scope and context ignore other statements.
	stmtStart, stmtEnd := 0, len(src)
	if sp, ok := StatementAt(src, caret, d); ok && caret >= sp.Start && caret <= sp.End {
		stmtStart, stmtEnd = sp.Start, sp.End
	} else {
		stmtStart = lastSemicolonBefore(src, caret, d)
	}
	toks := significant(src, stmtStart, stmtEnd, d)
	scope := tablesInScope(src, toks, schema)

	// Qualified: "<qualifier>." before the prefix.
	if start > 0 && src[start-1] == '.' {
		q := qualifierBefore(src, start-1)
		res.Items = rank(qualified(q, scope, schema), res.Prefix)
		return res
	}

	ctx := contextAt(src, toks, caret)
	var cands []Item
	switch ctx {
	case ctxTable:
		cands = append(cands, tableItems(schema, scoreTableCtx)...)
		cands = append(cands, schemaItems(schema, scoreSchemaCtx)...)
	case ctxColumn:
		cands = append(cands, columnItems(scope, true, scoreScopeColumn)...)
		// Columns of unreferenced tables are weaker suggestions.
		if len(scope) == 0 {
			cands = append(cands, columnItems(allTables(schema), true, scoreLooseColumn)...)
		}
		cands = append(cands, aliasItems(scope, scoreAlias)...)
		cands = append(cands, functionItems(d, scoreFunctionCol)...)
		cands = append(cands, keywordItems(d, scoreKeywordCol)...)
	default:
		cands = append(cands, keywordItems(d, scoreKeywordDefault)...)
		cands = append(cands, functionItems(d, scoreFunctionDefault)...)
		cands = append(cands, tableItems(schema, scoreTableDefault)...)
	}
	res.Items = rank(cands, res.Prefix)
	return res
}

func lastSemicolonBefore(src string, caret int, d Dialect) int {
	last := 0
	for _, t := range Tokenize(src[:caret], d) {
		if t.Kind == KindOperator && src[t.Start:t.End] == ";" {
			last = t.End
		}
	}
	return last
}

type sigTok struct {
	Token
	up string // uppercased text for words, raw text otherwise
}

func significant(src string, from, to int, d Dialect) []sigTok {
	var out []sigTok
	for _, t := range Tokenize(src[from:to], d) {
		if t.Kind == KindSpace || t.Kind == KindComment {
			continue
		}
		t.Start += from
		t.End += from
		txt := src[t.Start:t.End]
		out = append(out, sigTok{Token: t, up: strings.ToUpper(txt)})
	}
	return out
}

type ctxKind int

const (
	ctxKeyword ctxKind = iota
	ctxTable
	ctxColumn
)

func isTableKeyword(up string) bool {
	switch up {
	case "FROM", "JOIN", "UPDATE", "INTO", "TABLE", "TRUNCATE":
		return true
	}
	return false
}

func isColumnKeyword(up string) bool {
	switch up {
	case "SELECT", "WHERE", "ON", "BY", "HAVING", "SET", "AND", "OR", "NOT", "DISTINCT", "WHEN", "THEN", "ELSE", "CASE", "RETURNING", "USING":
		return true
	}
	return false
}

// contextAt looks at the last complete significant token before the caret (the
// token being typed is excluded).
func contextAt(src string, toks []sigTok, caret int) ctxKind {
	for i := len(toks) - 1; i >= 0; i-- {
		t := toks[i]
		if t.End > caret {
			continue
		}
		if t.End == caret && (t.Kind == KindIdent || t.Kind == KindKeyword || t.Kind == KindFunction || t.Kind == KindType) {
			continue // the word under the caret
		}
		if t.Kind == KindOperator {
			switch t.up {
			case ",":
				// A comma continues the nearest clause: find it.
				return clauseBefore(toks[:i])
			case "(", "=", "<", ">", "<=", ">=", "<>", "!=", "+", "-", "*", "/", "||":
				return ctxColumn
			}
			return ctxKeyword
		}
		if isTableKeyword(t.up) {
			return ctxTable
		}
		if isColumnKeyword(t.up) {
			return ctxColumn
		}
		if t.Kind == KindIdent || t.Kind == KindQuotedIdent || t.Kind == KindNumber || t.Kind == KindString {
			// After a table ref or expression: a keyword is the likely next word,
			// unless we're inside a column list handled by the comma case.
			return ctxKeyword
		}
		return ctxKeyword
	}
	return ctxKeyword
}

func clauseBefore(toks []sigTok) ctxKind {
	depth := 0
	for i := len(toks) - 1; i >= 0; i-- {
		t := toks[i]
		if t.Kind == KindOperator {
			switch t.up {
			case ")":
				depth++
			case "(":
				if depth == 0 {
					return ctxColumn
				}
				depth--
			}
			continue
		}
		if depth > 0 {
			continue
		}
		if isTableKeyword(t.up) {
			return ctxTable
		}
		if isColumnKeyword(t.up) {
			return ctxColumn
		}
	}
	return ctxKeyword
}

// scopeTable is a table referenced in the statement, with its alias.
type scopeTable struct {
	Alias string // empty when none
	Table Table
}

func isNotAlias(up string) bool {
	switch up {
	case "WHERE", "JOIN", "LEFT", "RIGHT", "INNER", "OUTER", "CROSS", "FULL", "NATURAL", "ON", "USING", "GROUP", "ORDER",
		"HAVING", "LIMIT", "OFFSET", "UNION", "INTERSECT", "EXCEPT", "SET", "VALUES", "SELECT", "RETURNING", "WINDOW",
		"FETCH", "FOR", "AS", "AND", "OR", "WITH":
		return true
	}
	return false
}

// tablesInScope parses FROM/JOIN/UPDATE/INTO references and their aliases.
func tablesInScope(src string, toks []sigTok, schema Schema) []scopeTable {
	var out []scopeTable
	word := func(t sigTok) string { return unquote(src[t.Start:t.End]) }
	isName := func(t sigTok) bool { return t.Kind == KindIdent || t.Kind == KindQuotedIdent }
	for i := 0; i < len(toks); i++ {
		if toks[i].up != "FROM" && toks[i].up != "JOIN" && toks[i].up != "UPDATE" && toks[i].up != "INTO" {
			continue
		}
		fromList := toks[i].up == "FROM"
		j := i + 1
		for j < len(toks) {
			if !isName(toks[j]) {
				break
			}
			schemaName, name := "", word(toks[j])
			j++
			if j+1 < len(toks) && toks[j].up == "." && isName(toks[j+1]) {
				schemaName, name = name, word(toks[j+1])
				j += 2
			} else if j < len(toks) && toks[j].up == "." {
				break // "schema." still being typed
			}
			alias := ""
			if j < len(toks) && toks[j].up == "AS" && j+1 < len(toks) && isName(toks[j+1]) {
				alias = word(toks[j+1])
				j += 2
			} else if j < len(toks) && isName(toks[j]) && !isNotAlias(toks[j].up) {
				alias = word(toks[j])
				j++
			}
			t, ok := findTable(schema, schemaName, name)
			if !ok {
				t = Table{Schema: schemaName, Name: name}
			}
			out = append(out, scopeTable{Alias: alias, Table: t})
			if fromList && j < len(toks) && toks[j].up == "," {
				j++
				continue
			}
			break
		}
	}
	return out
}

func unquote(s string) string {
	if len(s) >= quotePairLen {
		switch {
		case s[0] == '"' && s[len(s)-1] == '"':
			return strings.ReplaceAll(s[1:len(s)-1], `""`, `"`)
		case s[0] == '`' && s[len(s)-1] == '`', s[0] == '[' && s[len(s)-1] == ']':
			return s[1 : len(s)-1]
		}
	}
	return s
}

func findTable(schema Schema, sch, name string) (Table, bool) {
	for _, t := range schema.Tables {
		if strings.EqualFold(t.Name, name) && (sch == "" || strings.EqualFold(t.Schema, sch)) {
			return t, true
		}
	}
	return Table{}, false
}

// qualifierBefore returns the identifier directly before the dot at index dot.
func qualifierBefore(src string, dot int) string {
	end := dot
	if end > 0 && (src[end-1] == '"' || src[end-1] == '`' || src[end-1] == ']') {
		q := src[end-1]
		open := q
		if q == ']' {
			open = '['
		}
		i := strings.LastIndexByte(src[:end-1], open)
		if i >= 0 {
			return unquote(src[i:end])
		}
	}
	i := end
	for i > 0 && isIdentPart(src[i-1]) {
		i--
	}
	return src[i:end]
}

func qualified(q string, scope []scopeTable, schema Schema) []Item {
	var items []Item
	seen := false
	for _, s := range scope {
		if strings.EqualFold(s.Alias, q) || (s.Alias == "" && strings.EqualFold(s.Table.Name, q)) {
			items = append(items, columnItems([]scopeTable{s}, false, scoreQualified)...)
			seen = true
		}
	}
	if seen {
		return items
	}
	// A table referenced by name even though it has an alias.
	for _, s := range scope {
		if strings.EqualFold(s.Table.Name, q) {
			return columnItems([]scopeTable{s}, false, scoreQualified)
		}
	}
	// Table known in the catalog but not referenced yet.
	if t, ok := findTable(schema, "", q); ok {
		return columnItems([]scopeTable{{Table: t}}, false, scoreCatalogTable)
	}
	// Schema qualifier: list its tables.
	for _, t := range schema.Tables {
		if strings.EqualFold(t.Schema, q) {
			items = append(items, Item{Label: t.Name, Insert: quoteIfNeeded(t.Name), Kind: ItemTable, Detail: t.Schema, Score: scoreQualified})
		}
	}
	return items
}

func allTables(schema Schema) []scopeTable {
	out := make([]scopeTable, len(schema.Tables))
	for i, t := range schema.Tables {
		out[i] = scopeTable{Table: t}
	}
	return out
}

func tableItems(schema Schema, base int) []Item {
	out := make([]Item, 0, len(schema.Tables))
	for _, t := range schema.Tables {
		detail := ""
		if t.Schema != "" {
			detail = t.Schema
		}
		out = append(out, Item{Label: t.Name, Insert: quoteIfNeeded(t.Name), Kind: ItemTable, Detail: detail, Score: base})
	}
	return out
}

func schemaItems(schema Schema, base int) []Item {
	seen := map[string]bool{}
	var out []Item
	for _, t := range schema.Tables {
		if t.Schema == "" || seen[t.Schema] {
			continue
		}
		seen[t.Schema] = true
		out = append(out, Item{Label: t.Schema, Insert: quoteIfNeeded(t.Schema), Kind: ItemSchema, Score: base})
	}
	return out
}

func columnItems(scope []scopeTable, withTable bool, base int) []Item {
	var out []Item
	seen := map[string]bool{}
	for _, s := range scope {
		for _, c := range s.Table.Columns {
			key := strings.ToLower(c.Name)
			if seen[key] {
				continue
			}
			seen[key] = true
			detail := c.Type
			if withTable && s.Table.Name != "" {
				if detail != "" {
					detail += "  "
				}
				detail += s.Table.Name
			}
			out = append(out, Item{Label: c.Name, Insert: quoteIfNeeded(c.Name), Kind: ItemColumn, Detail: detail, Score: base})
		}
	}
	return out
}

func aliasItems(scope []scopeTable, base int) []Item {
	var out []Item
	for _, s := range scope {
		name := s.Alias
		if name == "" {
			name = s.Table.Name
		}
		if name != "" {
			out = append(out, Item{Label: name, Insert: quoteIfNeeded(name), Kind: ItemTable, Detail: "in scope", Score: base})
		}
	}
	return out
}

func keywordItems(d Dialect, base int) []Item {
	kws := Keywords(d)
	out := make([]Item, 0, len(kws))
	for _, k := range kws {
		out = append(out, Item{Label: k, Insert: k, Kind: ItemKeyword, Score: base})
	}
	return out
}

func functionItems(d Dialect, base int) []Item {
	names := map[string]bool{}
	for _, k := range strings.Fields(functionsWords) {
		names[k] = true
	}
	extra := sqliteFunctionsWords
	if d == Postgres {
		extra = pgFunctionsWords
	}
	for _, k := range strings.Fields(extra) {
		names[k] = true
	}
	out := make([]Item, 0, len(names))
	for k := range names {
		out = append(out, Item{Label: k, Insert: k, Kind: ItemFunction, Detail: "function", Score: base})
	}
	return out
}

func plainIdent(s string) bool {
	if s == "" || isDigit(s[0]) {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (!isIdentStart(c) && !isDigit(c)) || (c >= 'A' && c <= 'Z') {
			return false
		}
	}
	return true
}

// quoteIfNeeded double-quotes identifiers that are not plain lowercase names or
// that collide with a keyword.
func quoteIfNeeded(s string) string {
	if plainIdent(s) && !isKeyword(strings.ToUpper(s)) {
		return s
	}
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

// rank filters by prefix (prefix, word-start, then subsequence match), adds
// the context base score, de-duplicates, and returns the top items.
func rank(cands []Item, prefix string) []Item {
	lp := strings.ToLower(prefix)
	best := map[string]Item{}
	for _, it := range cands {
		bonus, ok := matchScore(strings.ToLower(it.Label), lp)
		if !ok {
			continue
		}
		it.Score += bonus
		if prefix != "" && strings.HasPrefix(it.Label, prefix) {
			it.Score += 5 // exact-case prefix
		}
		key := it.Kind.Badge() + "\x00" + strings.ToLower(it.Label)
		if old, dup := best[key]; !dup || it.Score > old.Score {
			best[key] = it
		}
	}
	out := make([]Item, 0, len(best))
	for _, it := range best {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if len(a.Label) != len(b.Label) {
			return len(a.Label) < len(b.Label)
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		return strings.ToLower(a.Label) < strings.ToLower(b.Label)
	})
	if len(out) > maxItems {
		out = out[:maxItems]
	}
	return out
}

func matchScore(label, prefix string) (int, bool) {
	if prefix == "" {
		return 0, true
	}
	if label == prefix {
		return matchExact, true
	}
	if strings.HasPrefix(label, prefix) {
		return matchPrefix, true
	}
	// Word-start match after '_' (e.g. "ca" matches "created_at").
	if i := strings.Index(label, "_"+prefix); i >= 0 {
		return matchWordStart, true
	}
	if strings.Contains(label, prefix) {
		return matchContains, true
	}
	// Subsequence match, weakest.
	j := 0
	for i := 0; i < len(label) && j < len(prefix); i++ {
		if label[i] == prefix[j] {
			j++
		}
	}
	if j == len(prefix) && len(prefix) >= minSubseqPrefix {
		return matchSubseq, true
	}
	return 0, false
}

// Apply returns src with the completion item inserted, plus the new caret.
func Apply(src string, c Completion, it Item) (string, int) {
	out := src[:c.ReplaceStart] + it.Insert + src[c.ReplaceEnd:]
	return out, c.ReplaceStart + len(it.Insert)
}
