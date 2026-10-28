package main

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/fsnotify/fsnotify"
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

	treeRowH     = 26
	treeHeadH    = 34
	treeFilterH  = 42
	treeFootH    = 32
	treeInnerPad = 5
	treeThumbMin = 18

	histHeadH    = 58
	histRowH     = 42
	histFileH    = 30
	histPreviewH = 18
	histSplitH   = 7
	histMinH     = 120
	changesHeadH = 28
	changesMinH  = 120

	historyMinRatio      = 0.18
	historyMaxRatio      = 0.82
	historyDefaultRatio  = 0.55
	defaultSplitRatio    = 0.5
	branchLogLimit       = 16
	branchPickerPad      = 122
	branchPickerRowH     = 26
	branchPickerMinRows  = 4
	changesRailTop       = 62
	paneScrollW          = 8
	paneThumbMin         = 24
	paneScrollLeft       = activityW + sideWidth - paneScrollW
	thumbCellH           = 14
	historyExpandedRows  = 3
	historyAutoRows      = 8
	historyAutoPad       = 4
	historyMetaH         = 22
	historyExpandedFiles = 5
	historyTipGap        = 2
	historyTipFlipH      = 96
	changeRowH           = 38
	changeListPad        = 6
	historyPad           = 8
	historyTipBody       = 280
	historyStep          = 0.03
	editorBorderPad      = 2
	actionPair           = 2
	halfDivisor          = 2
	roundHalf            = 0.5
	treeWheelRows        = 3.0

	maxTabs       = 8
	maxRecent     = 4
	treeParentPad = 2

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

	changes            changeSet
	branchChanges      changeSet
	branchPath         string
	statuses           map[string]string
	view               View
	fileNav            *modifiedNav
	summary            changeSummary
	changeNav          changeNavigation
	firstChangePending string

	mode           string
	filter         string
	folders        map[string]bool
	treeOffset     int
	treeDrag       bool
	treeDragAt     int
	treeScale      float64
	treeWheelAccum float64

	tabs   []string
	active int
	docs   map[string]*fileDoc

	split           bool
	focusGroup      int
	tabs2           []string
	active2         int
	splitRatio      float64
	scrollGroup     [2]int
	splitWheelAccum [2]float64

	editing    bool
	edit       *editBuffer
	dirty      map[string]bool
	editCaret  int
	editAnchor int
	editUndo   []editSnapshot
	editRedo   []editSnapshot
	dragSelect bool
	selX       float64
	selY       float64
	menuOpen   bool
	menuX      int
	menuY      int
	menuItems  []menuViewRow

	searchOpen      bool
	searchQuery     string
	searchCase      bool
	searchHits      []workspaceHit
	searchIndex     int
	searchScroll    historyScroll
	searchDrag      bool
	searchDragAt    int
	searchDragScale float64
	searchReturn    string
	searchNote      string
	searchTruncated bool
	searchToken     int
	searchBusy      bool
	searchResult    chan searchReply
	searchCancel    context.CancelFunc

	diffPath      string
	diffSplit     bool
	diffRows      []diffRow
	changesScroll changesScroll
	changesDrag   bool
	changesDragAt int
	changesScale  float64
	diffSource    string

	commits          commitLog
	branches         []branchRow
	headBranch       string
	historyRef       string
	historyLabel     string
	branchOpen       bool
	branchQuery      string
	commitDetails    map[string]*changedFile
	commitKeys       []string
	commitOpen       string
	commitFile       string
	commitRows       []diffRow
	commitPreview    []diffRow
	patchLoader      *commitPatchLoader
	branchLogs       *branchLogCache
	historyScroll    historyScroll
	historyLimit     int
	historyDone      bool
	historyListDrag  bool
	historyListAt    int
	historyListScale float64
	historyToken     int
	historyLoad      chan historyLoadReply

	historyRatio  float64
	historyManual bool
	historyStart  float64
	historyDrag   bool
	historyDy     float64
	hoverCommit   string
	hoverPath     string
	commitTip     commitFloatView
	inlineCard    inlineCard

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
	scrollX      int
	viewportCols int
	findOpen     bool
	findQuery    string
	findCase     bool
	findIndex    int
	findHits     []findHit
	findReturn   string

	watch        *fsnotify.Watcher
	watchCancel  context.CancelFunc
	watchUpdates chan watchUpdate
	watchNote    string

	sessionPath     string
	sessionLast     string
	sessionReady    bool
	sessionRestored bool
	scanLimit       int
}

// New scans the workspace, builds the page, and opens the README. An empty
// root uses the embedded sample project.
func New(root string) (*App, error) { return NewWithLimit(root, 0) }

// NewWithLimit is New with a file scan budget. A limit of zero or less uses
// the default maxWorkspaceFiles.
func NewWithLimit(root string, limit int) (*App, error) {
	var (
		ws  *workspace
		err error
	)

	if root == "" {
		ws, err = loadSampleWorkspace()
	} else {
		ws, err = loadDiskWorkspaceWithLimit(root, limit)
	}

	if err != nil {
		return nil, err
	}

	a := &App{
		ws:            ws,
		docs:          map[string]*fileDoc{},
		dirty:         map[string]bool{},
		folders:       map[string]bool{},
		statuses:      map[string]string{},
		commitDetails: map[string]*changedFile{},
		mode:          "explorer",
		diffSource:    "working",
		historyRatio:  historyDefaultRatio,
		splitRatio:    defaultSplitRatio,
		viewportH:     defaultHeight,
		devOn:         true,
		wrap:          true,
		branchLogs:    newBranchLogCache(branchLogLimit),
		searchResult:  make(chan searchReply, 1),
		scanLimit:     ws.Limit,
		historyLimit:  maxCommits,
		historyScroll: historyScroll{Limit: maxCommits},
		historyLoad:   make(chan historyLoadReply, 1),
	}

	if ws.Sample {
		a.folders["backend"] = true
		a.folders["web"] = true
	}

	if ws.Sample {
		a.changes = sampleChangeSet()
		a.useSampleHistory()
	} else {
		a.changes = changeSet{Note: "Open Source Control to read the working tree."}
	}

	a.statuses = statusMap(a.changes)
	a.syncFileNav()
	a.selectFirstDiff()

	page, err := ownframe.New(ownframe.Config{
		DevTools:         true,
		Title:            "Code Review",
		HTML:             buildHTML(),
		Width:            defaultWidth,
		Height:           defaultHeight,
		MinWidth:         minWidth,
		MinHeight:        minHeight,
		DisableWheelZoom: true,
		DisableKeyZoom:   true,
	})
	if err != nil {
		return nil, err
	}

	a.page = page
	if !ws.Sample {
		// New has no caller context. The patch worker lives for the app and
		// Close cancels its in-flight git read.
		a.patchLoader = NewCommitPatchLoader(ws.Root, maxCommitPatchCache)
	}

	page.Handle(ownframe.Handlers{
		Click:          a.onClick,
		Press:          a.onPress,
		Change:         a.onChange,
		KeyDown:        a.onKeyDown,
		KeyUp:          a.onKeyUp,
		Submit:         a.onSubmit,
		Hover:          a.onHover,
		DragStart:      a.onDragStart,
		DragMove:       a.onDragMove,
		DragEnd:        a.onDragEnd,
		Wheel:          a.onWheel,
		Type:           a.onType,
		Backspace:      a.onBackspace,
		DeleteWord:     a.onDeleteWord,
		Copy:           a.onCopy,
		Cut:            a.onCut,
		Paste:          a.onPaste,
		SelectAll:      a.onSelectAll,
		Undo:           a.onUndo,
		Redo:           a.onRedo,
		ContextMenuFor: a.contextMenuForPage,
		MenuAction:     a.MenuAction,
		PresentMenu:    a.presentMenu,
	})
	page.Scroll().SetWindowing(true)
	page.SetScrollWindow(a.onScroll)
	page.SetTick(a.tickWatch)

	a.openInitial()
	a.refresh()

	return a, nil
}

