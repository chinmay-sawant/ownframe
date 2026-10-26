package main

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

const (
	defaultWidth  = 1500
	defaultHeight = 900
	minWidth      = 840
	minHeight     = 700

	activityW    = 46
	sideWidth    = 268
	detailsWidth = 280
	chromeTop    = 42
	chromeBottom = 24
	tabStripH    = 38
	crumbH       = 30

	chromeLeft    = activityW + sideWidth
	editorBodyTop = chromeTop + tabStripH + crumbH

	rowH = 20

	codeGutter   = 54
	codePadRight = 22
	codeCellW    = 7.33 // monospace advance at the 12.2 px editor font
	diffCellW    = 7.2  // monospace advance at the 12 px diff font
	diffUnifiedW = 46 + 46 + 20 + 18
	diffSplitW   = 44 + 12
	wrapMinCols  = 8

	histHeadH    = 32
	histRowH     = 34
	histFileH    = 24
	histPreviewH = 18
	histSplitH   = 7
	histMinH     = 168
	diffMinH     = 240

	historyMinRatio     = 0.18
	historyMaxRatio     = 0.82
	historyDefaultRatio = 0.5
	historyExpandedRows = 3
	historyMaxRows      = 6
	historyPad          = 8
	historyTipBody      = 280
	historyStep         = 0.03
	editorBorderPad     = 2
	actionPair          = 2
	halfDivisor         = 2

	maxTabs      = 8
	maxTree      = 26
	diffPageRows = 13
	maxRecent    = 4

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

	diffPath   string
	diffSplit  bool
	diffRows   []diffRow
	changePage int
	diffSource string

	commits       commitLog
	commitDetails map[string]*changedFile
	commitKeys    []string
	commitOpen    string
	commitFile    string
	commitRows    []diffRow
	commitPreview []diffRow
	commitPage    int

	historyRatio float64
	historyStart float64
	historyDrag  bool
	historyDy    float64
	hoverCommit  string

	details bool
	devOn   bool

	scrollY       int
	viewportH     int
	scrollPending bool
	scrollTarget  int
	lastScroll    int
	toast         toastView

	paletteOpen  bool
	recent       []string
	paletteQuery string
	paletteIndex int
	ctrl         bool
	shift        bool
	preview      bool
	wrap         bool
	viewportCols int
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
		ws:            ws,
		docs:          map[string]*fileDoc{},
		folders:       map[string]bool{},
		statuses:      map[string]string{},
		commitDetails: map[string]*changedFile{},
		mode:          "explorer",
		diffSource:    "working",
		historyRatio:  historyDefaultRatio,
		viewportH:     defaultHeight,
		devOn:         true,
		wrap:          true,
	}

	if ws.Sample {
		a.folders["backend"] = true
		a.folders["web"] = true
	}

	if ws.Sample {
		a.changes = sampleChangeSet()
	} else {
		a.changes = changeSet{Note: "Open Source Control to read the working tree."}
	}

	a.statuses = statusMap(a.changes)
	a.selectFirstDiff()

	page, err := ownframe.New(ownframe.Config{
		DevTools:  true,
		Title:     "Code Review",
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
		Click:     a.onClick,
		Change:    a.onChange,
		KeyDown:   a.onKeyDown,
		KeyUp:     a.onKeyUp,
		Submit:    a.onSubmit,
		Hover:     a.onHover,
		DragStart: a.onDragStart,
		DragMove:  a.onDragMove,
		DragEnd:   a.onDragEnd,
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
	a.addRecent(target)
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

	if a.scrollPending {
		switch offsetY {
		case a.scrollTarget:
			a.scrollPending = false
		case a.lastScroll:
			return false
		default:
			a.scrollPending = false
		}
	}

	a.lastScroll = offsetY

	cols := 0
	if a.wrap {
		cols = a.codeCols()
	}

	if offsetY == a.scrollY && viewH == a.viewportH && cols == a.viewportCols {
		return false
	}

	a.scrollY = offsetY
	a.viewportH = viewH
	a.viewportCols = cols
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
		Wrap:      a.wrap,
	}

	v.Tree, v.TreeMore, v.TreeCount, v.Filtered = a.treeRows()
	v.Tabs = a.tabRows()
	v.Recent = a.recentRows()
	v.ActiveTab = a.active
	v.PaletteQuery = a.paletteQuery
	v.Palette = a.paletteView()

	if doc := a.activeDoc(); doc != nil {
		v.Active = doc.Path
		v.ActiveDir = dirName(doc.Path)
		v.ActiveName = baseName(doc.Path)
		v.Lang = doc.Lang
		v.LangLabel = languageInfo(doc.Lang).Label
		v.Badge = languageInfo(doc.Lang).Badge
		v.LineCount = len(doc.Lines)
		v.ByteSize = humanSize(doc.ByteSize)
		v.Encoding = doc.Encoding
		v.Indent = doc.Indent
		v.Truncated = doc.Truncated
		v.Status = a.statuses[doc.Path]
		v.StatusClass = "st-" + strings.ToLower(v.Status)
		v.Outline = doc.Outline
	}

	if a.mode == "changes" {
		a.fillDiff(&v)
		a.fillHistory(&v)
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

	v.Previewable = doc.Lang == langMarkdown
	v.Preview = a.preview && v.Previewable

	if v.Preview {
		doc.ensurePreview()
		start, end, track := a.windowSlice(len(doc.Preview))
		v.TrackTop = track

		if len(doc.Preview) > 0 {
			v.PreviewRows = doc.Preview[start:end]
		}

		return
	}

	start, end, track := a.windowSlice(len(doc.Lines))
	if a.wrap {
		plan := doc.wrapped(a.codeCols())
		start, end, track = a.windowSlice(len(plan.Rows))
		v.TrackTop = track
		v.Rows = plan.Rows[start:end]

		return
	}

	v.TrackTop = track
	v.Rows = make([]codeRow, 0, end-start)

	for i := start; i < end; i++ {
		v.Rows = append(v.Rows, codeRow{No: i + 1, Line: i, Segments: doc.Tokens[i]})
	}
}

// diffColumns is the text cell count for the unified or split diff.
func (a *App) diffColumns() (int, int) {
	w := a.editorWidth()

	unified := int(float64(w-diffUnifiedW) / diffCellW)
	split := int(float64(w/2-diffSplitW) / diffCellW)

	if unified < wrapMinCols {
		unified = wrapMinCols
	}

	if split < wrapMinCols {
		split = wrapMinCols
	}

	return unified, split
}

// currentDiffRows returns the rows for the working tree or the selected commit.
func (a *App) currentDiffRows() []diffRow {
	if a.diffSource == "commit" {
		return a.commitRows
	}

	return a.diffRows
}

// fillDiff fills the Source Control list and the selected diff rows.
func (a *App) fillDiff(v *View) {
	d := diffView{
		Split:      a.diffSplit,
		Sample:     a.changes.Sample,
		Note:       a.changes.Note,
		EmptyTitle: a.changes.EmptyTitle,
		HasChanges: len(a.changes.Files) > 0,
	}

	add, del := a.changes.totals()
	d.TotalAdd, d.TotalDel = add, del

	pages := 0
	if len(a.changes.Files) > 0 {
		pages = (len(a.changes.Files) + diffPageRows - 1) / diffPageRows
	}

	page := a.changePage
	if page >= pages {
		page = pages - 1
	}

	if page < 0 {
		page = 0
	}

	d.Page, d.Pages = page+1, pages
	from := page * diffPageRows
	to := min(from+diffPageRows, len(a.changes.Files))

	for _, file := range a.changes.Files[from:to] {
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
			Selected:    file.Path == a.diffPath && a.diffSource == "working",
		})
	}

	if a.diffSource == "commit" {
		a.fillCommitDiff(v, &d)
		v.Diff = d

		return
	}

	d.Source = "working"
	d.HasFiles = len(a.changes.Files) > 0

	if !d.HasFiles {
		v.Diff = d

		return
	}

	file := findFile(a.changes.Files, a.diffPath)
	if file == nil {
		file = &a.changes.Files[0]
	}

	a.fillDiffFile(v, &d, file, a.diffRows)
	d.Selected = file.Path
	v.Diff = d
}

