package styleinvalid

// Doc notes for scoped style invalidation:
//
// Supported: per-box dependency tracking, style dirty ranges, matched-selector
// skipping, fixed/sticky/absolute layer promotion, and anchor-based offsets.
//
// Unsupported under scoped invalidation: global percentage reflow that spans
// the whole document, more than two style dirty passes in one frame, and style
// properties that require a full relayout when an ancestor changes.
