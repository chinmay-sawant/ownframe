package ui

import "os"

// exec performs one backend call.
func (w *worker) exec(j job) result {
	r := result{kind: j.kind, gen: j.gen, sheet: j.sheet, area: j.area}

	switch j.kind {
	case jobFetch:
		r.cells, r.err = w.b.Range(j.sheet, j.area)
	case jobApply:
		r.rev, r.err = w.b.Apply(j.sheet, j.edits)
		r.edits = j.edits
	case jobUndo:
		r.und, r.err = w.b.Undo()
	case jobRedo:
		r.und, r.err = w.b.Redo()
	case jobUsed:
		r.area, r.ok = w.b.Used(j.sheet)
	case jobReadFile:
		r.data, r.err = os.ReadFile(j.path)
	case jobPreview:
		r.prev, r.err = w.b.PreviewCSV(j.sheet, j.data, j.flag)
	case jobCommitCSV:
		r.imp, r.err = w.b.CommitCSV(j.sheet, j.data, j.flag)
	case jobExport:
		r.path = j.path

		var text string
		if text, r.err = w.b.ExportCSV(j.sheet); r.err == nil {
			r.data = []byte(text)
			r.err = os.WriteFile(j.path, r.data, uiExportPerm) //nolint:gosec // reason: export path is the user-chosen dialog destination, writing there is intended
		}
	case jobPref:
		r.err = w.b.SetPref(j.key, j.value)
	}

	return r
}
