// Package driver is Quarry's database layer: a unified Driver interface with
// SQLite (modernc.org/sqlite), PostgreSQL (pgx/v5), and an offline mock
// PostgreSQL implementation. It has no UI or Ebiten dependency.
package driver

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EngineKind names a database engine.
type EngineKind string

const (
	EngineSQLite   EngineKind = "sqlite"
	EnginePostgres EngineKind = "postgres"
)

// Environment tags for ConnectionConfig.EnvTag.
const (
	EnvLocal   = "Local"
	EnvStaging = "Staging"
	EnvProd    = "Prod"
)

// ConnectionConfig describes one saved or ad-hoc connection.
type ConnectionConfig struct {
	ID       string
	Name     string
	Engine   EngineKind
	EnvTag   string // "Local", "Staging", "Prod"
	ReadOnly bool
	FilePath string // SQLite path (or ":memory:")
	Host     string // Postgres fields
	Port     int
	Database string
	User     string
	Password string // Stored in memory only unless user opts in
	SSLMode  string // disable, prefer, require, verify-full
	// Mock selects the deterministic offline PostgreSQL driver (Engine must be
	// EnginePostgres). Used by tests, CI, and demos.
	Mock bool
	// ConnectTimeout bounds Ping and pool creation; zero means 5 seconds.
	ConnectTimeout time.Duration
}

// ColumnMeta describes one table or result column.
type ColumnMeta struct {
	Name       string
	Type       string
	Nullable   bool
	PrimaryKey bool
	DefaultVal string
	// ForeignKey is true when the column participates in a foreign key.
	ForeignKey bool
}

// IndexMeta describes one index or key constraint.
type IndexMeta struct {
	Name    string
	Columns []string
	Unique  bool
	Primary bool
	// Method is the access method (btree, hash, gin) when known.
	Method string
	// Definition is the CREATE INDEX statement when known.
	Definition string
}

// ForeignKeyMeta describes one foreign key constraint.
type ForeignKeyMeta struct {
	Name       string
	Columns    []string
	RefSchema  string
	RefTable   string
	RefColumns []string
	OnUpdate   string
	OnDelete   string
}

// SchemaNodeKind classifies a SchemaNode.
type SchemaNodeKind string

const (
	NodeSchema   SchemaNodeKind = "schema"
	NodeTable    SchemaNodeKind = "table"
	NodeView     SchemaNodeKind = "view"
	NodeMatView  SchemaNodeKind = "matview"
	NodeIndex    SchemaNodeKind = "index"
	NodeTrigger  SchemaNodeKind = "trigger"
	NodeFunction SchemaNodeKind = "function"
	// NodeGroup is a synthetic folder ("Tables", "Views", ...) whose Children
	// are all of one kind.
	NodeGroup SchemaNodeKind = "group"
)

// SchemaNode is one entry in the schema explorer tree. ListSchemas returns
// schemas at the top level; each schema has NodeGroup children (Tables, Views,
// Materialized Views, Indexes, Triggers, Functions; empty groups omitted) whose
// children are leaf objects.
type SchemaNode struct {
	Name     string
	Kind     SchemaNodeKind
	Schema   string // owning schema ("main" for SQLite)
	Table    string // owning table for index and trigger nodes
	RowCount int64  // approximate rows; -1 when unknown or not applicable
	Detail   string // short hint, e.g. "UNIQUE", view definition head
	Children []SchemaNode
}

// QualifiedName returns schema.name quoted for display.
func (n SchemaNode) QualifiedName() string {
	if n.Schema == "" {
		return n.Name
	}
	return n.Schema + "." + n.Name
}

// QueryResult is a bounded result set.
type QueryResult struct {
	Columns      []ColumnMeta
	Rows         [][]CellValue
	RowsAffected int64
	Duration     time.Duration
	Truncated    bool // more rows existed than maxRows
}

