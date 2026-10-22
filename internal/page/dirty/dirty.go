// Package dirty holds the damage tracker, display diff, and rect-merge
// helpers used by package page. It does not import page.
package dirty

// Pad is the margin a dirty rect gains for shadows, outlines, and
// antialiasing.
const Pad = 4

// Gap is how near two dirty rects must sit to merge into one.
const Gap = 8

// MaxRects caps the damage list. Past the cap the list collapses to
// its bounding box, so one take never pays for an unbounded list.
const MaxRects = 8

// frameThird is the area divisor for a collapse to one box. A damage
// list or a display diff whose area exceeds the frame divided by
// frameThird, more than one third, stops being a partial repaint.
const frameThird = 3
