package ui

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

const queryTimeout = 5 * time.Minute

// runState tracks the one in-flight run so Cancel and the timer can reach it.
type runState struct {
	cancel  context.CancelFunc
	started time.Time
	shown   time.Time
	tabID   int
}

// stmtResult is the outcome of one statement of a script.
type stmtResult struct {
	SQL      string
	Res      driver.QueryResult
	Affected int64
	IsRead   bool
	Dur      time.Duration
	Err      error
}

// queryDone is posted from a worker goroutine and drained in onTick.
type queryDone struct {
	tabID    int
	kind     string // run | explain | table
	results  []stmtResult
	plan     *driver.ExplainPlan
	err      error
	sql      string // statement text, for error hints and history
	total    int64  // table row count, for table pages
	canceled bool
}

type lastQueryInfo struct {
	Dur  time.Duration
	Rows int
	Set  bool
	Err  bool
}

func (a *App) rowCap(t *workTab) int {
	if t.rowLimit <= 0 {
		return driver.DefaultMaxRows
	}
	return t.rowLimit
}

// runFromKey runs the statement at the caret (or the selection), or the whole
// script when all is true.
func (a *App) runFromKey(ctx context.Context, all bool) error {
	a.ac.open = false
	t := a.activeTabPtr()
	if t == nil || t.kind != tabQuery || t.ed == nil {
		return nil
	}
	if a.running != nil {
		return nil
	}
	ed := t.ed
	d := a.dialect()
	if all {
		stmts := sqlkit.Script(ed.text, d)
		if len(stmts) == 0 {
			return a.redraw(ctx)
		}
		return a.startRun(ctx, t, stmts)
	}
	lo, hi := ed.selection()
	tg, ok := sqlkit.RunTarget(ed.text, ed.caret, lo, hi, d)
	if !ok {
		return a.redraw(ctx)
	}
	t.runBase = tg.Span.Start
	return a.startRun(ctx, t, []string{tg.SQL})
}

// runSelection runs only the selected text.
func (a *App) runSelection(ctx context.Context) error {
	t := a.activeTabPtr()
	if t == nil || t.ed == nil || !t.ed.hasSelection() {
		a.notify("Select SQL to run")
		return a.redraw(ctx)
	}
	return a.runFromKey(ctx, false)
}

var destructiveRE = regexp.MustCompile(`(?is)^\s*(?:DROP\s+TABLE(?:\s+IF\s+EXISTS)?|TRUNCATE(?:\s+TABLE)?)\s+(?:ONLY\s+)?(?:"?[\w]+"?\.)?"?(\w+)"?`)

// destructiveTable returns the table a DROP TABLE or TRUNCATE statement hits.
func destructiveTable(stmts []string) (string, bool) {
	for _, s := range stmts {
		if m := destructiveRE.FindStringSubmatch(s); m != nil {
			return m[1], true
		}
	}
	return "", false
}

func (a *App) startRun(ctx context.Context, t *workTab, stmts []string) error {
	a.ac.open = false
	if name, ok := destructiveTable(stmts); ok && !a.cfg.ReadOnly && (a.confirmDestructive || a.cfg.EnvTag == driver.EnvProd) {
		return a.askConfirm(ctx, &confirmState{
			Title: "Destructive statement", Typed: name, Label: "Run anyway",
			Body: "This statement drops or empties table " + name + ". Type the table name to continue.",
			OK:   func(ctx context.Context) error { return a.doRun(ctx, t, stmts) },
		})
	}
	return a.doRun(ctx, t, stmts)
}

func (a *App) doRun(ctx context.Context, t *workTab, stmts []string) error {
	if a.drv == nil {
		t.err = &queryErr{Message: "No active connection"}
		t.resultView = "results"
		return a.redraw(ctx)
	}
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), queryTimeout)
	a.running = &runState{cancel: cancel, started: a.clock(), shown: a.clock(), tabID: t.id}
	t.err = nil
	t.resultView = "results"
	drv, limit, tabID := a.drv, a.rowCap(t), t.id
	go func() {
		out := queryDone{tabID: tabID, kind: "run"}
		for _, s := range stmts {
			r := execOne(rctx, drv, s, limit)
			out.results = append(out.results, r)
			out.sql = s
			if r.Err != nil {
				out.err = r.Err
				break
			}
		}
		out.canceled = rctx.Err() != nil
		cancel()
		a.doneCh <- out
	}()
	return a.redraw(ctx)
}

func execOne(ctx context.Context, drv driver.Driver, sql string, limit int) stmtResult {
	start := time.Now()
	r := stmtResult{SQL: sql}
	if driver.IsReadStatement(sql) {
		r.IsRead = true
		r.Res, r.Err = drv.Query(ctx, sql, limit)
	} else {
		r.Affected, r.Err = drv.Exec(ctx, sql)
	}
	r.Dur = time.Since(start)
	return r
}

