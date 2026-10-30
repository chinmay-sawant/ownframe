package main

import "strings"

// This file holds the per-component zoom rules. Every rule mirrors a fixed
// pixel value in components/*.css or a Go geometry constant, scaled by the
// layout zoom. Heights that Go scroll math depends on must match the a.zx()
// values or rows desync from hit testing.

// zoomRule scales one CSS rule: parts are the literal declaration chunks
// around each scaled pixel, and bases holds the 1x pixels in order. The
// bases mirror components/*.css 1:1; keeping them beside the selector
// preserves that mapping.
type zoomRule struct {
	sel   string
	parts []string
	bases []int
}

// writeZoomRules renders scaled rules into the theme override.
func writeZoomRules(b *strings.Builder, px func(int) string, rules []zoomRule) {
	for _, r := range rules {
		b.WriteString(r.sel)
		b.WriteString(" {")
		for i, part := range r.parts {
			b.WriteString(part)
			if i < len(r.bases) {
				b.WriteString(px(r.bases[i]))
			}
		}
		b.WriteString(" }")
	}
}

func zoomChromeCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		// Titlebar shell.
		{".titlebar", []string{" gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{8, 10, 14}},
		{".file-btn", []string{" height:", " !important; padding:0 ", " !important;"}, []int{28, 10}},
		// Brand block.
		{".brand", []string{" height:", " !important; gap:", " !important; padding:0 ", " !important;"}, []int{28, 8, 8}},
		{".brand-sub", []string{" padding-top:", " !important; margin-left:", " !important;"}, []int{1, 2}},
		{".brand-mark svg", []string{" width:", " !important; height:", " !important;"}, []int{19, 19}},
		// Command bar.
		{".cmdbar", []string{" flex-basis:", " !important; min-width:", " !important; height:", " !important; margin-left:", " !important; gap:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{430, 160, 26, 6, 7, 8, 6}},
		{".cmdbar svg", []string{" width:", " !important; height:", " !important;"}, []int{12, 12}},
		{".cmd-key", []string{" height:", " !important; min-width:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{16, 16, 4, 4}},
		// Window controls.
		{".win-controls", []string{" margin-right:-", " !important;"}, []int{10}},
		{".win-btn", []string{" width:", " !important;"}, []int{46}},
		{".win-btn svg", []string{" width:", " !important; height:", " !important;"}, []int{11, 11}},
		// File drawer and flyouts.
		{".file-drawer", []string{" left:", " !important; width:", " !important; padding:", " !important; border-radius:", " !important;"}, []int{14, 280, 6, 8}},
		{".file-item", []string{" min-height:", " !important; height:", " !important; line-height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{30, 30, 30, 10, 6}},
		{".project-input", []string{" height:", " !important; line-height:", " !important; padding:0 ", " !important;"}, []int{28, 28, 8}},
		{".locations-label", []string{" height:", " !important; line-height:", " !important;"}, []int{22, 22}},
		{".locations-list", []string{" max-height:", " !important;"}, []int{180}},
		{".recent-flyout", []string{" width:", " !important; padding:", " !important; border-radius:", " !important; max-height:", " !important;"}, []int{260, 6, 8, 220}},
		// Browse panel.
		{".browse-panel", []string{" top:", " !important; width:", " !important; margin-left:-", " !important; padding:", " !important; border-radius:", " !important;"}, []int{70, 560, 280, 10, 10}},
		{"#browse-panel", []string{" left:calc(50% - ", ") !important;"}, []int{280}},
		{".browse-path", []string{" padding:0 ", " !important;"}, []int{4}},
		{".browse-list", []string{" max-height:", " !important;"}, []int{180}},
		{".browse-panel .browse-list", []string{" max-height:", " !important;"}, []int{320}},
		// Alerts and toast chrome.
		{".file-alert", []string{" right:", " !important; width:", " !important; padding:", " ", " ", " ", " !important; border-radius:", " !important;"}, []int{14, 300, 8, 8, 8, 10, 8}},
		{".file-alert-x", []string{" min-height:", " !important; height:", " !important; line-height:", " !important;"}, []int{24, 24, 24}},
		{".toast-layer", []string{" left:", " !important;"}, []int{16}},
		{".toast", []string{" width:", " !important; min-height:", " !important; gap:", " !important; padding:", " ", " !important; border-radius:", " !important; border-left-width:", " !important;"}, []int{322, 52, 9, 10, 11, 8, 3}},
		{".toast-icon", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{20, 20, 6}},
	})
}

func zoomTreeCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".side-head", []string{" height:", " !important; flex-basis:", " !important; gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{34, 34, 8, 8, 14}},
		{".filter", []string{" height:", " !important; margin:", " ", " ", " !important; gap:", " !important; padding:0 ", " !important;"}, []int{28, 8, 9, 6, 7, 8}},
		{".filter-icon, .filter-icon svg", []string{" width:", " !important; height:", " !important;"}, []int{12, 12}},
		{".filter-input", []string{" height:", " !important; line-height:", " !important;"}, []int{26, 26}},
		{".filter-x", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{18, 18, 3}},
		{".tree", []string{" padding:", " 0 ", " !important;"}, []int{1, 4}},
		{".trow", []string{" gap:", " !important; padding-right:", " !important; border-left-width:", " !important;"}, []int{7, 10, 2}},
		{".chev, .chev svg", []string{" width:", " !important; height:", " !important; flex-basis:", " !important;"}, []int{10, 10, 10}},
		{".chev", []string{" margin-right:", " !important;"}, []int{8}},
		{".ficon", []string{" width:", " !important; height:", " !important; flex-basis:", " !important; border-radius:", " !important;"}, []int{18, 18, 18, 4}},
		{".ficon.sm", []string{" width:", " !important; height:", " !important; flex-basis:", " !important;"}, []int{16, 16, 16}},
		{".tree-note", []string{" padding:0 ", " !important; border-radius:", " !important; line-height:", " !important;"}, []int{4, 3, 13}},
		{".tname", []string{" max-height:", " !important; line-height:", " !important;"}, []int{26, 13}},
		{".fstat", []string{" width:", " !important; height:", " !important; flex-basis:", " !important; padding:", " 0 !important; border-radius:", " !important;"}, []int{16, 16, 16, 1, 3}},
		// Scrollbar rail and thumb.
		{".tree-scroll", []string{" top:", " !important; bottom:", " !important; width:", " !important;"}, []int{1, 4, 8}},
		{".tree-thumb", []string{" left:", " !important; width:", " !important; min-height:", " !important; border-radius:", " !important;"}, []int{1, 6, 18, 3}},
		{".tree-tag", []string{" padding:", " ", " ", " !important;"}, []int{5, 13, 7}},
		{".tree-more", []string{" margin:", " ", " 0 !important; padding:", " ", " !important; border-radius:", " !important;"}, []int{6, 10, 7, 8, 5}},
		{".side-foot", []string{" height:", " !important; flex-basis:", " !important; gap:", " !important; padding:0 ", " !important;"}, []int{32, 32, 9, 14}},
		{".side-btn", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{20, 20, 4}},
		{".side-tools", []string{" gap:", " !important;"}, []int{6}},
		{"#side-collapse svg", []string{" width:", " !important; height:", " !important;"}, []int{13, 13}},
		{"#side-refresh svg", []string{" width:", " !important; height:", " !important;"}, []int{12, 12}},
		// Changes list.
		{".changes-head", []string{" height:", " !important; flex-basis:", " !important; gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{28, 28, 6, 10, 12}},
		{".change", []string{" gap:", " !important; padding:", " ", " ", " ", " !important; border-left-width:", " !important;"}, []int{7, 4, 16, 4, 8, 2}},
		{".change-name", []string{" line-height:", " !important; max-height:", " !important;"}, []int{13, 26}},
		{".change-dir", []string{" line-height:", " !important; max-height:", " !important;"}, []int{11, 22}},
		{".change-num", []string{" width:", " !important; flex-basis:", " !important; gap:", " !important;"}, []int{30, 30, 1}},
		{".cstat", []string{" min-width:", " !important; height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{16, 16, 4, 4}},
		{".sc-stats", []string{" gap:", " !important;"}, []int{7}},
		{".changes-count, .side-tag", []string{" padding:0 ", " !important; border-radius:", " !important;"}, []int{4, 3}},
	})
}

func zoomEditorCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		// Tabs and close controls.
		{".tab-x", []string{" width:", " !important; height:", " !important; top:", " !important; right:", " !important; border-radius:", " !important;"}, []int{21, 21, 8, 4, 4}},
		{".tab-tool, .tab-nav", []string{" width:", " !important; height:", " !important; flex-basis:", " !important; border-radius:", " !important; margin-right:", " !important;"}, []int{26, 26, 26, 5, 3}},
		{".tab-tool svg, .tab-nav svg", []string{" width:", " !important; height:", " !important;"}, []int{13, 13}},
		{".tab-dot", []string{" width:", " !important; height:", " !important;"}, []int{6, 6}},
		{".tab-status", []string{" width:", " !important; height:", " !important;"}, []int{16, 16}},
		// Crumb bar and pills.
		{".crumbs", []string{" gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{8, 10, 13}},
		{".crumb-btn", []string{" height:", " !important; border-radius:", " !important; padding:0 ", " ", " !important;"}, []int{20, 4, 6, 2}},
		{".seg", []string{" height:", " !important; border-radius:", " !important; margin:0 ", " !important;"}, []int{22, 5, 5}},
		{".seg-btn", []string{" height:", " !important; padding:0 ", " 2px !important;"}, []int{20, 8}},
		{".lang-pill", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{16, 6, 4}},
		{".crumb-path, .diff-path, .file-path", []string{" line-height:", " !important;"}, []int{20}},
		{".crumb-meta", []string{" padding-bottom:", " !important;"}, []int{2}},
		// Find bar.
		{".find-input", []string{" width:", " !important; min-width:", " !important; height:", " !important; line-height:", " !important; border-radius:", " !important;"}, []int{190, 80, 21, 21, 4}},
		{".find-count", []string{" height:", " !important;"}, []int{21}},
		{".find-btn", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{20, 20, 4}},
		// Gutter and code geometry.
		{".gutter-mark", []string{" width:", " !important; flex-basis:", " !important;"}, []int{16, 16}},
		{".ln", []string{" width:", " !important; flex-basis:", " !important; padding-right:", " !important; line-height:var(--row-h) !important;"}, []int{44, 44, 10}},
		{".cline", []string{" padding-right:", " !important;"}, []int{22}},
		{".edit-caret", []string{" top:", " !important; width:", " !important; height:", " !important;"}, []int{2, 2, 13}},
		{".edit-surface, .edit-line-hit", []string{" left:", " !important;"}, []int{60}},
		{".edit-area", []string{" bottom:", " !important; padding:", " ", " !important;"}, []int{20, 6, 12}},
		{".edit-hint", []string{" height:", " !important; line-height:", " !important; padding:0 ", " !important;"}, []int{20, 20, 12}},
		{".edit-flag", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{16, 6, 4}},
		{".expand-btn", []string{" height:", " !important; line-height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{16, 14, 6, 3}},
		// Editor scroll rail and context menu.
		{".editor-scroll-thumb", []string{" right:", " !important; width:", " !important; border-radius:", " !important;"}, []int{8, 8, 4}},
		{".ctx-menu", []string{" min-width:", " !important; padding:", " !important; border-radius:", " !important;"}, []int{188, 4, 8}},
		{".ctx-item", []string{" height:", " !important; line-height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{26, 26, 10, 5}},
		{".keys em", []string{" min-width:", " !important; height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{24, 22, 4, 3}},
		// Welcome and empty states.
		{".welcome-card", []string{" width:", " !important; padding:", " ", " ", " !important; border-radius:", " !important; gap:", " !important;"}, []int{548, 28, 30, 22, 8, 12}},
		{".empty-mark", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{34, 34, 8}},
		{".empty .button", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{28, 12, 6}},
	})
}

func zoomRailsCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".list-scroll, .hist-scroll", []string{" width:", " !important;"}, []int{8}},
		{".list-thumb, .hist-thumb", []string{" width:", " !important; min-height:", " !important; border-radius:", " !important;"}, []int{8, 24, 4}},
		{".list-scroll svg, .hist-scroll svg, .tree-scroll svg, .editor-scroll svg", []string{" width:", " !important;"}, []int{8}},
		{".drawer-resize", []string{" width:", " !important;"}, []int{6}},
		{".drawer-resize svg", []string{" width:", " !important;"}, []int{6}},
		{".editor-scroll", []string{" width:", " !important;"}, []int{8}},
		{".search-scroll-track", []string{" flex-basis:", " !important; width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{8, 8, 352, 4}},
		{".search-thumb", []string{" width:", " !important; border-radius:", " !important;"}, []int{8, 4}},
		{".search-row", []string{" height:", " !important; min-height:", " !important; max-height:", " !important; padding:", " 0 !important;"}, []int{44, 44, 44, 5}},
		{".search-loc", []string{" width:", " !important; flex-basis:", " !important;"}, []int{48, 48}},
		{".search-line", []string{" height:", " !important; line-height:", " !important;"}, []int{14, 14}},
		{".search-text", []string{" max-height:", " !important;"}, []int{34}},
		{".search-list", []string{" max-height:", " !important;"}, []int{362}},
	})
}

func zoomHistoryCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".hist-head", []string{" height:", " !important; flex-basis:", " !important; gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{32, 32, 8, 10, 14}},
		{".history-current", []string{" height:", " !important; flex-basis:", " !important; gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{26, 26, 8, 10, 8}},
		{".branch-panel", []string{" padding:", " 0 !important;"}, []int{4}},
		// Compact side-history rail: qualified rules beat the generic ones.
		{".side-history .hist-head", []string{" gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{4, 6, 10}},
		{".side-history .hist-row", []string{" gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{8, 6, 7}},
		{".side-history .hist-hash", []string{" width:", " !important; flex-basis:", " !important;"}, []int{42, 42}},
		{".side-history .hist-meta", []string{" padding-left:", " !important;"}, []int{17}},
		{".side-history .hist-file", []string{" gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{5, 6, 17}},
		{".side-history .branch-filter", []string{" padding:0 ", " ", " ", " !important;"}, []int{10, 4, 12}},
		{".hist-row", []string{" gap:", " !important; padding:0 ", " 0 ", " !important;"}, []int{12, 10, 8}},
		{".hist-dot", []string{" width:", " !important; height:", " !important; border-width:", " !important;"}, []int{7, 7, 1}},
		{".hist-hash", []string{" width:", " !important;"}, []int{52}},
		{".hist-subject", []string{" max-height:calc(", " + ", ") !important;"}, []int{13, 13}},
		{".hist-line", []string{" line-height:", " !important;"}, []int{13}},
		{".hist-when, .hist-stat", []string{" gap:", " !important;"}, []int{5}},
		{".hist-chev", []string{" width:", " !important; flex-basis:", " !important;"}, []int{10, 10}},
		{".hist-meta", []string{" height:", " !important; line-height:", " !important; padding:0 ", " 0 ", " !important;"}, []int{22, 22, 12, 23}},
		{".hist-file", []string{" min-height:", " !important; gap:", " !important; padding:", " ", " ", " ", " !important;"}, []int{30, 7, 3, 10, 3, 23}},
		{".hist-graph", []string{" width:", " !important; height:", " !important; flex-basis:", " !important;"}, []int{42, 42, 42}},
		{".hist-graph svg", []string{" width:", " !important; height:", " !important; display:block !important;"}, []int{42, 42}},
		{".hist-btn", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{19, 6, 4}},
		// Branch picker.
		{".branch-filter", []string{" height:", " !important; flex-basis:", " !important; gap:", " !important; padding:0 ", " ", " ", " !important;"}, []int{26, 26, 5, 10, 4, 12}},
		{".branch-input", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{22, 7, 4}},
		{".branch-row", []string{" height:", " !important; gap:", " !important; padding:0 ", " !important;"}, []int{26, 7, 12}},
		{".branch-mark", []string{" width:", " !important; flex-basis:", " !important;"}, []int{9, 9}},
		{".branch-name", []string{" max-height:", " !important;"}, []int{24}},
		{".branch-track", []string{" max-height:", " !important;"}, []int{22}},
		{".branch-more", []string{" padding:", " ", " ", " ", " !important;"}, []int{4, 12, 6, 28}},
		{".sidebar .branch-actions", []string{" gap:", " !important; padding:0 ", " ", " ", " !important;"}, []int{4, 10, 4, 12}},
		// Commit tip float.
		{".commit-tip-float", []string{" max-width:", " !important; max-height:", " !important; padding:", " ", " !important; border-radius:", " !important;"}, []int{340, 160, 6, 10, 7}},
		{".commit-tip-float .tip-subject, .commit-tip-float b, .commit-tip-float span, .commit-tip-float em", []string{" line-height:", " !important;"}, []int{14}},
	})
}

func zoomDiffCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".dnum", []string{" width:", " !important; flex-basis:", " !important; padding-right:", " !important;"}, []int{46, 46, 9}},
		{".dsign", []string{" width:", " !important; flex-basis:", " !important;"}, []int{20, 20}},
		{".dcell", []string{" padding-right:", " !important;"}, []int{18}},
		{".drow.split .dnum", []string{" width:", " !important; flex-basis:", " !important;"}, []int{44, 44}},
		{".drow.split .dcell", []string{" padding-right:", " !important;"}, []int{12}},
		{".drow.current .dnum:first-child", []string{" border-left-width:", " !important;"}, []int{2}},
		{".dcell + .dnum", []string{" border-left-width:", " !important;"}, []int{1}},
		{".dhunk, .dmeta", []string{" padding-left:", " !important;"}, []int{12}},
		{".em-a, .em-d", []string{" border-radius:", " !important;"}, []int{2}},
		{".dcell-expand", []string{" gap:", " !important; padding:0 ", " !important;"}, []int{8, 8}},
	})
}

func zoomFloatsCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		// Bottom-center navigator.
		{".file-nav", []string{" left:calc(50% - ", ") !important; width:", " !important; height:", " !important; gap:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{210, 420, 34, 2, 6, 10}},
		{".file-nav-arrow", []string{" width:", " !important; height:", " !important; flex-basis:", " !important; border-radius:", " !important;"}, []int{26, 26, 26, 6}},
		{".file-nav-change", []string{" width:", " !important; flex-basis:", " !important;"}, []int{20, 20}},
		{".file-nav-current", []string{" width:", " !important; flex-basis:", " !important; height:", " !important; line-height:", " !important;"}, []int{150, 150, 18, 18}},
		{".file-nav-dir", []string{" width:", " !important; flex-basis:", " !important; height:", " !important; line-height:", " !important;"}, []int{60, 60, 18, 18}},
		// Inline card.
		{".inline-card", []string{" max-height:", " !important; padding:", " ", " ", " !important; border-radius:", " !important;"}, []int{220, 8, 10, 7, 8}},
		// Narrow windows keep the pill centered with scaled geometry.
		{"@media (max-width: 900px) {.file-nav", []string{" left:calc(50% - ", ") !important; width:", " !important;}"}, []int{160, 320}},
		{"@media (max-width: 900px) {.file-nav-current", []string{" width:", " !important; flex-basis:", " !important;}"}, []int{110, 110}},
		{"@media (max-width: 352px) {.file-nav", []string{" left:", " !important; width:calc(100vw - ", ") !important;}"}, []int{16, 32}},
		{"@media (max-width: 352px) {.file-nav-current", []string{" width:", " !important; flex-basis:", " !important;}"}, []int{60, 60}},
	})
}

func zoomActivityCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".activity", []string{" gap:", " !important; padding:", " 0 !important;"}, []int{4, 8}},
		{".act", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{34, 34, 8}},
		{".act svg", []string{" width:", " !important; height:", " !important;"}, []int{19, 19}},
		{".act.on::before", []string{" left:-", " !important; top:", " !important; bottom:", " !important; width:", " !important; border-radius:", " !important;"}, []int{6, 8, 8, 2, 2}},
		{".act-badge", []string{" top:", " !important; right:", " !important; min-width:", " !important; height:", " !important; padding:0 ", " !important; border-radius:999px !important;"}, []int{2, 1, 14, 14, 3}},
	})
}

func zoomOverlayCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		// Command palette.
		{".palette", []string{" top:", " !important; width:", " !important; margin-left:-", " !important; border-radius:", " !important;"}, []int{70, 600, 300, 10}},
		{"#palette-panel", []string{" left:calc(50% - ", ") !important;"}, []int{300}},
		{".palette-search", []string{" height:", " !important; gap:", " !important; padding:0 ", " !important;"}, []int{44, 8, 12}},
		{".palette-input", []string{" height:", " !important; padding:0 ", " !important; line-height:", " !important;"}, []int{30, 6, 28}},
		{".palette-close", []string{" width:", " !important; height:", " !important; border-radius:", " !important;"}, []int{22, 22, 4}},
		{".palette-row", []string{" height:", " !important; gap:", " !important; padding:0 ", " !important;"}, []int{38, 9, 11}},
		{".palette-text", []string{" width:", " !important;"}, []int{510}},
		{".search-panel", []string{" top:", " !important;"}, []int{56}},
		{".search-meta", []string{" height:", " !important; line-height:", " !important; padding:0 ", " !important;"}, []int{24, 24, 12}},
		{".palette-foot", []string{" height:", " !important; padding:0 ", " !important;"}, []int{28, 12}},
		{".palette-empty", []string{" padding:", " ", " !important;"}, []int{22, 12}},
		// Browse panel is in zoomChromeCSS. Settings panel.
		{".settings-panel", []string{" top:", " !important; left:calc(50% - ", ") !important; width:", " !important; border-radius:", " !important; max-width:calc(100vw - ", ") !important;"}, []int{50, 360, 720, 10, 40}},
		{".settings-nav-active", []string{" padding:", " ", " !important; border-radius:", " !important;"}, []int{9, 10, 4}},
		{".settings-caption", []string{" margin:", " 0 ", " !important;"}, []int{5, 12}},
		{".settings-theme", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important; margin-bottom:", " !important;"}, []int{48, 10, 4, 5}},
		{".settings-header", []string{" height:", " !important; padding:0 ", " !important;"}, []int{48, 16}},
		{".settings-nav", []string{" width:", " !important; flex-basis:", " !important; padding:", " ", " !important;"}, []int{180, 180, 16, 10}},
		{".settings-content", []string{" padding:", " ", " ", " !important;"}, []int{18, 20, 20}},
	})
}

func zoomDetailsCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".det-head", []string{" height:", " !important; flex-basis:", " !important; padding:0 ", " 0 ", " !important;"}, []int{34, 34, 8, 16}},
		{".det-file", []string{" gap:", " !important; padding:", " ", " !important;"}, []int{10, 14, 16}},
		{".ficon.lg", []string{" width:", " !important; height:", " !important; flex-basis:", " !important; border-radius:", " !important;"}, []int{30, 30, 30, 7}},
		{".det-grid", []string{" grid-template-columns:", " 1fr !important; gap:", " !important; padding:", " ", " !important;"}, []int{76, 10, 14, 16}},
		{".det-grid > span, .det-grid > b", []string{" padding:", " 0 !important;"}, []int{3}},
		{".det-sub", []string{" padding:", " ", " ", " !important;"}, []int{10, 16, 6}},
		{".outline", []string{" padding:", " 0 ", " !important;"}, []int{4, 8}},
		{".orow", []string{" gap:", " !important; padding:0 ", " !important;"}, []int{8, 16}},
		{".okind", []string{" width:", " !important; flex-basis:", " !important;"}, []int{26, 26}},
		{".det-empty", []string{" padding:", " !important;"}, []int{16}},
		// Status bar.
		{".statusbar", []string{" padding:0 ", " !important; gap:", " !important;"}, []int{4, 2}},
		{".sb-item", []string{" gap:", " !important; padding:0 ", " !important;"}, []int{7, 10}},
		{".sb-dot", []string{" width:", " !important; height:", " !important; flex-basis:", " !important;"}, []int{5, 5, 5}},
		{".statusbar svg", []string{" width:", " !important; height:", " !important;"}, []int{10, 11}},
		{".scan-badge", []string{" height:", " !important; padding:0 ", " !important; border-radius:", " !important;"}, []int{15, 5, 4}},
	})
}

func zoomPreviewCSS(b *strings.Builder, px func(int) string) {
	writeZoomRules(b, px, []zoomRule{
		{".prow", []string{" padding:0 ", " !important;"}, []int{22}},
		{".pln", []string{" width:", " !important; flex-basis:", " !important; padding-right:", " !important;"}, []int{44, 44, 10}},
		{".pmark", []string{" width:", " !important; flex-basis:", " !important; padding-right:", " !important;"}, []int{18, 18, 7}},
		{".prow-quote", []string{" border-left-width:", " !important;"}, []int{2}},
	})
}
