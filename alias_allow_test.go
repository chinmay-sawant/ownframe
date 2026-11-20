package ownframe

func aliasAllow() map[string]bool {
	return aliasSet(`api.go:Box
api.go:Config
api.go:Handlers
api.go:MenuItem
clipboard.go:Secret
elements.go:Element
elements.go:ElementKind
elements.go:ElementLayout
elements.go:ElementRegistry
events.go:Drop
events.go:HoverHandler
fetch.go:FetchResponse
page.go:ElementEdit
page.go:ElementPatch
page_access.go:AccessibilityActionKind
page_access.go:AccessibilityBounds
page_access.go:AccessibilityNode
page_access.go:AccessibilityNodeID
page_access.go:AccessibilityRequest
page_access.go:AccessibilityRole
page_access.go:AccessibilityTree
page_access.go:Accessible
page_probe.go:Display
page_probe.go:DisplayOp
page_probe.go:DisplayPatch
page_probe.go:MotionStats
page_probe.go:Stats
page_scroll.go:ScrollbarMode
page_scroll.go:TouchScrollOptions
run.go:WindowOptions
serve.go:ServeOptions`)
}

func aliasSet(s string) map[string]bool {
	m := make(map[string]bool)
	for _, n := range splitAliasLines(s) {
		m[n] = true
	}
	return m
}
