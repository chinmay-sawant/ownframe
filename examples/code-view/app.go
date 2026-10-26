package main

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

const (
	defaultWidth  = 1500
	defaultHeight = 900
	minWidth      = 1180
	minHeight     = 700

	activityW    = 46
	sideWidth    = 268
	chromeTop    = 42
	chromeBottom = 24
	tabStripH    = 38
	crumbH       = 30

	chromeLeft    = activityW + sideWidth
	editorBodyTop = chromeTop + tabStripH + crumbH

	rowH     = 20
	overscan = 8

	maxTabs   = 8
	maxTree   = 26
	maxRecent = 4

	scrollStep = 96
	pageScroll = 420
	treePad    = 8
	treeIndent = 14
)

// App owns the workspace, the open documents, and the page. Every handler
// ends in one refresh; no timer and no background goroutine is involved.
type App struct {
	page *ownframe.Page
	ws   *workspace

	changes  changeSet
	statuses map[string]string
	view     View

	mode    string
	filter  string
	folders map[string]bool

	tabs   []string
	active int
	docs   map[string]*fileDoc

	diffPath  string
	diffSplit bool
	diffRows  []diffRow

	details bool
	devOn   bool

	scrollY   int
	viewportH int
	toast     toastView
}

// New scans the workspace, builds the page, and opens the README. An empty
// root uses the embedded sample project.
func New(root string) (*App, error) {
	var (
		ws  *workspace
		err error
	)

	if root == "" {
		ws, err = loadSampleWorkspace()
	} else {
		ws, err = loadDiskWorkspace(root)
	}

	if err != nil {
		return nil, err
	}

	a := &App{
		ws:        ws,
		docs:      map[string]*fileDoc{},
		folders:   map[string]bool{},
		statuses:  map[string]string{},
		mode:      "explorer",
		viewportH: defaultHeight,
		devOn:     true,
	}

	a.folders["backend"] = true
	a.folders["web"] = true

	if ws.Sample {
		a.changes = sampleChangeSet()
	} else {
		a.changes = changeSet{Note: "Open Source Control to read the working tree."}
	}

	a.statuses = statusMap(a.changes)
	a.selectFirstDiff()

	page, err := ownframe.New(ownframe.Config{
		DevTools:  true,
		Title:     "Beacon — Code View",
		HTML:      buildHTML(),
		Width:     defaultWidth,
		Height:    defaultHeight,
		MinWidth:  minWidth,
		MinHeight: minHeight,
	})
	if err != nil {
		return nil, err
	}

	a.page = page
	page.Handle(ownframe.Handlers{
		Click:   a.onClick,
		Change:  a.onChange,
		KeyDown: a.onKeyDown,
	})
	page.Scroll().SetWindowing(true)
	page.SetScrollWindow(a.onScroll)

	a.openInitial()
	a.refresh()

	return a, nil
}

// Page returns the page Run opens.
func (a *App) Page() *ownframe.Page { return a.page }

// View returns the last data snapshot. Tests use it.
func (a *App) View() View { return a.view }

// SetDevTools turns the dock on or off and repaints the status chip.
func (a *App) SetDevTools(on bool) {
	a.devOn = on
	a.page.Window().SetDevTools(on)
	a.refresh()
}

// openInitial opens the README, or the first file when there is none.
func (a *App) openInitial() {
	target := "README.md"

	if a.ws.ByPath[target] == nil {
		if len(a.ws.Files) == 0 {
			return
		}

		target = a.ws.Files[0].Path
	}

	a.docs[target] = a.ws.openDoc(target)
	a.tabs = []string{target}
	a.active = 0
}

// refresh publishes one snapshot.
func (a *App) refresh() {
	a.view = a.buildView()
	a.page.SetData(&a.view)
}

