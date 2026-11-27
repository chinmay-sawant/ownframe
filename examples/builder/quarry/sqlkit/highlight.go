// Package sqlkit holds the pure-Go SQL helpers behind Quarry: a dialect-aware
// tokenizer, statement splitter, autocomplete engine, staged changeset SQL
// generator, and CSV/JSON/SQL import and export. It has no UI dependencies.
package sqlkit

import "strings"

// Dialect selects SQLite or PostgreSQL lexing and SQL generation rules.
type Dialect int

// Supported dialects.
const (
	SQLite Dialect = iota
	Postgres
)

// String returns a lowercase dialect name.
func (d Dialect) String() string {
	if d == Postgres {
		return "postgres"
	}
	return "sqlite"
}

// TokenKind classifies a lexed token. Each kind maps to one --qr-syn-* token.
type TokenKind int

// Token kinds.
const (
	KindSpace       TokenKind = iota // whitespace
	KindIdent                        // bare identifier (no highlight colour)
	KindQuotedIdent                  // "name", `name`, [name]
	KindKeyword                      // --qr-syn-kw
	KindFunction                     // --qr-syn-fn
	KindType                         // --qr-syn-type
	KindString                       // 'text', E'text', $$text$$, $tag$text$tag$ (--qr-syn-str)
	KindNumber                       // --qr-syn-num
	KindComment                      // -- and /* */ (--qr-syn-comment)
	KindOperator                     // operators and punctuation (--qr-syn-op)
	KindParam                        // ?, ?1, :name, @name, $1
)

// String returns a short name for the kind.
func (k TokenKind) String() string {
	names := [...]string{"space", "ident", "quoted-ident", "keyword", "function", "type", "string", "number", "comment", "operator", "param"}
	if int(k) < len(names) {
		return names[k]
	}
	return "unknown"
}

// Token is a half-open byte range [Start, End) of the source.
type Token struct {
	Kind       TokenKind
	Start, End int
}

// Text returns the token text within src.
func (t Token) Text(src string) string { return src[t.Start:t.End] }

// Tokenize lexes src into contiguous tokens covering every byte. Unterminated
// strings and comments extend to the end of the input.
func Tokenize(src string, d Dialect) []Token {
	l := lexer{src: src, d: d}
	for l.pos < len(src) {
		l.next()
	}
	return l.out
}

type lexer struct {
	src string
	d   Dialect
	pos int
	out []Token
}

func (l *lexer) emit(k TokenKind, end int) {
	l.out = append(l.out, Token{Kind: k, Start: l.pos, End: end})
	l.pos = end
}

func isSpace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v'
}
func isDigit(c byte) bool { return c >= '0' && c <= '9' }
func isIdentStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c >= 0x80
}
func isIdentPart(c byte) bool { return isIdentStart(c) || isDigit(c) || c == '$' }

func (l *lexer) at(i int) byte {
	if i < len(l.src) {
		return l.src[i]
	}
	return 0
}

func (l *lexer) next() {
	c := l.src[l.pos]
	switch {
	case isSpace(c):
		e := l.pos
		for e < len(l.src) && isSpace(l.src[e]) {
			e++
		}
		l.emit(KindSpace, e)
	case c == '-' && l.at(l.pos+1) == '-':
		e := strings.IndexByte(l.src[l.pos:], '\n')
		if e < 0 {
			e = len(l.src)
		} else {
			e += l.pos
		}
		l.emit(KindComment, e)
	case c == '/' && l.at(l.pos+1) == '*':
		l.emit(KindComment, l.blockCommentEnd())
	case c == '\'':
		l.emit(KindString, l.quotedEnd(l.pos, '\'', l.d == Postgres && l.escapePrefix()))
	case c == '"':
		l.emit(KindQuotedIdent, l.quotedEnd(l.pos, '"', false))
	case c == '`':
		l.emit(KindQuotedIdent, l.quotedEnd(l.pos, '`', false))
	case c == '[' && l.d == SQLite:
		e := strings.IndexByte(l.src[l.pos:], ']')
		if e < 0 {
			e = len(l.src)
		} else {
			e += l.pos + 1
		}
		l.emit(KindQuotedIdent, e)
	case c == '$' && l.d == Postgres:
		l.dollar()
	case isDigit(c) || (c == '.' && isDigit(l.at(l.pos+1))):
		l.emit(KindNumber, l.numberEnd())
	case c == '?':
		e := l.pos + 1
		for e < len(l.src) && isDigit(l.src[e]) {
			e++
		}
		l.emit(KindParam, e)
	case (c == ':' && l.at(l.pos+1) != ':' && isIdentStart(l.at(l.pos+1))) || (c == '@' && isIdentStart(l.at(l.pos+1))):
		e := l.pos + 1
		for e < len(l.src) && isIdentPart(l.src[e]) {
			e++
		}
		l.emit(KindParam, e)
	case isIdentStart(c):
		e := l.pos
		for e < len(l.src) && isIdentPart(l.src[e]) {
			e++
		}
		l.emit(l.classifyWord(l.src[l.pos:e], e), e)
	default:
		l.operator()
	}
}