// publishPreviewImages registers the diagram strips the current documents hold.
func (a *App) publishPreviewImages() {
	if a.page == nil {
		return
	}

	for _, doc := range a.docs {
		if doc == nil {
			continue
		}

		for _, row := range doc.Preview {
			if row.Image == "" || len(row.png) == 0 {
				continue
			}

			a.page.SetImage(row.Image, row.png)
		}
	}
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
	a.publishPreviewImages()
	a.page.SetData(&a.view)
	a.persistSession()
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

	if !a.split {
		a.scrollY = offsetY
		a.scrollGroup[0] = offsetY
	}
	a.viewportH = viewH
	a.viewportCols = cols
	a.inlineCard = inlineCard{}
	a.view = a.buildView()
	a.publishPreviewImages()
	a.page.SetData(&a.view)

	return true
}

// buildView assembles one frame from the current state.
func (a *App) buildView() View {
	v := View{
		Mode:        a.mode,
		Filter:      a.filter,
		Details:     a.details,
		DevOn:       a.devOn,
		MenuOpen:    a.menuOpen,
		MenuX:       a.menuX,
		MenuY:       a.menuY,
		MenuRows:    append([]menuViewRow(nil), a.menuItems...),
		Modal:       a.paletteOpen || a.branchOpen || a.searchOpen,
		Workspace:   baseName(a.ws.Root),
		Branch:      a.changes.Branch,
		Changed:     len(a.changes.Files),
		DevNote:     a.changes.Note,
		ScanBudget:  a.ws.ScanBudget(),
		WatchNote:   a.watchNote,
		HoverPath:   a.hoverPath,
		Dirty:       a.dirtyCount(),
		Toast:       a.toast,
		Wrap:        a.wrap,
		ScrollX:     0,
		FindQuery:   a.findQuery,
		FindOpen:    a.findOpen,
		FindCase:    a.findCase,
		FindCurrent: a.findCurrent(),
		FindTotal:   len(a.findHits),
		BranchQuery: a.branchQuery,
		Split:       a.split && a.mode == "explorer",
		FocusGroup:  a.focusGroup,
	}

	if !a.wrap {
		v.ScrollX = a.scrollX
	}

	v.Tree, v.TreeTotal, v.TreeCount, v.Filtered = a.treeRows()
	v.TreeStart = a.treeOffset
	v.TreeVisible = a.treeVisibleRows()
	a.fillTreeScroll(&v)
	v.Tabs = a.tabRows()
	v.Recent = a.recentRows()
	v.ActiveTab = a.active
	v.PaletteQuery = a.paletteQuery
	v.SearchQuery = a.searchQuery
	v.Palette = a.paletteView()
	v.Search = a.searchViewState()
	v.FileNav = a.fileNavViewState()
	v.ChangeLabel = a.fileNavSummary()

	v.PaneLeft, v.PaneWidth = a.paneGeometry()

	if doc := a.groupDoc(0); doc != nil {
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
		v.CanEdit = doc.Err == "" && !doc.Binary
		v.Editing = a.editing && a.editPath() == doc.Path
		v.EditDirty = a.dirty[doc.Path]

		if v.Editing && a.edit != nil {
			v.EditText = a.edit.Text()
		}
	}

	if a.mode == "changes" {
		a.fillDiff(&v)
		a.fillHistory(&v)
	} else {
		a.fillCode(&v)
	}

	if v.Split {
		v.Second = a.secondView()
	}

	v.CommitTip = a.commitTip
	v.InlineCard = a.inlineCard
	v.SpaceHeight = a.contentHeight()

	return v
}

