package sqlkit

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Rows yields result rows in order. Slices are only valid during the yield.
type Rows = iter.Seq[[]any]

// SliceRows adapts an in-memory result to Rows.
func SliceRows(rows [][]any) Rows {
	return func(yield func([]any) bool) {
		for _, r := range rows {
			if !yield(r) {
				return
			}
		}
	}
}

// ExportOptions configures the writers. The zero value is a comma-delimited
// CSV with a header row and empty NULLs.
type ExportOptions struct {
	Delimiter rune   // CSV only; default ','
	NoHeader  bool   // CSV only: omit the header row
	Null      string // CSV text for NULL; default empty
	Table     string // SQL: target table name
	Schema    string // SQL: optional schema
	Dialect   Dialect
	BatchSize int // SQL: rows per INSERT; default 1
}

// textValue formats v for CSV cells.
func textValue(v any, null string) string {
	switch x := v.(type) {
	case nil:
		return null
	case string:
		return x
	case []byte:
		if utf8.Valid(x) {
			return string(x)
		}
		return `\x` + hex.EncodeToString(x)
	case time.Time:
		return x.Format(time.RFC3339Nano)
	case bool:
		return strconv.FormatBool(x)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32)
	default:
		return fmt.Sprint(x)
	}
}

// WriteCSV writes rows as RFC 4180 CSV.
func WriteCSV(w io.Writer, cols []string, rows Rows, o ExportOptions) error {
	cw := csv.NewWriter(w)
	if o.Delimiter != 0 {
		cw.Comma = o.Delimiter
	}
	if !o.NoHeader {
		if err := cw.Write(cols); err != nil {
			return err
		}
	}
	rec := make([]string, len(cols))
	var werr error
	for r := range rows {
		for i := range rec {
			var v any
			if i < len(r) {
				v = r[i]
			}
			rec[i] = textValue(v, o.Null)
		}
		if werr = cw.Write(rec); werr != nil {
			break
		}
	}
	cw.Flush()
	if werr != nil {
		return werr
	}
	return cw.Error()
}

// jsonValue converts a cell into a value encoding/json renders faithfully.
func jsonValue(v any) any {
	switch x := v.(type) {
	case []byte:
		if utf8.Valid(x) {
			return string(x)
		}
		return `\x` + hex.EncodeToString(x)
	case time.Time:
		return x.Format(time.RFC3339Nano)
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return fmt.Sprint(x)
		}
	case float32:
		return jsonValue(float64(x))
	}
	return v
}

func jsonObject(cols []string, r []any, indent string) ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, c := range cols {
		if i > 0 {
			b.WriteByte(',')
		}
		if indent != "" {
			b.WriteString("\n" + indent + indent)
		}
		k, _ := json.Marshal(c)
		b.Write(k)
		b.WriteByte(':')
		if indent != "" {
			b.WriteByte(' ')
		}
		var v any
		if i < len(r) {
			v = jsonValue(r[i])
		}
		vb, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		b.Write(vb)
	}
	if indent != "" && len(cols) > 0 {
		b.WriteString("\n" + indent)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// WriteJSON writes rows as a pretty-printed array of objects, keeping column order.
func WriteJSON(w io.Writer, cols []string, rows Rows) error {
	bw := bufio.NewWriter(w)
	first := true
	_, _ = bw.WriteString("[")
	for r := range rows {
		ob, err := jsonObject(cols, r, "  ")
		if err != nil {
			return err
		}
		if !first {
			_, _ = bw.WriteString(",")
		}
		first = false
		_, _ = bw.WriteString("\n  ")
		_, _ = bw.Write(ob)
	}
	if first {
		_, _ = bw.WriteString("]\n")
	} else {
		_, _ = bw.WriteString("\n]\n")
	}
	return bw.Flush()
}

// WriteNDJSON writes one compact JSON object per line.
func WriteNDJSON(w io.Writer, cols []string, rows Rows) error {
	bw := bufio.NewWriter(w)
	for r := range rows {
		ob, err := jsonObject(cols, r, "")
		if err != nil {
			return err
		}
		_, _ = bw.Write(ob)
		_ = bw.WriteByte('\n')
	}
	return bw.Flush()
}

// Literal renders v as a SQL literal for dialect d.
func Literal(v any, d Dialect) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case bool:
		if d == Postgres {
			if x {
				return "TRUE"
			}
			return "FALSE"
		}
		if x {
			return "1"
		}
		return "0"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprint(x)
	case float32:
		return floatLiteral(float64(x), d)
	case float64:
		return floatLiteral(x, d)
	case []byte:
		if d == Postgres {
			return `'\x` + hex.EncodeToString(x) + `'`
		}
		return "X'" + hex.EncodeToString(x) + "'"
	case time.Time:
		return quoteString(x.Format(time.RFC3339Nano))
	case string:
		return quoteString(x)
	default:
		return quoteString(fmt.Sprint(x))
	}
}

