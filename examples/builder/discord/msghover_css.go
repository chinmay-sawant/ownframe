package main

// msghoverCSS restyles the Discord-clone message hover toolbar
// (.dc-msg .dc-msg-hover in messages.go messageRow) as Discord's floating
// action pill: one dark rounded bar with the reply, pin, and more icon
// buttons, hugging the right edge of the hovered message.
//
// Wiring (done by the integrator, not here): append msghoverCSS after appCSS
// in app.go themeSheet, so these rules win ties by source order. Every rule
// below also outranks styles.go on specificity (each selector repeats
// .dc-msg), so the pill paint holds wherever the sheet lands.
//
// In-flow design: the toolbar is a plain in-flow div after the reaction
// pills, and position:absolute does not apply to flex children in this
// engine, so the pill sits in the normal column flow right-aligned with
// margin-left:auto (plus align-self: flex-end as a second right-align
// path), never overlapping the message text.
//
// Always-visible fallback: the first draft hid the pill with display:none
// and revealed it with .dc-msg:hover .dc-msg-hover. Visual verification on
// -web (POST /hover over a message row, then frame.png) showed the row
// itself restyles on hover (.dc-msg:hover background applies) but no
// descendant-of-:hover selector ever takes effect: neither the toolbar nor
// a paint-only probe (.dc-msg:hover .of-bubble) changed. So the reveal rule
// is dropped and the pill stays visible. If the engine later supports
// ancestor :hover, re-adding
//
//	.dc-msg:hover .dc-msg-hover { display: flex; }
//
// over a display:none base restores hover-reveal with no other change.
const msghoverCSS = `
.dc-msg .dc-msg-hover {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 2px;
  align-self: flex-end;
  margin-left: auto;
  margin-top: 4px;
  padding: 2px 3px;
  border: 1px solid #1e1f22;
  border-radius: 4px;
  background: #2b2d31;
}
.dc-msg .dc-msg-hover .dc-iconbtn {
  width: 20px;
  height: 20px;
  padding: 0;
  border-radius: 3px;
  background: transparent;
  color: #b5bac1;
  font-size: 13px;
  line-height: 1;
}
.dc-msg .dc-msg-hover .dc-iconbtn:hover {
  background: #404249;
  color: #fff;
}
.of-scope--light .dc-msg .dc-msg-hover {
  border-color: #e3e5e8;
  background: #fff;
}
.of-scope--light .dc-msg .dc-msg-hover .dc-iconbtn {
  color: #5c5e66;
}
.of-scope--light .dc-msg .dc-msg-hover .dc-iconbtn:hover {
  background: #e3e5e8;
  color: #060607;
}
`
