package depthalert

const pageHTML = `<!doctype html>
<html><head><meta name="viewport" content="width=device-width, initial-scale=1">
<style>
*{box-sizing:border-box}body{margin:0;background:#111914;color:#f1f5f2;font:15px/1.3 sans-serif}
main{padding:10px;display:grid;gap:7px}.eyebrow{color:#b7c8bd;font-size:12px;letter-spacing:.08em}
h1{font-size:20px;line-height:1.2;margin:0}.state{padding:10px;border:2px solid #66776d;border-radius:16px;background:#1b2820;line-height:1.25}
.state strong{font-size:22px;display:block}#scan-state{display:block;min-height:1.3em}.metrics{display:flex;justify-content:space-between;gap:10px}
.metric{background:#1b2820;border-radius:12px;padding:8px;flex:1}.metric b{font-size:19px;display:block;margin-top:2px}
.note{font-size:12px;line-height:1.25;color:#d3ddd6;margin:0}
</style></head><body><main id="dashboard">
<div class="eyebrow">DEPTH ALERT · EXPERIMENTAL</div><h1>Center view</h1>
<section class="state"><span>Current signal</span><strong id="signal">{{.State}}</strong>
<div id="scan-state">{{if .Scanning}}Scanning is active{{else}}Scanning is paused{{end}}</div></section>
<div class="metrics"><div class="metric"><span>Nearest surface</span><b id="distance">{{.Distance}}</b></div>
<div class="metric"><span>Depth coverage</span><b id="coverage">{{.Coverage}}</b></div></div>
<p class="note">Experimental. Keep using your cane or guide dog.</p>
</main></body></html>`
