package sqlkit

import (
	"errors"
	"fmt"
	"strings"
)

// Errors returned when an edit cannot be staged.
var (
	ErrReadOnly = errors.New("connection is read-only")
	ErrNoRowID  = errors.New("table has no primary key or row identifier")
	ErrNoChange = errors.New("no pending changes")
)

// RowID identifies one row by its key columns. For tables without a primary
// key use RowIDFallback (SQLite rowid, Postgres ctid).
type RowID struct {
	Cols []string
	Vals []any
}

// Key is a stable map key for the row.
func (r RowID) Key() string {
	var b strings.Builder
	for i, c := range r.Cols {
		fmt.Fprintf(&b, "%s=%T:%v;", c, r.Vals[i], r.Vals[i])
	}
	return b.String()
}

// RowIDFallback returns the engine's hidden row identifier column for d.
func RowIDFallback(d Dialect) string {
	if d == Postgres {
		return "ctid"
	}
	return "rowid"
}

// Config describes the table a changeset edits.
type Config struct {
	Dialect  Dialect
	Schema   string // optional
	Table    string
	ReadOnly bool
	// HasRowID is false when the table has neither a primary key nor a usable
	// hidden row identifier (for example a view or WITHOUT ROWID table lacking
	// a key). All edits are then refused.
	HasRowID bool
}

// Statement is one parameterized statement of a commit.
type Statement struct {
	Kind string // "UPDATE", "INSERT" or "DELETE"
	SQL  string
	Args []any
}

type cellEdit struct {
	col      string
	old, val any
}

type opKind int

const (
	opUpdate opKind = iota
	opInsert
	opDelete
)

type op struct {
	kind opKind
	id   RowID
	// update
	edits []cellEdit
	// insert
	insID int
	cols  []string
	vals  []any
}

// Changeset tracks staged cell edits, inserted rows, and deleted rows before
// they are written inside one transaction. It is not safe for concurrent use.
type Changeset struct {
	cfg    Config
	ops    []*op
	nextID int
}

// NewChangeset creates an empty changeset for cfg.
func NewChangeset(cfg Config) *Changeset { return &Changeset{cfg: cfg} }

// Config returns the table configuration.
func (c *Changeset) Config() Config { return c.cfg }

func (c *Changeset) check(id *RowID) error {
	if c.cfg.ReadOnly {
		return ErrReadOnly
	}
	if id != nil && (len(id.Cols) == 0 || len(id.Cols) != len(id.Vals)) {
		return ErrNoRowID
	}
	if id != nil && !c.cfg.HasRowID {
		return ErrNoRowID
	}
	return nil
}

func (c *Changeset) find(kind opKind, key string) *op {
	for _, o := range c.ops {
		if o.kind == kind && o.id.Key() == key {
			return o
		}
	}
	return nil
}

// UpdateCell stages col from old to val on the row id. Setting a value equal to
// the original drops the edit; a row that is staged for deletion is rejected.
func (c *Changeset) UpdateCell(id RowID, col string, old, val any) error {
	if err := c.check(&id); err != nil {
		return err
	}
	key := id.Key()
	if c.find(opDelete, key) != nil {
		return fmt.Errorf("row is staged for deletion")
	}
	o := c.find(opUpdate, key)
	if o == nil {
		if valuesEqual(old, val) {
			return nil
		}
		o = &op{kind: opUpdate, id: id}
		c.ops = append(c.ops, o)
	}
	for i := range o.edits {
		if o.edits[i].col == col {
			if valuesEqual(o.edits[i].old, val) {
				o.edits = append(o.edits[:i], o.edits[i+1:]...)
				c.prune()
				return nil
			}
			o.edits[i].val = val
			return nil
		}
	}
	if valuesEqual(old, val) {
		return nil
	}
	o.edits = append(o.edits, cellEdit{col: col, old: old, val: val})
	return nil
}

func (c *Changeset) prune() {
	kept := c.ops[:0]
	for _, o := range c.ops {
		if o.kind == opUpdate && len(o.edits) == 0 {
			continue
		}
		kept = append(kept, o)
	}
	c.ops = kept
}

func valuesEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	return fmt.Sprintf("%T:%v", a, a) == fmt.Sprintf("%T:%v", b, b)
}

// InsertRow stages a new row and returns a handle for RemoveInsert. Columns not
// listed take their defaults.
func (c *Changeset) InsertRow(cols []string, vals []any) (int, error) {
	if err := c.check(nil); err != nil {
		return 0, err
	}
	if len(cols) != len(vals) {
		return 0, fmt.Errorf("insert: %d columns but %d values", len(cols), len(vals))
	}
	c.nextID++
	c.ops = append(c.ops, &op{kind: opInsert, insID: c.nextID, cols: append([]string(nil), cols...), vals: append([]any(nil), vals...)})
	return c.nextID, nil
}

// RemoveInsert unstages a row added with InsertRow.
func (c *Changeset) RemoveInsert(handle int) bool {
	for i, o := range c.ops {
		if o.kind == opInsert && o.insID == handle {
			c.ops = append(c.ops[:i], c.ops[i+1:]...)
			return true
		}
	}
	return false
}

// DeleteRow stages deletion of the row id and drops its pending cell edits.
func (c *Changeset) DeleteRow(id RowID) error {
	if err := c.check(&id); err != nil {
		return err
	}
	key := id.Key()
	if c.find(opDelete, key) != nil {
		return nil
	}
	if u := c.find(opUpdate, key); u != nil {
		u.edits = nil
		c.prune()
	}
	c.ops = append(c.ops, &op{kind: opDelete, id: id})
	return nil
}