// fillCode fills the first editor group, its visible rows, and the empty state.
func (a *App) fillCode(v *View) {
	doc := a.groupDoc(0)
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
	v.Preview = a.preview && v.Previewable && a.activeGroup() == 0
	cols := a.codeColsForGroup(0)

	if v.Preview {
		doc.ensurePreview(cols)
		start, end, track := a.windowSliceForGroup(0, len(doc.Preview))
		v.TrackTop = track

		if len(doc.Preview) > 0 {
			v.PreviewRows = doc.Preview[start:end]
		}

		return
	}

	start, end, track := a.windowSliceForGroup(0, len(doc.Lines))
	if a.wrap {
		plan := doc.wrapped(cols)
		start, end, track = a.windowSliceForGroup(0, len(plan.Rows))
		v.TrackTop = track
		v.Rows = a.copyFindRows(cloneCodeRows(plan.Rows[start:end]))
		a.decorateCodeRows(doc, v.Rows, 0)
		a.decorateEditRows(v.Rows)

		return
	}

	v.TrackTop = track
	v.Rows = make([]codeRow, 0, end-start)

	for i := start; i < end; i++ {
		row := codeRow{No: i + 1, Line: i, Segments: doc.Tokens[i]}
		if !a.wrap {
			row.ScrollX = a.scrollX
		}
		v.Rows = append(v.Rows, row)
	}

	v.Rows = a.copyFindRows(v.Rows)
	a.decorateCodeRows(doc, v.Rows, 0)
	a.decorateEditRows(v.Rows)
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

// diffDisplayRows expands the selected diff to the visual rows the page draws.
func (a *App) diffDisplayRows() []diffRow {
	rows := a.currentDiffRows()
	if len(rows) == 0 {
		return nil
	}

	if a.wrap {
		unified, split := a.diffColumns()
		return wrapDiffRows(rows, unified, split, true)
	}

	return rows
}

// fillDiff fills the Source Control list and the selected diff rows.
func (a *App) fillDiff(v *View) {
	files := a.sidebarChangeFiles()
	active := a.changes
	if a.diffSource == "branch" {
		active = a.branchChanges
	}
	d := diffView{
		Split:      a.diffSplit,
		Sample:     active.Sample,
		Note:       active.Note,
		EmptyTitle: active.EmptyTitle,
		HasChanges: len(files) > 0,
		Count:      len(files),
	}

	add, del := active.totals()
	d.TotalAdd, d.TotalDel = add, del
	selected := a.diffPath
	if a.diffSource == "branch" {
		selected = a.branchPath
	}
	if a.diffSource == "commit" {
		selected = a.commitFile
		d.Note = ""
		d.EmptyTitle = "No files in this commit"
		d.TotalAdd, d.TotalDel = 0, 0
		if commit := a.findCommit(a.commitOpen); commit != nil {
			d.TotalAdd, d.TotalDel = commit.Add, commit.Del
		}
	}

	total := len(files)
	visible := a.changeVisibleRows()
	from, to := a.changesScroll.Rows(total, visible)

	d.Scroll = total > visible
	d.AtTop = from == 0
	d.AtEnd = to >= total

	if d.Scroll {
		track := a.changesTrackHeight()
		d.ScrollTop, d.ScrollHeight, _ = a.changesScroll.Thumb(track, total, visible, paneThumbMin)
		d.ScrollTrack = track
		d.RailLeft = paneScrollLeft
		d.RailTop = chromeTop + changesRailTop
		d.RailThumbTop = d.RailTop + d.ScrollTop
		d.ThumbCells = max(1, d.ScrollHeight/thumbCellH)
	}

	for _, file := range files[from:to] {
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
			Selected:    file.Path == selected,
		})
	}

	if a.diffSource == "commit" {
		a.fillCommitDiff(v, &d)
		v.Diff = d

		return
	}

	d.Source = a.diffSource
	d.HasFiles = len(active.Files) > 0

	if !d.HasFiles {
		v.Diff = d

		return
	}

	file := findFile(active.Files, selected)
	if file == nil {
		file = &active.Files[0]
	}

	a.fillDiffFile(v, &d, file, a.diffRows)
	d.Selected = file.Path
	v.Diff = d
}

func (a *App) sidebarChangeFiles() []changedFile {
	if a.diffSource == "branch" {
		return a.branchChanges.Files
	}
	if a.diffSource != "commit" {
		return a.changes.Files
	}
	commit := a.findCommit(a.commitOpen)
	if commit == nil {
		return nil
	}
	files := make([]changedFile, 0, len(commit.Files))
	for _, file := range commit.Files {
		files = append(files, changedFile{
			Path: file.Path, OldPath: file.OldPath, Status: file.Status,
			Add: file.Add, Del: file.Del, Binary: file.Binary,
		})
	}
	return files
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
		if !a.ws.Sample && path != "" {
			d.Note = "Loading the patch for " + path + "…"
			d.Loading = true
		} else {
			d.Note = "Choose a file in the commit to read its patch."
		}

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

		if d.Source == "commit" || d.Source == "branch" {
			v.Status = file.Status
		} else {
			v.Status = a.statuses[doc.Path]
		}

		v.StatusClass = "st-" + strings.ToLower(v.Status)
	}

	if a.wrap {
		unified, split := a.diffColumns()
		rows = wrapDiffRows(rows, unified, split, true)
	} else {
		d.ScrollX = a.scrollX
	}

	start, end, track := a.windowSlice(len(rows))
	d.TrackTop = track

	if len(rows) > 0 {
		visible := append([]diffRow(nil), rows[start:end]...)
		current := a.scrollY / rowH
		if a.changeNav.valid && a.changeNav.path == file.Path &&
			a.changeNav.mode == a.mode && a.changeNav.scroll == a.scrollY {
			current = a.changeNav.row
		}
		markCurrentDiffChange(rows, visible, start, current)
		if !a.wrap {
			for i := range visible {
				visible[i].ScrollX = a.scrollX
			}
		}
		d.Rows = visible
	}
}

func diffChangedRow(row diffRow) bool {
	return row.Kind == rowAdd || row.Kind == rowDel ||
		row.Kind == "split" && (row.LeftKind == rowDel || row.RightKind == rowAdd)
}

func diffChangeStarts(rows []diffRow) []int {
	var starts []int
	inside := false
	for i, row := range rows {
		changed := diffChangedRow(row)
		if changed && !inside {
			starts = append(starts, i)
		}
		inside = changed
	}
	return starts
}

