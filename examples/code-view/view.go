package main

// View is the data the code view template prints. Every field is a plain
// value or a prebuilt row, so one SetData call publishes one frame.
type View struct {
	Mode    string
	Filter  string
	Details bool
	DevOn   bool

	Workspace string
	Branch    string
	Changed   int
	DevNote   string

	Tree      []treeRow
	TreeMore  int
	TreeCount int
	Filtered  bool

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
	Welcome     bool
	Preview     bool
	Previewable bool
	PreviewRows []previewRow
	Empty       bool
	EmptyKind   string
	EmptyTitle  string
	EmptyBody   string
	Recent      []recentRow

	Diff         diffView
	Outline      []symbolRow
	Toast        toastView
	Palette      paletteView
	PaletteQuery string
}

// treeRow is one file or folder in the explorer.
type treeRow struct {
	Path        string
	Name        string
	Dir         bool
	Open        bool
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

// codeRow is one source line with its prebuilt spans.
type codeRow struct {
	No       int
	Segments []segment
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
	Files      []diffFileRow
	Selected   string
	Split      bool
	Status     string
	StatusCls  string
	Path       string
	Name       string
	Dir        string
	Lang       string
	Badge      string
	Add        int
	Del        int
	TotalAdd   int
	TotalDel   int
	Hunks      int
	Rows       []diffRow
	TrackTop   int
	HasFiles   bool
	Page       int
	Pages      int
	Sample     bool
	Note       string
	EmptyTitle string
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