// fillCommitDiff publishes the selected file from one commit.
func (a *App) fillCommitDiff(v *View, d *diffView) {
	commit := a.findCommit(a.commitOpen)
	d.Source = "commit"
	d.HasFiles = false

	if commit == nil {
		d.Note = "The selected commit is no longer in history."

		return
	}

	d.CommitHash = commit.Hash
	d.CommitShort = commit.Short
	d.CommitSubject = commit.Subject
	d.CommitAuthor = commit.Author
	d.CommitDate = commit.Date

	path := a.commitFile
	if path == "" && len(commit.Files) > 0 {
		path = commit.Files[0].Path
	}

	file := a.commitDetails[commitDetailKey(commit.Hash, path)]
	if file == nil {
		d.Note = "Choose a file in the commit to read its patch."
		d.EmptyTitle = "Commit diff"

		return
	}

	d.HasFiles = true
	d.Selected = file.Path
	a.fillDiffFile(v, d, file, a.commitRows)
}

// fillDiffFile fills the diff header, document metadata, and visible rows.
func (a *App) fillDiffFile(v *View, d *diffView, file *changedFile, rows []diffRow) {
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

	doc := a.docs[file.Path]
	if doc == nil {
		doc = a.ws.openDoc(file.Path)
		a.docs[file.Path] = doc
	}

	if doc != nil {
		v.Active = doc.Path
		v.ActiveDir = dirName(doc.Path)
		v.ActiveName = baseName(doc.Path)
		v.Lang = doc.Lang
		v.LangLabel = languageInfo(doc.Lang).Label
		v.Badge = languageInfo(doc.Lang).Badge
		v.LineCount = len(doc.Lines)
		v.ByteSize = humanSize(doc.ByteSize)
		v.Encoding = doc.Encoding
		v.Indent = doc.Indent
		v.Truncated = doc.Truncated
		v.Outline = doc.Outline

		if d.Source == "commit" {
			v.Status = file.Status
		} else {
			v.Status = a.statuses[doc.Path]
		}

		v.StatusClass = "st-" + strings.ToLower(v.Status)
	}

	start, end, track := a.windowSlice(len(rows))
	d.TrackTop = track

	if len(rows) > 0 {
		visible := append([]diffRow(nil), rows[start:end]...)
		unified, split := a.diffColumns()
		clipDiffRows(visible, unified, split)
		d.Rows = visible
	}
}

