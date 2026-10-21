# The cost of an element patch against a full redraw

## What the harness measures

The fixture is `testdata/css-preview`, the whole 43 row catalog in the page,
at 1908x999, the desktop window size. The page is built with
`ownframe.Config.RetainState` on, because the engine keeps no layout state
without it and refuses every patch. `SetPerf(true)` records the stage times.

The change is the click state of one card, in two places.

- The card row `catalog-row-01`. Its class goes from `catalog-row` to
  `catalog-row is-selected`.
- The state span `catalog-state-01` inside that card. Its class goes from
  `catalog-state` to `catalog-state is-selected` and its text from `Ready` to
  `Started / selected`.

Both paths make the same tree edit through `Page.PatchElement`, because the
page API has no other way to change the tree. The full path redraws the page
after the edit. The patch path stops there. The difference between the two
runs is the redraw.

## The numbers

Intel i7-13700HX, Go 1.26.4, the Blinkless commit `c1054ff`.

| Path | Time | Allocated | Allocations | Patches applied |
|---|---:|---:|---:|---:|
| FullRedrawCardClass | 34.5 ms | 20.18 MB | 32188 | |
| ElementPatchCardClass | 6.08 ms | 6.31 MB | 6875 | 30 of 30 |
| FullRedrawStateClass | 39.2 ms | 19.86 MB | 31656 | |
| ElementPatchStateClass | 4.72 ms | 5.99 MB | 6347 | 30 of 30 |
| FullRedrawSameSource | 27.2 ms | 13.87 MB | 25304 | |

Both patch benchmarks apply 30 of 30 now, because the engine splices a card
row. The times are the median of five runs of 30 iterations. The full redraw
spread was 33.7 to 52.9 ms and the card patch spread was 5.1 to 7.5 ms, so
the wall clock moves with the load of the machine, and the allocation counts
do not.

The engine is the commit `c1054ff`, which is what the local checkout holds as
well, so these numbers are not comparable with a run against the previous
engine commit. The allocation counts are comparable.

The redraw splits into three stages, measured by the page itself.

| Stage | Time |
|---|---:|
| Template | 0.58 ms |
| Cascade | 0.09 ms |
| Layout | 27.7 ms |

## What the patch saves

The layout stage is the whole cost of a redraw here. Laying out 43 cards,
9178 operations and every text run takes 27.7 ms. The patch lays out one span
and leaves the other 42 cards alone.

For the state span the patch costs 4.72 ms against the 39.2 ms of the full
path, so the patch is about eight times faster. It allocates 5.99 MB against
19.86 MB, so about three times less. The cascade is not saved. The patch
re-cascades the whole document, but that stage costs 0.09 ms, so there is
nothing to win there.

## What the patch still costs

A patch of one span still allocates 5.99 MB, and a patch of the card row
allocates 6.31 MB. The host adopts the whole patched display the engine hands
back, because a patch that changes the operation count moves every later
operation index, so the engine rebuilds the operation list and the owners for
the whole page, and the page holds 9178 operations. The layout is skipped, the
copy is not. A page that patches often should expect that allocation on every
patch.

The page also keeps the engine's layout state for the whole page, the box
tree and the resolved styles, because a patch needs them. That memory is
paid for as long as the page lives, and `RetainState` is off by default.

## What the engine refuses today

`TestPatchTargetsReport` prints this table. It probes each element of the
fixture with a class change and reports what the engine says.

| Element | Engine outcome |
|---|---|
| `catalog-state-01` | spliced, 1 operation |
| `start-property-webkit_animation` | spliced, 2 operations |
| `preview-metadata` | spliced, 1 operation |
| `catalog-row-01` | refused: the element changed size, 128.84 to 112.34 high |
| `catalog` | refused: the element changed size |
| `catalog-grid` | refused: the element changed size |
| `catalog-target-webkit_animation` | refused: the element changed size |
| `play-all` | refused: the element changed size, 24.75 to 30 high |
| `metadata` | refused: the element changed size, 400.05 to 408.93 high |

The probe puts a class that carries no rules on the element, so a card loses
its padding and its border and changes size, and the engine refuses on the
size change. An out-of-flow box inside the subtree is laid out again now, so
the preview target no longer refuses the row. The class the example uses
carries rules, so the card row is spliced with 203 operations, and the state
span inside the card rides along with that one patch.

The metadata card is refused for a different reason. Its readout box scrolls,
so the card's height depends on the height of the readout, and a subtree
layout of the card does not reproduce it.

## The parity gate

`TestPatchElementMatchesAFullRedraw` is the element patch parity gate, in the
shape of the Blinkless parity test. It patches the state span, then redraws
the page, and compares the two displays operation by operation: the kind, the
geometry, the text, the color and the alpha of every operation. It compares
the operations of the card that holds the span, the operations of the span,
the box of the span, and every operation the card does not own. The
operation id is left out, because a full redraw assigns new ids.

`TestPatchCardRowMatchesAFullRedraw` is the same gate for the card row. The
card row patch applies today, so the test takes the patched operations,
redraws the page, and compares the card operations of the patch with the card
operations of the full redraw. If the engine refuses the patch, the test
compares the operations of the page before the refusal with the operations
after it, and checks that the page kept the operations it had.

## The commands

```sh
cd /home/chinmay/ChinmayPersonalProjects/ownframe/examples

# the parity gates and the table of what the engine refuses
go test -p 2 -run 'TestPatch' -v ./css-preview/

# the cost of a full redraw and of an element patch
go test -p 2 -run XXX -bench BenchmarkPatchCost -benchtime 30x -count 5 ./css-preview/
```

On a shared machine, wrap both in `flock /tmp/opencode/ownframe-heavy.lock`
and run them with `GOMAXPROCS=2`. The wall clock times move with the load of
the machine, the allocation counts do not.

## The files

| File | What it holds |
|---|---|
| `patch_cost_setup_test.go` | the fixture page at 1908x999 with the retained state |
| `patch_cost_card_test.go` | the card, the state span, and the class strings |
| `patch_cost_bench_test.go` | the benchmark and the full redraw of the card class |
| `patch_cost_state_bench_test.go` | the full redraw of the state span and the stage report |
| `patch_cost_patch_bench_test.go` | the two element patch benchmarks |
| `patch_cost_parity_test.go` | the parity gate for the state span |
| `patch_cost_card_parity_test.go` | the parity gate for the card row |
| `patch_cost_compare_test.go` | the operation and box comparison |
| `patch_cost_ops_test.go` | the operation lookup by element |
| `patch_cost_fixture_test.go` | the executed fixture template and the stylesheet |
| `patch_cost_targets_test.go` | the table of what the engine refuses |