// escapePrefix reports whether the quote at pos is preceded by an E/e prefix
// token (Postgres escape string, backslash escapes).
func (l *lexer) escapePrefix() bool {
	n := len(l.out)
	if n == 0 {
		return false
	}
	p := l.out[n-1]
	if p.End != l.pos || p.End-p.Start != 1 {
		return false
	}
	c := l.src[p.Start]
	return c == 'E' || c == 'e'
}

func (l *lexer) quotedEnd(start int, q byte, backslash bool) int {
	i := start + 1
	for i < len(l.src) {
		c := l.src[i]
		if backslash && c == '\\' {
			i += 2
			continue
		}
		if c == q {
			if l.at(i+1) == q {
				i += 2
				continue
			}
			return i + 1
		}
		i++
	}
	return len(l.src)
}

// blockCommentEnd returns the end of the /* */ comment at pos. Postgres
// comments nest; SQLite comments do not.
// commentOpenLen is the length of the "/*" opener.
const commentOpenLen = 2

func (l *lexer) blockCommentEnd() int {
	depth := 1
	i := l.pos + commentOpenLen
	for i < len(l.src) {
		switch {
		case l.d == Postgres && l.src[i] == '/' && l.at(i+1) == '*':
			depth++
			i += 2
		case l.src[i] == '*' && l.at(i+1) == '/':
			depth--
			i += 2
			if depth == 0 {
				return i
			}
		default:
			i++
		}
	}
	return len(l.src)
}

func (l *lexer) dollar() {
	// $1 positional parameter.
	if isDigit(l.at(l.pos + 1)) {
		e := l.pos + 1
		for e < len(l.src) && isDigit(l.src[e]) {
			e++
		}
		l.emit(KindParam, e)
		return
	}
	// $tag$ ... $tag$ with an optional tag.
	e := l.pos + 1
	for e < len(l.src) && isIdentPart(l.src[e]) && l.src[e] != '$' {
		e++
	}
	if e < len(l.src) && l.src[e] == '$' && (e <= l.pos+1 || !isDigit(l.src[l.pos+1])) {
		tag := l.src[l.pos : e+1]
		if i := strings.Index(l.src[e+1:], tag); i >= 0 {
			l.emit(KindString, e+1+i+len(tag))
		} else {
			l.emit(KindString, len(l.src))
		}
		return
	}
	l.emit(KindOperator, l.pos+1)
}

func (l *lexer) numberEnd() int {
	i := l.pos
	if l.src[i] == '0' && (l.at(i+1) == 'x' || l.at(i+1) == 'X') {
		i += 2
		for i < len(l.src) && isHex(l.src[i]) {
			i++
		}
		return i
	}
	for i < len(l.src) && isDigit(l.src[i]) {
		i++
	}
	if l.at(i) == '.' {
		i++
		for i < len(l.src) && isDigit(l.src[i]) {
			i++
		}
	}
	if c := l.at(i); c == 'e' || c == 'E' {
		j := i + 1
		if s := l.at(j); s == '+' || s == '-' {
			j++
		}
		if isDigit(l.at(j)) {
			for j < len(l.src) && isDigit(l.src[j]) {
				j++
			}
			i = j
		}
	}
	return i
}

func isHex(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}

