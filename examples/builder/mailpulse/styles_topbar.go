package main

// topbarFixCSS owns search, popover, and toast paint for the MailPulse top
// bar. ChadCN owns component paint and appCSS owns layout; this sheet only
// fixes alignment, panels, and elevation the other sheets leave out.
const topbarFixCSS = `
.mp-search { align-items: center; }
.mp-search-icon {
  position: relative;
  width: 17px;
  height: 17px;
  flex-shrink: 0;
  line-height: 0;
  align-self: center;
}
.mp-search-mag {
  position: absolute;
  left: 0;
  top: 0;
  width: 13px;
  height: 13px;
  border: 2px solid #4b5563;
  border-radius: 50%;
  box-sizing: border-box;
}
.mp-search-handle {
  position: absolute;
  left: 10px;
  top: 9px;
  width: 3px;
  height: 10px;
  border-radius: 2px;
  background: #4b5563;
  transform: rotate(45deg);
}
.mp-search .of-input { height: 24px; line-height: 24px; padding-top: 0; padding-bottom: 0; align-self: center; flex: 1; min-width: 0; }
.mp-search-clear { cursor: pointer; color: #6b7280; font-size: 14px; line-height: 1; padding: 0 2px; flex-shrink: 0; align-self: center; }
.mp-topbar .of-button { line-height: 1; display: inline-flex; align-items: center; justify-content: center; }
.mp-bell-wrap { position: relative; display: inline-block; width: 28px; flex-shrink: 0; line-height: 0; align-self: center; }
.mp-dev-wrap { position: relative; display: inline-flex; flex-shrink: 0; line-height: 0; align-self: center; }
.mp-dev-dot {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #8b5cf6;
  border: 2px solid #ffffff;
  box-sizing: border-box;
}
.mp-bell-icon { display: flex; flex-direction: column; align-items: center; line-height: 0; }
.mp-bell-dome { width: 11px; height: 8px; border-radius: 6px 6px 2px 2px; background: #4b5563; }
.mp-bell-skirt { width: 14px; height: 2px; border-radius: 1px; background: #4b5563; }
.mp-bell-clap { width: 4px; height: 4px; border-radius: 50%; background: #4b5563; }
.mp-bell-dot {
  position: absolute;
  top: 3px;
  right: 3px;
  width: 9px;
  height: 9px;
  border-radius: 50%;
  background: #ef4444;
  border: 2px solid #ffffff;
  box-sizing: border-box;
}
.mp-topbar-right { position: relative; display: flex; align-items: center; gap: 12px; min-width: 0; }
.mp-search-wrap { position: relative; }
.mp-search-anchor { position: relative; display: block; height: 0; }
.mp-search-drop {
  position: absolute;
  top: 8px;
  left: 0;
  width: 100%;
  min-width: 200px;
  background: #ffffff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
  z-index: 70;
  overflow: hidden;
  padding: 4px;
}
.mp-search-row { padding: 8px 10px; border-radius: 8px; cursor: pointer; font-size: 13px; line-height: 1.4; }
.mp-search-row:hover { background: #f7f7f4; }
.mp-search-empty { padding: 10px 12px; color: #6b7280; font-size: 13px; line-height: 1.4; }
.mp-avatar-btn { position: relative; display: inline-block; width: 40px; cursor: pointer; line-height: 0; flex-shrink: 0; align-self: center; }
.mp-panel-anchor { position: relative; display: block; height: 0; }
.mp-float-panel {
  position: absolute;
  top: 8px;
  right: 0;
  width: 320px;
  background: #ffffff;
  border: 1px solid #e8e8e6;
  border-radius: 12px;
  overflow: hidden;
  padding: 8px;
  z-index: 70;
}
.mp-hidden { display: none; }
.mp-notif-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 8px; padding: 10px; border-radius: 8px; }
.mp-notif-row:hover { background: #f7f7f4; }
.mp-notif-main { display: flex; flex-direction: column; gap: 2px; min-width: 0; flex: 1; }
.mp-notif-title { font-size: 13px; font-weight: 600; line-height: 1.4; }
.mp-notif-body { font-size: 12px; color: #6b7280; line-height: 1.4; }
.mp-notif-age { font-size: 11px; color: #a3a3a3; line-height: 1.4; }
.mp-notif-x { cursor: pointer; color: #6b7280; font-size: 14px; line-height: 1; padding: 2px; flex-shrink: 0; }
.mp-notif-empty { padding: 16px 12px 4px; text-align: center; font-size: 13px; font-weight: 600; }
.mp-notif-caught { padding: 0 12px 12px; text-align: center; font-size: 12px; color: #6b7280; }
.mp-user-head { display: flex; align-items: center; gap: 10px; padding: 10px; border-radius: 8px; }
.mp-user-name { font-size: 13px; font-weight: 600; }
.mp-user-muted { font-size: 12px; color: #6b7280; }
.mp-user-row { padding: 8px 10px; border-radius: 8px; cursor: pointer; font-size: 13px; line-height: 1.4; }
.mp-user-row:hover { background: #f7f7f4; }
.mp-toast { z-index: 200; box-shadow: none; }
`
