package main

// appCSS is Discord-clone layout and paint. Dark-first; chadcn.Dark supplies
// component paint, so this sheet owns the shell columns and chat dressing.
const appCSS = `
.discord-shell {
  --of-button-radius: 4px;
  --of-radius-sm: 4px;
  --of-radius-lg: 8px;
  display: flex;
  flex-direction: column;
  width: 100vw;
  min-width: 960px;
  height: 100vh;
  min-height: 600px;
  overflow: hidden;
  background: #313338;
  color: #dbdee1;
  font-family: sans-serif;
  font-size: 14px;
  line-height: 1.4;
}
.discord-shell * { box-sizing: border-box; }
.dc-body {
  display: flex;
  flex-direction: row;
  flex: 1;
  flex-shrink: 0;
  height: calc(100vh - 48px);
  min-height: 0;
  min-width: 0;
}
.dc-brand {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  white-space: nowrap;
}
.dc-topbar-spacer { flex: 1; min-width: 8px; }
#dc-search-tip { flex-shrink: 0; }
#dc-top-search {
  width: 420px;
  max-width: 40vw;
  flex-shrink: 0;
  justify-content: flex-start;
}
#dc-top-search .of-button-label { flex: 1; text-align: left; color: #949ba4; }
.dc-window-controls {
  display: flex;
  align-items: center;
  gap: 2px;
  flex-shrink: 0;
}

.dc-rail {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  width: 72px;
  flex-shrink: 0;
  padding: 12px 0;
  overflow-y: auto;
  overflow-x: hidden;
  background: #1e1f22;
  scrollbar-width: thin;
  scrollbar-color: #1e1f22 #1e1f22;
}
.dc-rail::-webkit-scrollbar { width: 0; }
.dc-rail-sep { width: 32px; height: 2px; border-radius: 1px; background: #35363c; flex-shrink: 0; }
.dc-server {
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #313338;
  color: #dbdee1;
  font-size: 17px;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease;
}
.dc-server:hover {
  background: #5865f2;
  color: #fff;
}
.dc-server--active {
  border-radius: 16px;
  background: #5865f2;
  color: #fff;
}
.dc-server--unread::after {
  content: "";
  position: absolute;
  left: -12px;
  top: 50%;
  width: 8px;
  height: 8px;
  margin-top: -4px;
  border-radius: 0 4px 4px 0;
  background: #fff;
}
.dc-server--active::before {
  content: "";
  position: absolute;
  left: -14px;
  top: 4px;
  bottom: 4px;
  width: 4px;
  border-radius: 0 4px 4px 0;
  background: #fff;
}
.dc-server--home {
  background: #313338;
  color: #23a559;
  font-size: 22px;
}
.dc-server--home:hover { background: #23a559; color: #fff; }
.dc-server-tip {
  position: absolute;
  left: 60px;
  top: 50%;
  transform: translateY(-50%) scale(0.9);
  padding: 6px 12px;
  border-radius: 6px;
  background: #111214;
  color: #fff;
  font-size: 13px;
  font-weight: 600;
  white-space: nowrap;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.1s ease, transform 0.1s ease;
  z-index: 100;
}
.dc-server:hover .dc-server-tip {
  opacity: 1;
  transform: translateY(-50%) scale(1);
}

.dc-side {
  display: flex;
  flex-direction: column;
  width: 240px;
  flex-shrink: 0;
  min-height: 0;
  background: #2b2d31;
  border-radius: 8px 0 0 0;
}
.dc-guild {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  height: 48px;
  flex-shrink: 0;
  padding: 0 16px;
  border-bottom: 1px solid #232428;
  font-size: 15px;
  font-weight: 700;
  color: #fff;
  cursor: pointer;
}
.dc-guild:hover { background: #35373c; }
.dc-guild-chev { color: #b5bac1; font-size: 12px; }
.dc-channels {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 12px 8px;
  scrollbar-width: thin;
  scrollbar-color: #1e1f22 transparent;
}
.dc-channels::-webkit-scrollbar { width: 8px; }
.dc-channels::-webkit-scrollbar-thumb {
  background: #1e1f22;
  border-radius: 4px;
}
.dc-channels::-webkit-scrollbar-track { background: transparent; }
.dc-cat {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 12px 8px 4px;
  color: #949ba4;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
  cursor: pointer;
}
.dc-cat:hover { color: #dbdee1; }
/* Raw ui.Button faces paint white by default; these rows reset to the
   Discord flat look. Width + left alignment keep every row full-bleed. */
.dc-server, .dc-guild, .dc-cat, .dc-chan, .dc-iconbtn {
  background: transparent;
  border: none;
  font-family: inherit;
}
.dc-guild, .dc-cat, .dc-chan { width: 100%; text-align: left; }
.dc-iconbtn { padding: 0; }
.dc-chan {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 8px;
  margin: 1px 0;
  border-radius: 4px;
  color: #949ba4;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
}
.dc-chan:hover { background: #35373c; color: #dbdee1; }
.dc-chan--active { background: #404249; color: #fff; }
.dc-chan--active:hover { background: #404249; color: #fff; }
.dc-chan--unread { color: #fff; font-weight: 600; }
.dc-chan-hash { color: #80848e; font-size: 18px; font-weight: 400; flex-shrink: 0; }
.dc-chan--unread .dc-chan-hash { color: #fff; }
.dc-chan-label {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dc-chan-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  border-radius: 8px;
  background: #f23f43;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.dc-main {
  display: flex;
  flex-direction: column;
  flex: 1;
  min-width: 0;
  min-height: 0;
  background: #313338;
}
.dc-topbar {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 48px;
  flex-shrink: 0;
  padding: 0 16px;
  border-bottom: 1px solid #232428;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
}
/* Channel header: same bar geometry as the chrome topbar, separate class so
   the two never share paint. */
.dc-chanhead {
  display: flex;
  align-items: center;
  gap: 8px;
  height: 48px;
  flex-shrink: 0;
  padding: 0 16px;
  border-bottom: 1px solid #232428;
  color: #fff;
  font-size: 14px;
  font-weight: 700;
}
.dc-topbar-hash { color: #80848e; font-size: 22px; font-weight: 400; }
.dc-topic {
  flex: 1;
  min-width: 0;
  margin: 0;
  padding-left: 8px;
  border-left: 1px solid #3f4147;
  color: #949ba4;
  font-size: 12px;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.dc-topbar-icons { display: flex; align-items: center; gap: 12px; flex-shrink: 0; }
.dc-iconbtn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 4px;
  color: #b5bac1;
  font-size: 16px;
  cursor: pointer;
}
.dc-iconbtn:hover { color: #fff; background: #35373c; }
.dc-iconbtn[aria-pressed="true"] { color: #f23f43; }
.dc-iconbtn img { width: 16px; height: 16px; display: block; }
.dc-chan-hash img { width: 16px; height: 16px; display: block; }

.dc-scroll {
  flex: 1;
  min-height: 0;
  position: relative;
  overflow: hidden;
}
#dc-scroll-content { padding: 16px 16px 8px 0; }
.dc-msg {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 6px 16px;
  margin-top: 12px;
  border-radius: 2px;
}
.dc-msg:hover { background: #2e3035; }
.dc-msg--mention { background: rgba(240, 178, 50, 0.08); }
.dc-msg--mention:hover { background: rgba(240, 178, 50, 0.12); }
.dc-msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #5865f2;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 700;
  line-height: 1;
  text-align: center;
  cursor: pointer;
}
.dc-msg-avatar:hover { filter: brightness(1.1); }
.dc-msg-body { flex: 1; min-width: 0; }
.dc-msg-head { display: flex; align-items: baseline; gap: 8px; }
.dc-msg-user { color: #fff; font-size: 14px; font-weight: 600; cursor: pointer; }
.dc-msg-user:hover { text-decoration: underline; }
.dc-msg--bot .dc-msg-user { color: #5865f2; }
.dc-msg-time { color: #949ba4; font-size: 11px; font-weight: 500; }
.dc-msg-tag { padding: 1px 5px; border-radius: 4px; background: #5865f2; color: #fff; font-size: 10px; font-weight: 700; }
.dc-msg-text { margin: 2px 0 0; color: #dbdee1; font-size: 14px; overflow-wrap: break-word; }
.dc-msg-text a { color: #00a8fc; text-decoration: none; }
.dc-msg-text a:hover { text-decoration: underline; }
.dc-msg-mention {
  padding: 0 4px;
  border-radius: 3px;
  background: rgba(88, 101, 242, 0.3);
  color: #c9cdfb;
  font-weight: 600;
  cursor: pointer;
}
.dc-msg-mention:hover { background: #5865f2; color: #fff; }
.dc-reacts { display: flex; flex-wrap: wrap; gap: 4px; margin-top: 6px; }
.dc-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 3px 8px;
  border: 1px solid #4e5058;
  border-radius: 8px;
  background: #2b2d31;
  color: #dbdee1;
  font-size: 13px;
  cursor: pointer;
}
.dc-pill:hover { border-color: #949ba4; }
.dc-pill--me { border-color: #5865f2; background: rgba(88, 101, 242, 0.25); }
.dc-pill-count { font-size: 12px; font-weight: 700; }
.dc-divider {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 16px 16px 4px;
  color: #949ba4;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.dc-divider::before, .dc-divider::after {
  content: "";
  flex: 1;
  height: 1px;
  background: #3f4147;
}
.dc-divider--new { color: #f23f43; }
.dc-divider--new::before, .dc-divider--new::after { background: #f23f43; }
.dc-typing {
  display: flex;
  align-items: center;
  gap: 6px;
  height: 22px;
  flex-shrink: 0;
  padding: 0 16px;
  background: #313338;
  position: relative;
  z-index: 5;
  color: #dbdee1;
  font-size: 12px;
  font-weight: 500;
}
.dc-typing strong { font-weight: 700; }
.dc-typing-dots { display: inline-flex; align-items: center; gap: 3px; }
.dc-typing-dots span {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #b5bac1;
  animation: dc-bounce 1.2s ease-in-out infinite;
}
.dc-typing-dots span:nth-child(2) { animation-delay: 0.15s; }
.dc-typing-dots span:nth-child(3) { animation-delay: 0.3s; }
@keyframes dc-bounce {
  0%, 60%, 100% { transform: translateY(0); opacity: 0.5; }
  30% { transform: translateY(-4px); opacity: 1; }
}

.dc-composer {
  flex-shrink: 0;
  padding: 0 16px 20px;
}
.dc-composer-box {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  border-radius: 8px;
  background: #383a40;
  min-height: 44px;
}
.dc-composer-box:focus-within { outline: none; }
.dc-composer-plus { display: inline-flex; align-items: center; justify-content: center; width: 24px; height: 24px; border-radius: 50%; flex-shrink: 0; background: #b5bac1; color: #383a40; font-size: 18px; font-weight: 700; cursor: pointer; }
.dc-composer-plus:hover { background: #fff; }
.dc-composer-input {
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  color: #dbdee1;
  font-size: 14px;
  padding: 11px 0;
}
.dc-composer-input::placeholder { color: #6d6f78; }
.dc-composer-input:focus { outline: none; }
.dc-composer-icons { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }
.dc-composer-icons .dc-iconbtn:hover { background: transparent; color: #fff; }

.dc-members {
  display: flex;
  flex-direction: column;
  width: 240px;
  flex-shrink: 0;
  min-height: 0;
  background: #2b2d31;
}
.dc-members-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 16px 8px;
  scrollbar-width: thin;
  scrollbar-color: #1e1f22 transparent;
}
.dc-members-list::-webkit-scrollbar { width: 8px; }
.dc-members-list::-webkit-scrollbar-thumb {
  background: #1e1f22;
  border-radius: 4px;
}
.dc-members-list::-webkit-scrollbar-track { background: transparent; }
.dc-role {
  padding: 12px 8px 4px;
  color: #949ba4;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}
.dc-member {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 5px 8px;
  border-radius: 4px;
  color: #949ba4;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  opacity: 0.7;
}
.dc-member:hover { background: #35373c; color: #dbdee1; opacity: 1; }
.dc-member-namerow { display: flex; align-items: center; gap: 6px; min-width: 0; }
.dc-member-name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-member-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  background: #80848e;
}
.dc-member-activity { color: #949ba4; font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-member-ava {
  position: relative;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #4e5058;
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}
.dc-dot {
  position: absolute;
  right: -2px;
  bottom: -2px;
  width: 12px;
  height: 12px;
  border: 3px solid #2b2d31;
  border-radius: 50%;
  background: #80848e;
}
.dc-dot--on { background: #23a559; }
.dc-dot--idle { background: #f0b232; }
.dc-dot--busy { background: #f23f43; }

.dc-voice {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 8px 8px;
  padding: 8px;
  border-radius: 4px;
  background: #232428;
  border: 1px solid #1e1f22;
  font-size: 12px;
}
.dc-voice-info { flex: 1; min-width: 0; line-height: 1.3; }
.dc-voice-state { color: #23a559; font-weight: 700; }
.dc-voice-name { color: #b5bac1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-voice-btns { display: flex; gap: 2px; flex-shrink: 0; }

.dc-userbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  margin-top: auto;
  padding: 8px;
  background: #232428;
  font-size: 13px;
}
.dc-userbar-ava {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: #5865f2;
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}
.dc-userbar-names { flex: 1; min-width: 0; line-height: 1.25; text-align: left; }
.dc-userbar-name { color: #fff; font-weight: 700; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-userbar-sub { color: #949ba4; font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-userbar .dc-iconbtn { width: 28px; height: 28px; font-size: 15px; }
.dc-userbar-profile {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  min-width: 0;
  padding: 0;
  background: transparent;
  border: none;
  font-family: inherit;
  text-align: left;
  cursor: pointer;
  border-radius: 4px;
}
.dc-userbar-profile:hover { background: #35373c; }

.dc-search {
  display: flex;
  align-items: center;
  width: 160px;
  padding: 2px 8px;
  border-radius: 4px;
  background: #1e1f22;
  color: #949ba4;
  font-size: 12px;
  cursor: pointer;
}
.dc-search:hover { color: #dbdee1; }

/* Overlay, palette, profile, and transcript dressing. */
.dc-dismiss {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100vh;
  background: transparent;
}
.dc-search-sections { display: flex; flex-direction: column; gap: 8px; }
.dc-search-results { display: flex; flex-direction: column; gap: 2px; }
.dc-search-results button {
  display: block;
  width: 100%;
  text-align: left;
  background: transparent;
  border: none;
  font-family: inherit;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: pointer;
}
.dc-search-results button:hover { background: #35373c; }
.dc-search-name { display: block; color: #dbdee1; font-size: 14px; font-weight: 600; }
.dc-search-detail { display: block; color: #949ba4; font-size: 12px; }
.dc-profile {
  position: fixed;
  width: 300px;
  z-index: 120;
  border-radius: 8px;
  background: #111214;
  border: 1px solid #1e1f22;
  padding: 16px;
}
.dc-replybar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 4px;
  color: #b5bac1;
  font-size: 12px;
}
.dc-replybar-text { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-msg-hover { display: flex; gap: 2px; margin-top: 4px; }
.dc-msg-reply { color: #949ba4; font-size: 12px; padding: 0 16px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.dc-msg-flag { color: #949ba4; font-size: 11px; }
.dc-empty { padding: 24px 16px; color: #949ba4; font-size: 14px; }
.dc-settings-tabs { display: flex; gap: 8px; margin-bottom: 16px; }
.dc-settings-section { display: flex; flex-direction: column; gap: 8px; margin-bottom: 16px; }

.discord-shell.of-scope--light { background: #e3e5e8; color: #313338; }
.of-scope--light .dc-rail { background: #e3e5e8; }
.of-scope--light .dc-side { background: #f2f3f5; }
.of-scope--light .dc-main { background: #fff; }
.of-scope--light .dc-members { background: #f2f3f5; }
.of-scope--light .dc-role { color: #5c5e66; }
.of-scope--light .dc-member { color: #313338; opacity: 1; }
.of-scope--light .dc-member:hover { background: #e3e5e8; color: #060607; }
.of-scope--light .dc-member-name { color: #060607; }
.of-scope--light .dc-member-activity { color: #5c5e66; }
.of-scope--light .dc-topbar, .of-scope--light .dc-chanhead { color: #060607; }
.of-scope--light .dc-chan { color: #5c5e66; }
.of-scope--light .dc-chan:hover { background: #e3e5e8; color: #060607; }
.of-scope--light .dc-chan--active { background: #d7d9dc; color: #060607; }
.of-scope--light .dc-chan--unread, .of-scope--light .dc-chan--unread .dc-chan-hash { color: #060607; }
.of-scope--light .dc-chan-badge { background: #f23f43; color: #fff; }
.of-scope--light .dc-server--unread::after, .of-scope--light .dc-server--active::before { background: #060607; }
.of-scope--light .dc-rail .of-tooltip-content { background: #fff; color: #060607; border: 1px solid #e3e5e8; }
.of-scope--light .dc-msg-mention { background: rgba(88, 101, 242, 0.15); color: #5865f2; }
.of-scope--light .dc-msg-mention:hover { background: #5865f2; color: #fff; }
.of-scope--light .dc-msg-text { color: #313338; }
.of-scope--light .dc-msg:hover { background: #f6f6f7; }
.of-scope--light .dc-topic { color: #5c5e66; }
.of-scope--light .dc-typing { background: #fff; color: #313338; }
.of-scope--light .dc-typing-dots span { background: #5c5e66; }
.of-scope--light .dc-userbar { background: #e3e5e8; }
.of-scope--light .dc-userbar-name { color: #060607; }
.of-scope--light .dc-userbar-sub { color: #5c5e66; }
.of-scope--light .dc-userbar-profile:hover { background: #d7d9dc; }
.of-scope--light .dc-voice { background: #fff; border-color: #e3e5e8; }
.of-scope--light .dc-voice-name { color: #5c5e66; }
.of-scope--light .dc-vc-name { color: #5c5e66; }
.of-scope--light .dc-vc-users > .dc-vc-user:first-child .dc-vc-name { color: #060607; }
.of-scope--light .dc-composer-box { background: #ebedef; }
.of-scope--light .dc-composer-input { color: #313338; }
`
