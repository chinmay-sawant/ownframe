import React from 'react';
import { repository } from './project';

const metrics = [
  { n: '94.7 ns', t: 'Animated card update', d: 'Patched animation update with 0 allocs/op (670× faster than the 63.4 µs full redraw).' },
  { n: '486 ns', t: 'Small-fixture warm redraw', d: 'Unchanged warm redraw check with 0 allocs/op, down from 143 µs in v0.0.2.' },
  { n: '140 µs', t: 'CSS animation gallery tick', d: 'Full 534-op animation gallery tick (96.4 KB, 3.3× faster), auto-paced to 144 Hz on Windows GPU.' },
];

const corePipelineRows = [
  ['FlappyTick', '59.2 ns', '0 B', '0', 'Per-frame game/tick loop'],
  ['DrawRectScan (dirty / full)', '183 / 172 ns', '0 B', '0', 'Display-list op culling scan'],
  ['RedrawWarm / RedrawHeavyWarm', '486 / 489 ns', '0 B', '0', 'Unchanged warm frame check'],
  ['NormalRedraw', '528 ns', '0 B', '0', 'Normal UI warm redraw'],
  ['Stages/TemplateExecute', '182 ns', '560 B', '5', 'Go html/template execution'],
  ['Stages/HTMLParse', '5.00 µs', '27.9 KB', '16', 'Blinkless HTML parse'],
  ['Stages/CSSApply', '2.31 µs', '3.18 KB', '37', 'Blinkless CSS selector match/apply'],
  ['Stages/DisplayList', '38.4 µs', '33.8 KB', '107', 'Layout, text measure, and display-list emit'],
  ['RedrawCold', '62.9 µs', '64.6 KB', '202', 'Full cold parse, style, and layout'],
  ['ClickCount (platform example)', '223 µs', '118 KB', '530', 'Click handler plus full redraw'],
  ['LargeRedraw', '17.99 ms', '2.75 MB', '8,188', 'Large unwindowed full rebuild (2.7× fewer allocs)'],
];

const listWorkloadRows = [
  ['ScrollOffset (PerfOff)', '23.2 ns', '0 B', '0', 'Offset update without relayout'],
  ['Unchanged (PerfOff)', '528 ns', '0 B', '0', 'Warm no-op redraw on 1,000 rows'],
  ['WindowedScroll (overscan 4)', '573 µs', '176 KB', '1,197', 'Visible-window rebuild on scroll'],
  ['PatchElement1Row', '645 µs', '209 KB', '102', '27.8× faster, 444× fewer allocs vs. full rebuild'],
  ['PatchElements10Rows', '851 µs', '278 KB', '624', '21.1× faster, 72.5× fewer allocs vs. full rebuild'],
  ['Resize (all 1,000 rows)', '13.5 ms', '2.64 MB', '5,258', 'Unwindowed full-list relayout (3.67× fewer allocs)'],
  ['FullRebuild1Row', '17.92 ms', '5.92 MB', '45,248', '1,000-row full template + parse + layout rebuild'],
  ['TextUpdate (all 1,000 rows)', '18.2 ms', '6.23 MB', '36,297', 'Unwindowed template rebuild on text change'],
];

const scaleStageRows = [
  ['TemplateExecute', '2,583 ns/row', '2,487 ns/row', '2.53 ms', '26,730', '13.2%'],
  ['HTMLParse', '1,428 ns/row', '1,338 ns/row', '1.36 ms', '6,123', '7.1%'],
  ['CSSApply', '145.8 ns/row', '63.5 ns/row', '43.2 µs', '75', '0.2%'],
  ['CSSRelayout', '1.32 ns/row', '0.32 ns/row', '62.2 ns', '1', '<0.01%'],
  ['DisplayList', '13,310 ns/row', '12,134 ns/row', '15.19 ms', '5,296', '79.4%'],
];

const layoutModelRows = [
  ['Block (inline-block cells)', '2.38 ms', '11,910 ns/row', '1.03 MB', '1,179', '1.00×'],
  ['Flex', '2.79 ms', '13,945 ns/row', '1.05 MB', '1,204', '1.17×'],
  ['Grid (3-column track)', '2.81 ms', '14,030 ns/row', '1.23 MB', '2,589', '1.18×'],
];