func floatLiteral(f float64, d Dialect) string {
	switch {
	case math.IsNaN(f):
		if d == Postgres {
			return "'NaN'"
		}
		return "NULL"
	case math.IsInf(f, 1):
		if d == Postgres {
			return "'Infinity'"
		}
		return "9e999"
	case math.IsInf(f, -1):
		if d == Postgres {
			return "'-Infinity'"
		}
		return "-9e999"
	}
	s := strconv.FormatFloat(f, 'g', -1, 64)
	if !strings.ContainsAny(s, ".eE") {
		s += ".0"
	}
	return s
}

func quoteString(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// WriteSQL writes INSERT statements, each terminated by ";\n".
func WriteSQL(w io.Writer, cols []string, rows Rows, o ExportOptions) error {
	if o.Table == "" {
		return errors.New("sql export: table name required")
	}
	target := QuoteIdent(o.Table)
	if o.Schema != "" {
		target = QuoteIdent(o.Schema) + "." + target
	}
	names := make([]string, len(cols))
	for i, c := range cols {
		names[i] = QuoteIdent(c)
	}
	head := "INSERT INTO " + target + " (" + strings.Join(names, ", ") + ") VALUES\n"
	batch := max(o.BatchSize, 1)
	bw := bufio.NewWriter(w)
	n := 0
	for r := range rows {
		if n%batch == 0 {
			if n > 0 {
				_, _ = bw.WriteString(";\n")
			}
			_, _ = bw.WriteString(head)
		} else {
			_, _ = bw.WriteString(",\n")
		}
		_, _ = bw.WriteString("  (")
		for i := range cols {
			if i > 0 {
				_, _ = bw.WriteString(", ")
			}
			var v any
			if i < len(r) {
				v = r[i]
			}
			_, _ = bw.WriteString(Literal(v, o.Dialect))
		}
		_, _ = bw.WriteString(")")
		n++
	}
	if n > 0 {
		_, _ = bw.WriteString(";\n")
	}
	return bw.Flush()
}

// ColumnType is an inferred import column type.
type ColumnType int

// Inferred types, narrowest first.
const (
	TypeBoolean ColumnType = iota
	TypeInteger
	TypeReal
	TypeText
)

// SQL returns the column type name for dialect d.
func (t ColumnType) SQL(d Dialect) string {
	switch t {
	case TypeBoolean:
		return "BOOLEAN"
	case TypeInteger:
		if d == Postgres {
			return "BIGINT"
		}
		return "INTEGER"
	case TypeReal:
		if d == Postgres {
			return "DOUBLE PRECISION"
		}
		return "REAL"
	}
	return "TEXT"
}

// ImportColumn is a parsed CSV column.
type ImportColumn struct {
	Name string
	Type ColumnType
}

// ImportData is a parsed, type-inferred CSV file. Rows hold nil, int64,
// float64, bool, or string per the column type.
type ImportData struct {
	Columns []ImportColumn
	Rows    [][]any
}

// ImportOptions configures ParseCSV.
type ImportOptions struct {
	Delimiter rune   // default ','
	NoHeader  bool   // first row is data; columns are named column_1...
	Null      string // text read as NULL in addition to empty fields; default none
}

// utf8BOM is skipped at the start of imported files.
const utf8BOMString = "\xEF\xBB\xBF"

// ParseCSV reads a CSV file and infers each column's type from every value:
// BOOLEAN (true/false), INTEGER, REAL, else TEXT. Empty fields become NULL.
func ParseCSV(r io.Reader, o ImportOptions) (*ImportData, error) {
	br := bufio.NewReader(r)
	if b, err := br.Peek(len(utf8BOMString)); err == nil && bytes.Equal(b, []byte(utf8BOMString)) {
		_, _ = br.Discard(len(utf8BOMString))
	}
	cr := csv.NewReader(br)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = true
	if o.Delimiter != 0 {
		cr.Comma = o.Delimiter
	}
	var recs [][]string
	for {
		rec, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("csv: %w", err)
		}
		recs = append(recs, rec)
	}
	if len(recs) == 0 {
		return nil, errors.New("csv: file is empty")
	}
	width := 0
	for _, r := range recs {
		width = max(width, len(r))
	}
	names := make([]string, width)
	body := recs
	if !o.NoHeader {
		for i := range names {
			if i < len(recs[0]) {
				names[i] = strings.TrimSpace(recs[0][i])
			}
		}
		body = recs[1:]
	}
	used := map[string]bool{}
	for i := range names {
		n := names[i]
		if n == "" {
			n = fmt.Sprintf("column_%d", i+1)
		}
		base := n
		for k := 2; used[strings.ToLower(n)]; k++ {
			n = fmt.Sprintf("%s_%d", base, k)
		}
		used[strings.ToLower(n)] = true
		names[i] = n
	}
	isNull := func(s string) bool { return s == "" || (o.Null != "" && s == o.Null) }
	cols := make([]ImportColumn, width)
	for c := range cols {
		t, seen, bools := TypeBoolean, false, false
		for _, rec := range body {
			if c >= len(rec) || isNull(rec[c]) {
				continue
			}
			seen = true
			it := inferType(strings.TrimSpace(rec[c]))
			if it == TypeBoolean {
				bools = true
			}
			t = max(t, it)
			if t == TypeText {
				break
			}
		}
		// Booleans mixed with numbers are not a boolean column.
		if bools && t != TypeBoolean {
			t = TypeText
		}
		if !seen {
			t = TypeText
		}
		cols[c] = ImportColumn{Name: names[c], Type: t}
	}
	out := &ImportData{Columns: cols, Rows: make([][]any, 0, len(body))}
	for _, rec := range body {
		row := make([]any, width)
		for c := range row {
			if c >= len(rec) || isNull(rec[c]) {
				continue
			}
			row[c] = convert(rec[c], cols[c].Type)
		}
		out.Rows = append(out.Rows, row)
	}
	return out, nil
}

