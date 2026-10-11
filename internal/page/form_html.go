package page

const formCSS = `<style>[data-ownframe-field]{display:inline-block;border:1px solid #c8c2b4;padding:4px 6px;min-width:10em;min-height:1.6em;white-space:pre-wrap}[data-ownframe-field][data-ownframe-focus="1"]{border:1px solid #1a56db;outline:1px solid #1a56db}[data-ownframe-selection]{background:#d6e2ff}[data-ownframe-composition]{text-decoration:underline}[data-ownframe-field][data-ownframe-placeholder="1"]{color:#6b7280}[data-ownframe-caret]{display:inline-block;width:0;height:1em;border-left:1px solid currentColor;margin-right:-1px;z-index:1}</style>`

// rewriteControls copies source, swapping each kept control for paintable
// HTML. It paints a select-all when selected is true and puts the caret at
// the end of the value. The page uses rewriteControlsState, which carries
// the caret and the range.
func rewriteControls(source string, spans []controlSpan, live map[string]Control, focusID string, selected bool) string {
	st := caretState{caret: -1}
	if focusID != "" {
		st.focus = focusID
		st.all = selected
	}

	return rewriteControlsState(source, spans, live, st)
}