// ActiveQuery is one running statement on the server.
type ActiveQuery struct {
	PID       int
	User      string
	Database  string
	State     string // active, idle, idle in transaction, ...
	Query     string
	StartedAt time.Time
	Duration  time.Duration
	WaitEvent string
	Blocked   bool
}

// StorageItem is one row of the storage breakdown (tables, indexes, WAL, ...).
type StorageItem struct {
	Name  string
	Bytes int64
}

// MetricSnapshot is one telemetry sample. Counter fields (CommitTotal,
// RollbackTotal, BlocksHit, BlocksRead) are cumulative so the collector can
// derive rates; TPS is the driver's own best estimate and may be zero for the
// first sample.
type MetricSnapshot struct {
	CapturedAt    time.Time
	ActiveConns   int
	IdleConns     int
	WaitingLocks  int
	TPS           float64
	CacheHitRatio float64 // 0..1; negative when the engine cannot report it
	DBSizeBytes   int64
	CommitTotal   int64
	RollbackTotal int64
	BlocksHit     int64
	BlocksRead    int64
	Storage       []StorageItem
	ActiveQueries []ActiveQuery
	// Extra carries engine-specific numbers (page_count, wal_pages, ...).
	Extra map[string]float64
}

// ScanKind classifies how a plan node reads data.
type ScanKind string

const (
	ScanNone   ScanKind = ""
	ScanFull   ScanKind = "SCAN"   // full table scan (SQLite SCAN, PG Seq Scan)
	ScanSearch ScanKind = "SEARCH" // indexed lookup (SQLite SEARCH, PG Index Scan)
)

// PlanNode is one node of an EXPLAIN tree.
type PlanNode struct {
	ID          int
	ParentID    int // -1 for the root
	NodeType    string
	Relation    string
	Index       string
	Detail      string
	Scan        ScanKind
	StartupCost float64
	TotalCost   float64
	PlanRows    float64
	ActualRows  float64
	ActualMs    float64 // total actual time in milliseconds; -1 when not analyzed
	Loops       float64
	SharedHit   int64
	SharedRead  int64
	Children    []*PlanNode
}

// ExplainPlan is the shared EXPLAIN result for every engine.
type ExplainPlan struct {
	Engine        EngineKind
	Root          *PlanNode
	Analyzed      bool
	PlanningMs    float64
	ExecutionMs   float64
	Raw           string // raw text or JSON as returned by the engine
	FullScans     int    // count of ScanFull nodes
	IndexSearches int    // count of ScanSearch nodes
}

// Driver is the engine-neutral database interface. Implementations are safe
// for concurrent use.
type Driver interface {
	Engine() EngineKind
	Config() ConnectionConfig
	// Version returns the server version string for the status bar.
	Version(ctx context.Context) (string, error)
	Ping(ctx context.Context) error
	Close() error
	ListSchemas(ctx context.Context) ([]SchemaNode, error)
	TableColumns(ctx context.Context, schema, table string) ([]ColumnMeta, error)
	TableIndexes(ctx context.Context, schema, table string) ([]IndexMeta, error)
	TableForeignKeys(ctx context.Context, schema, table string) ([]ForeignKeyMeta, error)
	TableDDL(ctx context.Context, schema, table string) (string, error)
	// Query runs a read statement and returns at most maxRows rows, setting
	// Truncated when more existed. maxRows <= 0 means DefaultMaxRows.
	Query(ctx context.Context, sqlText string, maxRows int) (QueryResult, error)
	// Exec runs a write or DDL statement and returns rows affected.
	Exec(ctx context.Context, sqlText string) (int64, error)
	Explain(ctx context.Context, sqlText string, analyze bool) (ExplainPlan, error)
	CollectMetrics(ctx context.Context) (MetricSnapshot, error)
	// CancelQuery asks the server to cancel the statement of backend pid.
	// For SQLite pid identifies a statement running in this process (see
	// MetricSnapshot.ActiveQueries); unknown pids return an error.
	CancelQuery(ctx context.Context, pid int) error
}

