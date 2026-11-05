package main

// searchCSS restyles the quick-switcher palette built by searchOverlay in
// overlays.go as Discord's Ctrl+K panel: a dark flat card with the query
// input on top, grouped channel/message rows, and a footer hint row.
//
// Wiring: append after appCSS in themeSheet (app.go), so these rules win
// over styles.go on equal specificity:
//
//	return chadcn.Dark + chadcn.Stylesheet + appCSS + searchCSS
//
// Class inventory (all markup already exists; this sheet only repaints it):
//
//	#dc-overlays       overlay root from shell.go overlays(); stretched over
//	                   the window so the absolutely positioned panel,
//	                   dismiss layer, dialog, sheet, menu, and toast below
//	                   resolve against the viewport instead of the root's
//	                   end-of-flow position (without this the panel's
//	                   top:72px lands a full viewport below the fold and the
//	                   dismiss layer collapses to zero height).
//	#dc-search-panel   absolute wrapper (inline styles set its 560px width,
//	                   top, and z-index; paint only here)
//	.dc-search-head    header row: title plus the search:close control
//	.dc-search-field   wrapper around the #dc-search Field input
//	.dc-search-sections group stack holding the Channels/Messages labels,
//	                   result lists, separators, and FieldDescription notes
//	.dc-search-results  one result list per group (rows dispatch
//	                   chan:select:<id> / msg:jump:<id>)
//	.dc-search-name    row primary line (channel name or message snippet)
//	.dc-search-detail  row secondary line (server/topic or channel/time)
//
// chadcn base paint overridden below (scoped under #dc-search-panel):
// card.css .of-card/.of-card-header/.of-card-title/.of-card-content
// (light surface + 14px radius), field.css .of-field/.of-field-label /
// .of-field-description, input.css .of-input, label.css .of-label.
const searchCSS = `
#dc-overlays {
  position: static;
  width: 100%;
  height: 0;
  overflow: visible;
}
#dc-search-panel {
  /* The inline recipe centers with left:50% plus a -280px margin-left,
     which the layout engine drops, leaving the panel's left edge glued to
     the viewport center. This shift finishes the centering in paint. */
  transform: translateX(-50%);
  background: #2b2d31;
  border: 1px solid #1e1f22;
  border-radius: 8px;
  overflow: hidden;
}
#dc-search-panel .of-card {
  gap: 0;
  padding: 0;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #dbdee1;
}
#dc-search-panel .of-card-header { padding: 12px 16px 0; }
#dc-search-panel .dc-search-head {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  min-height: 24px;
}
#dc-search-panel .dc-search-head .of-card-title { display: none; }
#dc-search-panel .dc-search-head button {
  padding: 2px 8px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: #b5bac1;
  font-size: 16px;
  line-height: 1;
  cursor: pointer;
}
#dc-search-panel .dc-search-head button:hover { background: #35373c; color: #fff; }
#dc-search-panel .of-card-content {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 8px 16px 12px;
  color: #dbdee1;
}
#dc-search-panel .dc-search-field .of-field { gap: 0; }
#dc-search-panel .dc-search-field .of-field-label { display: none; }
#dc-search-panel .dc-search-field .of-input {
  min-height: 44px;
  padding: 11px 12px;
  border: 1px solid #1e1f22;
  border-radius: 4px;
  background: #1e1f22;
  color: #dbdee1;
  font-size: 16px;
  line-height: 22px;
}
#dc-search-panel .dc-search-field .of-input::placeholder { color: #6d6f78; }
#dc-search-panel .dc-search-field .of-input:focus,
#dc-search-panel .dc-search-field .of-input:focus-visible,
#dc-search-panel .dc-search-field .of-input[data-ownframe-focus="1"] {
  border-color: #5865f2;
  outline: none;
}
#dc-search-panel .dc-search-sections {
  display: flex;
  flex-direction: column;
  gap: 4px;
  /* Fixed list height with an internal scroll, quick-switcher style: the
     panel keeps one stable size across query, capped, and empty states
     instead of jumping as rows filter. */
  max-height: 340px;
  overflow-y: auto;
  scrollbar-width: thin;
  scrollbar-color: #1e1f22 transparent;
}
#dc-search-panel .dc-search-sections .of-label {
  padding: 8px 8px 2px;
  color: #949ba4;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}
#dc-search-panel .dc-search-results { display: flex; flex-direction: column; gap: 1px; }
#dc-search-panel .dc-search-results button {
  display: flex;
  flex-direction: column;
  gap: 1px;
  width: 100%;
  padding: 7px 10px;
  border: none;
  border-radius: 4px;
  background: transparent;
  text-align: left;
  font-family: inherit;
  cursor: pointer;
}
#dc-search-panel .dc-search-results button:hover,
#dc-search-panel .dc-search-results button:focus-visible {
  background: #35373c;
  outline: none;
}
#dc-search-panel .dc-search-name {
  display: block;
  color: #dbdee1;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.35;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
#dc-search-panel .dc-search-detail {
  display: block;
  color: #949ba4;
  font-size: 12px;
  line-height: 1.35;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
#dc-search-panel .dc-search-sections > hr,
#dc-search-panel .dc-search-sections > .of-separator {
  height: 1px;
  margin: 6px 8px;
  border: none;
  background: #3f4147;
}
#dc-search-panel .dc-search-sections .of-field-description {
  margin: 4px 0 0;
  padding: 8px 10px 2px;
  border-top: 1px solid #3f4147;
  color: #949ba4;
  font-size: 12px;
  line-height: 1.4;
}
`
