package main

import "testing"

func TestCatalogTransitionPropertiesReachOwnFrameFrames(t *testing.T) {
	cases := append(append([]transitionRuntimeCase{}, standardTransitionRuntimeCases...), webkitTransitionRuntimeCases...)
	for _, c := range cases {
		t.Run(c.property, func(t *testing.T) {
			before, _ := sampleTransition(t, c, c.before)
			after, _ := sampleTransition(t, c, c.after)
			if before == after {
				t.Fatalf("sample = %+v for both values", after)
			}
		})
	}
}
