package main

const editorScrollMinThumb = 24

type editorScrollView struct {
	Show        bool
	Top, Height int
}

type editorScrollDrag struct {
	active                      bool
	group, start, limit, travel int
	delta                       float64
}

func (a *App) editorScroll(group int) editorScrollView {
	track := a.bodyHeight()
	limit, offset := a.pageScrollLimit(), a.scrollY
	if a.split && a.mode == "explorer" {
		limit, offset = a.maxGroupScroll(group), a.scrollGroup[group]
	}
	if limit <= 0 || track <= 0 {
		return editorScrollView{}
	}
	height := max(editorScrollMinThumb, track*track/(track+limit))
	height = min(height, track)
	top := (track - height) * min(max(offset, 0), limit) / limit
	return editorScrollView{Show: true, Top: top, Height: height}
}

func (a *App) startEditorScrollDrag(group int) bool {
	bar := a.editorScroll(group)
	limit, offset := a.pageScrollLimit(), a.scrollY
	if a.split && a.mode == "explorer" {
		limit, offset = a.maxGroupScroll(group), a.scrollGroup[group]
	}
	travel := a.bodyHeight() - bar.Height
	if !bar.Show || travel <= 0 {
		return false
	}
	a.editorDrag = editorScrollDrag{active: true, group: group, start: offset, limit: limit, travel: travel}
	return true
}

func (a *App) moveEditorScrollDrag(dy float64) {
	a.editorDrag.delta += dy
	d := a.editorDrag
	target := d.start + int(d.delta*float64(d.limit)/float64(d.travel))
	if a.split && a.mode == "explorer" {
		a.scrollGroupTo(d.group, target)
		return
	}
	a.scrollTo(target)
}
