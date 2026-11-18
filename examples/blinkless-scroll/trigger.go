package main

// TimelineTrigger is one timeline-trigger. The activation range turns the
// trigger on when the timeline position enters it. The active range keeps it
// on: once on, it turns off only when the position leaves the active range.
// An auto active range reuses the activation range, so the trigger then
// follows the position with no gap.
type TimelineTrigger struct {
	Name            string
	Source          string   // first source, kept for convenience
	Sources         []string // timelines it may watch, first one that resolves wins
	ActivationStart RangeBound
	ActivationEnd   RangeBound
	ActiveStart     RangeBound
	ActiveEnd       RangeBound
}

// TriggerState holds the on/off memory of one trigger between evaluations.
type TriggerState struct{ Active bool }

func within(pos, lo, hi float64) bool { return pos >= lo && pos < hi }

// Evaluate moves the state for the position pos on ruler space and returns
// the new active value. A position that equals the end of the whole ruler
// counts as inside the last range so a trigger can still hold at 100%.
func (t TimelineTrigger) Evaluate(st *TriggerState, pos float64, space RangeSpace) bool {
	aLo, aHi := space.Resolve(t.ActivationStart, false), space.Resolve(t.ActivationEnd, true)
	bLo, bHi := aLo, aHi

	if !autoBound(t.ActiveStart) || !autoBound(t.ActiveEnd) {
		bLo, bHi = space.Resolve(t.ActiveStart, false), space.Resolve(t.ActiveEnd, true)
	}

	in := func(lo, hi float64) bool {
		return within(pos, lo, hi) || (pos == hi && hi >= space.Total && hi > lo)
	}

	switch {
	case !st.Active && in(aLo, aHi):
		st.Active = true
	case st.Active && !in(bLo, bHi):
		st.Active = false
	}

	return st.Active
}

// autoBound is true for an edge that says "same as the activation range".
func autoBound(b RangeBound) bool { return b.Auto || b.Normal || b == RangeBound{} }