// findCommit returns one commit by full hash.
func (a *App) findCommit(hash string) *gitCommit {
	for i := range a.commits.Commits {
		if a.commits.Commits[i].Hash == hash {
			return &a.commits.Commits[i]
		}
	}

	return nil
}

// bodyHeight is the editor body below the tab strip and breadcrumbs.
func (a *App) bodyHeight() int {
	h := a.viewportH - editorBodyTop - chromeBottom
	if h < rowH {
		h = rowH
	}

	return h
}

// historyVisible reports whether the commit rail shares the editor.
func (a *App) historyVisible() bool {
	return a.mode == "changes" && !a.ws.Sample && !a.commits.Unavailable && a.commits.Branch != ""
}

// historyHeight is the clamped top pane height.
func (a *App) historyHeight() int {
	if !a.historyVisible() {
		return 0
	}

	body := a.bodyHeight() - histSplitH
	maxTop := body - diffMinH

	h := int(float64(body) * a.historyRatio)
	if h < histMinH {
		h = histMinH
	}

	if maxTop < histMinH {
		return body / halfDivisor
	}

	if h > maxTop {
		h = maxTop
	}

	return h
}

// diffBodyHeight is the remaining space for diff rows.
func (a *App) diffBodyHeight() int {
	split := 0
	if a.historyVisible() {
		split = histSplitH
	}

	h := a.bodyHeight() - a.historyHeight() - split
	if h < rowH {
		h = rowH
	}

	return h
}

// commitPageSize is the number of commit rows that fit on one history page.
func (a *App) commitPageSize() int {
	avail := a.historyHeight() - histHeadH - historyPad
	if a.commitOpen != "" {
		return historyExpandedRows
	}

	size := avail / histRowH
	if size < 1 {
		size = 1
	}

	if size > historyMaxRows {
		size = 6
	}

	return size
}