func multiOps() []string {
	return []string{"::", "<>", "!=", "<=", ">=", "||", "->>", "->", "#>>", "#>", "<<", ">>", "@>", "<@", "&&", "!~*", "!~", "~*", "==", "?|", "?&"}
}

func (l *lexer) operator() {
	rest := l.src[l.pos:]
	for _, op := range multiOps() {
		if strings.HasPrefix(rest, op) {
			l.emit(KindOperator, l.pos+len(op))
			return
		}
	}
	// Advance by one whole rune so multi-byte input never splits.
	n := 1
	for n < len(rest) && rest[n]&0xC0 == 0x80 {
		n++
	}
	l.emit(KindOperator, l.pos+n)
}

// classifyWord decides keyword/function/type/ident for a bare word ending at end.
func (l *lexer) classifyWord(w string, end int) TokenKind {
	up := strings.ToUpper(w)
	// An identifier directly after '.' is a column or table, never a keyword.
	if n := len(l.out); n > 0 && l.out[n-1].Kind == KindOperator && l.src[l.out[n-1].Start:l.out[n-1].End] == "." {
		return KindIdent
	}
	// After a cast operator the word is a type.
	if n := len(l.out); n > 0 && l.out[n-1].Kind == KindOperator && l.src[l.out[n-1].Start:l.out[n-1].End] == "::" {
		return KindType
	}
	paren := l.parenFollows(end)
	if isType(up) || (l.d == Postgres && isPGType(up)) {
		// date('now'), time(...), datetime(...) are SQLite functions.
		if l.d != SQLite || !paren || !isSQLiteFunction(up) {
			return KindType
		}
	}
	// A function is a known name followed by '(' (possibly after spaces).
	if isFunction(up) || (l.d == Postgres && isPGFunction(up)) || (l.d == SQLite && isSQLiteFunction(up)) {
		if paren {
			return KindFunction
		}
	}
	if isKeyword(up) || (l.d == Postgres && isPGKeyword(up)) || (l.d == SQLite && isSQLiteKeyword(up)) {
		return KindKeyword
	}
	return KindIdent
}

func (l *lexer) parenFollows(end int) bool {
	j := end
	for j < len(l.src) && isSpace(l.src[j]) {
		j++
	}
	return l.at(j) == '('
}

const keywordsWords = `ABORT ADD ALL ALTER ANALYZE AND AS ASC BEGIN BETWEEN BY CASCADE CASE CHECK COLLATE COLUMN COMMIT
CONFLICT CONSTRAINT CREATE CROSS CURRENT_DATE CURRENT_TIME CURRENT_TIMESTAMP DEFAULT DELETE DESC DISTINCT DO DROP
ELSE END ESCAPE EXCEPT EXISTS EXPLAIN FALSE FETCH FILTER FIRST FOREIGN FROM FULL GROUP HAVING IF IN INDEX INNER INSERT
INTERSECT INTO IS ISNULL JOIN KEY LAST LEFT LIKE LIMIT NATURAL NOT NOTHING NOTNULL NULL NULLS OFFSET ON OR ORDER OUTER
OVER PARTITION PRIMARY REFERENCES RELEASE RENAME REPLACE RETURNING RIGHT ROLLBACK ROW ROWS SAVEPOINT SELECT SET TABLE
TEMP TEMPORARY THEN TO TRANSACTION TRIGGER TRUE UNION UNIQUE UPDATE USING VACUUM VALUES VIEW WHEN WHERE WINDOW WITH`

