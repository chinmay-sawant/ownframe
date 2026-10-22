package ui

// jobKind is one worker request.
type jobKind uint8

const (
	jobFetch jobKind = iota
	jobApply
	jobUndo
	jobRedo
	jobUsed
	jobReadFile
	jobPreview
	jobCommitCSV
	jobExport
	jobPref
)

// job is one request for the backend worker.
type job struct {
	kind  jobKind
	gen   uint64
	sheet string
	area  Area
	edits []Edit
	data  []byte
	path  string
	key   string
	value string
	flag  bool
}

// result is one answer from the backend worker.
type result struct {
	kind  jobKind
	gen   uint64
	sheet string
	area  Area
	ok    bool
	cells []Cell
	edits []Edit
	rev   uint64
	und   UndoResult
	prev  Preview
	imp   ImportResult
	data  []byte
	path  string
	err   error
}