// fillHistory builds the commit rail and its expanded preview.
func (a *App) fillHistory(v *View) {
	h := historyView{Visible: a.historyVisible()}
	if !h.Visible {
		v.History = h

		return
	}

	h.Branch = a.commits.Branch
	h.Note = a.commits.Note
	h.Unborn = a.commits.Unborn
	h.Height = a.historyHeight()

	size := a.commitPageSize()
	h.Pages = (len(a.commits.Commits) + size - 1) / size
	if h.Pages < 1 {
		h.Pages = 1
	}

	if a.commitPage >= h.Pages {
		a.commitPage = h.Pages - 1
	}

	if a.commitPage < 0 {
		a.commitPage = 0
	}

	h.Page = a.commitPage + 1
	from := a.commitPage * size
	to := min(from+size, len(a.commits.Commits))

	for _, commit := range a.commits.Commits[from:to] {
		row := commitRow{
			Hash:     commit.Hash,
			Short:    commit.Short,
			Subject:  commit.Subject,
			Author:   commit.Author,
			Date:     commit.Date,
			Body:     commit.Body,
			Add:      commit.Add,
			Del:      commit.Del,
			More:     commit.More,
			Head:     commit.Hash == a.commits.Head,
			Expanded: commit.Hash == a.commitOpen,
			Selected: commit.Hash == a.commitOpen && a.diffSource == "commit",
		}

		if row.Expanded {
			left := a.historyHeight() - histHeadH - historyPad - size*histRowH
			fileRows := left / histFileH
			if fileRows < 0 {
				fileRows = 0
			}

			shown := min(len(commit.Files), fileRows)
			row.More += len(commit.Files) - shown
			left -= shown * histFileH

			for _, file := range commit.Files[:shown] {
				lang := languageOf(file.Path)
				row.Files = append(row.Files, commitFileRow{
					Path:        file.Path,
					Name:        baseName(file.Path),
					Dir:         dirName(file.Path),
					Status:      file.Status,
					StatusClass: "st-" + strings.ToLower(file.Status),
					Lang:        lang,
					Badge:       languageInfo(lang).Badge,
					Add:         file.Add,
					Del:         file.Del,
					Selected:    file.Path == a.commitFile && a.diffSource == "commit",
				})
			}

			if a.commitFile != "" && a.diffSource == "commit" {
				previewRows := left / histPreviewH
				if previewRows < 0 {
					previewRows = 0
				}

				if previewRows > len(a.commitPreview) {
					previewRows = len(a.commitPreview)
				}

				row.PreviewPath = a.commitFile

				if previewRows > 0 {
					preview := append([]diffRow(nil), a.commitPreview[:previewRows]...)
					unified, _ := a.diffColumns()
					clipDiffRows(preview, unified, unified)
					row.Preview = preview
				}
			}
		}

		h.Commits = append(h.Commits, row)
	}

	if a.hoverCommit != "" {
		for _, commit := range a.commits.Commits {
			if commit.Short != a.hoverCommit {
				continue
			}

			body := strings.Join(strings.Fields(commit.Body), " ")
			if runes := []rune(body); len(runes) > historyTipBody {
				body = string(runes[:historyTipBody]) + "…"
			}

			h.Tip = commitTipView{
				Show:    true,
				Subject: commit.Subject,
				Body:    body,
				Meta:    commit.Short + " · " + commit.Author + " · " + commit.Date,
			}
		}
	}

	v.History = h
}

// windowSlice returns the row range the fixed body shows and the track offset.
func (a *App) windowSlice(total int) (int, int, int) {
	bodyH := a.diffBodyHeight()
	if bodyH < rowH {
		bodyH = rowH
	}

	first := a.scrollY / rowH
	start := first
	if start > total {
		start = total
	}

	end := start + bodyH/rowH + 1
	if end > total {
		end = total
	}

	// Rows start at the top of the body. A negative overscan row would paint
	// and hit-test over the tab strip and breadcrumbs because hit testing
	// does not apply CSS overflow clips.
	return start, end, 0
}