// isKeyword reports whether up is in the keywords word list.
func isKeyword(up string) bool {
	switch up {
	case "ABORT", "ADD", "ALL", "ALTER", "ANALYZE", "AND",
		"AS", "ASC", "BEGIN", "BETWEEN", "BY", "CASCADE",
		"CASE", "CHECK", "COLLATE", "COLUMN", "COMMIT", "CONFLICT",
		"CONSTRAINT", "CREATE", "CROSS", "CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP",
		"DEFAULT", "DELETE", "DESC", "DISTINCT", "DO", "DROP",
		"ELSE", "END", "ESCAPE", "EXCEPT", "EXISTS", "EXPLAIN",
		"FALSE", "FETCH", "FILTER", "FIRST", "FOREIGN", "FROM",
		"FULL", "GROUP", "HAVING", "IF", "IN", "INDEX",
		"INNER", "INSERT", "INTERSECT", "INTO", "IS", "ISNULL",
		"JOIN", "KEY", "LAST", "LEFT", "LIKE", "LIMIT",
		"NATURAL", "NOT", "NOTHING", "NOTNULL", "NULL", "NULLS",
		"OFFSET", "ON", "OR", "ORDER", "OUTER", "OVER",
		"PARTITION", "PRIMARY", "REFERENCES", "RELEASE", "RENAME", "REPLACE",
		"RETURNING", "RIGHT", "ROLLBACK", "ROW", "ROWS", "SAVEPOINT",
		"SELECT", "SET", "TABLE", "TEMP", "TEMPORARY", "THEN",
		"TO", "TRANSACTION", "TRIGGER", "TRUE", "UNION", "UNIQUE",
		"UPDATE", "USING", "VACUUM", "VALUES", "VIEW", "WHEN",
		"WHERE", "WINDOW", "WITH":
		return true
	}
	return false
}

const sqliteKeywordsWords = `AUTOINCREMENT ATTACH DATABASE DETACH GLOB INDEXED INSTEAD MATCH OF PRAGMA REGEXP REINDEX
ROWID STRICT VIRTUAL WITHOUT`

// isSQLiteKeyword reports whether up is in the sqliteKeywords word list.
func isSQLiteKeyword(up string) bool {
	switch up {
	case "AUTOINCREMENT", "ATTACH", "DATABASE", "DETACH", "GLOB", "INDEXED",
		"INSTEAD", "MATCH", "OF", "PRAGMA", "REGEXP", "REINDEX",
		"ROWID", "STRICT", "VIRTUAL", "WITHOUT":
		return true
	}
	return false
}

const pgKeywordsWords = `ARRAY ANY ILIKE LATERAL RETURNS SIMILAR SOME TABLESPACE UNLOGGED CONCURRENTLY MATERIALIZED
OWNER SCHEMA SEQUENCE EXTENSION FUNCTION LANGUAGE PROCEDURE GRANT REVOKE TRUNCATE COPY ONLY RECURSIVE
VERBOSE ANALYSE`

// isPGKeyword reports whether up is in the pgKeywords word list.
func isPGKeyword(up string) bool {
	switch up {
	case "ARRAY", "ANY", "ILIKE", "LATERAL", "RETURNS", "SIMILAR",
		"SOME", "TABLESPACE", "UNLOGGED", "CONCURRENTLY", "MATERIALIZED", "OWNER",
		"SCHEMA", "SEQUENCE", "EXTENSION", "FUNCTION", "LANGUAGE", "PROCEDURE",
		"GRANT", "REVOKE", "TRUNCATE", "COPY", "ONLY", "RECURSIVE",
		"VERBOSE", "ANALYSE":
		return true
	}
	return false
}

// isType reports whether up is in the types word list.
func isType(up string) bool {
	switch up {
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT", "MEDIUMINT",
		"REAL", "DOUBLE", "FLOAT", "NUMERIC", "DECIMAL", "BOOLEAN",
		"BOOL", "TEXT", "VARCHAR", "CHAR", "CHARACTER", "BLOB",
		"DATE", "TIME", "TIMESTAMP", "DATETIME":
		return true
	}
	return false
}

// isPGType reports whether up is in the pgTypes word list.
func isPGType(up string) bool {
	switch up {
	case "BIGSERIAL", "SERIAL", "SMALLSERIAL", "BYTEA", "JSON", "JSONB",
		"UUID", "TIMESTAMPTZ", "TIMETZ", "INTERVAL", "INET", "CIDR",
		"MACADDR", "MONEY", "OID", "XML", "TSVECTOR", "TSQUERY",
		"INT2", "INT4", "INT8", "FLOAT4", "FLOAT8", "CITEXT",
		"HSTORE":
		return true
	}
	return false
}

const functionsWords = `ABS AVG CAST COALESCE COUNT GROUP_CONCAT IFNULL LENGTH LOWER LTRIM MAX MIN NULLIF REPLACE ROUND RTRIM
SUBSTR SUBSTRING SUM TRIM UPPER TOTAL RANK ROW_NUMBER DENSE_RANK LAG LEAD`