const microRows = [
  ['PatchOpsPoolRentFree', '11.2 ns', '0 B', '0', 'Display-op patch slice pool rent/free'],
  ['AnimCardPatched', '94.7 ns', '4 B', '0', 'Retained-op animation tick (vs. 63.4 µs redraw)'],
  ['Frame_Dirty_InvalidateAndRedraw', '146 ns', '0 B', '0', 'Dirty invalidation and redraw check'],
  ['SampleTransform', '158 ns', '63 B', '1', 'CSS transform interpolation sample'],
  ['Frame_Tick_Idle', '234 ns', '0 B', '0', 'Idle frame tick loop'],
  ['SampleOneEffect / SampleTenEffects', '382 ns / 2.76 µs', '0 B', '0', 'CSS animation effect sampling'],
  ['CaretLongRun (offsetIn / lineText)', '495 ns / 1.59 µs', '1.79 / 10.2 KB', '1 / 5', 'Long text-run caret lookup'],
  ['Frame_WarmRedraw (Card / Dashboard)', '499 / 512 ns', '0 B', '0', 'Warm frame redraw microbenchmarks'],
  ['Frame_Click_Dispatch', '2.19 µs', '496 B', '5', 'Pointer hit-test and click dispatch'],
  ['AnimationTick (534-op CSS gallery)', '140 µs', '96.4 KB', '147', 'Full gallery restamp (3.3× faster than 466 µs)'],
];

const windowsGpuAppRows = [
  ['animations.exe (534 ops, active CSS)', '144.0 FPS', '2.76 ms', '4.79 / 5.72 ms', '21.0 MB', 'Auto-paced to 144 Hz monitor, 0 long frames, −43% CPU'],
  ['wispr-flow-dashboard.exe (398 ops)', '442.3 FPS', '0.75 ms', '1.12 / 7.45 ms', '21.4 MB', '0 long frames, 1638×950'],
  ['spotify-player.exe (169 ops, 46 images)', '68.2 FPS', '3.57 ms', '5.35 / 9.59 ms', '22.0 MB', '0 long frames, timer-paced, 1440×960'],
  ['departure-board.exe (628 ops, idle)', '2.1 FPS', '1.14 ms', '1.48 / 12.19 ms', '17.7 MB', 'Idle-parked at 4.1% CPU, 1000×640'],
];

const windowsReplayRows = [
  ['ChunkCardDamage (direct → chunked)', '4.22 → 2.00 µs', '849 B', '8', '2.1× faster (1 blit vs. 5 ops)'],
  ['BlitRaster', '406 ns', '581 B', '4', 'Cached raster blit'],
  ['CachedRoundedFill (warm / cold)', '668 ns / 5.36 µs', '596 B / 5.97 KB', '4 / 43', '8.0× faster warm'],
  ['CachedStroke (warm / cold)', '605 ns / 3.45 µs', '588 B / 4.33 KB', '4 / 31', '5.7× faster warm'],
  ['Text_NoSpacing (warm / direct)', '612 ns / 18.5 µs', '623 B / 18.9 KB', '4 / 91', '30× faster warm vs. direct'],
  ['Text_LetterSpacing (warm / direct)', '902 ns / 40.7 µs', '540 B / 29.6 KB', '4 / 215', '45× faster warm vs. direct'],
  ['CompositeLayer (normal / multiply / screen)', '463 ns / 1.93 / 1.82 µs', '536 B / 2.1–2.5 KB', '3 / 18', 'Kage mix-blend shader'],
  ['LayerDrawLayers (100 / 500 / 1,000 ops)', '1.99 / 3.95 / 4.84 µs', '2.14 KB', '12', 'Retained layer compositing'],
  ['TextDamageRedraw (warm / scroll / miss)', '37.4 / 48.6 µs / 1.44 ms', '27.3 KB / 1.14 MB', '200 / 5,162', 'Text damage cache hit vs. miss'],
  ['LayerRotatedCard', '8.56 µs', '2.04 KB', '11', '1 blit/frame after 40-op bake'],
];

