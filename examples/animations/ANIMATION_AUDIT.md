# Animation card audit

This audit covers all 33 cards in `animations.html`. Expectations describe the
page clock: scrolling does not restart delays or completed animations. Negative
start delays mean several looping examples open partway through a pass. Each
card's Play button restarts its own clock and pauses the other cards; Replay
starts it again. Play all resumes shared playback.

| Card | Declared behavior | Audit result |
|---|---|---|
| ease | 180px travel, 2.4s per pass, ease, alternate; blue/orange/green keyframes | Description names travel and colour changes. |
| linear | Same positions and colours at constant speed, alternating | Matches the declaration. |
| overshoot | Cubic-bezier easing with negative and above-one output, applied between each glide keyframe | Description covers pullback and overshoot around keyframe positions. |
| stepped | Five jump-end steps across a 0 to 180px pass, alternating | Uses two endpoint keyframes so five steps are not repeated across two intervals. |
| normal | 2s left-to-right pass, then reset | Matches. |
| reverse | 2s right-to-left pass, then reset | Matches. |
| alternate | Forward and reverse passes alternate | Matches. |
| alternate-reverse | Reverse pass first, then forward | Description says 180px right, rather than the stage's right edge. |
| positive delay | Wait once for 0.8s at opacity 0.3, then translate and pulse | Description distinguishes the initial delay from looping playback. |
| negative delay | Begin 1s into a 2s pass, at 90px and full opacity | Initial pose checked directly. |
| fill none | Underlying grey before/after the delayed purple-to-green pass | Underlying colour differs from keyframe colours, making the reset visible. |
| fill forwards | Underlying grey during delay; retain green endpoint afterward | Before/after paint checked. |
| fill both | Purple start keyframe during delay; retain green endpoint afterward | Before/after paint checked. |
| once | One 1.2s, 130px pass; hold endpoint | Matches. |
| three times | Three 1.2s passes, resetting between passes; hold last endpoint | Matches. |
| infinite | Repeat a 1.2s, 130px pass with a reset | Removed the incorrect claim that the restart is invisible. |
| translate | 150px right over 3s, then reset | Description states the reset rather than implying a reverse pass. |
| rotate | 0 to 150 degrees clockwise over 3s, then reset | Removed the incorrect claim that it unwinds. |
| scale | 0.72 to 1.22 over 3s, then reset small | Description distinguishes reset from reverse playback. |
| skew | 0 to 22 degree skewX, then reset | Description identifies top-left and bottom-right displacement correctly. |
| matrix | Translation plus about 10 degrees clockwise rotation and slight uniform growth | Removed the incorrect shear description. |
| rotateX | Alternate 0 to 70 degrees around the horizontal axis | Angle label corrected; projected glyphs now compress with the chip. |
| rotateY | Alternate 0 to 70 degrees around the vertical axis | Angle label corrected; projected glyphs now narrow with the chip. |
| translateZ | Alternate 0 to 70px toward a camera 520px away | Added alternate playback to match the stated return; glyphs grow with the chip. |
| matrix3d | Fixed 30-degree Y turn; alternate depth between 24px and 64px | Description decodes the matrix; snippet wraps and glyphs follow projection. |
| camera pull | Perspective distance changes 320 to 920px; child has a fixed 35-degree Y turn | Keyframes change the child's projected paint. |
| origin slide | Perspective origin changes top-left to bottom-right at distance 600px | Keyframes change the child's projected paint. |
| fade | Opacity 0.2 to 1 over 3s, then reset | Matches. |
| fill colour | Blue to red over 3s, then reset; label remains white | Fill/label ownership checked separately. |
| ink colour | White to amber over 3s, then reset; dark fill stays fixed | Keyframes change text paint. |
| edge colour | Blue to red border over 3s, then reset; fill/label stay fixed | Keyframes change border paint. |
| corner fill | Blue to orange beveled fill; dark border and path stay fixed | Fill/stroke ownership checked separately. |
| vendor alias | Purple/down to pink/10px up and back within each 1.4s pass; alternate direction | Description explains the complete sequence; long prefixed declaration wraps. |

## Verification

`TestEveryCardKeyframeChangesPaint` freezes each card at its start, midpoint,
and end with controlled timing overrides, then checks that its declared effect
changes retained paint. It checks all cards' effects, not every original timing
combination. The normal declarations are also checked for binding errors and
unsupported-property gaps. Separate tests check delay initial poses, fill-mode
paint before/after a run, the stepped interval count, depth-card alternation,
and code-snippet bounds.

Native snapshots capture the first drawn frame after selected scroll changes
while CSS animation continues. Snapshot mode performs readback and PNG encoding,
so its FPS values are not performance measurements.

GPU tests compare projected text against an explicitly transformed local text
run, and compare scroll-away/return frames against freshly painted frames with
and without retained layers. The dangerous full-size DevTools-close action is
not part of this audit or its native capture sequence.
