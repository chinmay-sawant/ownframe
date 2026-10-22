package page

func synonymAllow() map[string]bool {
	return splitNames("DisplayPatch,ElementRegistry,FocusNext,FocusPrev,LastFocusID,MarkRect,PNG,PatchReport,PressedID,Redo,Redraw,RedrawIfChanged,SelectLineAt,SelectWordAt,Submit,Undo,WantFrameIn")
}

func splitNames(s string) map[string]bool {
	m := make(map[string]bool)
	for _, n := range splitComma(s) {
		m[n] = true
	}
	return m
}