export default function Benchmarks() {
  return <div className="p-page">
    <section id="compare" className="container p-head" aria-labelledby="compare-title">
      <p className="eyebrow">Measured, with context</p>
      <h1 id="compare-title" className="p-title">Small updates should stay small.</h1>
      <p className="lede">Retained operations and in-place element patching let small updates avoid rebuilding the whole document. These display-free Go CPU benchmarks cover specific fixtures and do not include GPU rasterization, software GL drivers, or display vsync.</p>
    </section>
    <section className="container p-block">
      <ul className="p-metrics">
        {metrics.map((m) => <li key={m.n} className="p-metric">
          <span className="p-metric-num">{m.n}</span>
          <h2 className="p-metric-title">{m.t}</h2>
          <p>{m.d}</p>
        </li>)}
      </ul>
      <p className="p-note">Measured sequentially (<code>-p 1</code>) after the post-v0.0.6 performance pass with Blinkless <code>0140a5f</code>, Intel Core i7-13700HX, Linux amd64 &amp; native Windows GPU (RTX 4060 @ 144 Hz), Go 1.26.11. Full tables and reproduction commands live in the <a href={`${repository}/blob/master/documentation/performance-baseline.md`}>performance baseline</a> and <a href={`${repository}/blob/master/documentation/performance.md`}>measurement procedure</a>.</p>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Core pipeline and page benchmarks</h2>
        <p className="p-bench-desc">Median of 3 sequential runs via <code>scripts/perf-baseline.sh --count 3 --benchtime 500ms</code>.</p>
        <div className="p-table-wrap" role="region" aria-label="Core pipeline and page benchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">Benchmark</th><th scope="col" className="p-col-hl">Median time</th><th scope="col">Bytes/op</th><th scope="col">Allocs/op</th><th scope="col">What it measures</th></tr></thead>
            <tbody>
              {corePipelineRows.map(([name, time, bytes, allocs, desc]) => <tr key={name}>
                <th scope="row">{name}</th>
                <td className="p-col-hl p-num"><strong>{time}</strong></td>
                <td className="p-num">{bytes}</td>
                <td className="p-num">{allocs}</td>
                <td>{desc}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">1,000-row list workloads and in-place patching</h2>
        <p className="p-bench-desc">Median of 3 sequential runs from <code>scripts/perf-baseline.sh --suite list</code> and <code>BenchmarkListPatchVsRebuild</code>.</p>
        <div className="p-table-wrap" role="region" aria-label="1,000-row list workloads and patching benchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">Workload</th><th scope="col" className="p-col-hl">Median time</th><th scope="col">Bytes/op</th><th scope="col">Allocs/op</th><th scope="col">Notes</th></tr></thead>
            <tbody>
              {listWorkloadRows.map(([name, time, bytes, allocs, notes]) => <tr key={name}>
                <th scope="row">{name}</th>
                <td className="p-col-hl p-num"><strong>{time}</strong></td>
                <td className="p-num">{bytes}</td>
                <td className="p-num">{allocs}</td>
                <td>{notes}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Pipeline stage scaling (50 → 200 → 1,000 rows)</h2>
        <p className="p-bench-desc">Per-stage cost and linearity from <code>BenchmarkScaleStages</code>. Every stage scales linearly with row count; <code>DisplayList</code> accounts for 78.8% of a 1,000-row rebuild.</p>
        <div className="p-table-wrap" role="region" aria-label="Pipeline stage scaling benchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">Stage</th><th scope="col">50 rows</th><th scope="col">200 rows</th><th scope="col" className="p-col-hl">1,000 rows total</th><th scope="col">1,000 rows allocs</th><th scope="col">Share</th></tr></thead>
            <tbody>
              {scaleStageRows.map(([stage, r50, r200, total, allocs, share]) => <tr key={stage}>
                <th scope="row">{stage}</th>
                <td className="p-num">{r50}</td>
                <td className="p-num">{r200}</td>
                <td className="p-col-hl p-num"><strong>{total}</strong></td>
                <td className="p-num">{allocs}</td>
                <td className="p-num">{share}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Layout model overhead (200 rows × 3 columns)</h2>
        <p className="p-bench-desc"><code>DisplayList</code> cost across Block, Flex, and Grid on the same 200-row dataset from <code>BenchmarkLayoutModels</code>.</p>
        <div className="p-table-wrap" role="region" aria-label="Layout model overhead benchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">Layout model</th><th scope="col" className="p-col-hl">Total time</th><th scope="col">Per row</th><th scope="col">Bytes/op</th><th scope="col">Allocs/op</th><th scope="col">vs. Block</th></tr></thead>
            <tbody>
              {layoutModelRows.map(([model, total, perRow, bytes, allocs, ratio]) => <tr key={model}>
                <th scope="row">{model}</th>
                <td className="p-col-hl p-num"><strong>{total}</strong></td>
                <td className="p-num">{perRow}</td>
                <td className="p-num">{bytes}</td>
                <td className="p-num">{allocs}</td>
                <td className="p-num">{ratio}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Frame and internal microbenchmarks</h2>
        <p className="p-bench-desc">Sequential runs across <code>./test</code>, <code>./internal/motion</code>, <code>./internal/frame</code>, <code>./internal/textrun</code>, and <code>./internal/page/patch</code>.</p>
        <div className="p-table-wrap" role="region" aria-label="Frame and internal microbenchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">Benchmark</th><th scope="col" className="p-col-hl">Time/op</th><th scope="col">Bytes/op</th><th scope="col">Allocs/op</th><th scope="col">What it measures</th></tr></thead>
            <tbody>
              {microRows.map(([name, time, bytes, allocs, desc]) => <tr key={name}>
                <th scope="row">{name}</th>
                <td className="p-col-hl p-num"><strong>{time}</strong></td>
                <td className="p-num">{bytes}</td>
                <td className="p-num">{allocs}</td>
                <td>{desc}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Native Windows GPU window benchmarks</h2>
        <p className="p-bench-desc">Pure-Go (<code>CGO_ENABLED=0</code>) Windows binaries measured sequentially for 8 s each via <code>-tags ownframe_debug -perf -debug-addr</code> on NVIDIA GeForce RTX 4060 Laptop GPU (143 Hz) + Intel Core i7-13700HX.</p>
        <div className="p-table-wrap" role="region" aria-label="Native Windows GPU window benchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">App</th><th scope="col" className="p-col-hl">Present FPS</th><th scope="col">Avg frame</th><th scope="col">p95 / p99</th><th scope="col">Go heap</th><th scope="col">Notes</th></tr></thead>
            <tbody>
              {windowsGpuAppRows.map(([app, fps, avg, pct, heap, notes]) => <tr key={app}>
                <th scope="row">{app}</th>
                <td className="p-col-hl p-num"><strong>{fps}</strong></td>
                <td className="p-num">{avg}</td>
                <td className="p-num">{pct}</td>
                <td className="p-num">{heap}</td>
                <td>{notes}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Native Windows Ebiten replay benchmarks</h2>
        <p className="p-bench-desc">Sequential run of <code>replay.test.exe</code> (<code>internal/replay</code>) on native Windows (<code>goos: windows</code>, <code>goarch: amd64</code>).</p>
        <div className="p-table-wrap" role="region" aria-label="Native Windows Ebiten replay benchmarks" tabIndex="0">
          <table className="p-table">
            <thead><tr><th scope="col">Benchmark</th><th scope="col" className="p-col-hl">Time/op</th><th scope="col">Bytes/op</th><th scope="col">Allocs/op</th><th scope="col">Notes</th></tr></thead>
            <tbody>
              {windowsReplayRows.map(([name, time, bytes, allocs, notes]) => <tr key={name}>
                <th scope="row">{name}</th>
                <td className="p-col-hl p-num"><strong>{time}</strong></td>
                <td className="p-num">{bytes}</td>
                <td className="p-num">{allocs}</td>
                <td>{notes}</td>
              </tr>)}
            </tbody>
          </table>
        </div>
      </div>

      <div className="p-bench-group">
        <h2 className="p-bench-title">Runtime architecture comparison</h2>
        <div className="p-table-wrap" role="region" aria-label="Runtime architecture comparison" tabIndex="0">
          <table className="p-table">
            <caption className="visually-hidden">Runtime architecture of ownframe and browser-based approaches</caption>
            <thead><tr><th scope="col">Inside the app</th><th scope="col">Electron</th><th scope="col">Wails v2 / Tauri</th><th scope="col" className="p-col-hl">ownframe</th></tr></thead>
            <tbody>
              <tr><th scope="row">Layout</th><td>Bundled Chromium</td><td>System WebView</td><td className="p-col-hl"><strong>Blinkless, in Go</strong></td></tr>
              <tr><th scope="row">UI event handlers</th><td>JavaScript</td><td>JavaScript and a native bridge</td><td className="p-col-hl"><strong>Go functions</strong></td></tr>
              <tr><th scope="row">Web compatibility</th><td>Chromium web platform</td><td>Host WebView web platform</td><td className="p-col-hl"><strong>Supported HTML and CSS subset</strong></td></tr>
            </tbody>
          </table>
        </div>
      </div>
      <p className="p-links"><a href="#docs/compare-electron">Read the Electron comparison</a> <span aria-hidden="true">·</span> <a href="#docs/why-ownframe">Why we built ownframe</a></p>
    </section>
  </div>;
}