// contentHeight is the page height for the current mode. The fixed chrome
// covers the rest, so a short file leaves the page at one viewport.
func (a *App) contentHeight() int {
	rows := 0

	if a.mode == "changes" {
		rows = len(a.currentDiffRows())
	} else if doc := a.activeDoc(); doc != nil {
		switch {
		case a.preview && doc.Lang == langMarkdown && doc.Err == "" && !doc.Binary:
			doc.ensurePreview()
			rows = len(doc.Preview)
		case a.wrap && doc.Err == "" && !doc.Binary:
			rows = len(doc.wrapped(a.codeCols()).Rows)
		default:
			rows = len(doc.Lines)
		}
	}

	top := a.historyHeight()
	if top > 0 {
		top += histSplitH
	}

	h := editorBodyTop + top + rows*rowH + chromeBottom
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

// editorWidth is the live editor column after the details drawer.
func (a *App) editorWidth() int {
	w := defaultWidth
	if a.page != nil {
		w, _ = a.page.Size()
	}

	if a.details {
		w -= detailsWidth
	}

	return w - chromeLeft
}

// codeCols is the number of monospace cells that fit in the code body.
func (a *App) codeCols() int {
	avail := a.editorWidth() - codeGutter - codePadRight - editorBorderPad
	cols := int(float64(avail) / codeCellW)
	if cols < wrapMinCols {
		cols = wrapMinCols
	}

	return cols
}

// topSourceLine is the source line at the top of the code view.
func (a *App) topSourceLine() int {
	if doc := a.activeDoc(); a.wrap && doc != nil {
		return doc.wrapped(a.codeCols()).lineAt(a.scrollY / rowH)
	}

	return a.scrollY / rowH
}

// toggleWrap flips soft wrap and keeps the top source line in view.
func (a *App) toggleWrap() {
	if a.mode != "explorer" {
		a.showToast("info", "Soft wrap", "Wrap applies to the file view. Switch to Explorer.")
		a.refresh()

		return
	}

	line := a.topSourceLine()
	a.wrap = !a.wrap
	a.viewportCols = 0

	if doc := a.activeDoc(); doc != nil && a.wrap {
		a.scrollTo(doc.wrapped(a.codeCols()).firstRow(line) * rowH)

		return
	}

	a.scrollTo(line * rowH)
}

// scrollTo moves the page and rebuilds the visible slice.
func (a *App) scrollTo(y int) {
	target := a.clampScroll(y)
	a.scrollY = target
	a.scrollTarget = target
	a.scrollPending = true
	a.page.ScrollTo(0, target)
	a.refresh()
}

// resetScroll returns to the top when a new file or mode opens.
func (a *App) resetScroll() {
	a.scrollY = 0
	a.scrollTarget = 0
	a.scrollPending = true
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

		status := a.statuses[path]

		rows = append(rows, tabRow{
			Path:        path,
			Name:        baseName(path),
			Lang:        lang,
			Badge:       languageInfo(lang).Badge,
			Status:      status,
			StatusClass: "st-" + strings.ToLower(status),
			Active:      i == a.active && a.mode == "explorer",
		})
	}

	return rows
}

