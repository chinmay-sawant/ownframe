# CSS masking

Run the desktop example with `go run ./examples/css-masking`, or serve its
picture with `go run ./examples/css-masking -web`.

The panels use photos of blossoms and a beach, with the mask deciding which
photo shows through. The local copies come from the Spotify example's
`track-02.jpg` and `track-21.jpg`; both are CC0 1.0. The page also uses the
copied `alpha-half.png` and `luminance-half.png` fixtures from Blinkless, plus
`mask-type.svg`. The page names `mask`, `mask-clip`,
`mask-composite`, `mask-image`, `mask-mode`, `mask-origin`, `mask-position`,
`mask-repeat`, `mask-size`, and SVG `mask-type`. Its test checks the transparent
and opaque halves against the unmasked photos, the visible `exclude` band,
the SVG mask, and retained display selection (`page_test.go`). Desktop replay
paints each masked element into a GPU group and applies its mask texture before
compositing (`../../internal/replay/group_layer.go:16-36`). `-web` serves a
PNG, so that export uses the software painter
(`../../internal/page/page_image.go:35-40`).

The four panels are visual scenarios, not four properties. They use the ten
unprefixed names above. Blinkless's first wave also tracks eight `-webkit-`
aliases, which this OwnFrame example does not use.

This painter uses the element border box. It accepts percentage or pixel mask
sizes and positions, and `no-repeat`, `repeat`, `repeat-x`, and `repeat-y`.
Other geometry boxes, `cover`, `contain`, `space`, and `round` return an
unsupported-value error.

The photo credits are in `../spotify-player/player/assets/audio/README.md`:
[blossoms](https://www.flickr.com/photos/137643065@N06/24298701426) and
[Marshall Beach sunset](https://www.flickr.com/photos/24046097@N00/6753666737).
