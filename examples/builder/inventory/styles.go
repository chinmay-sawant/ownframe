package main

const appCSS = `
.inv-shell {
  position: relative;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  height: 100vh;
  min-height: 100vh;
  padding-top: 60px;
  overflow: hidden;
  background: #f6f6f3;
  color: #1c1c1a;
  font-family: sans-serif;
  font-size: 14px;
}
.inv-topbar {
  position: fixed; top: 0; left: 0; right: 0; z-index: 40;
  display: flex;
  align-items: center;
  gap: 10px;
  height: 60px;
  padding: 0 16px;
  border-bottom: 1px solid #e6e4dc;
  background: rgba(255, 255, 255, 0.92);
}
.inv-brand { display: flex; align-items: center; gap: 9px; flex-shrink: 0; font-weight: 700; font-size: 14px; }
.inv-brand-mark {
  display: flex; align-items: center; justify-content: center;
  width: 28px; height: 28px; border-radius: 8px;
  background: #1d5c3f; color: #fafaf7;
  font-weight: 700; font-size: 14px;
}
.inv-brand-name { font-weight: 700; }
.inv-nav { display: flex; align-items: center; gap: 4px; margin-left: 8px; }
.inv-nav-item, .inv-nav-link {
  padding: 6px 12px; border-radius: 999px;
  color: #6b6963; font-size: 13px; font-weight: 500;
  white-space: nowrap;
}
.inv-nav-item--active, .inv-nav-link--active { background: #e8f0e8; color: #1d5c3f; font-weight: 600; }
.inv-spacer, .inv-topbar-spacer { flex: 1; }
.inv-search {
  display: flex; align-items: center; gap: 8px;
  width: 240px; padding: 7px 12px;
  border: 1px solid #e6e4dc; border-radius: 8px;
  background: #f6f6f3; color: #6b6963; font-size: 13px;
}
.inv-search-wrap { position: relative; }
.inv-search-panel {
  position: absolute; top: 40px; left: 0;
  width: 340px; z-index: 60;
  background: #fff; border: 1px solid #e6e4dc; border-radius: 10px;
  box-shadow: 0 12px 32px rgba(0,0,0,.14);
  padding: 6px; max-height: 320px; overflow: hidden;
}
.inv-search-backdrop {
  position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 50;
  background: rgba(28, 27, 23, 0.14);
}
.inv-search-picked { color: #8a8780; font-size: 11px; padding: 2px 12px 0; }
.inv-search-more { color: #8a8780; font-size: 11px; padding: 4px 8px 2px; }
.inv-bell { color: #6b6963; }
.inv-alerts-btn { position: relative; display: inline-flex; align-items: center; justify-content: center; }
.inv-alerts-dot {
  position: absolute; top: 4px; right: 4px;
  width: 8px; height: 8px; border-radius: 99px;
  background: #dc2626; border: 2px solid #fff;
}
.inv-window-controls { display: flex; align-items: center; gap: 2px; flex-shrink: 0; }
.inv-body-viewport {
  flex: 1;
  min-height: 0;
  min-width: 0;
  position: relative;
  overflow: hidden;
}
.inv-body-content {
  position: relative;
  display: flex; flex-direction: column; gap: 10px;
  padding: 12px 28px 12px;
  min-width: 0;
}
.inv-page-head { display: flex; align-items: flex-end; justify-content: space-between; gap: 12px; flex-wrap: wrap; }
.inv-page-head h1 { font-size: 20px; line-height: 1.2; margin: 0; }
.inv-page-titles { display: flex; flex-direction: column; gap: 1px; }
.inv-page-sub { margin: 0; color: #8a8780; font-size: 11px; }
.inv-page-actions { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; }
.inv-primary { background: #1d5c3f; border-color: #1d5c3f; color: #fafaf7; }
.inv-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.65fr) minmax(0, 1fr);
  gap: 12px;
  align-items: start;
  padding: 0;
}
.inv-left { display: flex; flex-direction: column; gap: 12px; min-width: 0; }
.inv-right { display: flex; flex-direction: column; gap: 12px; min-width: 0; }
.inv-card {
  background: #ffffff;
  border: 1px solid #e6e4dc;
  border-radius: 10px;
  padding: 10px 12px;
  min-width: 0;
}
.inv-card-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.inv-card-title { font-size: 13px; font-weight: 700; }
.inv-card-sub { margin: 2px 0 0; color: #8a8780; font-size: 11px; }
.inv-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.inv-row--between { justify-content: space-between; }
.inv-muted, .app-muted { color: #8a8780; font-size: 11px; }
.app-row { display: flex; align-items: center; gap: 10px; flex-wrap: wrap; }
.app-row--between { justify-content: space-between; }
.app-col { display: flex; flex-direction: column; gap: 8px; min-width: 0; }
.inv-kpis { display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 0; }
.inv-kpi { display: flex; flex-direction: column; gap: 4px; padding: 0 10px; min-width: 0; }
.inv-kpi + .inv-kpi { border-left: 1px solid #e6e4dc; }
.inv-kpi:first-child { padding-left: 0; }
.inv-kpi-value { font-size: 22px; font-weight: 700; line-height: 1.1; }
.inv-kpi-value--danger { color: #b91c1c; }
.inv-kpi-label { font-size: 10px; letter-spacing: 0.05em; text-transform: uppercase; color: #8a8780; }
.inv-kpi-row { display: flex; align-items: baseline; gap: 6px; }
.inv-kpi-row .inv-muted { font-size: 10px; }
.inv-kpi-foot { display: flex; align-items: center; gap: 6px; flex-wrap: wrap; font-size: 10px; color: #8a8780; }
.inv-kpi-foot .inv-muted { font-size: 10px; }
.inv-track { position: relative; height: 6px; border-radius: 999px; background: #eeeeea; }
.inv-track-fill { position: absolute; top: 0; left: 0; bottom: 0; border-radius: 999px; background: #1d5c3f; }
.inv-marker { display: none; }
.inv-segs { display: flex; height: 6px; border-radius: 999px; overflow: hidden; background: #eeeeea; }
.inv-seg { height: 100%; }
.inv-seg--red { background: #dc2626; }
.inv-seg--amber { background: #f59e0b; }
.inv-seg--green { background: #1d5c3f; }
.inv-seg--gray { background: #d6d3cb; }
.inv-filters { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; padding: 8px 0 6px; }
.inv-pill {
  padding: 5px 12px; border-radius: 999px; font-size: 12px; font-weight: 500;
  background: #f1efe9; color: #6b6963;
}
.inv-pill-active, .inv-pill--active { background: #e8f0e8; color: #1d5c3f; font-weight: 700; }
.inv-table-wrap { width: 100%; overflow: hidden; }
.inv-table-wrap .of-table-head { font-size: 10px; letter-spacing: 0.07em; padding: 6px 10px 6px 0; }
.inv-table-wrap .of-table-cell { padding: 8px 10px 8px 0; font-size: 12px; vertical-align: middle; }
.inv-table-wrap .of-table-cell .app-muted { font-size: 10px; }
.inv-bar { position: relative; height: 10px; min-width: 80px; border-radius: 999px; background: #eeeeea; overflow: visible; }
.inv-bar-fill { position: absolute; top: 0; left: 0; bottom: 0; border-radius: 999px; }
.inv-bar-fill--stockout { background: #dc2626; }
.inv-bar-fill--reorder { background: #f59e0b; }
.inv-bar-fill--healthy { background: #16a34a; }
.inv-bar-fill--overstock { background: #94a3b8; }
.inv-bar-marker { display: none; }
.pill-stockout, .pill-reorder, .pill-healthy, .pill-overstock {
  display: inline-flex; align-items: center; justify-content: center;
  padding: 2px 8px; border-radius: 999px;
  font-size: 10px; font-weight: 700; letter-spacing: 0.04em; text-transform: uppercase;
  white-space: nowrap;
}
.pill-stockout { background: rgba(220, 38, 38, 0.12); color: #b91c1c; }
.pill-reorder { background: rgba(245, 158, 11, 0.16); color: #b45309; }
.pill-healthy { background: rgba(22, 163, 74, 0.12); color: #15803d; }
.pill-overstock { background: rgba(100, 116, 139, 0.14); color: #475569; }
.inv-bins-grid { display: grid; gap: 3px; }
.inv-bins { display: grid; grid-template-columns: repeat(10, minmax(0, 1fr)); gap: 3px; }
.inv-bin {
  display: block; text-align: center;
  min-height: 22px; border-radius: 4px;
  font-size: 11px; font-weight: 600; line-height: 22px;
}
.inv-bin-head { display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 700; color: #8a8780; min-height: 16px; }
.inv-bin:hover { outline: 2px solid #1c1c1a; outline-offset: 1px; }
.inv-bin-key { width: 14px; height: 10px; border-radius: 3px; }
.inv-legend { display: flex; align-items: center; gap: 12px; font-size: 11px; color: #8a8780; }
.inv-legend-key { display: flex; align-items: center; gap: 5px; }
.inv-legend-stop { display: flex; align-items: center; gap: 5px; font-size: 11px; color: #6b6963; }
.inv-legend-swatch { width: 10px; height: 10px; border-radius: 3px; }
.inv-bins-note { font-size: 11px; color: #6b6963; }
.inv-bins-warn { color: #b91c1c; font-weight: 700; }
.inv-po { display: flex; flex-direction: column; gap: 0; }
.inv-po-side { display: flex; align-items: center; gap: 6px; }
.inv-dot { width: 10px; height: 10px; border-radius: 999px; border: 2px solid #1d5c3f; background: transparent; flex-shrink: 0; }
.inv-dot--filled { background: #1d5c3f; }
.inv-badge { display: inline-flex; align-items: center; padding: 2px 8px; border-radius: 999px; font-size: 10px; font-weight: 700; white-space: nowrap; }
.inv-badge--arriving { background: #1d5c3f; color: #fafaf7; }
.inv-badge--transit { background: rgba(29, 92, 63, 0.12); color: #1d5c3f; }
.inv-badge--water { background: rgba(59, 130, 246, 0.14); color: #1d4ed8; }
.inv-catbar { display: flex; height: 12px; border-radius: 999px; overflow: hidden; background: #eeeeea; }
.inv-catseg { height: 100%; min-width: 0; }
.inv-cat-legend { display: flex; flex-wrap: wrap; gap: 6px 10px; font-size: 11px; }
`
