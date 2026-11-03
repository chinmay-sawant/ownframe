package main

// appCSS is dashboard layout only. ChadCN owns component paint and state.
const appCSS = `
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--of-background, #0a0a0a);
  color: var(--of-text, #fafafa);
  font-family: sans-serif;
  font-size: 14px;
}
.app-body {
  display: flex;
  flex: 1;
  min-height: 0;
  min-width: 0;
  padding-top: 60px;
  padding-left: 232px;
  background: var(--of-background, #0a0a0a);
}
.app-shell--collapsed .app-body { padding-left: 64px; }
#app-sidebar {
  position: fixed;
  top: 0;
  left: 0;
  height: 100vh;
  padding-top: 60px;
  box-sizing: border-box;
  z-index: 30;
}
.app-topbar {
  position: fixed; top: 0; left: 0; right: 0; z-index: 40;
  display: flex;
  align-items: center;
  gap: 18px;
  height: 60px;
  flex-shrink: 0;
  padding: 0 20px;
  border-bottom: 1px solid var(--of-border, #262626);
  background: var(--of-surface, #171717);
}
.app-brand { display: flex; align-items: center; gap: 9px; flex-shrink: 0; }
.app-brand-mark {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 8px;
  background: var(--of-primary, #fafafa); color: var(--of-on-primary, #171717);
  font-weight: 700; font-size: 14px;
}
.app-brand-text { display: flex; flex-direction: column; line-height: 1.2; }
.app-topbar-spacer { flex: 1; }
.app-window-controls { display: flex; align-items: center; gap: 1px; }
.app-sidebar-head {
  display: flex; align-items: center; justify-content: space-between; gap: 8px;
  padding-bottom: 12px; border-bottom: 1px solid var(--of-border, #262626);
}
.app-sidebar-project { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.app-sidebar-nav { display: flex; flex-direction: column; gap: 3px; padding-top: 8px; }
.app-sidebar-foot {
  display: flex; align-items: center; gap: 9px;
  padding-top: 12px; border-top: 1px solid var(--of-border, #262626);
}
.app-sidebar-user { display: flex; flex-direction: column; flex: 1; min-width: 0; line-height: 1.25; }
.app-main {
  display: flex; flex-direction: column; flex: 1; gap: 14px;
  min-width: 0; padding: 22px 28px 40px;
}
.app-page-head { flex-shrink: 0; }
.app-page-head { display: flex; flex-direction: column; gap: 6px; }
.app-page-title { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.app-page-badges { display: flex; align-items: center; gap: 7px; }
.app-page-sub { margin: 0; color: var(--of-muted, #a3a3a3); font-size: 13px; }
.app-page-actions { display: flex; align-items: center; gap: 9px; padding: 8px 0 0; }
.app-stack { display: flex; flex-direction: column; gap: 14px; min-width: 0; }
.app-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.app-row--between { justify-content: space-between; }
.app-col { display: flex; flex-direction: column; gap: 8px; min-width: 0; }
.app-muted { color: var(--of-muted, #a3a3a3); font-size: 12px; }
.app-kpis {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}
.app-grid-2 { display: grid; grid-template-columns: minmax(0, 2fr) minmax(0, 1fr); gap: 12px; }
.app-grid-3 { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; }
.app-card-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
.app-kpi-value { font-size: 28px; font-weight: 700; line-height: 1.1; }
.app-kpi-label { font-size: 11px; letter-spacing: 0.04em; text-transform: uppercase; color: var(--of-muted, #a3a3a3); }
.app-section-label {
  font-size: 10px; font-weight: 700; letter-spacing: 0.08em;
  text-transform: uppercase; color: var(--of-muted, #a3a3a3);
}
.app-filter-row {
  display: flex; align-items: flex-start; gap: 9px; flex-wrap: wrap;
  padding: 12px; border: 1px solid var(--of-border, #262626);
  border-radius: var(--of-radius, 10px); background: var(--of-surface, #171717);
}
.app-filter { display: block; min-width: 120px; }
.app-filter > .of-label { display: block; }
.app-filter > * + * { margin-top: 4px; }
.app-filter--wide { min-width: 240px; flex: 1; }
.app-filter--severity { min-width: 230px; }
.app-table-wrap { width: 100%; overflow: hidden; }
.app-log-message { max-width: 420px; overflow: hidden; white-space: nowrap; }
.app-level {
  display: inline-flex; align-items: center; justify-content: center;
  padding: 2px 7px; border-radius: 999px; font-size: 10px;
  font-weight: 700; letter-spacing: 0.05em; text-transform: uppercase;
}
.app-level--error { background: rgba(239, 68, 68, 0.16); color: #f87171; }
.app-level--warn { background: rgba(245, 158, 11, 0.16); color: #fbbf24; }
.app-level--info { background: rgba(59, 130, 246, 0.14); color: #93c5fd; }
.app-level--debug { background: rgba(163, 163, 163, 0.14); color: #a3a3a3; }
.app-overlay {
  position: fixed; top: 0; left: 0;
  width: 100%; height: 100vh; box-sizing: border-box;
  z-index: 90; display: flex; align-items: flex-start; justify-content: center;
  padding-top: 96px; background-color: #000000b3;
}
.app-command { width: 620px; max-width: 90vw; }
.app-drawer-body { display: flex; flex-direction: column; gap: 12px; padding: 16px 20px 24px; }
.app-canvas {
  display: flex; align-items: center; justify-content: center;
  min-height: 100%; background: var(--of-control, #262626);
  color: var(--of-muted, #a3a3a3); font-size: 12px;
}
.app-rtl { direction: rtl; text-align: right; }
.app-chat { max-height: 320px; }
.app-chat-row { display: flex; gap: 9px; padding: 6px 0; }
.app-chat-row--end { flex-direction: row-reverse; }
.app-chat-body { display: flex; flex-direction: column; gap: 3px; max-width: 78%; }
.app-chat-meta { font-size: 10px; color: var(--of-muted, #a3a3a3); }
.app-timeline { display: flex; flex-direction: column; gap: 0; }
.app-toolbar { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.app-split { min-height: 380px; }
.app-empty { padding: 28px 16px; }
.app-spacer { flex: 1; }
.app-kpi-foot { display: flex; align-items: center; justify-content: space-between; gap: 8px; }
.app-heat { display: flex; align-items: flex-end; gap: 2px; height: 34px; }
.app-heat span { flex: 1; background: var(--of-primary, #fafafa); border-radius: 2px; opacity: 0.8; }
`
