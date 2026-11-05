package main

// profileCSS dresses the member profile popover (members.go profileCard) like
// a real Discord user card. Integrator: append it to the theme in app.go
// themeSheet, after appCSS so these rules win:
//
//	return chadcn.Dark + chadcn.Stylesheet + appCSS + profileCSS
//
// Every class below is emitted by profileCard; nothing else uses them.
const profileCSS = `
.dc-profile {
  position: fixed;
  right: 252px;
  top: 64px;
  width: 300px;
  z-index: 150;
  border-radius: 8px;
  background: #111214;
  border: 1px solid #1e1f22;
  overflow: hidden;
  padding: 0 0 16px;
  font-family: sans-serif;
}
.dc-profile-banner {
  display: block;
  box-sizing: border-box;
  width: 100%;
  height: 90px;
  padding: 8px;
  background: #5865f2;
  text-align: right;
}
.dc-profile-banner .of-button {
  color: #fff;
  background: transparent;
}
.dc-profile-banner .of-button:hover {
  color: #fff;
}
.dc-profile-ava {
  margin: -28px 0 0 16px;
}
.dc-profile-ava .of-avatar-frame {
  border: 4px solid #111214;
  border-radius: 50%;
}
.dc-profile-body {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 8px 16px 0;
}
.dc-profile-namerow {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}
.dc-profile-name {
  color: #fff;
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dc-profile-user {
  display: block;
  margin-top: -8px;
  color: #949ba4;
  font-size: 13px;
  font-weight: 500;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.dc-profile-sec {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.dc-profile-sechead {
  color: #b5bac1;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.dc-profile-text {
  color: #dbdee1;
  font-size: 13px;
  line-height: 1.4;
  overflow-wrap: break-word;
}
.dc-profile-roles {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}
.dc-profile-rolechip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  border-radius: 4px;
  background: #2b2d31;
}
.dc-profile-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #80848e;
}
.dc-profile-chiptext {
  color: #dbdee1;
  font-size: 12px;
  font-weight: 600;
}
.dc-profile-actions {
  display: flex;
  gap: 8px;
  margin-top: 2px;
}
.dc-profile-actions .of-button {
  flex: 1;
}
`
