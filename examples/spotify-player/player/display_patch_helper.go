package player

import "github.com/chinmay-sawant/ownframe"

type DisplayPatchHelper struct{ app *App }

func NewDisplayPatchHelper(app *App) *DisplayPatchHelper {
	return &DisplayPatchHelper{app: app}
}

func (h *DisplayPatchHelper) page() *ownframe.Page {
	if h == nil || h.app == nil {
		return nil
	}
	return h.app.page
}

func (h *DisplayPatchHelper) Report() ownframe.DisplayPatch {
	if p := h.page(); p != nil {
		return p.PatchReport()
	}
	return ownframe.DisplayPatch{}
}

func (h *DisplayPatchHelper) LastReport() ownframe.DisplayPatch { return h.Report() }

func (h *DisplayPatchHelper) PatchReport() ownframe.DisplayPatch { return h.Report() }

func (h *DisplayPatchHelper) PatchOp(i int, fn func(*ownframe.DisplayOp)) bool {
	if p := h.page(); p != nil && p.Display() != nil {
		return p.PatchOp(i, fn)
	}
	return false
}

func (h *DisplayPatchHelper) PatchDisplay(fn func(*ownframe.Display)) ownframe.DisplayPatch {
	if p := h.page(); p != nil {
		return p.PatchDisplay(fn)
	}
	return ownframe.DisplayPatch{}
}

func (a *App) DisplayPatchHelper() *DisplayPatchHelper { return NewDisplayPatchHelper(a) }

func (a *App) PatchReport() ownframe.DisplayPatch {
	return NewDisplayPatchHelper(a).Report()
}

func (a *App) PatchOp(i int, fn func(*ownframe.DisplayOp)) bool {
	return NewDisplayPatchHelper(a).PatchOp(i, fn)
}

func (a *App) PatchDisplay(fn func(*ownframe.Display)) ownframe.DisplayPatch {
	return NewDisplayPatchHelper(a).PatchDisplay(fn)
}

func DisplayPatchReport(app *App) ownframe.DisplayPatch {
	return NewDisplayPatchHelper(app).Report()
}

func PatchOpSafe(app *App, i int, fn func(*ownframe.DisplayOp)) bool {
	return NewDisplayPatchHelper(app).PatchOp(i, fn)
}
