package ownframe

func aliasAllow() map[string]bool {
	return aliasSet(`accessibility.go:AccessibilityActionKind
accessibility.go:AccessibilityBounds
accessibility.go:AccessibilityNode
accessibility.go:AccessibilityNodeID
accessibility.go:AccessibilityRequest
accessibility.go:AccessibilityRole
accessibility.go:AccessibilityTree
accessibility.go:Accessible
api.go:Box
api.go:Config
api.go:Handlers
api.go:MenuItem
display.go:Display
display.go:DisplayOp
display_patch.go:DisplayPatch
drop.go:Drop
element_edit.go:ElementEdit
elements.go:Element
elements.go:ElementKind
elements.go:ElementLayout
elements.go:ElementRegistry
fetch.go:FetchResponse
hover.go:HoverHandler
page_probe.go:MotionStats
scrollbar.go:ScrollbarMode
secret.go:Secret
serve.go:ServeOptions
stats.go:Stats
touch_scroll.go:TouchScrollOptions
window_options.go:WindowOptions`)
}

func aliasSet(s string) map[string]bool {
	m := make(map[string]bool)
	for _, n := range splitAliasLines(s) {
		m[n] = true
	}
	return m
}
