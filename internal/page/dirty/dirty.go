// Package dirty holds the damage tracker and rect-merge helpers used by
// package page. It does not import page.
package dirty

// Pad is the margin a dirty rect gains for shadows, outlines, and
// antialiasing.
const Pad = 4

// Gap is how near two dirty rects must sit to merge into one.
const Gap = 8

// MaxRects caps the damage list. Past the cap the list collapses to
// its bounding box, so one take never pays for an unbounded list.
const MaxRects = 8
