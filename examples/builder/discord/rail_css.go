package main

// railCSS paints the voice-channel occupant rows built by voicePanel in
// servers.go: the dc-vc-users list, dc-vc-user rows, dc-vc-ava initial
// circles, and dc-vc-name labels. The first row is the local demo user
// (dcMeName); this sheet gives it a green presence dot and a brighter name.
// All four classes are display-only: voicePanel emits no data-action hooks
// on them.
//
// Const name: railCSS. Wiring: append it after appCSS in themeSheet
// (app.go), which this file must not touch:
// chadcn.Dark + chadcn.Stylesheet + appCSS + railCSS.
const railCSS = `
.dc-vc-users {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin: 0 8px 8px;
  padding: 2px 0 2px 16px;
}
.dc-vc-user {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 28px;
  min-height: 28px;
  min-width: 0;
}
.dc-vc-ava {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  background: #5865f2;
  color: #fff;
  font-size: 9px;
  font-weight: 700;
  line-height: 1;
  text-align: center;
}
.dc-vc-users > .dc-vc-user:first-child .dc-vc-ava::after {
  content: "";
  position: absolute;
  right: -1px;
  bottom: -1px;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #23a559;
  border: 1px solid #2b2d31;
}
.dc-vc-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  color: #b5bac1;
  font-size: 12px;
}
.dc-vc-users > .dc-vc-user:first-child .dc-vc-name {
  color: #dbdee1;
  font-weight: 600;
}
`

// railCSSAnchor keeps railCSS tied to the markup that needs it until
// themeSheet (app.go) appends the sheet; without a reference the unused
// linter flags the const.
var _ = railCSS
