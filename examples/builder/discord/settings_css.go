package main

// settingsCSS paints the fullscreen Discord-style settings overlay built by
// settingsSheet in overlays.go, plus the #dc-float wrapper from shell.go.
//
// Wiring: append after appCSS in themeSheet (app.go):
//
//	chadcn.Dark + chadcn.Stylesheet + appCSS + settingsCSS
//
// #dc-float is a plain block wrapper holding the dismiss layer and the
// profile card. It stays position:static with height:0 so it never moves
// layout; fixed children (the viewport-unit dismiss layer, the profile
// popover) resolve against the viewport.
const settingsCSS = `
#dc-float {
  position: static;
  width: 100%;
  height: 0;
  overflow: visible;
}
#dc-settings {
  top: 48px;
  height: calc(100vh - 48px);
  overflow-y: auto;
  background: #313338;
}
.of-scope--light #dc-settings { background: #e3e5e8; }
.of-scope--light .dc-st-title { color: #060607; }
.of-scope--light .dc-st-navsection { color: #5c5e66; }
.of-scope--light .dc-st-navitem { color: #5c5e66; }
.of-scope--light .dc-st-navitem:hover { background: #d7d9dc; color: #060607; }
.of-scope--light .dc-st-navitem--active { background: #d7d9dc; color: #060607; }
.of-scope--light .dc-st-card { background: #f2f3f5; }
.of-scope--light .dc-st-card .dc-st-row div span:first-child { color: #060607; }
.of-scope--light .dc-st-card .dc-st-row div span:last-child { color: #5c5e66; }
.of-scope--light .dc-st-esc { border-color: #5c5e66; color: #5c5e66; }
.of-scope--light .dc-st-esc:hover { border-color: #060607; color: #060607; }
.of-scope--light .dc-st-key { color: #5c5e66; }
.dc-st {
  display: flex;
  flex-direction: row;
  justify-content: center;
  gap: 20px;
  width: 100%;
  min-height: 100%;
}
.dc-st-nav {
  width: 218px;
  flex-shrink: 0;
  padding: 24px 20px 60px 0;
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.dc-st-navsection {
  padding: 16px 10px 4px;
  color: #949ba4;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.dc-st-navitem {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  font-family: inherit;
  padding: 8px 10px;
  border-radius: 4px;
  color: #b5bac1;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
}
.dc-st-navitem:hover { background: #35373c; color: #dbdee1; }
.dc-st-navitem--active { background: #404249; color: #fff; }
.dc-st-main {
  flex: 0 1 740px;
  min-width: 0;
  padding: 24px 40px 80px 0;
}
.dc-st-title {
  color: #fff;
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 20px;
}
.dc-st-card {
  background: #2b2d31;
  border-radius: 8px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.dc-st-row {
  display: flex;
  align-items: center;
  gap: 16px;
}
.dc-st-card .dc-st-row button {
  background: #4e5058;
  border: none;
  border-radius: 4px;
  color: #fff;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  padding: 8px 16px;
  cursor: pointer;
  margin-left: auto;
  flex-shrink: 0;
}
.dc-st-card .dc-st-row button[aria-pressed="true"] {
  background: #5865f2;
}
.dc-st-card .dc-st-row button:hover {
  filter: brightness(1.1);
}
.dc-st-card .dc-st-row div {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.dc-st-card .dc-st-row div span:first-child {
  color: #fff;
  font-size: 14px;
  font-weight: 600;
}
.dc-st-card .dc-st-row div span:last-child {
  color: #b5bac1;
  font-size: 12px;
}
.dc-st-escwrap {
  position: fixed;
  top: 108px;
  right: 60px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.dc-st-esc {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: transparent;
  border: 2px solid #b5bac1;
  color: #b5bac1;
  font-size: 16px;
  cursor: pointer;
  display: block;
  text-align: center;
}
.dc-st-esc:hover { border-color: #fff; color: #fff; }
.dc-st-key {
  display: block;
  text-align: center;
  color: #b5bac1;
  font-size: 11px;
  font-weight: 700;
  margin-top: 4px;
}
.dc-st-keycap {
  padding: 4px 10px;
  border-radius: 4px;
  background: #1e1f22;
  border: 1px solid #4e5058;
  color: #dbdee1;
  font-size: 12px;
  font-weight: 700;
  white-space: nowrap;
}
`