// onScroll is the windowing callback. It slices the code or diff rows around
// the viewport and reports whether the frame needs a layout.
func (a *App) onScroll(offsetY, viewH int) bool {
	if viewH <= 0 {
		viewH = a.viewportH
	}

	if offsetY == a.scrollY && viewH == a.viewportH {
		return false
	}

	a.scrollY = offsetY
	a.viewportH = viewH
	a.view = a.buildView()
	a.page.SetData(&a.view)

	return true
}

// buildView assembles one frame from the current state.
func (a *App) buildView() View {
	v := View{
		Mode:      a.mode,
		Filter:    a.filter,
		Details:   a.details,
		DevOn:     a.devOn,
		Workspace: baseName(a.ws.Root),
		Branch:    a.changes.Branch,
		Changed:   len(a.changes.Files),
		DevNote:   a.changes.Note,
		Toast:     a.toast,
	}

	v.Tree, v.TreeMore, v.TreeCount, v.Filtered = a.treeRows()
	v.Tabs = a.tabRows()
	v.Recent = a.recentRows()
	v.ActiveTab = a.active

	if doc := a.activeDoc(); doc != nil {
		v.Active = doc.Path
		v.ActiveDir = dirName(doc.Path)
		v.ActiveName = baseName(doc.Path)
		v.Lang = doc.Lang
		v.LangLabel = languageInfo(doc.Lang).Label
		v.Badge = languageInfo(doc.Lang).Badge
		v.LineCount = len(doc.Lines)
		v.ByteSize = humanSize(doc.ByteSize)
		v.Truncated = doc.Truncated
		v.Status = a.statuses[doc.Path]
		v.StatusClass = "st-" + strings.ToLower(v.Status)
		v.Outline = doc.Outline
	}

	if a.mode == "changes" {
		a.fillDiff(&v)
	} else {
		a.fillCode(&v)
	}

	v.SpaceHeight = a.contentHeight()

	return v
}

// fillCode fills the active document, its visible rows, and the empty state.
func (a *App) fillCode(v *View) {
	doc := a.activeDoc()
	if doc == nil {
		v.Welcome = true
		v.Active = ""

		return
	}

	if doc.Err != "" {
		v.Empty = true
		v.EmptyKind = "error"
		v.EmptyTitle = "Cannot open " + baseName(doc.Path)
		v.EmptyBody = doc.Err

		return
	}

	if doc.Binary {
		v.Empty = true
		v.EmptyKind = "binary"
		v.EmptyTitle = "Binary file"
		v.EmptyBody = "The viewer reads text files only. Pick a source file in the explorer."

		return
	}

	start, end, track := a.windowSlice(len(doc.Lines))
	v.TrackTop = track
	v.Rows = make([]codeRow, 0, end-start)

	for i := start; i < end; i++ {
		v.Rows = append(v.Rows, codeRow{No: i + 1, Segments: doc.Tokens[i]})
	}
}

// fillDiff fills the Source Control list and the selected diff rows.
func (a *App) fillDiff(v *View) {
	d := diffView{
		Split:    a.diffSplit,
		TotalAdd: 0,
		TotalDel: 0,
		Sample:   a.changes.Sample,
		Note:     a.changes.Note,
		HasFiles: len(a.changes.Files) > 0,
	}

	add, del := a.changes.totals()
	d.TotalAdd, d.TotalDel = add, del

	for _, file := range a.changes.Files {
		d.Files = append(d.Files, diffFileRow{
			Path:        file.Path,
			Name:        baseName(file.Path),
			Dir:         dirName(file.Path),
			Status:      file.Status,
			StatusClass: "st-" + strings.ToLower(file.Status),
			Lang:        languageOf(file.Path),
			Badge:       languageInfo(languageOf(file.Path)).Badge,
			Add:         file.Add,
			Del:         file.Del,
			Selected:    file.Path == a.diffPath,
		})
	}

	if !d.HasFiles {
		v.Diff = d

		return
	}

	file := findFile(a.changes.Files, a.diffPath)
	if file == nil {
		file = &a.changes.Files[0]
	}

	d.Selected = file.Path
	d.Path = file.Path
	d.Name = baseName(file.Path)
	d.Dir = dirName(file.Path)
	d.Lang = languageOf(file.Path)
	d.Badge = languageInfo(languageOf(file.Path)).Badge
	d.Status = file.Status
	d.StatusCls = "st-" + strings.ToLower(file.Status)
	d.Add = file.Add
	d.Del = file.Del
	d.Hunks = len(file.Hunks)

	start, end, track := a.windowSlice(len(a.diffRows))
	d.TrackTop = track

	if len(a.diffRows) > 0 {
		d.Rows = a.diffRows[start:end]
	}

	v.Diff = d
}