// startExplain requests a plan for the statement at the caret.
func (a *App) startExplain(ctx context.Context, analyze bool) error {
	t := a.activeTabPtr()
	if t == nil || t.kind != tabQuery || a.running != nil || a.drv == nil {
		return nil
	}
	lo, hi := t.ed.selection()
	tg, ok := sqlkit.RunTarget(t.ed.text, t.ed.caret, lo, hi, a.dialect())
	if !ok {
		return a.redraw(ctx)
	}
	rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), queryTimeout)
	a.running = &runState{cancel: cancel, started: a.clock(), shown: a.clock(), tabID: t.id}
	drv, tabID := a.drv, t.id
	go func() {
		plan, err := drv.Explain(rctx, tg.SQL, analyze)
		cancel()
		a.doneCh <- queryDone{tabID: tabID, kind: "explain", plan: &plan, err: err, sql: tg.SQL}
	}()
	return a.redraw(ctx)
}

func (a *App) cancelRunning() {
	if a.running != nil {
		a.running.cancel()
	}
}

func (a *App) tabByID(id int) *workTab {
	for _, t := range a.tabs {
		if t.id == id {
			return t
		}
	}
	return nil
}

// applyQueryDone folds a finished run into tab, status bar, history, and slow log.
func (a *App) applyQueryDone(r queryDone) {
	a.running = nil
	t := a.tabByID(r.tabID)
	switch r.kind {
	case "explain":
		if t == nil {
			return
		}
		if r.err != nil {
			t.err = a.describeErr(r.err, r.sql, t)
			t.resultView = "results"
			return
		}
		t.plan, t.err = r.plan, nil
		t.resultView = "explain"
		return
	case "table":
		a.applyTablePage(t, r)
		return
	}
	var total time.Duration
	var rows int
	var msgs []string
	for i, s := range r.results {
		total += s.Dur
		switch {
		case s.Err != nil:
			msgs = append(msgs, fmt.Sprintf("Statement %d: error: %v", i+1, s.Err))
		case s.IsRead:
			n := len(s.Res.Rows)
			rows = n
			msgs = append(msgs, fmt.Sprintf("Statement %d: %d rows in %s", i+1, n, fmtDur(s.Dur)))
		default:
			msgs = append(msgs, fmt.Sprintf("Statement %d: %d rows affected in %s", i+1, s.Affected, fmtDur(s.Dur)))
		}
		a.recordHistory(s, r.err != nil && i == len(r.results)-1)
		a.captureSlow(s)
	}
	a.lastQuery = lastQueryInfo{Dur: total, Rows: rows, Set: true, Err: r.err != nil}
	if t == nil {
		return
	}
	t.messages = msgs
	t.execMs = float64(total.Microseconds()) / msPerSecond
	if r.err != nil {
		if r.canceled {
			t.err = &queryErr{Message: "Query canceled"}
		} else {
			t.err = a.describeErr(r.err, r.sql, t)
		}
		return
	}
	t.err = nil
	t.rowErrLine = -1
	for i := len(r.results) - 1; i >= 0; i-- {
		if r.results[i].IsRead {
			t.grid = newGridState(r.results[i].Res)
			t.sortCol, t.sortDir, t.quick, t.colf = 0, 0, "", nil
			t.grid.colf = nil
			t.cs = nil
			t.editing = nil
			t.resultView = "results"
			return
		}
	}
	t.grid = nil
	t.resultView = "messages"
}

var nearRE = regexp.MustCompile(`near "([^"]+)"`)

// describeErr builds the structured banner: SQLSTATE or code, message, and a
// line/column hint located in the editor.
func (a *App) describeErr(err error, stmt string, t *workTab) *queryErr {
	q := &queryErr{Message: err.Error()}
	var st interface{ SQLState() string }
	if errors.As(err, &st) {
		q.Code = st.SQLState()
	}
	var codeErr interface{ Code() int }
	if q.Code == "" && errors.As(err, &codeErr) {
		q.Code = fmt.Sprintf("SQLITE %d", codeErr.Code())
	}
	if m := nearRE.FindStringSubmatch(q.Message); m != nil && stmt != "" {
		if i := strings.Index(stmt, m[1]); i >= 0 {
			src := stmt
			base := 0
			if t != nil && t.ed != nil {
				if j := strings.Index(t.ed.text, stmt); j >= 0 {
					src, base = t.ed.text, j
				}
			}
			q.Line, q.Col = sqlkit.LineCol(src, base+i)
		}
	}
	if q.Line == 0 {
		q.Line = 1
	}
	if t != nil {
		t.rowErrLine = q.Line - 1
	}
	return q
}

func (a *App) jumpToError(ctx context.Context) error {
	t := a.activeTabPtr()
	if t == nil || t.err == nil || t.ed == nil {
		return nil
	}
	off := t.ed.offsetAt(t.err.Line-1, max(0, t.err.Col-1))
	t.ed.caret, t.ed.anchor = off, off
	t.ed.ensureVisible(a.editorRows())
	a.focus = "editor"
	return a.redraw(ctx)
}