func inferType(s string) ColumnType {
	if l := strings.ToLower(s); l == "true" || l == "false" {
		return TypeBoolean
	}
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return TypeInteger
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil && !math.IsNaN(f) && !math.IsInf(f, 0) && !strings.ContainsAny(s, "xXpP_") {
		return TypeReal
	}
	return TypeText
}

func convert(s string, t ColumnType) any {
	trimmed := strings.TrimSpace(s)
	switch t {
	case TypeBoolean:
		return strings.EqualFold(trimmed, "true")
	case TypeInteger:
		n, _ := strconv.ParseInt(trimmed, 10, 64)
		return n
	case TypeReal:
		f, _ := strconv.ParseFloat(trimmed, 64)
		return f
	}
	return s
}

// Preview returns up to n leading rows.
func (d *ImportData) Preview(n int) [][]any { return d.Rows[:min(max(n, 0), len(d.Rows))] }

// CreateTableSQL renders CREATE TABLE for the inferred columns.
func (d *ImportData) CreateTableSQL(table string, dialect Dialect) string {
	parts := make([]string, len(d.Columns))
	for i, c := range d.Columns {
		parts[i] = QuoteIdent(c.Name) + " " + c.Type.SQL(dialect)
	}
	return "CREATE TABLE " + QuoteIdent(table) + " (" + strings.Join(parts, ", ") + ")"
}

// InsertSQL renders the parameterized INSERT used for each imported row.
func (d *ImportData) InsertSQL(table string, dialect Dialect) string {
	names := make([]string, len(d.Columns))
	phs := make([]string, len(d.Columns))
	p := newPlaceholders(dialect)
	for i, c := range d.Columns {
		names[i] = QuoteIdent(c.Name)
		phs[i] = p.next()
	}
	return "INSERT INTO " + QuoteIdent(table) + " (" + strings.Join(names, ", ") + ") VALUES (" + strings.Join(phs, ", ") + ")"
}