// UnstageDelete restores a row staged for deletion.
func (c *Changeset) UnstageDelete(id RowID) bool {
	key := id.Key()
	for i, o := range c.ops {
		if o.kind == opDelete && o.id.Key() == key {
			c.ops = append(c.ops[:i], c.ops[i+1:]...)
			return true
		}
	}
	return false
}

// Discard drops every staged change.
func (c *Changeset) Discard() { c.ops = nil }

// Pending returns the number of staged cell edits, inserted rows, and deleted
// rows (each edited cell counts as one change).
func (c *Changeset) Pending() int {
	n := 0
	for _, o := range c.ops {
		if o.kind == opUpdate {
			n += len(o.edits)
		} else {
			n++
		}
	}
	return n
}

// EditedValue returns the staged value of a cell, if any.
func (c *Changeset) EditedValue(id RowID, col string) (any, bool) {
	if o := c.find(opUpdate, id.Key()); o != nil {
		for _, e := range o.edits {
			if e.col == col {
				return e.val, true
			}
		}
	}
	return nil, false
}

// IsDeleted reports whether the row is staged for deletion.
func (c *Changeset) IsDeleted(id RowID) bool { return c.find(opDelete, id.Key()) != nil }

// InsertedRow is a read-only view of a staged insert.
type InsertedRow struct {
	Handle int
	Cols   []string
	Vals   []any
}

// Inserts returns staged inserts in staging order.
func (c *Changeset) Inserts() []InsertedRow {
	var out []InsertedRow
	for _, o := range c.ops {
		if o.kind == opInsert {
			out = append(out, InsertedRow{Handle: o.insID, Cols: o.cols, Vals: o.vals})
		}
	}
	return out
}

// Statements renders every staged change as a parameterized statement in
// staging order (updates, inserts and deletes interleaved as staged).
func (c *Changeset) Statements() ([]Statement, error) {
	if c.cfg.ReadOnly {
		return nil, ErrReadOnly
	}
	if len(c.ops) == 0 {
		return nil, ErrNoChange
	}
	var out []Statement
	for _, o := range c.ops {
		switch o.kind {
		case opUpdate:
			p := newPlaceholders(c.cfg.Dialect)
			var sets []string
			var args []any
			for _, e := range o.edits {
				sets = append(sets, QuoteIdent(e.col)+" = "+p.next())
				args = append(args, e.val)
			}
			where, wargs := c.where(o.id, p)
			out = append(out, Statement{Kind: "UPDATE", SQL: "UPDATE " + c.table() + " SET " + strings.Join(sets, ", ") + " WHERE " + where, Args: append(args, wargs...)})
		case opInsert:
			if len(o.cols) == 0 {
				out = append(out, Statement{Kind: "INSERT", SQL: "INSERT INTO " + c.table() + " DEFAULT VALUES"})
				continue
			}
			p := newPlaceholders(c.cfg.Dialect)
			names := make([]string, len(o.cols))
			phs := make([]string, len(o.cols))
			for i, col := range o.cols {
				names[i] = QuoteIdent(col)
				phs[i] = p.next()
			}
			out = append(out, Statement{Kind: "INSERT", SQL: "INSERT INTO " + c.table() + " (" + strings.Join(names, ", ") + ") VALUES (" + strings.Join(phs, ", ") + ")", Args: append([]any(nil), o.vals...)})
		case opDelete:
			p := newPlaceholders(c.cfg.Dialect)
			where, args := c.where(o.id, p)
			out = append(out, Statement{Kind: "DELETE", SQL: "DELETE FROM " + c.table() + " WHERE " + where, Args: args})
		}
	}
	return out, nil
}

// Preview renders the changeset as executable SQL with literal values inlined,
// for the SQL preview modal. Never execute it; use Statements.
func (c *Changeset) Preview() (string, error) {
	sts, err := c.Statements()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, s := range sts {
		b.WriteString(Inline(s.SQL, s.Args, c.cfg.Dialect))
		b.WriteString(";\n")
	}
	return b.String(), nil
}

func (c *Changeset) table() string {
	if c.cfg.Schema != "" {
		return QuoteIdent(c.cfg.Schema) + "." + QuoteIdent(c.cfg.Table)
	}
	return QuoteIdent(c.cfg.Table)
}

func (c *Changeset) where(id RowID, p *placeholders) (string, []any) {
	var parts []string
	var args []any
	for i, col := range id.Cols {
		ph := p.next()
		if c.cfg.Dialect == Postgres && col == "ctid" {
			ph += "::tid"
		}
		parts = append(parts, QuoteIdent(col)+" = "+ph)
		args = append(args, id.Vals[i])
	}
	return strings.Join(parts, " AND "), args
}

type placeholders struct {
	d Dialect
	n int
}

func newPlaceholders(d Dialect) *placeholders { return &placeholders{d: d} }

func (p *placeholders) next() string {
	p.n++
	if p.d == Postgres {
		return fmt.Sprintf("$%d", p.n)
	}
	return "?"
}

// QuoteIdent double-quotes an identifier, doubling embedded quotes.
func QuoteIdent(s string) string { return `"` + strings.ReplaceAll(s, `"`, `""`) + `"` }

// Inline substitutes args into a parameterized statement for display only.
func Inline(sql string, args []any, d Dialect) string {
	var b strings.Builder
	n := 0
	toks := Tokenize(sql, d)
	for _, t := range toks {
		txt := sql[t.Start:t.End]
		if t.Kind == KindParam {
			if d == Postgres {
				var idx int
				if _, err := fmt.Sscanf(txt, "$%d", &idx); err == nil && idx >= 1 && idx <= len(args) {
					b.WriteString(Literal(args[idx-1], d))
					continue
				}
			} else if txt == "?" && n < len(args) {
				b.WriteString(Literal(args[n], d))
				n++
				continue
			}
		}
		b.WriteString(txt)
	}
	return b.String()
}