// windowSlice returns the row range the fixed body shows and the track offset.
func (a *App) windowSlice(total int) (int, int, int) {
	bodyH := a.viewportH - editorBodyTop - chromeBottom
	if bodyH < rowH {
		bodyH = rowH
	}

	first := a.scrollY / rowH
	start := first - overscan
	if start < 0 {
		start = 0
	}

	count := bodyH/rowH + overscan*2 + 1
	end := start + count
	if end > total {
		end = total
	}

	if start > total {
		start = total
		end = total
	}

	track := start*rowH - a.scrollY
	if maxRows := (bodyH - track) / rowH; end-start > maxRows {
		end = start + maxRows
	}

	return start, end, track
}

// contentHeight is the page height for the current mode. The fixed chrome
// covers the rest, so a short file leaves the page at one viewport.
func (a *App) contentHeight() int {
	rows := 0

	if a.mode == "changes" {
		rows = len(a.diffRows)
	} else if doc := a.activeDoc(); doc != nil {
		rows = len(doc.Lines)
	}

	h := editorBodyTop + rows*rowH + chromeBottom
	if h < a.viewportH {
		h = a.viewportH
	}

	return h
}

// clampScroll keeps the target inside the page.
func (a *App) clampScroll(y int) int {
	maxY := a.contentHeight() - a.viewportH
	if maxY < 0 {
		maxY = 0
	}

	if y < 0 {
		return 0
	}

	if y > maxY {
		return maxY
	}

	return y
}

// scrollTo moves the page and rebuilds the visible slice.
func (a *App) scrollTo(y int) {
	target := a.clampScroll(y)
	a.scrollY = target
	a.page.ScrollTo(0, target)
	a.refresh()
}

// resetScroll returns to the top when a new file or mode opens.
func (a *App) resetScroll() {
	a.scrollY = 0
	a.page.ScrollTo(0, 0)
}

// treeRows returns the explorer list, the number of hidden rows, and the
// workspace file count.
func (a *App) treeRows() ([]treeRow, int, int, bool) {
	count := len(a.ws.Files)

	if a.filter != "" {
		return a.filteredRows()
	}

	total := countTreeOpen(a.ws.Tree, a.folders)
	rows := make([]treeRow, 0, min(total, maxTree))
	a.walkTree(a.ws.Tree, 0, &rows)

	more := total - len(rows)
	if more < 0 {
		more = 0
	}

	return rows, more, count, false
}

// walkTree appends visible nodes until the row cap.
func (a *App) walkTree(nodes []*fileNode, depth int, rows *[]treeRow) {
	for _, node := range nodes {
		if len(*rows) >= maxTree {
			return
		}

		row := treeRow{
			Path:   node.Path,
			Name:   node.Name,
			Dir:    node.Dir,
			Indent: treePad + depth*treeIndent,
		}

		if node.Dir {
			row.Open = a.folders[node.Path]
		} else if node.Meta != nil {
			row.Lang = node.Meta.Lang
			row.Badge = languageInfo(node.Meta.Lang).Badge
			row.Status = a.statuses[node.Path]
			row.StatusClass = "st-" + strings.ToLower(row.Status)
		}

		*rows = append(*rows, row)

		if node.Dir && row.Open {
			a.walkTree(node.Children, depth+1, rows)
		}
	}
}

