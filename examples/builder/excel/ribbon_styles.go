package main

// ribbonCSS styles the ribbon, dropdowns, pickers, dialogs, context menu and
// charts. Every class carries the rb- prefix so it cannot collide with the
// shell and grid sheets in styles.go.
const ribbonCSS = `
.rb-ribbon { position: relative; display: flex; flex-direction: column; height: 88px; background: var(--xl-ribbon-bg); }
.rb-tabs { display: flex; align-items: center; gap: 2px; height: 28px; padding: 0 8px; flex-shrink: 0; }
.rb-tab {
  display: flex; align-items: center; justify-content: center; box-sizing: border-box; margin: 0;
  height: 24px; padding: 0 12px; border: 0; border-radius: 4px; background: transparent;
  color: var(--xl-header-fg); font-size: 12px; white-space: nowrap; flex-shrink: 0;
}
.rb-tab:hover { background-color: var(--xl-brand-tint); }
.rb-tab-on { background-color: var(--xl-brand-tint); color: var(--xl-brand-strong); font-weight: 700; }
.rb-tab-file { background-color: var(--xl-brand); color: #ffffff; font-weight: 700; }
.rb-tab-file:hover { background-color: var(--xl-brand-hover); }
.rb-tabs-fill { flex: 1; }
.rb-filter-sort { display: flex; gap: 8px; margin-bottom: 8px; }
.rb-ribbon-body {
  display: flex; align-items: stretch; gap: 0; height: 60px; padding: 0 6px;
  background: var(--xl-ribbon-group-bg); border-top: 1px solid var(--xl-ribbon-border);
}
.rb-rg { display: flex; flex-direction: column; padding: 2px 8px 0 8px; border-right: 1px solid var(--xl-ribbon-border); }
.rb-rg-body { display: flex; flex-direction: column; flex: 1; gap: 2px; justify-content: center; }
.rb-rb-col { display: flex; flex-direction: column; gap: 2px; }
.rb-rg-name { font-size: 10px; text-align: center; color: var(--xl-cell-muted); height: 12px; }
.rb-rb-row { display: flex; align-items: center; gap: 2px; }
.rb-rb {
  display: flex; align-items: center; justify-content: center; box-sizing: border-box; margin: 0; line-height: 1;
  height: 22px; min-height: 22px; min-width: 22px; padding: 0 6px; border: 1px solid transparent; border-radius: 3px;
  background: transparent; color: var(--xl-cell-text); font-size: 11px; white-space: nowrap;
}
.rb-rb:hover { background: var(--xl-brand-tint); }
.rb-rb-on { background-color: var(--xl-header-active-bg); border-color: var(--xl-brand-strong); color: var(--xl-header-active-fg); }
.rb-rb-off { opacity: 0.4; }
.rb-rb-qa { height: 22px; }
.rb-rb-bold { font-weight: 700; }
.rb-rb-italic { font-style: italic; }
.rb-rb-underline { text-decoration: underline; }
.rb-rb-strike { text-decoration: line-through; }
.rb-rb-size { min-width: 42px; }
.rb-rb-nf { min-width: 110px; text-align: left; }
.rb-rb-lg { flex-direction: column; height: 44px; min-height: 44px; min-width: 44px; }
.rb-rb-glyph { font-size: 12px; font-weight: 700; }
.rb-rb-cap { font-size: 10px; }
.rb-dd { position: relative; display: flex; }
.rb-dd-panel {
  position: absolute; left: 0; top: 24px; z-index: 80; display: flex; flex-direction: column; min-width: 190px;
  padding: 4px; background: var(--xl-ribbon-group-bg); border: 1px solid var(--xl-ribbon-border); border-radius: 6px;
  box-shadow: 0 2px 6px #00000026; color: var(--xl-cell-text);
}
.rb-dd-scroll { max-height: 240px; overflow-y: auto; min-width: 70px; }
.rb-dd-item {
  display: flex; box-sizing: border-box; margin: 0; align-items: center; justify-content: space-between; gap: 16px; height: 26px; padding: 0 8px;
  border: 0; border-radius: 4px; background: transparent; color: var(--xl-cell-text); font-size: 12px; text-align: left;
}
.rb-dd-item:hover { background: var(--xl-brand-tint); }
.rb-dd-hint { color: var(--xl-cell-muted); font-size: 11px; }
.rb-dd-sep { height: 1px; margin: 4px 0; background: var(--xl-ribbon-border); }
.rb-file-menu { position: absolute; left: 8px; top: 26px; z-index: 90; }
.rb-picker { width: 226px; padding: 8px; }
.rb-picker-title { margin: 6px 0 3px 0; font-size: 11px; font-weight: 700; color: var(--xl-cell-muted); }
.rb-sw-row { display: flex; gap: 2px; margin-bottom: 2px; }
.rb-sw { width: 20px; height: 18px; padding: 0; border: 1px solid #00000033; border-radius: 2px; }
.rb-sw:hover { border-color: var(--xl-brand-strong); }
.rb-styles { flex-direction: row; flex-wrap: wrap; width: 300px; gap: 4px; }
.rb-style-chip { height: 28px; padding: 0 10px; border: 1px solid var(--xl-ribbon-border); border-radius: 3px; font-size: 12px; }
.rb-ctx-layer { position: fixed; left: 0; top: 0; right: 0; bottom: 0; z-index: 120; }
.rb-ctx { z-index: 121; min-width: 220px; }
.rb-opt {
  display: flex; align-items: center; box-sizing: border-box;
  height: 26px; padding: 0 10px; margin: 0 4px 4px 0; border: 1px solid var(--xl-ribbon-border); border-radius: 4px;
  background: transparent; color: var(--xl-cell-text); font-size: 12px;
}
.rb-opt:hover { background: var(--xl-brand-tint); }
.rb-opt-on { background-color: var(--xl-brand); border-color: var(--xl-brand); color: #ffffff; }
.rb-primary { background-color: var(--xl-brand); border-color: var(--xl-brand); color: #ffffff; }
.rb-field { display: flex; flex-wrap: wrap; align-items: center; gap: 4px; margin-bottom: 8px; }
.rb-field-label { width: 140px; font-size: 12px; font-weight: 700; flex-shrink: 0; }
.rb-find-list, .rb-filter-list, .rb-rules { display: flex; flex-direction: column; max-height: 160px; overflow-y: auto; }
.rb-find-hit { height: 24px; padding: 0 8px; border: 0; background: transparent; text-align: left; font-size: 12px; color: var(--xl-cell-text); }
.rb-find-hit:hover { background: var(--xl-brand-tint); }
.rb-find-msg { margin: 4px 0; }
.rb-fc-tabs { display: flex; gap: 4px; margin-bottom: 8px; }
.rb-fc-pane { min-height: 96px; }
.rb-fc-preview { margin-top: 8px; padding: 12px; border: 1px solid var(--xl-ribbon-border); text-align: center; }
.rb-dec-n { min-width: 20px; text-align: center; }
.rb-rule { display: flex; align-items: center; gap: 4px; padding: 2px 0; font-size: 12px; }
.rb-rule-text { flex: 1; }
.rb-sc-group { margin-bottom: 8px; }
.rb-sc-row { display: flex; align-items: center; gap: 12px; height: 24px; font-size: 12px; }
.rb-chart-layer { z-index: 25; }
.rb-chart {
  display: flex; flex-direction: column; box-sizing: border-box; overflow: hidden;
  background: var(--xl-sheet-bg); border: 1px solid var(--xl-gridline); border-radius: 4px; box-shadow: 0 4px 12px #0000002e;
}
.rb-chart-selected { border: 2px solid var(--xl-brand-strong); }
.rb-chart-head {
  display: flex; align-items: center; justify-content: space-between; height: 26px; padding: 0 8px; flex-shrink: 0;
  background: var(--xl-header-bg); color: var(--xl-header-fg); font-size: 12px; font-weight: 700;
}
.rb-chart-close { width: 20px; height: 20px; padding: 0; border: 0; background: transparent; color: var(--xl-header-fg); }
.rb-chart-body { flex: 1; min-height: 0; padding: 8px; overflow: hidden; }
.rb-chart-empty { color: var(--xl-cell-muted); font-size: 12px; }
.rb-hbars { display: flex; flex-direction: column; gap: 6px; }
.rb-hbar-row { display: flex; align-items: center; gap: 8px; font-size: 11px; }
.rb-hbar-label { width: 80px; flex-shrink: 0; }
.rb-hbar-track { flex: 1; height: 14px; background: var(--xl-header-bg); }
.rb-hbar-fill { height: 14px; }
.rb-pie-strip { display: flex; height: 28px; margin: 12px 0; }
.rb-pie-seg { height: 28px; }
.rb-pie-legend { display: flex; flex-direction: column; gap: 4px; font-size: 11px; }
.rb-pie-item { display: flex; align-items: center; gap: 6px; }
.rb-pie-dot { width: 10px; height: 10px; border-radius: 5px; }
`
