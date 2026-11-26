package ui

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/ownframe/ui"

	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/driver"
	"github.com/chinmay-sawant/ownframe/examples/builder/quarry/sqlkit"
)

type tabKind int

// Workspace tab kinds.
const (
	tabQuery tabKind = iota
	tabTable
	tabStructure
	tabMonitor
)

func (k tabKind) label() string {
	return [...]string{"Query", "Table", "Structure", "Monitor"}[k]
}

// queryErr is the structured error shown in the results pane.
type queryErr struct {
	Code    string
	Message string
	Line    int // 1-based line in the statement/editor, 0 when unknown
	Col     int
}

// workTab is one workspace tab of any kind.
type workTab struct {
	id    int
	kind  tabKind
	title string
	dirty bool

	// Query tabs.
	ed         *editor
	rowLimit   int
	resultView string // results | explain | messages
	plan       *driver.ExplainPlan
	rawPlan    bool
	messages   []string
	err        *queryErr
	rowErrLine int
	execMs     float64
	runBase    int // editor offset the last run statement began at

	// Table and query data grid.
	schema, table string
	grid          *gridState
	filter        string
	sortCol       int
	sortDir       int            // 0 none, 1 asc, 2 desc
	freeze        bool           // keep the first column fixed while scrolling sideways
	quick         string         // client-side filter of a query result
	colf          map[int]string // per-column filters of a query result, by column index
	showFilters   bool           // the filter row under the header is open
	page          int
	pageSize      int
	totalRows     int64
	cs            *sqlkit.Changeset
	writable      bool
	pk            []string
	csErr         string

	inspectorOpen bool
	editing       *cellEditing

	// Structure tabs.
	structSub string
	cols      []driver.ColumnMeta
	indexes   []driver.IndexMeta
	fks       []driver.ForeignKeyMeta
	ddl       string
}

func (a *App) activeTabPtr() *workTab {
	if a.activeTab < 0 || a.activeTab >= len(a.tabs) {
		return nil
	}
	return a.tabs[a.activeTab]
}

func (a *App) addTab(t *workTab) *workTab {
	a.nextTabID++
	t.id = a.nextTabID
	a.tabs = append(a.tabs, t)
	a.activeTab = len(a.tabs) - 1
	a.ac.open = false
	return t
}

// newQueryTab opens a SQL editor tab seeded with sql.
func (a *App) newQueryTab(sql string) *workTab {
	n := 1
	for _, t := range a.tabs {
		if t.kind == tabQuery {
			n++
		}
	}
	t := a.addTab(&workTab{kind: tabQuery, title: fmt.Sprintf("Query %d", n), ed: newEditor(sql), rowLimit: a.defLimit, resultView: "results", rowErrLine: -1})
	a.focus = "editor"
	return t
}

func (a *App) findTab(kind tabKind, schema, table string) int {
	for i, t := range a.tabs {
		if t.kind == kind && t.schema == schema && t.table == table {
			return i
		}
	}
	return -1
}

func (a *App) openMonitorTab() {
	if i := a.findTab(tabMonitor, "", ""); i >= 0 {
		a.activeTab = i
		return
	}
	a.addTab(&workTab{kind: tabMonitor, title: "Monitor"})
}

func (a *App) closeTab(ctx context.Context, i int) error {
	if i < 0 || i >= len(a.tabs) {
		return nil
	}
	if a.tabs[i].kind == tabMonitor {
		a.mon.stopOnClose()
	}
	a.tabs = append(a.tabs[:i], a.tabs[i+1:]...)
	switch {
	case len(a.tabs) == 0:
		a.newQueryTab("")
	case a.activeTab >= len(a.tabs):
		a.activeTab = len(a.tabs) - 1
	case i < a.activeTab:
		a.activeTab--
	}
	a.ac.open = false
	return a.redraw(ctx)
}

func (a *App) cycleTab(ctx context.Context, dir int) error {
	if len(a.tabs) == 0 {
		return nil
	}
	a.activeTab = (a.activeTab + dir + len(a.tabs)) % len(a.tabs)
	a.ac.open = false
	return a.redraw(ctx)
}

func (a *App) selectTab(ctx context.Context, i int) error {
	a.activeTab = i
	a.ac.open = false
	return a.redraw(ctx)
}

// tabStrip renders the 34 px tab bar.
func (a *App) tabStrip() *ui.Element {
	bar := ui.Div().ID("tab-strip").Class("qr-tabs").Attr("role", ariaRole("tablist")).Attr("aria-label", "Workspace tabs")
	for i, t := range a.tabs {
		i, t := i, t
		cls := "qr-tab"
		if i == a.activeTab {
			cls += " qr-tab-active"
		}
		tab := ui.Div().ID(fmt.Sprintf("tab-%d", t.id)).Class(cls).Attr("role", ariaRole("tab")).
			Attr("aria-label", t.kind.label()+" tab "+t.title).
			Attr("aria-selected", fmt.Sprint(i == a.activeTab)).
			Children(a.icon(tabIcon(t.kind), toneMuted, iconSmall))
		tab.Children(ui.Span().Text(t.title))
		if t.dirty {
			tab.Children(ui.Span().ID(fmt.Sprintf("tab-dirty-%d", t.id)).Class("qr-tab-dirty").Text("*"))
		}
		tab.Children(ui.Span().ID(fmt.Sprintf("tab-close-%d", t.id)).Class("qr-tab-close").
			Attr("role", "button").Attr("aria-label", "Close "+t.title).Children(a.icon("x", toneMuted, iconSmall)).
			OnClick(func(ctx context.Context) error { return a.closeTab(ctx, a.indexOfTab(t)) }))
		tab.OnClick(func(ctx context.Context) error { return a.selectTab(ctx, a.indexOfTab(t)) })
		bar.Children(tab)
	}
	bar.Children(ui.Div().ID("tab-new").Class("qr-tab-new").Attr("role", "button").Attr("aria-label", "New query tab").
		Children(a.icon("plus", toneBrand, iconSmall), ui.Span().Text("New Query")).OnClick(func(ctx context.Context) error {
		a.newQueryTab("")
		return a.redraw(ctx)
	}))
	return bar
}

func (a *App) indexOfTab(t *workTab) int {
	for i, x := range a.tabs {
		if x == t {
			return i
		}
	}
	return -1
}

// workspace renders the active tab's content.
func (a *App) workspace() *ui.Element {
	ws := ui.Div().ID("workspace").Class("qr-workspace")
	if banner := a.prodBanner(); banner != nil {
		ws.Children(banner)
	}
	ws.Children(a.tabStrip())
	t := a.activeTabPtr()
	if t == nil {
		return ws
	}
	switch t.kind {
	case tabQuery:
		ws.Children(a.queryTabView(t))
	case tabTable:
		ws.Children(a.tableTabView(t))
	case tabStructure:
		ws.Children(a.structureView(t))
	case tabMonitor:
		ws.Children(a.monitorView(t))
	}
	return ws
}

// queryTabView is the editor above the results pane.
func (a *App) queryTabView(t *workTab) *ui.Element {
	return ui.Div().ID("query-tab").Class("qr-tab-body").Children(
		a.editorPane(t),
		ui.Div().ID("split-editor").Class("qr-drag-handle-h").Attr("aria-label", "Resize editor"),
		a.resultsPane(t),
	)
}

func tabIcon(k tabKind) string {
	switch k {
	case tabTable:
		return "table"
	case tabStructure:
		return "list"
	case tabMonitor:
		return "activity"
	}
	return "file-code"
}
