package main

// View is the data the code view template prints. Every field is a plain
// value or a prebuilt row, so one SetData call publishes one frame.
type View struct {
	Mode    string
	Filter  string
	Details bool
	DevOn   bool
	Modal   bool

	Workspace   string
	Branch      string
	BranchQuery string
	Changed     int
	DevNote     string
	ScanBudget  string
	WatchNote   string
	HoverPath   string

	Tree         []treeRow
	TreeCount    int
	TreeTotal    int
	TreeStart    int
	TreeVisible  int
	TreeScroll   bool
	TreeThumbTop int
	TreeThumbH   int
	TreeTrackH   int
	Filtered     bool

	Tabs        []tabRow
	ActiveTab   int
	Active      string
	ActiveDir   string
	ActiveName  string
	Lang        string
	LangLabel   string
	Badge       string
	LineCount   int
	ByteSize    string
	Encoding    string
	Indent      string
	Status      string
	StatusClass string
	Truncated   bool

	Rows        []codeRow
	TrackTop    int
	SpaceHeight int
	Wrap        bool
	FindQuery   string
	FindOpen    bool
	FindCase    bool
	FindCurrent int
	FindTotal   int
	Welcome     bool
	Preview     bool
	Previewable bool
	PreviewRows []previewRow
	Empty       bool
	EmptyKind   string
	EmptyTitle  string
	EmptyBody   string
	Recent      []recentRow

	Editing   bool
	EditText  string
	EditDirty bool
	CanEdit   bool
	SaveNote  string
	Dirty     int

	Split      bool
	FocusGroup int
	PaneLeft   int
	PaneWidth  int
	Second     secondaryView

	Diff         diffView
	History      historyView
	Outline      []symbolRow
	Toast        toastView
	Palette      paletteView
	PaletteQuery string
	SearchQuery  string
	Search       searchView
	FileNav      fileNavView
	ChangeLabel  string
	CommitTip    commitFloatView
	InlineCard   inlineCard
}

// commitFloatView is the page-space commit message card that floats to the
// right of the history rail.
type commitFloatView struct {
	Show  bool
	Left  int
	Top   int
	Width int
	Lines []string
}

// fileNavView is the bottom-center changed-file navigator.
type fileNavView struct {
	Open     bool
	Label    string
	Path     string
	Name     string
	Dir      string
	Position int
	Total    int
	Add      int
	Del      int
	Multiple bool
}

// secondaryView is the right-hand editor group. It carries only the fields
// that pane draws; the left group keeps the top-level View fields.
type secondaryView struct {
	Open        bool
	Left        int
	Width       int
	Tabs        []tabRow
	ActiveTab   int
	Active      string
	ActiveDir   string
	ActiveName  string
	Lang        string
	LangLabel   string
	Badge       string
	LineCount   int
	ByteSize    string
	Status      string
	StatusClass string
	Truncated   bool

	Rows        []codeRow
	TrackTop    int
	Welcome     bool
	Preview     bool
	Previewable bool
	PreviewRows []previewRow
	Empty       bool
	EmptyKind   string
	EmptyTitle  string
	EmptyBody   string

	Editing   bool
	EditText  string
	EditDirty bool
	CanEdit   bool
}

// searchView is the workspace content search panel.
type searchView struct {
	Open      bool
	Query     string
	Case      bool
	Index     int
	Files     int
	Note      string
	Truncated bool
	Busy      bool
	Rows      []searchRow
}

// searchRow is one content-search match.
type searchRow struct {
	Index    int
	Path     string
	Name     string
	Dir      string
	Line     int
	Col      int
	Before   string
	Match    string
	After    string
	Selected bool
}

// treeRow is one file or folder in the explorer.
type treeRow struct {
	Path        string
	Name        string
	Dir         bool
	Open        bool
	Ignored     bool
	Indent      int
	Lang        string
	Badge       string
	Status      string
	StatusClass string
}

// tabRow is one open editor tab.
type tabRow struct {
	Path        string
	Name        string
	Lang        string
	Badge       string
	Status      string
	StatusClass string
	Active      bool
}

// codeRow is one visual source row with its prebuilt spans. No is zero on a
// wrapped continuation row.
type codeRow struct {
	No          int
	Line        int
	Start       int
	Segments    []segment
	Mark        string
	MarkClass   string
	Hint        string
	ChangeClass string
	ChangeLine  int
	Caret       bool
	CaretLeft   int
}

// recentRow is one quick-open entry on the welcome panel.
type recentRow struct {
	Path  string
	Name  string
	Dir   string
	Lang  string
	Badge string
}

// diffView is the Source Control panel and the diff header.
type diffView struct {
	Files         []diffFileRow
	Selected      string
	Split         bool
	Status        string
	StatusCls     string
	Path          string
	Name          string
	Dir           string
	Lang          string
	Badge         string
	Add           int
	Del           int
	TotalAdd      int
	TotalDel      int
	Hunks         int
	Rows          []diffRow
	TrackTop      int
	HasFiles      bool
	HasChanges    bool
	Page          int
	Pages         int
	Scroll        bool
	ScrollTop     int
	ScrollHeight  int
	ScrollTrack   int
	RailThumbTop  int
	ThumbCells    int
	AtTop         bool
	AtEnd         bool
	Sample        bool
	Note          string
	EmptyTitle    string
	Source        string
	CommitHash    string
	CommitShort   string
	CommitSubject string
	CommitAuthor  string
	CommitDate    string
	Loading       bool
}

// historyView is the commit rail above the diff.
type historyView struct {
	Visible      bool
	Sample       bool
	Height       int
	Branch       string
	HeadBranch   string
	Browsing     bool
	Detached     bool
	Branches     []branchRow
	BranchOpen   bool
	BranchQuery  string
	BranchMore   int
	Commits      []commitRow
	Page         int
	Pages        int
	Tip          commitTipView
	Note         string
	Unborn       bool
	Loading      bool
	Scroll       bool
	ScrollTop    int
	ScrollHeight int
	ScrollTrack  int
	AtTop        bool
	AtEnd        bool
	LoadingMore  bool
	RailTop      int
	RailThumbTop int
	ThumbCells   int
}

// commitRow is one commit in the history rail.
type commitRow struct {
	Hash        string
	Short       string
	Subject     string
	Author      string
	Date        string
	Body        string
	Add         int
	Del         int
	Files       []commitFileRow
	More        int
	Expanded    bool
	Selected    bool
	Head        bool
	Preview     []diffRow
	PreviewPath string
	PreviewErr  string
	Graph       []graphCell
	Lanes       int
}

// commitFileRow is one changed file inside an expanded commit.
type commitFileRow struct {
	Path        string
	Name        string
	Dir         string
	Status      string
	StatusClass string
	Lang        string
	Badge       string
	Add         int
	Del         int
	Selected    bool
}

// commitTipView is the hover card for one commit.
type commitTipView struct {
	Show    bool
	Top     int
	Subject string
	Body    string
	Meta    string
}

// diffFileRow is one changed file in the source control list.
type diffFileRow struct {
	Path        string
	Name        string
	Dir         string
	Status      string
	StatusClass string
	Lang        string
	Badge       string
	Add         int
	Del         int
	Selected    bool
}

// symbolRow is one outline entry in the details panel.
type symbolRow struct {
	Name  string
	Kind  string
	Line  int
	No    int
	Depth int
}

// toastView is the transient message card.
type toastView struct {
	Show    bool
	Kind    string
	Title   string
	Message string
}