// filteredRows returns a flat path list for the filter box.
func (a *App) filteredRows() ([]treeRow, int, int, bool) {
	query := strings.ToLower(a.filter)
	rows := make([]treeRow, 0, maxTree)
	matched := 0

	for _, file := range a.ws.Files {
		if !strings.Contains(strings.ToLower(file.Path), query) {
			continue
		}

		matched++

		if len(rows) >= maxTree {
			continue
		}

		rows = append(rows, treeRow{
			Path:        file.Path,
			Name:        file.Name,
			Indent:      treePad,
			Lang:        file.Lang,
			Badge:       languageInfo(file.Lang).Badge,
			Status:      a.statuses[file.Path],
			StatusClass: "st-" + strings.ToLower(a.statuses[file.Path]),
		})
	}

	more := matched - len(rows)
	if more < 0 {
		more = 0
	}

	return rows, more, len(a.ws.Files), true
}

// countTreeOpen counts every visible node with the current fold state.
func countTreeOpen(nodes []*fileNode, open map[string]bool) int {
	total := 0

	for _, node := range nodes {
		total++

		if node.Dir && open[node.Path] {
			total += countTreeOpen(node.Children, open)
		}
	}

	return total
}

// tabRows returns the open tabs.
func (a *App) tabRows() []tabRow {
	rows := make([]tabRow, 0, len(a.tabs))

	for i, path := range a.tabs {
		doc := a.docs[path]
		lang := languageOf(path)
		if doc != nil {
			lang = doc.Lang
		}

		rows = append(rows, tabRow{
			Path:   path,
			Name:   baseName(path),
			Lang:   lang,
			Badge:  languageInfo(lang).Badge,
			Active: i == a.active && a.mode == "explorer",
		})
	}

	return rows
}

// recentRows lists the welcome shortcuts that exist in this workspace.
func (a *App) recentRows() []recentRow {
	paths := []string{"README.md", "backend/probe.go", "web/index.html", "tools/report.py", "docs/architecture.md"}
	rows := make([]recentRow, 0, maxRecent)

	for _, path := range paths {
		if a.ws.ByPath[path] == nil {
			continue
		}

		lang := languageOf(path)
		rows = append(rows, recentRow{
			Path:  path,
			Name:  baseName(path),
			Dir:   dirName(path),
			Lang:  lang,
			Badge: languageInfo(lang).Badge,
		})

		if len(rows) >= maxRecent {
			break
		}
	}

	return rows
}

// activeDoc returns the document behind the active tab.
func (a *App) activeDoc() *fileDoc {
	if a.active < 0 || a.active >= len(a.tabs) {
		return nil
	}

	return a.docs[a.tabs[a.active]]
}

// openFile opens or focuses one file and resets the scroll.
func (a *App) openFile(path string) {
	if a.ws.ByPath[path] == nil {
		a.showToast("error", "Not in workspace", path+" is not part of this workspace.")

		return
	}

	if _, ok := a.docs[path]; !ok {
		a.docs[path] = a.ws.openDoc(path)
	}

	found := -1

	for i, open := range a.tabs {
		if open == path {
			found = i

			break
		}
	}

	if found < 0 {
		a.tabs = append(a.tabs, path)
		if len(a.tabs) > maxTabs {
			a.tabs = append(a.tabs[:0], a.tabs[1:]...)
		}

		found = len(a.tabs) - 1
	}

	a.active = found
	a.mode = "explorer"
	a.resetScroll()
	a.refresh()
}

// selectTab focuses one tab.
func (a *App) selectTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}

	a.active = index
	a.mode = "explorer"
	a.resetScroll()
	a.refresh()
}

// closeTab closes one tab and picks the next one.
func (a *App) closeTab(index int) {
	if index < 0 || index >= len(a.tabs) {
		return
	}

	a.tabs = append(a.tabs[:index], a.tabs[index+1:]...)

	if index < a.active {
		a.active--
	}

	if a.active >= len(a.tabs) {
		a.active = len(a.tabs) - 1
	}

	if a.active < 0 {
		a.active = 0
	}

	a.resetScroll()
	a.refresh()
}

