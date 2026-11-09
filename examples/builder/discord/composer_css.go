package main

// composerCSS is the message-composer polish sheet for the discord clone.
//
// Const name: composerCSS. Wiring (done by the integrator, not here):
// append composerCSS after appCSS in app.go themeSheet, so these rules win
// ties by source order. Every rule below also outranks styles.go on
// specificity (each selector repeats the dc-composer-box or dc-replybar
// class it refines), so the polish holds wherever the sheet lands.
//
// The markup lives in composer.go and pairs every dc-cp- (composer polish)
// class with the older dc-composer* / dc-replybar* classes from styles.go,
// so the composer keeps its base paint until this sheet is wired in. No
// rule here reaches outside .dc-composer, and none adds a data-action hook.
const composerCSS = `
.dc-composer-box.dc-cp-box { border-radius: 12px; background: #383a40; }
.dc-composer-box.dc-cp-box:focus-within { outline: 1px solid #5865f2; outline-offset: -1px; }
/* chadcn Input paints .of-input, not .dc-composer-input, so flatten it to
   the Discord transparent-field look inside the grey bar. */
.dc-composer-box .of-input {
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  box-shadow: none;
  color: #dbdee1;
}
.dc-composer-box .of-input::placeholder { color: #6d6f78; opacity: 1; }
.dc-composer-box .of-input:focus { outline: none; box-shadow: none; }
.dc-composer-box .dc-cp-plus { line-height: 1; }
.dc-composer-box .dc-cp-glyphs {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  flex-shrink: 0;
  color: #b5bac1;
}
.dc-composer-box .dc-cp-glyph { font-size: 19px; line-height: 1; cursor: pointer; }
.dc-composer-box .dc-cp-glyph:hover { color: #fff; }
.dc-composer-box .dc-cp-gif {
  font-size: 10px;
  font-weight: 800;
  letter-spacing: 0.5px;
  border: 1.5px solid currentColor;
  border-radius: 4px;
  padding: 1px 4px;
  line-height: 1.4;
  cursor: pointer;
}
.dc-composer-box .dc-cp-gif:hover { color: #fff; }
.dc-composer-box .dc-cp-send { display: inline-flex; flex-shrink: 0; }
.dc-composer-box .dc-cp-send .of-button--tonal { background-color: #5865f2; color: #fff; border: none; }
.dc-composer-box .dc-cp-send .of-button--tonal:hover { background-color: #4752c4; }
/* Reply preview: same grey as the bar, rounded top, snapped to the box. */
.dc-replybar.dc-cp-reply {
  background: #383a40;
  border-radius: 8px 8px 0 0;
  border-bottom: 1px solid #2b2d31;
  padding: 8px 12px 6px;
}
.dc-replybar.dc-cp-reply + .dc-composer-box.dc-cp-box { border-radius: 0 0 8px 8px; }
.dc-replybar .dc-cp-reply-label { color: #fff; font-weight: 700; flex-shrink: 0; }
.dc-replybar .dc-cp-reply-snippet { color: #b5bac1; }
.dc-replybar .dc-cp-reply-clear { margin-left: auto; flex-shrink: 0; }
.of-scope--light .dc-composer-box.dc-cp-box, .of-scope--light .dc-replybar.dc-cp-reply { background: #ebedef; }
.of-scope--light .dc-composer-box .of-input { color: #313338; }
.of-scope--light .dc-replybar .dc-cp-reply-label { color: #060607; }
.of-scope--light .dc-replybar .dc-cp-reply-snippet, .of-scope--light .dc-composer-box .dc-cp-glyphs { color: #5c5e66; }
`