// TxExecer is implemented by drivers that can run a batch of statements, each
// with bound arguments, in one transaction. On failure everything is rolled
// back and failed is the index of the offending statement (-1 on success).
type TxExecer interface {
	ExecTx(ctx context.Context, sqls []string, args [][]any) (failed int, err error)
}

// DefaultMaxRows caps results when callers pass maxRows <= 0.
const DefaultMaxRows = 100000

var (
	// ErrReadOnly is returned for write statements on read-only connections.
	ErrReadOnly = errors.New("connection is read-only: write statements are blocked")
	// ErrUnsupported is returned for operations an engine cannot perform.
	ErrUnsupported = errors.New("operation not supported by this engine")
)

// Open returns the Driver for cfg: SQLite, pgx PostgreSQL, or the mock when
// cfg.Mock is set.
func Open(ctx context.Context, cfg ConnectionConfig) (Driver, error) {
	switch cfg.Engine {
	case EngineSQLite:
		return OpenSQLite(ctx, cfg)
	case EnginePostgres:
		if cfg.Mock {
			return NewPGMock(cfg), nil
		}
		return OpenPostgres(ctx, cfg)
	default:
		return nil, fmt.Errorf("driver: unknown engine %q", cfg.Engine)
	}
}

// CellKind is the dynamic type of a CellValue.
type CellKind int

const (
	KindNull CellKind = iota
	KindInt
	KindFloat
	KindText
	KindBool
	KindTime
	KindJSON
	KindBlob
)

// String names the kind for badges and type columns.
func (k CellKind) String() string {
	switch k {
	case KindNull:
		return "null"
	case KindInt:
		return "int"
	case KindFloat:
		return "float"
	case KindText:
		return "text"
	case KindBool:
		return "bool"
	case KindTime:
		return "time"
	case KindJSON:
		return "json"
	case KindBlob:
		return "blob"
	}
	return "unknown"
}

// CellValue is a typed cell. Only the field matching Kind is meaningful; JSON
// values keep their text in S.
type CellValue struct {
	Kind CellKind
	I    int64
	F    float64
	S    string
	B    bool
	T    time.Time
	Blob []byte
}

// Constructors.
func Null() CellValue               { return CellValue{Kind: KindNull} }
func Int(v int64) CellValue         { return CellValue{Kind: KindInt, I: v} }
func Float(v float64) CellValue     { return CellValue{Kind: KindFloat, F: v} }
func Text(v string) CellValue       { return CellValue{Kind: KindText, S: v} }
func Bool(v bool) CellValue         { return CellValue{Kind: KindBool, B: v} }
func Time(v time.Time) CellValue    { return CellValue{Kind: KindTime, T: v} }
func JSON(v string) CellValue       { return CellValue{Kind: KindJSON, S: v} }
func Blob(v []byte) CellValue       { return CellValue{Kind: KindBlob, Blob: v} }
func (c CellValue) IsNull() bool    { return c.Kind == KindNull }
func (c CellValue) IsNumeric() bool { return c.Kind == KindInt || c.Kind == KindFloat }

// String formats the cell for display. NULL renders as "NULL"; blobs render as
// "<n bytes> 0x..." with a short hex preview; times use RFC 3339.
func (c CellValue) String() string {
	switch c.Kind {
	case KindNull:
		return "NULL"
	case KindInt:
		return strconv.FormatInt(c.I, 10)
	case KindFloat:
		return strconv.FormatFloat(c.F, 'g', -1, 64)
	case KindText, KindJSON:
		return c.S
	case KindBool:
		if c.B {
			return "true"
		}
		return "false"
	case KindTime:
		if c.T.Nanosecond() == 0 {
			return c.T.Format("2006-01-02 15:04:05")
		}
		return c.T.Format("2006-01-02 15:04:05.999999")
	case KindBlob:
		n := len(c.Blob)
		pre := c.Blob
		if len(pre) > blobPreview {
			pre = pre[:blobPreview]
		}
		suffix := ""
		if n > len(pre) {
			suffix = "..."
		}
		return fmt.Sprintf("<%d bytes> 0x%s%s", n, hex.EncodeToString(pre), suffix)
	}
	return ""
}

