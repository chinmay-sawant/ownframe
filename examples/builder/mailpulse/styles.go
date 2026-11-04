package main

// appCSS is MailPulse layout only. ChadCN owns component paint and state.
const appCSS = `
.mp-shell {
  display: flex;
  align-items: flex-start;
  min-width: 960px;
  min-height: 100vh;
  background: #f6f6f5;
  color: #1a1a1a;
  font-family: sans-serif;
  font-size: 14px;
}
.mp-shell #mp-sidebar {
  position: sticky;
  top: 0;
  height: 100vh;
  flex-shrink: 0;
}
.mp-shell #mp-sidebar .of-sidebar-content {
  overflow-y: auto;
  flex-grow: 1;
  flex-basis: 0;
}
.mp-brand { display: flex; align-items: center; gap: 9px; padding: 2px 0; }
.mp-brand-mark {
  display: inline-flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 8px; flex-shrink: 0;
  background: #1a1a1a; color: #fff; font-weight: 700; font-size: 14px;
}
.mp-nav { display: flex; flex-direction: column; min-height: 0; }
.mp-user { display: flex; align-items: center; gap: 9px; }
.mp-user-text { display: flex; flex-direction: column; line-height: 1.25; min-width: 0; }
.mp-right {
  display: flex;
  flex: 1;
  flex-direction: column;
  min-width: 0;
}
.mp-topbar {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  padding: 14px 28px;
  background: #f6f6f5;
  flex-wrap: nowrap;
}
.mp-window-controls { display: flex; align-items: center; gap: 1px; flex-shrink: 0; }
.mp-topbar-spacer { flex: 1; min-width: 12px; }
.mp-search { width: 220px; flex-shrink: 0; }
.mp-search .of-input {
  border: none;
  background: transparent;
  box-shadow: none;
  padding: 0;
  width: 100%;
}
.mp-search .of-input:focus { outline: none; box-shadow: none; }
.mp-content { min-width: 0; padding: 0 28px 32px; }
.mp-card {
  min-width: 0;
  overflow: hidden;
  border: 1px solid #e8e8e6;
  border-radius: 16px;
  background: #fff;
}
.mp-pagehead {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 22px 24px 16px;
}
.mp-pagehead-text { display: flex; flex-direction: column; gap: 4px; min-width: 0; }
.mp-pagehead-actions { display: flex; align-items: center; gap: 8px; flex-shrink: 0; }
.mp-title { margin: 0; font-size: 20px; font-weight: 700; line-height: 1.25; }
.mp-sub { margin: 0; color: #6b7280; font-size: 13px; line-height: 1.4; }
.mp-muted { color: #6b7280; font-size: 12px; line-height: 1.4; }
.mp-kpis {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  border-top: 1px solid #eee;
  border-bottom: 1px solid #eee;
}
.mp-kpi {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-width: 180px;
  padding: 16px 20px;
  border-left: 1px solid #eee;
}
.mp-kpi:first-child { border-left: none; }
.mp-kpi-top { display: flex; align-items: center; gap: 8px; }
.mp-kpi-label { color: #1a1a1a; font-size: 13px; font-weight: 500; }
.mp-kpi-more { margin-left: auto; color: #a3a3a3; letter-spacing: 2px; }
.mp-kpi-value-row { display: flex; align-items: center; gap: 8px; }
.mp-kpi-value { font-size: 24px; font-weight: 700; line-height: 1.1; }
.mp-kpi-foot { display: flex; align-items: center; gap: 8px; }
.mp-kpi-leader { flex-grow: 1; border-top: 1px dotted #d4d4d2; min-width: 8px; }
.mp-kpi-dot {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 8px;
  flex-shrink: 0;
  font-size: 13px;
  line-height: 1;
  overflow: hidden;
}
.mp-kpi-dot--orange { background: #fff1de; color: #ff8a00; }
.mp-kpi-dot--blue { background: #e3edff; color: #2f7bff; }
.mp-kpi-dot--purple { background: #ece4ff; color: #7c3aed; }
.mp-kpi-dot--green { background: #ddf5e5; color: #16a34a; }
.mp-delta {
  display: inline-flex;
  align-items: center;
  padding: 3px 8px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.3;
  white-space: nowrap;
}
.mp-delta--up { background: #e6f9ed; color: #0d7a3f; }
.mp-delta--down { background: #fdecec; color: #c03636; }
.mp-grid-mid {
  display: grid;
  grid-template-columns: minmax(0, 7fr) minmax(0, 5fr);
  border-bottom: 1px solid #eee;
}
.mp-grid-mid > * { min-width: 0; padding: 20px 24px; }
.mp-funnel { border-left: 1px solid #eee; }
.mp-funnel-row { display: flex; flex-direction: column; gap: 6px; padding: 7px 0; }
.mp-funnel-top { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; }
.mp-funnel-label { font-size: 13px; white-space: nowrap; }
.mp-funnel-stats { display: flex; align-items: baseline; gap: 8px; flex-shrink: 0; }
.mp-funnel-count { color: #6b7280; font-size: 12px; white-space: nowrap; }
.mp-funnel-pct { font-size: 13px; font-weight: 700; min-width: 52px; text-align: right; white-space: nowrap; }
.mp-grid-bot {
  display: grid;
  grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
}
.mp-grid-bot > * { min-width: 0; padding: 20px 24px; }
.mp-grid-bot > *:nth-child(2) { border-left: 1px solid #eee; }
.mp-section-title { margin: 0 0 4px; font-size: 14px; font-weight: 700; }
.mp-section-sub { margin: 0 0 12px; color: #6b7280; font-size: 12px; }
.mp-heat-badges { display: flex; align-items: center; gap: 8px; margin-bottom: 14px; }
.mp-heat-grid {
  display: grid;
  grid-template-columns: 44px repeat(18, minmax(0, 1fr));
  align-items: center;
  gap: 5px;
}
.mp-heat-day { color: #6b7280; font-size: 11px; white-space: nowrap; }
.mp-heat-hour { color: #6b7280; font-size: 11px; text-align: center; }
.mp-heat-readout {
  margin: 12px 0 4px;
  padding: 8px 12px;
  border-radius: 8px;
  background: #f7f5ff;
  border: 1px solid #e6defc;
  color: #4c1d95;
  font-size: 12px;
  line-height: 1.5;
  min-height: 34px;
}
.mp-heat-cell:hover { outline: 2px solid #7c3aed; outline-offset: -2px; cursor: pointer; }
.mp-heat-cell { width: 100%; aspect-ratio: 1; border-radius: 5px; background: #f1f1ef; }
.mp-heat-cell--l0 { background: #f1f1ef; }
.mp-heat-cell--l1 { background: #e2dcfd; }
.mp-heat-cell--l2 { background: #c9bcfb; }
.mp-heat-cell--l3 { background: #a78bfa; }
.mp-heat-cell--l4 { background: #7c3aed; }
.mp-heat--0 { background: #f1f1ef; }
.mp-heat--1 { background: #e2dcfd; }
.mp-heat--2 { background: #c9bcfb; }
.mp-heat--3 { background: #a78bfa; }
.mp-heat--4 { background: #7c3aed; }
.mp-heat-swatches { display: flex; align-items: center; gap: 4px; }
.mp-heat-swatches .mp-heat-cell { width: 14px; height: 14px; aspect-ratio: auto; }
.mp-heat-legend { display: flex; align-items: center; gap: 6px; margin-top: 12px; color: #6b7280; font-size: 11px; }
.mp-heat-legend-label { white-space: nowrap; }
.mp-dots { display: flex; align-items: center; gap: 4px; }
.mp-dot { width: 14px; height: 8px; border-radius: 4px; background: #ececec; flex-shrink: 0; transition: background-color 0.35s ease; }
.mp-dot--on { background: #8b5cf6; }
.mp-dot--off { background: #ececec; }
.mp-linkrow {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-top: 1px solid #f0f0ee;
}
.mp-linkrow:first-of-type { border-top: none; }
.mp-bars { display: flex; align-items: flex-end; gap: 6px; min-height: 200px; min-width: 0; overflow: hidden; }
.mp-bar-col { display: flex; flex: 1 1 0; flex-direction: column; align-items: center; gap: 6px; min-width: 0; }
.mp-bar { width: 100%; max-width: 40px; border-radius: 4px 4px 0 0; background: #8b5cf6; transition: height 0.45s ease, background-color 0.45s ease; }
.mp-report-row { padding: 16px 8px; gap: 20px; }
.mp-report-row + .mp-report-row { border-top: 1px solid #f0f0ee; }
.mp-link-meta {
  display: flex; flex-direction: column; gap: 6px; min-width: 0; flex: 1;
}
.mp-linkicon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: 8px; flex-shrink: 0;
  background: #f4f4f5; font-size: 14px;
}
.mp-linkmain { display: flex; flex-direction: column; gap: 2px; min-width: 0; flex: 1; }
.mp-linkmeta { display: flex; flex-direction: column; gap: 2px; min-width: 0; flex: 1; }
.mp-linkstats { display: flex; align-items: baseline; gap: 12px; flex-shrink: 0; text-align: right; }
.mp-linkmeta-row { display: flex; flex-direction: column; gap: 2px; }
.mp-engage-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: 8px; flex-shrink: 0;
  background: #f4f4f5; font-size: 14px;
}
.mp-engage-label { font-size: 13px; font-weight: 500; margin-bottom: 6px; }
.mp-engage-stats { display: flex; align-items: baseline; gap: 8px; flex-shrink: 0; margin-left: auto; }
.mp-engage-open { font-size: 13px; }
.mp-toast {
  display: flex; align-items: flex-start; justify-content: space-between; gap: 12px;
  box-sizing: border-box; width: 520px; max-width: calc(100vw - 32px); margin: 0; padding: 10px 12px;
  border: 1px solid #e8e8e6; border-radius: 12px; background: #1a1a1a; color: #fff;
  font-size: 12px; line-height: 1.5;
  z-index: 200;
}
.mp-toast-msg {
  flex: 1 1 auto; min-width: 0;
  white-space: normal; overflow-wrap: break-word;
}
.mp-toast .of-button { flex-shrink: 0; }
.mp-insight {
  margin-top: 12px;
  padding: 10px 12px;
  border: 1px solid #e6defc;
  border-radius: 10px;
  background: #f7f5ff;
  font-size: 12px;
  line-height: 1.5;
}
`