func markCurrentDiffChange(rows, visible []diffRow, start, current int) {
	starts := diffChangeStarts(rows)
	if len(starts) == 0 {
		return
	}
	selected := starts[0]
	for _, row := range starts {
		if row > current {
			break
		}
		selected = row
	}
	end := selected
	for end < len(rows) && diffChangedRow(rows[end]) {
		end++
	}
	for i := range visible {
		visible[i].Current = start+i >= selected && start+i < end
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
	return a.mode == "changes" && !a.commits.Unavailable && a.historyLabel != ""
}

// historyHeight is the clamped top pane height.
func (a *App) historyHeight() int {
	if !a.historyVisible() {
		return 0
	}

	stack := a.drawerStackHeight()
	maxH := stack - changesMinH

	h := 0
	if !a.historyManual {
		rows := min(len(a.commits.Commits), historyAutoRows)
		if rows < 1 {
			rows = 1
		}

		h = histHeadH + historyPad + rows*histRowH + historyAutoPad
	} else {
		h = int(float64(stack) * a.historyRatio)
	}
	if h < histMinH {
		h = histMinH
	}

	if maxH < histMinH {
		return stack / halfDivisor
	}

	if h > maxH {
		h = maxH
	}

	return h
}

// historyPaneTop is the page y of the history rail's top edge.
func (a *App) historyPaneTop() int {
	return a.viewportH - chromeBottom - treeFootH - a.historyHeight()
}

// historyVisibleRows is how many 42 px commit rows fit in the rail.
func (a *App) historyVisibleRows() int {
	h := a.historyHeight() - histHeadH - historyPad
	rows := h / histRowH
	if rows < 1 {
		rows = 1
	}

	return rows
}

// historyTrackHeight is the rail's pixel height below its header.
func (a *App) historyTrackHeight() int {
	h := a.historyHeight() - histHeadH - historyPad
	if h < 1 {
		h = 1
	}

	return h
}

// pageWidth is the current page width, with the default as a fallback.
func (a *App) pageWidth() int {
	w, _ := a.page.Size()
	if w < 1 {
		w = defaultWidth
	}

	return w
}

// buildCommitTip places the floating commit card beside the history rail.
func (a *App) buildCommitTip(commit *gitCommit, rowTop int) commitFloatView {
	body := strings.Join(strings.Fields(commit.Body), " ")
	meta := commit.Short + " · " + commit.Author + " · " + commit.Date
	pageTop := a.historyPaneTop() + rowTop

	lines := commitTipText(commit.Subject, body, meta, commitTipCols(commitTipWidth))
	layout := layoutCommitTip(a.pageWidth(), a.viewportH, activityW, sideWidth, pageTop, histRowH, len(lines))
	if !layout.Show {
		return commitFloatView{}
	}

	lines = commitTipText(commit.Subject, body, meta, commitTipCols(layout.Width))
	layout = layoutCommitTip(a.pageWidth(), a.viewportH, activityW, sideWidth, pageTop, histRowH, len(lines))

	return commitFloatView{
		Show:  layout.Show,
		Left:  layout.Left,
		Top:   layout.Top,
		Width: layout.Width,
		Lines: lines,
	}
}

// diffBodyHeight is the editor body. History lives in the sidebar.
func (a *App) diffBodyHeight() int { return a.bodyHeight() }

// drawerStackHeight is the Source Control space between the side header and
// footer, after the changes header and divider.
func (a *App) drawerStackHeight() int {
	h := a.viewportH - chromeTop - chromeBottom - treeHeadH - treeFootH - changesHeadH - histSplitH
	if h < 1 {
		h = 1
	}

	return h
}

// changeVisibleRows is how many 38 px change rows fit above History.
func (a *App) changeVisibleRows() int {
	stack := a.drawerStackHeight()
	body := stack - a.historyHeight()
	rows := (body - changeListPad) / changeRowH
	if rows < 1 {
		rows = 1
	}

	return rows
}

// revealChangeIndex scrolls the changed-file list so one row is visible.
func (a *App) revealChangeIndex(index int) {
	total := len(a.sidebarChangeFiles())
	visible := a.changeVisibleRows()
	start, end := a.changesScroll.Rows(total, visible)

	switch {
	case index < start:
		a.changesScroll.SetOffset(index, total, visible)
	case index >= end:
		a.changesScroll.SetOffset(index-visible+1, total, visible)
	}
}

// revealChangePath reveals one changed file by path.
func (a *App) revealChangePath(path string) {
	for i, file := range a.sidebarChangeFiles() {
		if file.Path == path {
			a.revealChangeIndex(i)

			return
		}
	}
}

// scrollChanges moves the changed-file list by rows.
func (a *App) scrollChanges(rows int) {
	a.changesScroll.ScrollBy(rows, len(a.sidebarChangeFiles()), a.changeVisibleRows())
	a.refresh()
}

// changesTrackHeight is the pixel height of the changed-file rail.
func (a *App) changesTrackHeight() int {
	h := a.drawerStackHeight() - a.historyHeight() - changeListPad
	if h < 1 {
		h = 1
	}

	return h
}

// fillHistory builds the commit rail and its expanded preview.
func (a *App) fillHistory(v *View) {
	a.commitTip = commitFloatView{}
	h := historyView{
		Visible:     a.historyVisible(),
		Sample:      a.commits.Sample,
		Branch:      a.historyLabel,
		HeadBranch:  a.headBranch,
		Browsing:    a.historyRef != "",
		Detached:    strings.HasPrefix(a.headBranch, "HEAD detached"),
		BranchOpen:  a.branchOpen,
		BranchQuery: a.branchQuery,
	}
	if !h.Visible {
		v.History = h

		return
	}

	if h.Branch == "" {
		h.Branch = a.commits.Branch
	}

	h.Note = a.commits.Note
	h.Unborn = a.commits.Unborn
	h.Height = a.historyHeight()
	h.Loading = v.Diff.Loading
	h.RailLeft = paneScrollLeft
	h.RailTop = a.historyPaneTop() + histHeadH

	if a.branchOpen {
		a.fillBranchList(&h)
		v.History = h

		return
	}

	visible := a.historyVisibleRows()
	total := len(a.commits.Commits)
	from, to := a.historyScroll.Rows(total, visible)

	h.Scroll = total > visible
	h.AtTop = from == 0
	h.AtEnd = to >= total
	h.LoadingMore = a.historyScroll.Loading

	if h.Scroll {
		track := a.historyTrackHeight()
		h.ScrollTop, h.ScrollHeight, _ = a.historyScroll.Thumb(track, total, visible, paneThumbMin)
		h.ScrollTrack = track
		h.RailThumbTop = h.RailTop + h.ScrollTop
		h.ThumbCells = max(1, h.ScrollHeight/thumbCellH)
	}

	graph := buildCommitGraph(a.commits.Commits)
	tipTop := histHeadH + historyPad
	tipRowTop := 0

	for i := from; i < to; i++ {
		commit := a.commits.Commits[i]
		rowTop := tipTop
		row := commitRow{
			Hash:       commit.Hash,
			Short:      commit.Short,
			Subject:    commit.Subject,
			Lines:      historySubjectLines(commit.Subject),
			Author:     commit.Author,
			Date:       commit.Date,
			Body:       commit.Body,
			Add:        commit.Add,
			Del:        commit.Del,
			More:       commit.More,
			Head:       commit.Hash == a.commits.Head,
			Expanded:   commit.Hash == a.commitOpen,
			Selected:   commit.Hash == a.commitOpen && a.diffSource == "commit",
			Graph:      graph.Rows[commit.Hash],
			GraphEdges: graph.Edges[commit.Hash],
			Lanes:      graph.Lanes,
		}
		for _, cell := range row.Graph {
			if cell.Kind == graphNode {
				row.GraphX = graphX(cell.Lane)
				row.GraphColor = graphColor(cell.Lane)
				break
			}
		}

		h.Commits = append(h.Commits, row)
		tipTop += histRowH

		if a.hoverCommit == commit.Short {
			tipRowTop = rowTop
			tip := rowTop + histRowH + historyTipGap
			if tip+historyTipFlipH > h.Height {
				tip = rowTop - historyTipFlipH
			}

			if tip < histHeadH {
				tip = histHeadH
			}

			h.Tip.Top = tip
		}
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
			a.commitTip = a.buildCommitTip(&commit, tipRowTop)
		}
	}

	v.History = h
}