// Equal reports whether two cells hold the same kind and value.
func (c CellValue) Equal(o CellValue) bool {
	if c.Kind != o.Kind {
		return false
	}
	switch c.Kind {
	case KindNull:
		return true
	case KindInt:
		return c.I == o.I
	case KindFloat:
		return c.F == o.F
	case KindText, KindJSON:
		return c.S == o.S
	case KindBool:
		return c.B == o.B
	case KindTime:
		return c.T.Equal(o.T)
	case KindBlob:
		return string(c.Blob) == string(o.Blob)
	}
	return false
}

// SQLLiteral renders the cell as a SQL literal usable in generated statements.
func (c CellValue) SQLLiteral() string {
	switch c.Kind {
	case KindNull:
		return "NULL"
	case KindInt:
		return strconv.FormatInt(c.I, 10)
	case KindFloat:
		return strconv.FormatFloat(c.F, 'g', -1, 64)
	case KindBool:
		if c.B {
			return "TRUE"
		}
		return "FALSE"
	case KindBlob:
		return "X'" + hex.EncodeToString(c.Blob) + "'"
	case KindTime:
		return "'" + c.String() + "'"
	default:
		return "'" + strings.ReplaceAll(c.S, "'", "''") + "'"
	}
}

// FromAny converts a database/sql scanned value into a CellValue. declType is
// the column's declared type name (may be empty) and is used to tag JSON.
func FromAny(v any, declType string) CellValue {
	switch x := v.(type) {
	case nil:
		return Null()
	case int64:
		if strings.EqualFold(declType, "BOOLEAN") || strings.EqualFold(declType, "BOOL") {
			return Bool(x != 0)
		}
		return Int(x)
	case int:
		return Int(int64(x))
	case int32:
		return Int(int64(x))
	case int16:
		return Int(int64(x))
	case float64:
		return Float(x)
	case float32:
		return Float(float64(x))
	case bool:
		return Bool(x)
	case time.Time:
		return Time(x)
	case []byte:
		if isJSONType(declType) {
			return JSON(string(x))
		}
		if isTextType(declType) {
			return Text(string(x))
		}
		return Blob(append([]byte(nil), x...))
	case string:
		if isJSONType(declType) {
			return JSON(x)
		}
		return Text(x)
	default:
		return Text(fmt.Sprint(x))
	}
}

func isJSONType(t string) bool {
	t = strings.ToLower(t)
	return t == "json" || t == "jsonb"
}

func isTextType(t string) bool {
	t = strings.ToLower(t)
	return strings.Contains(t, "char") || strings.Contains(t, "text") || t == "name" || t == "clob"
}

// CheckReadOnly returns ErrReadOnly when sqlText contains a statement that can
// modify data or schema. Allowed leading keywords are SELECT, WITH (provided no
// write keyword appears outside strings and comments), VALUES, TABLE, SHOW,
// DESCRIBE, EXPLAIN (without ANALYZE of a write), and PRAGMA without an
// assignment. Every statement of a multi-statement script is checked.
func CheckReadOnly(sqlText string) error {
	stmts := splitStatements(sqlText)
	for _, st := range stmts {
		toks := keywordTokens(st)
		if len(toks) == 0 {
			continue
		}
		switch toks[0] {
		case "SELECT", "VALUES", "TABLE", "SHOW", "DESCRIBE", "DESC":
			if hasWriteKeyword(toks[1:], true) {
				return fmt.Errorf("%w (%s)", ErrReadOnly, toks[0])
			}
		case "WITH", "EXPLAIN":
			if hasWriteKeyword(toks[1:], false) {
				return fmt.Errorf("%w (%s)", ErrReadOnly, toks[0])
			}
		case "PRAGMA":
			if strings.Contains(st, "=") {
				return fmt.Errorf("%w (PRAGMA assignment)", ErrReadOnly)
			}
		default:
			return fmt.Errorf("%w (%s)", ErrReadOnly, toks[0])
		}
	}
	return nil
}

