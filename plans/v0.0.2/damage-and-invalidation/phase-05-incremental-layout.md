# Phase 05, incremental layout

Closed until phases 1 to 4 land and a measurement asks for it. This is the
Blink-sized project: a persistent document with stable node identity, dirty
flags, incremental cascade, and dirty-subtree layout.

## The gate

Open this phase only when a large example reports layout or display-list time
as the top frame stage while phases 1 to 4 hold. Candidates: `perf-complex`
(480 rows) and `perf-stress`. The current numbers live in
`documentation/performance-baseline.md` and
`plans/v0.0.2/performance/`. No gate, no work.

## Model

Chrome's shape, from the BlinkNG deep dive: content is marked "possibly needs
paint invalidation" during style and layout, the pre-paint walk resolves the
marks, and only affected tiles re-raster.

The ownframe equivalent needs a mutation model first, because blinkless is a
pure function today: template in, display list out, no state between calls.

- Persistent document with stable node identity across iterations.
- Dirty flags from `SetData` and layout inputs, resolved per subtree for style
  and layout.
- An invalidation output: which elements' paint changed, so ownframe can skip
  `diffDisplay` and feed the phase 01 damage list directly.
- `DisplayList` stays as the from-scratch path. The incremental path is
  additive, and the from-scratch list is the parity reference.

## Checklist

- [ ] Re-measure first: `BenchmarkRedrawStages`, `BenchmarkRedrawHeavyWarm`,
      and the large example on the release machine, and record which stage
      dominates.
- [ ] Write the design doc in the blinkless checkout; keep the pure function
      available.
- [ ] Prototype dirty-subtree layout on one container type, with byte-parity
      tests against the from-scratch list.
- [ ] Wire the invalidation output through ownframe; the diff stays as the
      correctness fallback.
- [ ] Measure again. Land only if the large example's top stage drops.

## Exit

- A one-field data change on a 500 row page reports zero parse, zero cascade,
  layout of the changed subtree only, and one changed op in the invalidation
  output, with byte parity against the full path.
- The release machine's large-example frame stage drops, recorded next to the
  phase 01 numbers.

## Risks

- This is the largest item in this folder and the least justified by current
  numbers. The board's cost was paint, not layout, and the fix did not need the
  engine. Keep it parked until the gate opens.
- Incremental style and layout parity is the classic source of subtle paint
  bugs. The byte-parity harness in `incremental-repaint.md` phase 8 is the
  pattern to reuse.