// fillBranchList builds the filtered branch picker rows.
func (a *App) fillBranchList(h *historyView) {
	limit := (h.Height - branchPickerPad) / branchPickerRowH
	if limit < branchPickerMinRows {
		limit = branchPickerMinRows
	}

	rows, more := filterBranches(a.branches, a.branchQuery, limit)
	h.Branches = rows
	h.BranchMore = more
}

// windowSlice returns the row range the fixed body shows and the track offset.
func (a *App) windowSlice(total int) (int, int, int) {
	return a.windowSliceForGroup(0, total)
}

// contentHeight is the page height for the current mode. The fixed chrome
// covers the rest, so a short file leaves the page at one viewport.
func (a *App) contentHeight() int {
	rows := 0

	if a.mode == "changes" {
		rows = len(a.diffDisplayRows())
	} else {
		rows = a.paneRowCount(0, a.groupDoc(0))

		if a.split {
			rows = max(rows, a.paneRowCount(1, a.groupDoc(1)))
		}
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

// editorWidth is the live editor column after the details drawer. The window
// already sizes the page to the strip beside the devtools dock.
func (a *App) editorWidth() int {
	w := defaultWidth
	if a.page != nil {
		w, _ = a.page.Size()
	}

	if a.details {
		w -= detailsWidth
	}

	if w < chromeLeft+paneMinWidth {
		w = chromeLeft + paneMinWidth
	}

	return w - chromeLeft
}

// codeCols is the number of monospace cells that fit in the focused pane.
func (a *App) codeCols() int {
	return a.codeColsForGroup(a.activeGroup())
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
	if a.mode == "changes" {
		a.wrap = !a.wrap
		if a.wrap {
			a.scrollX = 0
		}
		a.resetScroll()
		a.refresh()

		return
	}

	line := a.topSourceLine()
	a.wrap = !a.wrap
	if a.wrap {
		a.scrollX = 0
	}
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
	a.inlineCard = inlineCard{}
	a.scrollY = target
	a.scrollGroup[0] = target
	a.scrollTarget = target
	a.scrollPending = true
	a.page.ScrollTo(0, target)
	a.refresh()
}

// resetScroll returns to the top when a new file or mode opens.
func (a *App) resetScroll() {
	a.resetFind()
	a.changeNav = changeNavigation{}
	a.scrollY = 0
	a.scrollGroup[0] = 0
	a.scrollGroup[1] = 0
	a.splitWheelAccum = [2]float64{}
	a.scrollTarget = 0
	a.scrollPending = true
	a.scrollX = 0
	a.page.ScrollTo(0, 0)
}

// resetFind clears the find bar without touching focus.
func (a *App) resetFind() {
	a.findOpen = false
	a.findQuery = ""
	a.findCase = false
	a.findHits = nil
	a.findIndex = 0
	a.findReturn = ""
}

// treeVisibleRows is the number of 26 px rows that fit in the sidebar.
func (a *App) treeVisibleRows() int {
	h := a.viewportH - chromeTop - chromeBottom - treeHeadH - treeFilterH - treeFootH - treeInnerPad
	rows := h / treeRowH
	if rows < 1 {
		rows = 1
	}

	return rows
}

// treeRows returns the visible explorer rows, the total row count, the file
// count, and whether the filter is active.
func (a *App) treeRows() ([]treeRow, int, int, bool) {
	count := len(a.ws.Files)
	if a.filter != "" {
		return a.filteredRows()
	}

	total := countTreeOpen(a.ws.Tree, a.folders)
	visible := a.treeVisibleRows()
	a.clampTreeOffset(total, visible)

	rows := make([]treeRow, 0, min(total, visible))
	index := 0
	a.walkTree(a.ws.Tree, 0, &rows, &index, a.treeOffset, a.treeOffset+visible)

	return rows, total, count, false
}

// walkTree appends the visible nodes in the current offset window.
func (a *App) walkTree(nodes []*fileNode, depth int, rows *[]treeRow, index *int, start, end int) {
	for _, node := range nodes {
		if *index >= end {
			return
		}

		pos := *index
		*index++

		if pos >= start {
			row := treeRow{
				Path:    node.Path,
				Name:    node.Name,
				Dir:     node.Dir,
				Ignored: node.Ignored,
				Indent:  treePad + depth*treeIndent,
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
		}

		if node.Dir && a.folders[node.Path] {
			a.walkTree(node.Children, depth+1, rows, index, start, end)
		}
	}
}

// filteredRows returns the visible slice of matching paths.
func (a *App) filteredRows() ([]treeRow, int, int, bool) {
	query := strings.ToLower(a.filter)
	visible := a.treeVisibleRows()
	all := make([]treeRow, 0, min(len(a.ws.Files), visible))

	for _, file := range a.ws.Files {
		if !strings.Contains(strings.ToLower(file.Path), query) {
			continue
		}

		all = append(all, treeRow{
			Path:        file.Path,
			Name:        file.Name,
			Indent:      treePad,
			Ignored:     file.Ignored,
			Lang:        file.Lang,
			Badge:       languageInfo(file.Lang).Badge,
			Status:      a.statuses[file.Path],
			StatusClass: "st-" + strings.ToLower(a.statuses[file.Path]),
		})
	}

	a.clampTreeOffset(len(all), visible)
	from := min(a.treeOffset, len(all))
	to := min(from+visible, len(all))

	return all[from:to], len(all), len(a.ws.Files), true
}

// clampTreeOffset keeps the first visible row inside the list.
func (a *App) clampTreeOffset(total, visible int) {
	maxOffset := total - visible
	if maxOffset < 0 {
		maxOffset = 0
	}

	if a.treeOffset > maxOffset {
		a.treeOffset = maxOffset
	}

	if a.treeOffset < 0 {
		a.treeOffset = 0
	}
}

// fillTreeScroll publishes the scrollbar thumb geometry.
func (a *App) fillTreeScroll(v *View) {
	if v.TreeTotal <= v.TreeVisible || v.TreeVisible < 1 {
		v.TreeScroll = false

		return
	}

	track := v.TreeVisible * treeRowH
	thumb := track * v.TreeVisible / v.TreeTotal
	if thumb < treeThumbMin {
		thumb = 18
	}

	if thumb > track {
		thumb = track
	}

	maxOffset := v.TreeTotal - v.TreeVisible
	room := track - thumb
	top := 0
	if maxOffset > 0 && room > 0 {
		top = v.TreeStart * room / maxOffset
	}

	v.TreeScroll = true
	v.TreeThumbH = thumb
	v.TreeThumbTop = top
	v.TreeTrackH = track
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

// tabRows returns the open tabs in the first group.
func (a *App) tabRows() []tabRow { return a.tabRowsFor(0) }

// tabRowsFor returns one group's open tabs.
func (a *App) tabRowsFor(group int) []tabRow {
	tabs := a.groupTabs(group)
	active := a.groupActive(group)
	rows := make([]tabRow, 0, len(tabs))

	for i, path := range tabs {
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
			Active:      i == active && a.mode == "explorer",
			Dirty:       a.dirty[path],
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

// activeDoc returns the document behind the focused group's active tab.
func (a *App) activeDoc() *fileDoc {
	return a.groupDoc(a.activeGroup())
}

// openFile opens or focuses one file in the group that has focus.
func (a *App) openFile(path string) { a.openFileIn(a.activeGroup(), path) }

// openFileIn opens or focuses one file in a specific editor group.
func (a *App) openFileIn(group int, path string) {
	if a.ws.ByPath[path] == nil {
		a.showToast("error", "Not in workspace", path+" is not part of this workspace.")
		a.refresh()

		return
	}

	a.finishEdit()
	a.firstChangePending = ""

	if _, ok := a.docs[path]; !ok {
		a.docs[path] = a.ws.openDoc(path)
	}

	tabs := &a.tabs
	active := &a.active

	if group == 1 && a.split {
		tabs = &a.tabs2
		active = &a.active2
	} else {
		group = 0
	}

	found := -1

	for i, open := range *tabs {
		if open == path {
			found = i

			break
		}
	}

	if found < 0 {
		*tabs = append(*tabs, path)
		if len(*tabs) > maxTabs {
			*tabs = append((*tabs)[:0], (*tabs)[1:]...)
		}

		found = len(*tabs) - 1
	}

	*active = found
	a.focusGroup = group
	a.mode = "explorer"
	a.preview = false
	a.inlineCard = inlineCard{}
	a.paletteOpen = false
	a.paletteQuery = ""
	a.paletteIndex = 0
	if a.fileNav != nil {
		a.fileNav.Select(path)
	}
	a.addRecent(path)
	if a.split {
		a.scrollGroup[group] = 0
		a.splitWheelAccum[group] = 0
		if group == 0 {
			a.scrollY = 0
			if a.page != nil {
				a.page.ScrollTo(0, 0)
			}
		}
	} else {
		a.resetScroll()
	}
	a.refresh()
	a.jumpToFirstChange(path)
}

// selectTab focuses one tab in the first group.
func (a *App) selectTab(index int) { a.selectTabIn(0, index) }

// selectTabIn focuses one tab in a specific group.
func (a *App) selectTabIn(group, index int) {
	tabs := a.groupTabs(group)
	if index < 0 || index >= len(tabs) {
		return
	}

	a.finishEdit()

	if group == 1 && a.split {
		a.active2 = index
	} else {
		group = 0
		a.active = index
	}

	a.focusGroup = group
	a.mode = "explorer"
	a.preview = false
	if a.split {
		a.scrollGroup[group] = 0
		a.splitWheelAccum[group] = 0
		if group == 0 {
			a.scrollY = 0
			if a.page != nil {
				a.page.ScrollTo(0, 0)
			}
		}
	} else {
		a.resetScroll()
	}
	a.refresh()
}

// closeTab closes one tab in the first group.
func (a *App) closeTab(index int) { a.closeTabIn(0, index) }

// closeTabIn closes one tab in a specific group.
func (a *App) closeTabIn(group, index int) {
	tabs := a.groupTabs(group)
	if index < 0 || index >= len(tabs) {
		return
	}

	a.finishEdit()

	next := append([]string(nil), tabs[:index]...)
	next = append(next, tabs[index+1:]...)

	if group == 1 && a.split {
		a.tabs2 = next
		if index < a.active2 {
			a.active2--
		}

		if a.active2 >= len(a.tabs2) {
			a.active2 = len(a.tabs2) - 1
		}

		if a.active2 < 0 {
			a.active2 = 0
		}
	} else {
		group = 0
		a.tabs = next

		if index < a.active {
			a.active--
		}

		if a.active >= len(a.tabs) {
			a.active = len(a.tabs) - 1
		}

		if a.active < 0 {
			a.active = 0
		}
	}

	if a.split && len(a.tabs2) == 0 && group == 1 {
		a.split = false
		a.focusGroup = 0
	}

	a.preview = false
	a.resetScroll()
	a.refresh()
}

// useSampleHistory installs the fixed sample rail without running git.
func (a *App) useSampleHistory() {
	history, details := sampleCommitLog()
	a.commits = history
	a.historyLabel = history.Branch
	a.headBranch = history.Branch
	a.commitDetails = details
	a.commitKeys = nil
	a.commitOpen = ""
	a.commitFile = ""
	a.commitRows = nil
	a.commitPreview = nil
	a.historyScroll.SetOffset(0, len(history.Commits), a.historyVisibleRows())
	a.historyScroll.Limit = len(history.Commits)
	a.historyDone = true
	a.hoverCommit = ""
	a.branchOpen = false
	a.branchQuery = ""
}

// closeWorkingTree leaves Source Control and returns to the open file. When
// every tab is closed it opens the initial file first.
func (a *App) closeWorkingTree(ctx context.Context) {
	if len(a.tabs) == 0 {
		a.openInitial()
	}

	a.setMode(ctx, "explorer")
}

// toggleFolder flips one directory and reveals its children when it opens.
func (a *App) toggleFolder(path string) {
	a.folders[path] = !a.folders[path]

	if a.folders[path] {
		a.revealTreeFolder(path)
	}

	a.refresh()
}

// revealTreeFolder puts a newly expanded folder at the top of the tree.
func (a *App) revealTreeFolder(path string) {
	index := 0
	if !treeRowIndex(a.ws.Tree, a.folders, path, &index) {
		return
	}

	visible := a.treeVisibleRows()
	if index > a.treeOffset+visible-4 {
		a.treeOffset = index
	}

	a.clampTreeOffset(countTreeOpen(a.ws.Tree, a.folders), visible)
}

// treeRowIndex finds the row index of one path in the current fold state.
func treeRowIndex(nodes []*fileNode, open map[string]bool, path string, index *int) bool {
	for _, node := range nodes {
		pos := *index
		*index++

		if node.Path == path {
			*index = pos

			return true
		}

		if node.Dir && open[node.Path] && treeRowIndex(node.Children, open, path, index) {
			return true
		}
	}

	return false
}

// collapseAll closes every folder.
func (a *App) collapseAll() {
	for path := range a.folders {
		delete(a.folders, path)
	}

	a.treeOffset = 0

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
	a.inlineCard = inlineCard{}
	if mode == "changes" && a.diffPath != "" && a.fileNav != nil {
		a.fileNav.Select(a.diffPath)
	}

	a.resetScroll()
	a.refresh()
}

// reloadChanges reads live git when -root is set, or the bundled fixture.
func (a *App) reloadChanges(ctx context.Context, announce bool) {
	if a.ws.Sample {
		a.changes = sampleChangeSet()
		a.useSampleHistory()
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
			a.syncFileNav()
			a.selectFirstDiff()
			a.rebuildDiffRows()
			a.refresh()

			return
		}

		a.changes = next

		branches, head, branchErr := loadBranches(ctx, a.ws.Root)
		if branchErr == nil {
			a.branches = branches
			a.headBranch = head

			if a.changes.Branch == "detached" && head != "" {
				a.changes.Branch = head
			}

			if a.historyRef != "" {
				if _, ok := branchByRef(a.branches, a.historyRef); !ok {
					a.historyRef = ""
					a.historyLabel = head
				}
			}
		}

		history, historyErr := loadCommitLog(ctx, a.ws.Root, a.historyRef)
		if historyErr != nil {
			history = commitLog{
				Branch:      next.Branch,
				Ref:         a.historyRef,
				Note:        "History unavailable: " + historyErr.Error(),
				Unavailable: false,
			}
		} else if a.historyRef != "" {
			a.branchLogs.Put(a.historyRef, history)
		}

		if history.Head != a.commits.Head {
			a.commitDetails = map[string]*changedFile{}
			a.commitKeys = nil
		}

		a.commits = history
		a.historyLabel = history.Branch
		a.historyScroll.Limit = len(history.Commits)
		a.historyDone = len(history.Commits) < a.historyLimit || a.historyLimit >= maxHistoryCommits

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

	a.historyLimit = maxCommits
	a.historyDone = false
	a.historyScroll.SetOffset(0, len(a.commits.Commits), a.historyVisibleRows())
	a.statuses = statusMap(a.changes)
	a.syncFileNav()
	a.changesScroll.SetOffset(0, len(a.sidebarChangeFiles()), a.changeVisibleRows())
	if a.diffSource == "commit" {
		a.revealChangePath(a.commitFile)
	}
	if a.diffSource == "branch" {
		a.refreshBranchChanges(ctx)
	}
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

	if a.fileNav != nil {
		a.fileNav.Select(a.diffPath)
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
	if a.diffSource == "branch" {
		file = findFile(a.branchChanges.Files, a.branchPath)
	}
	if file == nil {
		return
	}

	if file.TotalLines == 0 && a.diffSource != "branch" {
		file.TotalLines = a.fileTotalLines(file.Path)
	}

	lang := languageOf(file.Path)
	if a.diffSplit {
		a.diffRows = diffRowsSplit(*file, lang)
	} else {
		a.diffRows = diffRowsUnified(*file, lang)
	}
	if a.diffSource == "branch" {
		rows := a.diffRows[:0]
		for _, row := range a.diffRows {
			if row.Kind != rowExpand {
				rows = append(rows, row)
			}
		}
		a.diffRows = rows
	}
}

func (a *App) readBranchChanges(ctx context.Context) (changeSet, error) {
	var changes changeSet
	if a.ws.Sample {
		changes = sampleChangeSet()
		changes.Note = "Bundled sample: combined changes from the sample branch commits."
		return changes, nil
	}
	base, err := defaultBranchRef(ctx, a.ws.Root)
	if err != nil {
		return changes, err
	}
	ref := a.historyRef
	if ref == "" {
		ref = "HEAD"
		for _, branch := range a.branches {
			if branch.CheckedOut {
				ref = branch.Ref
				break
			}
		}
	}
	return loadBranchChanges(ctx, a.ws.Root, base, ref)
}

func (a *App) showBranchChanges(ctx context.Context) {
	changes, err := a.readBranchChanges(ctx)
	if err != nil {
		a.showToast("error", "Branch changes unavailable", err.Error())
		a.refresh()
		return
	}
	a.branchChanges = changes
	a.branchPath = ""
	if len(changes.Files) > 0 {
		a.branchPath = changes.Files[0].Path
	}
	a.diffSource = "branch"
	a.firstChangePending = ""
	a.changesScroll.SetOffset(0, len(changes.Files), a.changeVisibleRows())
	a.resetScroll()
	a.rebuildDiffRows()
	a.refresh()
	a.jumpToFirstChange(a.branchPath)
}

func (a *App) refreshBranchChanges(ctx context.Context) {
	changes, err := a.readBranchChanges(ctx)
	if err != nil {
		a.showToast("error", "Branch changes unavailable", err.Error())
		return
	}
	a.branchChanges = changes
	if findFile(changes.Files, a.branchPath) == nil {
		a.branchPath = ""
		if len(changes.Files) > 0 {
			a.branchPath = changes.Files[0].Path
		}
	}
	a.rebuildDiffRows()
}

func (a *App) selectBranchFile(path string) {
	if findFile(a.branchChanges.Files, path) == nil {
		return
	}
	a.branchPath = path
	a.revealChangePath(path)
	a.resetScroll()
	a.rebuildDiffRows()
	a.refresh()
	a.jumpToFirstChange(path)
}

// pickBranch loads history for a ref without changing the checkout.
func (a *App) pickBranch(ctx context.Context, ref string) {
	if ref != "" {
		if _, ok := branchByRef(a.branches, ref); !ok {
			return
		}
	}

	history, cached := a.branchLogs.Get(ref)
	if !cached {
		next, err := loadCommitLog(ctx, a.ws.Root, ref)
		if err != nil {
			a.showToast("error", "Branch history unavailable", err.Error())
			a.refresh()

			return
		}

		history = next
		a.branchLogs.Put(ref, history)
	}

	a.historyRef = ref
	a.historyLabel = history.Branch
	a.commits = history
	a.commitOpen = ""
	a.commitFile = ""
	a.commitRows = nil
	a.commitPreview = nil
	a.historyLimit = maxCommits
	a.historyDone = len(history.Commits) < maxCommits
	a.historyScroll.SetOffset(0, len(history.Commits), a.historyVisibleRows())
	a.historyScroll.Limit = len(history.Commits)
	a.diffSource = "working"
	a.branchOpen = false
	a.branchQuery = ""
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
}

// selectDiff shows one file's working-tree diff.
func (a *App) selectDiff(path string) {
	if findFile(a.changes.Files, path) == nil {
		return
	}

	a.diffPath = path
	a.firstChangePending = ""
	a.changeNav = changeNavigation{}
	a.diffSource = "working"
	a.mode = "changes"
	a.inlineCard = inlineCard{}
	a.revealChangePath(path)
	if a.fileNav != nil {
		a.fileNav.Select(path)
	}
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
	a.jumpToFirstChange(path)
}

// showWorkingDiff returns the lower pane to the working tree.
func (a *App) showWorkingDiff() {
	a.firstChangePending = ""
	a.changeNav = changeNavigation{}
	a.diffSource = "working"
	a.revealChangePath(a.diffPath)
	a.rebuildDiffRows()
	a.resetScroll()
	a.refresh()
}

// toggleCommit expands one commit in the history rail.
func (a *App) toggleCommit(ctx context.Context, hash string) {
	a.firstChangePending = ""
	commit := a.findCommit(hash)
	if commit == nil {
		return
	}
	a.changeNav = changeNavigation{}

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
			a.revealChangePath(a.diffPath)
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
	a.changesScroll.SetOffset(0, len(commit.Files), a.changeVisibleRows())
	a.revealChangePath(a.commitFile)
	a.rebuildCommitRows()
	a.resetScroll()
	a.refresh()
	if !a.jumpToFirstChange(a.commitFile) && a.commitDetails[commitDetailKey(hash, a.commitFile)] == nil && a.commitFile != "" {
		a.firstChangePending = commitDetailKey(hash, a.commitFile)
	}
}

// selectCommitFile selects one file inside an expanded commit.
func (a *App) selectCommitFile(ctx context.Context, hash, path string) {
	a.firstChangePending = ""
	if a.findCommit(hash) == nil {
		return
	}
	a.changeNav = changeNavigation{}

	a.commitOpen = hash
	a.commitFile = path
	a.diffSource = "commit"
	a.revealChangePath(path)
	a.focusCommitPage(hash)
	a.ensureCommitFile(ctx, hash, path)
	a.rebuildCommitRows()
	a.resetScroll()
	a.refresh()
	if !a.jumpToFirstChange(path) && a.commitDetails[commitDetailKey(hash, path)] == nil && path != "" {
		a.firstChangePending = commitDetailKey(hash, path)
	}
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

// ensureCommitFile requests one commit patch. The worker reads git off the
// page loop; pollPatches stores the result when it arrives.
func (a *App) ensureCommitFile(_ context.Context, hash, path string) {
	if path == "" || a.ws.Sample {
		return
	}

	key := commitDetailKey(hash, path)
	if a.commitDetails[key] != nil {
		return
	}

	if a.patchLoader == nil {
		// This fallback exists for an App built by a test without New. The
		// worker lives for the app and Close cancels its in-flight read.
		a.patchLoader = NewCommitPatchLoader(a.ws.Root, maxCommitPatchCache) //nolint:contextcheck // app-lifetime patch worker
	}

	// The loader owns a process-lifetime context; this call is not tied to a
	// request, and Close cancels an in-flight read.
	a.patchLoader.Request(context.Background(), hash, path) //nolint:contextcheck // app-lifetime patch worker
}

// pollPatches drains finished commit patches on the page loop.
func (a *App) pollPatches(ctx context.Context) bool {
	if a.patchLoader == nil {
		return false
	}

	changed := false

	for {
		result, ok := a.patchLoader.Poll()
		if !ok {
			break
		}

		if result.Err != nil {
			if result.Key == a.firstChangePending {
				a.firstChangePending = ""
			}
			a.showToast("error", "Commit diff unavailable", result.Err.Error())
			changed = true

			continue
		}

		if result.File == nil {
			continue
		}

		a.commitDetails[result.Key] = result.File
		a.commitKeys = append(a.commitKeys, result.Key)

		for len(a.commitKeys) > maxCommitPatchCache {
			old := a.commitKeys[0]
			a.commitKeys = a.commitKeys[1:]
			delete(a.commitDetails, old)
		}

		changed = true
	}

	if changed {
		if a.diffSource == "commit" {
			a.rebuildCommitRows()
			if a.firstChangePending == commitDetailKey(a.commitOpen, a.commitFile) && a.commitDetails[a.firstChangePending] != nil {
				if a.scrollY == 0 {
					a.jumpToFirstChange(a.commitFile)
				}
				a.firstChangePending = ""
			}
			a.refresh()
		}

		_ = a.page.Redraw(ctx)
	}

	return changed
}

// focusCommitPage scrolls the history rail to the commit with hash.
func (a *App) focusCommitPage(hash string) {
	for i, commit := range a.commits.Commits {
		if commit.Hash == hash {
			a.revealHistoryIndex(i)

			return
		}
	}
}

// revealHistoryIndex scrolls the commit rail so one row is visible.
func (a *App) revealHistoryIndex(index int) {
	total := len(a.commits.Commits)
	visible := a.historyVisibleRows()
	start, end := a.historyScroll.Rows(total, visible)

	switch {
	case index < start:
		a.historyScroll.SetOffset(index, total, visible)
	case index >= end:
		a.historyScroll.SetOffset(index-visible+1, total, visible)
	}
}

// scrollHistory moves the commit rail by rows and may load the next batch.
func (a *App) scrollHistory(ctx context.Context, rows int) {
	a.historyScroll.ScrollBy(rows, len(a.commits.Commits), a.historyVisibleRows())
	a.maybeLoadMoreHistory(ctx)
	a.refresh()
}

// maybeLoadMoreHistory asks for a larger commit log when the rail crosses
// half of the commits it already has. The git read runs off the page loop.
func (a *App) maybeLoadMoreHistory(ctx context.Context) {
	a.startHistoryLoad(ctx)
}

// rebuildCommitRows renders the selected commit file's patch.
func (a *App) rebuildCommitRows() {
	a.commitRows = nil
	a.commitPreview = nil

	file := a.commitDetails[commitDetailKey(a.commitOpen, a.commitFile)]
	if file == nil {
		return
	}

	if file.TotalLines == 0 {
		file.TotalLines = a.fileTotalLines(file.Path)
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
	rows := a.diffDisplayRows()
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
		doc.ensurePreview(a.previewCols())
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