// IsReadStatement reports whether sqlText would pass CheckReadOnly. The UI
// uses it to pick between Query and Exec.
func IsReadStatement(sqlText string) bool {
	return CheckReadOnly(sqlText) == nil && hasStatement(sqlText)
}

func hasStatement(sqlText string) bool {
	for _, st := range splitStatements(sqlText) {
		if len(keywordTokens(st)) > 0 {
			return true
		}
	}
	return false
}

// blobPreview is how many leading blob bytes CellValue.String shows in hex.
const blobPreview = 8

func isWriteKeyword(t string) bool {
	switch t {
	case "INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "TRUNCATE", "CREATE", "MERGE", "GRANT", "REVOKE",
		"ATTACH", "DETACH", "VACUUM", "REINDEX", "COPY":
		return true
	}
	return false
}

// hasWriteKeyword scans keyword tokens after the leading one. For a plain
// SELECT only INTO is suspect (SELECT ... INTO creates a table in PostgreSQL);
// WITH and EXPLAIN also flag any data-modifying keyword, because a CTE can
// carry INSERT, UPDATE, or DELETE.
func hasWriteKeyword(toks []string, selectOnly bool) bool {
	for _, t := range toks {
		if t == "INTO" {
			return true
		}
		if !selectOnly && isWriteKeyword(t) {
			return true
		}
	}
	return false
}

// splitStatements splits on top-level semicolons, honoring quotes and comments.
func splitStatements(s string) []string {
	var out []string
	var b strings.Builder
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			j := skipQuoted(s, i)
			b.WriteString(s[i:j])
			i = j - 1
		case c == '-' && i+1 < n && s[i+1] == '-':
			for i < n && s[i] != '\n' {
				i++
			}
			b.WriteByte('\n')
		case c == '/' && i+1 < n && s[i+1] == '*':
			j := strings.Index(s[i+2:], "*/")
			if j < 0 {
				i = n
			} else {
				i += j + len("*/") + 1
			}
			b.WriteByte(' ')
		case c == ';':
			out = append(out, b.String())
			b.Reset()
		default:
			b.WriteByte(c)
		}
	}
	if strings.TrimSpace(b.String()) != "" {
		out = append(out, b.String())
	}
	return out
}

func skipQuoted(s string, i int) int {
	q := s[i]
	j := i + 1
	for j < len(s) {
		if s[j] == q {
			if j+1 < len(s) && s[j+1] == q {
				j += 2
				continue
			}
			return j + 1
		}
		j++
	}
	return len(s)
}

// keywordTokens returns upper-cased bare words outside quotes of a single
// statement (comments already removed by splitStatements).
func keywordTokens(st string) []string {
	var toks []string
	n := len(st)
	for i := 0; i < n; i++ {
		c := st[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			i = skipQuoted(st, i) - 1
		case isWordStart(c):
			j := i
			for j < n && (isWordStart(st[j]) || (st[j] >= '0' && st[j] <= '9')) {
				j++
			}
			toks = append(toks, strings.ToUpper(st[i:j]))
			i = j - 1
		}
	}
	return toks
}

func isWordStart(c byte) bool {
	return c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// QuoteIdent quotes an identifier with double quotes (valid for SQLite and
// PostgreSQL).
func QuoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// QualifiedTable returns a quoted schema.table, omitting a blank schema.
func QualifiedTable(schema, table string) string {
	if schema == "" {
		return QuoteIdent(table)
	}
	return QuoteIdent(schema) + "." + QuoteIdent(table)
}
