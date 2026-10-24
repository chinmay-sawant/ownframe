package motion

// Stats is a snapshot of the runtime counters. The page copies it into the
// probe block; motion logs nothing on the hot path.
type Stats struct {
	// Bound is the number of players the last bind installed.
	Bound int
	// Active is the number of players that reported work at the last Step.
	Active int
	// Sampled is the number of players the last Step sampled.
	Sampled int
	// NeedsRestamp is the number of sampled players whose transform or
	// perspective could not be written as a 2D matrix at the last Step,
	// because the element or the keyframe is 3D, or because it sampled a
	// perspective. Step collects one Request per element for them and the
	// page applies them through layout.RestampTransforms.
	NeedsRestamp int
	// NeedsRetainedState reports whether the last bind found motion that needs
	// the re-projection pass. Without retained layout state the engine pass
	// cannot run, so the page turns it on and lays the page out once more
	// before the first restamp.
	NeedsRetainedState bool
	// Restamped is the number of elements the page re-projected at the last
	// tick.
	Restamped int
	// RestampError is the last re-projection failure. The page keeps drawing
	// the display it already had, so a bad override leaves the last good frame
	// on screen and the next sample tries again.
	RestampError string
	// Rebinds counts the binds that replaced at least one bound player.
	Rebinds int
	// LastError is the last bind failure. It is empty when the last bind
	// succeeded; a failed box does not stop the other boxes from binding.
	LastError string
	// Gaps lists the fidelity gaps the last bind accepted, one entry per
	// element and reason. A gap is a property or approximation the runtime
	// knows it cannot reproduce exactly yet: the declaration stays at the
	// element's base paint, or the approximation is written and reported, such
	// as per-operation opacity on overlapping content until phase 8. The list
	// is rebuilt at every bind and empty after Bind(nil).
	Gaps []string
	// UnresolvedTimelines counts the dashed-ident animation-timeline
	// references the last bind could not resolve: a name no ancestor
	// declares, a nested element scroller this engine does not implement, or
	// a view timeline whose subject is not in the display. Those animations
	// stay inactive; they never fall back to wall-clock time.
	UnresolvedTimelines int
	// InactiveTimelines counts the bound players the last Step left inactive
	// because a resolved named timeline had no range at that page geometry,
	// such as a page that does not scroll. It is reset by every Step.
	InactiveTimelines int
	// LayoutBound is the number of elements whose layout-tier keyframes the
	// last bind installed. Each one reaches the engine's override relayout
	// path through the per-frame budget.
	LayoutBound int
	// LayoutRelayouts is the number of elements the page relaid out through
	// the engine override pass at the last tick.
	LayoutRelayouts int
	// LayoutQueued is the number of elements still waiting for a relayout
	// after the last tick's budget. A nonzero value means the budget deferred
	// work; the elements sample when their turn runs, not when they queued.
	LayoutQueued int
	// LayoutBudgetHit reports that the last tick left at least one element
	// queued because the budget was spent. A page that keeps hitting it wants
	// a larger budget or a lower sample rate for layout animations.
	LayoutBudgetHit bool
	// LayoutError is the last override relayout failure. The page keeps
	// drawing the display it already had, so a bad value leaves the last good
	// frame on screen and the next queued sample tries again.
	LayoutError string
	// SVGPaintRequests is the number of inline-SVG paint values the last Step
	// collected, one per element whose paint changed. Each one is a candidate
	// for layout.RestyleInlineSVG.
	SVGPaintRequests int
	// SVGPaintRestyles is the number of elements the page restyled at the last
	// tick. It is one at most unless the budget was raised.
	SVGPaintRestyles int
	// SVGPaintQueued is the number of collected restyles the last take left
	// waiting for the per-frame budget. The next Step collects them again, so
	// they are never lost.
	SVGPaintQueued int
	// SVGPaintBudgetHit reports that the last take left at least one element
	// queued because the budget was spent. A page that keeps hitting it wants a
	// larger budget or fewer animated icons.
	SVGPaintBudgetHit bool
	// SVGPaintDeferred counts the samples whose paint property could not be
	// restyled, such as a url(#gradient) paint server. The element keeps its
	// base paint and SVGPaintDeferredReason names the declaration.
	SVGPaintDeferred int
	// SVGPaintDeferredReason is the first deferral reason of the last Step.
	SVGPaintDeferredReason string
	// SVGPaintError is the last restyle failure. The page keeps drawing the
	// display it already had, and the next tick collects the request again.
	SVGPaintError string
	// TransitionsBound is the number of transition renderers the last bind
	// installed.
	TransitionsBound int
	// TransitionsStarted is the number of new renderers the last bind started
	// from a captured style change.
	TransitionsStarted int
	// TransitionsCancelled is the number of live renderers the last bind
	// dropped: a removed element, a removed declaration, or a replaced run.
	TransitionsCancelled int
	// TransitionsSkipped is the number of captures or live entries the last
	// bind could not use because the element's operation count or kinds
	// changed, or because an animation owns the field. Those properties snap
	// to the committed value.
	TransitionsSkipped int
	// TransitionsBlocked is the number of property changes the last capture
	// saw but could not transition because the display does not expose the
	// value or the engine path cannot express it yet. A blocked change snaps
	// to the committed value.
	TransitionsBlocked int
	// TransitionSkipReason names the first property change the last capture
	// left alone with a reason: an outgoing presence with no destination
	// operations, a display cross-fade that needs the phase 04 layer, an
	// integer that needs a relayout property the engine does not expose, or a
	// layout length whose computed value is not in the display. It is empty
	// when nothing was blocked.
	TransitionSkipReason string
	// TransitionsNoID counts the transition declarations the last capture
	// skipped because the element has no id to match across displays.
	TransitionsNoID int
	// TransitionsNoOld counts the declarations the last capture skipped
	// because the element was not in the previous display, which is an
	// appearing element rather than a style change.
	TransitionsNoOld int
	// TransitionsActive is the number of renderers that reported work at the
	// last step.
	TransitionsActive int
	// TransitionsSampled is the number of renderers the last step sampled.
	TransitionsSampled int
	// TransitionError is the last bind failure for a captured transition. The
	// other transitions still bind; a failed one snaps to the committed paint.
	TransitionError string
}