// recentRows lists the welcome shortcuts that exist in this workspace.
func (a *App) recentRows() []recentRow {
	rows := make([]recentRow, 0, maxRecent)

	for _, path := range a.recent {
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

// addRecent moves path to the front of the recent list.
func (a *App) addRecent(path string) {
	next := []string{path}

	for _, old := range a.recent {
		if old == path || a.ws.ByPath[old] == nil || len(next) >= maxRecent {
			continue
		}

		next = append(next, old)
	}

	a.recent = next
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
		a.refresh()

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
	a.preview = false
	a.paletteOpen = false
	a.paletteQuery = ""
	a.paletteIndex = 0
	a.addRecent(path)
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
	a.preview = false
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

	a.preview = false
	a.resetScroll()
	a.refresh()
}

// closeWorkingTree leaves Source Control and returns to the open file. When
// every tab is closed it opens the initial file first.
func (a *App) closeWorkingTree(ctx context.Context) {
	if len(a.tabs) == 0 {
		a.openInitial()
	}

	a.setMode(ctx, "explorer")
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

	if a.filter != "" {
		a.filter = ""
		a.page.Form().SetFormValue("filter", "")
	}

	a.refresh()
}

// setMode switches the activity bar destination.
func (a *App) setMode(_ context.Context, mode string) {
	if mode != "explorer" && mode != "changes" {
		return
	}

	a.mode = mode

	a.resetScroll()
	a.refresh()
}

// reloadChanges reads live git when -root is set, or the bundled fixture.
func (a *App) reloadChanges(ctx context.Context, announce bool) {
	if a.ws.Sample {
		a.changes = sampleChangeSet()
		a.commits = commitLog{Unavailable: true, Note: "The bundled sample is not a git repository."}
		a.diffSource = "working"
	} else {
		next, err := loadLiveChanges(ctx, a.ws.Root)
		if err != nil {
			if announce {
				a.showToast("error", "Git unavailable", err.Error())
				a.refresh()

				return
			}

			a.changes = changeSet{
				Branch:     "no git",
				EmptyTitle: "Git unavailable",
				Note:       "Git is unavailable at " + a.ws.Root + ". The explorer still reads the folder.",
			}
			a.commits = commitLog{Unavailable: true, Note: "Git is unavailable at " + a.ws.Root + "."}
			a.diffSource = "working"
			a.commitOpen = ""
			a.commitFile = ""
			a.commitRows = nil
			a.commitPreview = nil
			a.statuses = statusMap(a.changes)
			a.selectFirstDiff()
			a.rebuildDiffRows()
			a.refresh()

			return
		}

		a.changes = next

		history, historyErr := loadCommitLog(ctx, a.ws.Root)
		if historyErr != nil {
			history = commitLog{
				Branch:      next.Branch,
				Note:        "History unavailable: " + historyErr.Error(),
				Unavailable: false,
			}
		}

		if history.Head != a.commits.Head {
			a.commitDetails = map[string]*changedFile{}
			a.commitKeys = nil
		}

		a.commits = history

		if history.Branch != "" {
			a.changes.Branch = history.Branch
		}

		if a.commitOpen != "" && a.findCommit(a.commitOpen) == nil {
			a.commitOpen = ""
			a.commitFile = ""
			a.commitRows = nil
			a.commitPreview = nil

			if a.diffSource == "commit" {
				a.diffSource = "working"
			}
		}
	}

	a.statuses = statusMap(a.changes)
	a.changePage = 0
	a.selectFirstDiff()
	a.rebuildDiffRows()
	a.resetScroll()

	if announce {
		switch {
		case a.ws.Sample:
			a.showToast("info", "Sample changes", "The bundled fixture is fixed. Use -root for a live repository.")
		case len(a.changes.Files) == 0:
			a.showToast("info", "Working tree clean", "No tracked or untracked changes were found.")
		default:
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
	if a.diffSource == "commit" {
		a.rebuildCommitRows()

		return
	}

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

// selectChangePage selects the first file on the current page.
func (a *App) selectChangePage() {
	from := a.changePage * diffPageRows
	if from < len(a.changes.Files) {
		a.diffPath = a.changes.Files[from].Path
		a.rebuildDiffRows()
	}

	a.resetScroll()
	a.refresh()
}

// selectDiff shows one file's working-tree diff.
func (a *App) selectDiff(path string) {
	if findFile(a.changes.Files, path) == nil {
		return
	}

	a.diffPath = path
	a.diffSource = "working"
	a.mode = "changes"
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
}

// showWorkingDiff returns the lower pane to the working tree.
func (a *App) showWorkingDiff() {
	a.diffSource = "working"
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
}

// toggleCommit expands one commit in the history rail.
func (a *App) toggleCommit(ctx context.Context, hash string) {
	commit := a.findCommit(hash)
	if commit == nil {
		return
	}

	if a.commitOpen == hash {
		if a.diffSource != "commit" && len(commit.Files) > 0 {
			a.selectCommitFile(ctx, hash, commit.Files[0].Path)

			return
		}

		a.commitOpen = ""
		a.commitFile = ""
		a.commitRows = nil
		a.commitPreview = nil
		a.hoverCommit = ""

		if a.diffSource == "commit" {
			a.diffSource = "working"
			a.rebuildDiffRows()
		}

		a.resetScroll()
		a.refresh()

		return
	}

	a.commitOpen = hash
	a.commitFile = ""
	a.focusCommitPage(hash)

	if len(commit.Files) > 0 {
		a.commitFile = commit.Files[0].Path
		a.ensureCommitFile(ctx, hash, a.commitFile)
	}

	a.diffSource = "commit"
	a.rebuildCommitRows()
	a.resetScroll()
	a.refresh()
}

// selectCommitFile selects one file inside an expanded commit.
func (a *App) selectCommitFile(ctx context.Context, hash, path string) {
	if a.findCommit(hash) == nil {
		return
	}

	a.commitOpen = hash
	a.commitFile = path
	a.diffSource = "commit"
	a.focusCommitPage(hash)
	a.ensureCommitFile(ctx, hash, path)
	a.rebuildCommitRows()
	a.resetScroll()
	a.refresh()
}

// openCommitDiff puts the selected commit file in the lower diff pane.
func (a *App) openCommitDiff(ctx context.Context, hash string) {
	commit := a.findCommit(hash)
	if commit == nil {
		return
	}

	path := a.commitFile
	if a.commitOpen != hash || path == "" {
		path = ""

		if len(commit.Files) > 0 {
			path = commit.Files[0].Path
		}
	}

	a.selectCommitFile(ctx, hash, path)
}

// ensureCommitFile reads and caches one commit patch.
func (a *App) ensureCommitFile(ctx context.Context, hash, path string) {
	if path == "" || a.ws.Sample {
		return
	}

	key := commitDetailKey(hash, path)
	if a.commitDetails[key] != nil {
		return
	}

	file, err := loadCommitFile(ctx, a.ws.Root, hash, path)
	if err != nil {
		a.showToast("error", "Commit diff unavailable", err.Error())

		return
	}

	a.commitDetails[key] = file
	a.commitKeys = append(a.commitKeys, key)

	for len(a.commitKeys) > maxCommitPatchCache {
		old := a.commitKeys[0]
		a.commitKeys = a.commitKeys[1:]
		delete(a.commitDetails, old)
	}
}

// focusCommitPage moves the history pager to the page holding hash.
func (a *App) focusCommitPage(hash string) {
	size := a.commitPageSize()
	if size < 1 {
		size = 1
	}

	for i, commit := range a.commits.Commits {
		if commit.Hash == hash {
			a.commitPage = i / size

			return
		}
	}
}

// rebuildCommitRows renders the selected commit file's patch.
func (a *App) rebuildCommitRows() {
	a.commitRows = nil
	a.commitPreview = nil

	file := a.commitDetails[commitDetailKey(a.commitOpen, a.commitFile)]
	if file == nil {
		return
	}

	lang := languageOf(file.Path)
	a.commitPreview = diffRowsUnified(*file, lang)

	if a.diffSplit {
		a.commitRows = diffRowsSplit(*file, lang)
	} else {
		a.commitRows = a.commitPreview
	}
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
	rows := a.currentDiffRows()
	if len(rows) == 0 {
		return
	}

	current := a.scrollY / rowH
	target := -1

	for i, row := range rows {
		if row.Kind != rowHunk {
			continue
		}

		if delta > 0 && i > current {
			target = i

			break
		}

		if delta < 0 && i < current {
			target = i

			break
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

// togglePreview switches a Markdown file between source and rendered rows.
func (a *App) togglePreview() {
	doc := a.activeDoc()
	if doc == nil || doc.Lang != langMarkdown || doc.Err != "" || doc.Binary {
		a.showToast("info", "Markdown preview", "Preview is available for Markdown files.")
		a.refresh()

		return
	}

	a.preview = !a.preview
	if a.preview {
		doc.ensurePreview()
	}

	a.resetScroll()
	a.refresh()
}

// outlineLine jumps from the details outline to the source line. In changes
// mode the outline belongs to the selected diff file, so open it first.
func (a *App) outlineLine(line int) {
	if a.mode == "changes" && a.diffPath != "" {
		a.openFile(a.diffPath)
	}

	a.scrollToSourceLine(line)
}

// scrollToSourceLine jumps to a source line, using the preview anchor when the
// rendered view is active.
func (a *App) scrollToSourceLine(line int) {
	target := line

	if doc := a.activeDoc(); a.preview && doc != nil && doc.Lang == langMarkdown && len(doc.PreviewLine) > 0 {
		if line >= 0 && line < len(doc.PreviewLine) && doc.PreviewLine[line] >= 0 {
			target = doc.PreviewLine[line]
		}
	} else if doc := a.activeDoc(); a.wrap && doc != nil {
		if line < 0 {
			line = 0
		}

		if line >= len(doc.Lines) {
			line = len(doc.Lines) - 1
		}

		target = doc.wrapped(a.codeCols()).firstRow(line)
	}

	a.scrollTo(target * rowH)
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