// toggleFolder flips one directory.
func (a *App) toggleFolder(path string) {
	a.folders[path] = !a.folders[path]
	a.refresh()
}

// collapseAll closes every folder.
func (a *App) collapseAll() {
	for path := range a.folders {
		delete(a.folders, path)
	}

	a.refresh()
}

// setMode switches the activity bar destination.
func (a *App) setMode(ctx context.Context, mode string) {
	if mode != "explorer" && mode != "changes" {
		return
	}

	a.mode = mode

	if mode == "changes" && !a.ws.Sample {
		a.reloadChanges(ctx, false)
	}

	a.resetScroll()
	a.refresh()
}

// reloadChanges reads live git when -root is set, or the bundled fixture.
func (a *App) reloadChanges(ctx context.Context, announce bool) {
	if a.ws.Sample {
		a.changes = sampleChangeSet()
	} else {
		next, err := loadLiveChanges(ctx, a.ws.Root)
		if err != nil {
			a.showToast("error", "Git unavailable", err.Error())
			a.refresh()

			return
		}

		a.changes = next
	}

	a.statuses = statusMap(a.changes)
	a.selectFirstDiff()
	a.rebuildDiffRows()
	a.resetScroll()

	if announce {
		if len(a.changes.Files) == 0 {
			a.showToast("info", "Working tree clean", "No tracked or untracked changes were found.")
		} else {
			a.showToast("success", "Changes refreshed", "Read the working tree from git.")
		}
	}

	a.refresh()
}

// selectFirstDiff keeps the current file when it still exists.
func (a *App) selectFirstDiff() {
	if len(a.changes.Files) == 0 {
		a.diffPath = ""
		a.diffRows = nil

		return
	}

	if findFile(a.changes.Files, a.diffPath) == nil {
		a.diffPath = a.changes.Files[0].Path
	}

	a.rebuildDiffRows()
}

// rebuildDiffRows renders every row for the selected file and split mode.
func (a *App) rebuildDiffRows() {
	a.diffRows = nil

	file := findFile(a.changes.Files, a.diffPath)
	if file == nil {
		return
	}

	lang := languageOf(file.Path)
	if a.diffSplit {
		a.diffRows = diffRowsSplit(*file, lang)
	} else {
		a.diffRows = diffRowsUnified(*file, lang)
	}
}

// selectDiff shows one file's diff.
func (a *App) selectDiff(path string) {
	if findFile(a.changes.Files, path) == nil {
		return
	}

	a.diffPath = path
	a.mode = "changes"
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
}

// setSplit switches between unified and split diff rows.
func (a *App) setSplit(split bool) {
	a.diffSplit = split
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
}

// jumpHunk moves to the previous or next hunk header.
func (a *App) jumpHunk(delta int) {
	if len(a.diffRows) == 0 {
		return
	}

	current := a.scrollY / rowH
	target := -1

	for i, row := range a.diffRows {
		if row.Kind != rowHunk {
			continue
		}

		if delta > 0 && i > current {
			target = i

			break
		}

		if delta < 0 && i < current {
			target = i
		}
	}

	if target < 0 {
		a.showToast("info", "No more hunks", "The selection is already at the edge of this diff.")
		a.refresh()

		return
	}

	a.scrollTo(target * rowH)
}

// showToast publishes one message card.
func (a *App) showToast(kind, title, message string) {
	a.toast = toastView{Show: true, Kind: kind, Title: title, Message: message}
}

// clearToast hides the message card.
func (a *App) clearToast() {
	a.toast = toastView{}
	a.refresh()
}

// toggleDetails opens or closes the details drawer.
func (a *App) toggleDetails() {
	a.details = !a.details
	a.refresh()
}

// statusMap turns the change list into the badges the explorer shows.
func statusMap(cs changeSet) map[string]string {
	out := make(map[string]string, len(cs.Files))

	for _, file := range cs.Files {
		out[file.Path] = file.Status
	}

	return out
}