// isFunction reports whether up is in the functions word list.
func isFunction(up string) bool {
	switch up {
	case "ABS", "AVG", "CAST", "COALESCE", "COUNT", "GROUP_CONCAT",
		"IFNULL", "LENGTH", "LOWER", "LTRIM", "MAX", "MIN",
		"NULLIF", "REPLACE", "ROUND", "RTRIM", "SUBSTR", "SUBSTRING",
		"SUM", "TRIM", "UPPER", "TOTAL", "RANK", "ROW_NUMBER",
		"DENSE_RANK", "LAG", "LEAD":
		return true
	}
	return false
}

const sqliteFunctionsWords = `DATE DATETIME JULIANDAY STRFTIME TIME UNIXEPOCH RANDOM RANDOMBLOB HEX INSTR PRINTF QUOTE
LAST_INSERT_ROWID CHANGES TYPEOF IIF JSON JSON_EXTRACT JSON_ARRAY JSON_OBJECT JSON_EACH JSON_GROUP_ARRAY ZEROBLOB
UNICODE CHAR LIKELY`

// isSQLiteFunction reports whether up is in the sqliteFunctions word list.
func isSQLiteFunction(up string) bool {
	switch up {
	case "DATE", "DATETIME", "JULIANDAY", "STRFTIME", "TIME", "UNIXEPOCH",
		"RANDOM", "RANDOMBLOB", "HEX", "INSTR", "PRINTF", "QUOTE",
		"LAST_INSERT_ROWID", "CHANGES", "TYPEOF", "IIF", "JSON", "JSON_EXTRACT",
		"JSON_ARRAY", "JSON_OBJECT", "JSON_EACH", "JSON_GROUP_ARRAY", "ZEROBLOB", "UNICODE",
		"CHAR", "LIKELY":
		return true
	}
	return false
}

const pgFunctionsWords = `NOW CURRENT_TIMESTAMP DATE_TRUNC DATE_PART EXTRACT TO_CHAR TO_DATE TO_TIMESTAMP STRING_AGG
ARRAY_AGG JSONB_BUILD_OBJECT JSON_BUILD_OBJECT JSONB_AGG JSON_AGG JSONB_EXTRACT_PATH GENERATE_SERIES UNNEST
POSITION SPLIT_PART REGEXP_REPLACE REGEXP_MATCHES CONCAT CONCAT_WS INITCAP LEFT RIGHT AGE GREATEST LEAST
PG_SIZE_PRETTY PG_TOTAL_RELATION_SIZE PG_DATABASE_SIZE RANDOM MD5 CEIL FLOOR TRUNC`

// isPGFunction reports whether up is in the pgFunctions word list.
func isPGFunction(up string) bool {
	switch up {
	case "NOW", "CURRENT_TIMESTAMP", "DATE_TRUNC", "DATE_PART", "EXTRACT", "TO_CHAR",
		"TO_DATE", "TO_TIMESTAMP", "STRING_AGG", "ARRAY_AGG", "JSONB_BUILD_OBJECT", "JSON_BUILD_OBJECT",
		"JSONB_AGG", "JSON_AGG", "JSONB_EXTRACT_PATH", "GENERATE_SERIES", "UNNEST", "POSITION",
		"SPLIT_PART", "REGEXP_REPLACE", "REGEXP_MATCHES", "CONCAT", "CONCAT_WS", "INITCAP",
		"LEFT", "RIGHT", "AGE", "GREATEST", "LEAST", "PG_SIZE_PRETTY",
		"PG_TOTAL_RELATION_SIZE", "PG_DATABASE_SIZE", "RANDOM", "MD5", "CEIL", "FLOOR",
		"TRUNC":
		return true
	}
	return false
}

// Keywords returns the sorted-by-nothing list of keyword names for dialect d.
func Keywords(d Dialect) []string {
	out := strings.Fields(keywordsWords)
	if d == Postgres {
		return append(out, strings.Fields(pgKeywordsWords)...)
	}
	return append(out, strings.Fields(sqliteKeywordsWords)...)
}
